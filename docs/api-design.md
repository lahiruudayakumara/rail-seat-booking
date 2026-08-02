# REST API Design

## Conventions

Base path `/api/v1`; JSON uses camelCase, UUID strings, RFC 3339 UTC instants, ISO dates, integer `amountMinor` plus `currency`/`currencyScale`. Collection responses use `items` and cursor pagination. Unknown fields are rejected on writes. Every response includes `X-Request-ID`; callers may supply it, but the server validates/generates it.

Public search/quote/create endpoints are initially anonymous and rate-limited. UUID booking lookup is reserved for authenticated passenger/support contexts when authentication arrives. Guest reference lookup/cancel requires a second factor such as normalized contact verification or a signed management token—reference alone is not authentication. Admin configuration is not in v1. `Idempotency-Key` is required on create and recommended on cancel; same key+payload replays the original response, different payload returns 409.

## Endpoint summary

| Endpoint | Purpose and parameters | Success | Errors | Auth/idempotency |
|---|---|---|---|---|
| `GET /health` | Process liveness; no parameters | 200 health | 500 only on process failure | Public; safe/idempotent |
| `GET /ready` | DB/migration readiness | 200 or 503 | `SERVICE_UNAVAILABLE` | Infrastructure/public-limited; safe |
| `GET /api/v1/routes` | Active routes; `cursor`,`limit` | 200 route page | 400 | Public; safe |
| `GET /api/v1/routes/{routeId}` | Route detail | 200 route | 400,404 | Public; safe |
| `GET /api/v1/routes/{routeId}/stations` | Ordered route stations | 200 station page | 400,404 | Public; safe |
| `GET /api/v1/stations` | Active stations; cursor pagination | 200 station page | 400 | Public; safe |
| `GET /api/v1/train-runs` | Filters `travelDate` required, `routeId`,`originStationId`,`destinationStationId`,`cursor`,`limit`; endpoints must be supplied together and ordered | 200 run page | 400,404 | Public; safe |
| `GET /api/v1/train-runs/{trainRunId}` | Run with route/train summary | 200 run | 400,404 | Public; safe |
| `GET /api/v1/train-runs/{trainRunId}/available-seats` | Required origin/destination UUIDs; optional `coachClass` | 200 available seat list | 400,404 | Public/rate-limited; snapshot only |
| `GET /api/v1/train-runs/{trainRunId}/seat-map` | Required origin/destination UUIDs; optional `coachClass` | 200 all reserved seats with segment status | 400,404,422 | Public/rate-limited; snapshot only |
| `POST /api/v1/fare-quotes` | Quote run, seat and endpoints | 201 quote | 400,404,422 | Public/rate-limited; semantically idempotent but creates quote ID |
| `POST /api/v1/booking-holds` | Atomically hold a quoted seat/segment | 201 hold and expiry | 400,409,422,429 | Public/rate-limited; short-lived signed hold token |
| `POST /api/v1/payments/sandbox` | Complete local sandbox payment and issue ticket | 201 checkout result | 400,401,404,409,422 | Hold token; idempotency key required |
| `POST /api/v1/bookings` | Convert a valid hold into a confirmed booking | 201 booking; replay may be 200/201 with replay header | 400,401,404,409,422,503 | Hold token; **Idempotency-Key required** |
| `POST /api/v1/bookings/access` | Verify reference plus booking email/phone | 200 booking and short-lived management token | 400,404,429 | Public/rate-limited; contact verification required |
| `POST /api/v1/bookings/{bookingId}/cancel` | Optional `{reason}` | 200 updated booking | 400,401/403,404,409,422 | Owner/support; repeat-safe, key recommended |

Path UUIDs must parse; `limit` is 1–100 (default 20); cursors are opaque. `travelDate` uses local service date. Availability validates that endpoints belong to the run route, origin precedes destination, class exists and run is future/bookable. Fare and booking additionally validate enabled reserved coach/seat. Passenger name is 1–120 characters; email/phone limits and normalization are server-side.

## Representative JSON

Availability:

```json
{
  "trainRunId": "30000000-0000-4000-8000-000000000001",
  "originStationId": "20000000-0000-4000-8000-000000000001",
  "destinationStationId": "20000000-0000-4000-8000-000000000005",
  "items": [{
    "id": "60000000-0000-4000-8000-000000000012",
    "label": "12A", "coachId": "50000000-0000-4000-8000-000000000001",
    "coachCode": "R1", "coachClass": "FIRST", "attributes": ["WINDOW"]
  }]
}
```

The seat-map response has the same envelope but includes every reserved seat and an
`availabilityStatus` of `AVAILABLE` or `BOOKED`. Status is calculated only for the
requested half-open segment, so a seat booked on an adjacent leg remains available.

Quote request/response:

```json
{"trainRunId":"30000000-0000-4000-8000-000000000001","seatId":"60000000-0000-4000-8000-000000000012","originStationId":"20000000-0000-4000-8000-000000000001","destinationStationId":"20000000-0000-4000-8000-000000000005"}
```

```json
{"id":"70000000-0000-4000-8000-000000000001","distanceKm":"120.000","amountMinor":46000,"currency":"LKR","currencyScale":2,"breakdown":{"baseFeeMinor":10000,"distanceFeeMinor":36000},"expiresAt":"2026-08-01T10:05:00Z"}
```

Booking request (with `Idempotency-Key: 65c...`):

```json
{
  "fareQuoteId":"70000000-0000-4000-8000-000000000001",
  "trainRunId":"30000000-0000-4000-8000-000000000001",
  "seatId":"60000000-0000-4000-8000-000000000012",
  "originStationId":"20000000-0000-4000-8000-000000000001",
  "destinationStationId":"20000000-0000-4000-8000-000000000005",
  "passenger":{"fullName":"Example Passenger","email":"passenger@example.com","phone":"+94770000000"}
}
```

```json
{
  "id":"80000000-0000-4000-8000-000000000001","reference":"BK-7J4M9Q2X","status":"CONFIRMED",
  "trainRunId":"30000000-0000-4000-8000-000000000001","seat":{"id":"60000000-0000-4000-8000-000000000012","label":"12A","coachCode":"R1","coachClass":"FIRST","attributes":["WINDOW"]},
  "originStationId":"20000000-0000-4000-8000-000000000001","destinationStationId":"20000000-0000-4000-8000-000000000005",
  "fare":{"amountMinor":46000,"currency":"LKR","currencyScale":2},"createdAt":"2026-08-01T10:00:00Z","confirmedAt":"2026-08-01T10:00:00Z"
}
```

Cancel body is optional: `{"reason":"Plans changed"}`. Successful repeat returns the same `CANCELLED` resource without another state transition.

## Error envelope and catalogue

```json
{
  "code":"VALIDATION_ERROR","message":"The request contains invalid fields.",
  "details":[{"field":"destinationStationId","message":"Destination must come after origin."}],
  "requestId":"90000000-0000-4000-8000-000000000001"
}
```

`details` is an array for field errors or an object for domain context; clients must branch on `code`, not message.

| Code | HTTP | Meaning/when | Retry and UI |
|---|---:|---|---|
| `VALIDATION_ERROR` | 400 | Malformed/invalid fields | Fix highlighted fields |
| `ROUTE_NOT_FOUND` | 404 | Route absent/inactive | Refresh search |
| `STATION_NOT_FOUND` | 404 | Station absent/inactive | Refresh stations |
| `TRAIN_RUN_NOT_FOUND` | 404 | Run absent/inaccessible | Return to results |
| `SEAT_NOT_FOUND` | 404 | Seat absent | Refresh seats |
| `SEAT_NOT_RESERVABLE` | 422 | Disabled/unreserved/not on train | Refresh/select another |
| `INVALID_JOURNEY_SEGMENT` | 422 | Endpoints missing/reversed/equal | Correct journey |
| `SEAT_NO_LONGER_AVAILABLE` | 409 | Commit overlap conflict | Do not auto-submit; refresh and suggest |
| `BOOKING_NOT_FOUND` | 404 | No accessible booking | Verify credentials/reference |
| `BOOKING_ALREADY_CANCELLED` | 409* | Already cancelled (*only if strict policy chosen) | Show cancelled; recommended API instead returns 200 |
| `BOOKING_CANNOT_BE_CANCELLED` | 422 | Completed/policy cutoff | Explain policy/contact support |
| `FARE_RULE_NOT_FOUND` | 422 | No unambiguous effective fare | Retry later only after configuration fix |
| `IDEMPOTENCY_CONFLICT` | 409 | Key reused with changed payload/in progress | New key only for new intent; poll/retry same payload |
| `INTERNAL_ERROR` | 500 | Unexpected fault | Safe bounded retry; show request ID |
| `SERVICE_UNAVAILABLE` | 503 | DB/dependency not ready | Retry with backoff/`Retry-After` |

Authentication failures use generic 401/403 messages. Booking access returns the same 404 response for an unknown reference or mismatched contact to resist enumeration. Management tokens are short-lived HMAC-signed bearer credentials scoped to one booking.
