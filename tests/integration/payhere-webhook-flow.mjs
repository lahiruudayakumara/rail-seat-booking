import { createHash, randomUUID } from "node:crypto";

const apiBaseUrl = process.env.API_BASE_URL ?? "http://localhost:8080";
const merchantId = process.env.PAYHERE_MERCHANT_ID ?? "1210000";
const merchantSecret = process.env.PAYHERE_MERCHANT_SECRET ?? "sandbox-merchant-secret";
const travelDate =
  process.env.TRAVEL_DATE ??
  new Intl.DateTimeFormat("en-CA", { timeZone: "Asia/Colombo" }).format(new Date());

const request = async (path, init) => {
  const response = await fetch(`${apiBaseUrl}${path}`, init);
  if (!response.ok) {
    throw new Error(`${init?.method ?? "GET"} ${path} returned ${response.status}`);
  }
  return response.headers.get("content-type")?.includes("application/json")
    ? response.json()
    : undefined;
};

const postJson = (path, body, idempotencyKey) =>
  request(path, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...(idempotencyKey ? { "Idempotency-Key": idempotencyKey } : {}),
    },
    body: JSON.stringify(body),
  });

const assert = (condition, message) => {
  if (!condition) throw new Error(message);
};

const md5Upper = (value) => createHash("md5").update(value).digest("hex").toUpperCase();

const routes = await request("/api/v1/routes");
const routeId = routes.items[0].id;
const stations = await request(`/api/v1/routes/${routeId}/stations`);
const originStationId = stations.items[0].id;
const destinationStationId = stations.items.at(-1).id;
const runs = await request(
  `/api/v1/train-runs?travelDate=${travelDate}&routeId=${routeId}`,
);
const trainRunId = runs.items[0].id;
const availability = await request(
  `/api/v1/train-runs/${trainRunId}/available-seats?originStationId=${originStationId}&destinationStationId=${destinationStationId}`,
);
const seatId = availability.items[0].id;

const quote = await postJson("/api/v1/fare-quotes", {
  trainRunId,
  seatId,
  originStationId,
  destinationStationId,
});
const hold = await postJson("/api/v1/booking-holds", { fareQuoteId: quote.id });
const booking = await postJson(
  "/api/v1/bookings",
  {
    holdId: hold.id,
    holdToken: hold.managementToken,
    fareQuoteId: quote.id,
    trainRunId,
    seatId,
    originStationId,
    destinationStationId,
    passenger: {
      fullName: "PayHere Integration",
      email: "payhere@example.test",
      phone: "+94770000001",
    },
  },
  `payhere-booking-${randomUUID()}`,
);
const checkout = await postJson(
  "/api/v1/payments/payhere",
  {
    bookingId: booking.id,
    bookingToken: booking.managementToken,
    billingAddress: "1 Integration Road",
    city: "Colombo",
  },
  `payhere-checkout-${randomUUID()}`,
);

assert(checkout.status === "PENDING", "checkout must start pending");
assert(
  checkout.actionUrl === "https://sandbox.payhere.lk/pay/checkout",
  "checkout must use PayHere Sandbox",
);
assert(checkout.fields.merchant_id === merchantId, "merchant ID does not match");
assert(checkout.fields.hash, "signed checkout hash is missing");
assert(!("merchant_secret" in checkout.fields), "merchant secret was exposed");

const secretHash = md5Upper(merchantSecret);
const signature = md5Upper(
  `${merchantId}${checkout.paymentId}${checkout.fields.amount}${checkout.fields.currency}2${secretHash}`,
);
const callback = new URLSearchParams({
  merchant_id: merchantId,
  order_id: checkout.paymentId,
  payment_id: `${Date.now()}${Math.floor(Math.random() * 1000)}`,
  payhere_amount: checkout.fields.amount,
  payhere_currency: checkout.fields.currency,
  status_code: "2",
  md5sig: signature,
});
const callbackRequest = (body) =>
  fetch(`${apiBaseUrl}/api/v1/webhooks/payhere`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body,
  });

assert((await callbackRequest(callback)).status === 200, "valid callback was rejected");
const payment = await request(`/api/v1/payments/${checkout.paymentId}`, {
  headers: { Authorization: `Bearer ${booking.managementToken}` },
});
assert(payment.status === "PAID", "payment was not marked paid");
assert(payment.booking.status === "CONFIRMED", "booking was not confirmed");
assert(payment.ticket?.status === "ACTIVE", "active ticket was not issued");
assert((await callbackRequest(callback)).status === 200, "duplicate callback was not idempotent");

callback.set("md5sig", "00000000000000000000000000000000");
assert((await callbackRequest(callback)).status === 401, "forged callback was not rejected");

console.log("PayHere webhook integration flow passed.");
