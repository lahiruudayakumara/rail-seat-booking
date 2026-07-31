# Fare Design

## Initial model

```text
distance_km = (destination.cumulative_distance_m - origin.cumulative_distance_m) / 1000
distance_component = travelled_distance × class_rate
subtotal = base_fee + distance_component
class_adjusted = subtotal × class_multiplier
fare = max(minimum_fare, class_adjusted)
```

All configuration is effective-dated by route and coach class. `base_fee_minor`, `rate_per_km_minor`, `minimum_fare_minor`, currency, scale and multiplier basis points are stored—not hardcoded. For precision, compute from metres using integer/rational arithmetic and round only at the configured boundary; default rule is half-up to the nearest currency minor unit. Never use binary floating point for money.

The brief's simple formula `base_fee + travelled_distance × class_rate` is normative. A class multiplier can either be 1.0000 when class-specific rates already encode class, or adjust a common rate—configuration must avoid double charging.

## Quote and snapshot

A quote includes an opaque UUID, train run, seat/class, endpoints, distance, selected rule ID/version, itemized values, total minor units, ISO 4217 currency, currency scale, and `expiresAt`. During booking the server validates or recalculates the quote and rejects an expired/mismatched one.

The booking stores the complete calculation snapshot because rules, station distances and tax policy change. Recalculation would alter historical receipts and refunds. Snapshot JSON has a schema version and includes base, distance component, multiplier, minimum adjustment, discount, tax, rounding mode and final total; the normalized total remains queryable.

## Examples (demonstration only)

Assume LKR scale 2, base LKR 100, rate LKR 3/km, minimum LKR 250:

- Colombo Fort cumulative 0 km → Kandy demo 120 km: `100 + 120×3 = LKR 460.00`.
- A 30 km trip: raw `100 + 30×3 = LKR 190.00`; minimum produces `LKR 250.00`.
- First-class multiplier 1.50 on the 120 km subtotal: `460×1.5 = LKR 690.00`.

These distances and fares are illustrative and must not be presented as official Sri Lanka Railways data.

## Extension points

Seasonal and peak/off-peak modifiers, passenger discounts and tax are placeholders behind versioned policies. Define deterministic ordering: base+distance → class multiplier → peak/season → discount → tax → final rounding/minimum (or explicitly choose another ordering and version it). Discounts need eligibility evidence; taxes need jurisdiction/rate snapshots. Unsupported ambiguity yields `FARE_RULE_NOT_FOUND`, never a guessed price.

Admin changes create new effective rules rather than mutate historical meaning. Validate no ambiguous active rules for route/class/date/priority; record author and audit event. Currency conversion is outside scope—each booking uses exactly one configured currency.
