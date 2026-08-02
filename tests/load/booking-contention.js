import http from "k6/http";
import { check } from "k6";
import { Counter, Rate } from "k6/metrics";

const bookingSuccesses = new Counter("booking_successes");
const unexpectedResponses = new Rate("unexpected_responses");

export const options = {
  scenarios: {
    contenders: {
      executor: "shared-iterations",
      vus: 12,
      iterations: 12,
      maxDuration: "30s",
    },
  },
  thresholds: {
    booking_successes: ["count==1"],
    unexpected_responses: ["rate==0"],
  },
};

export function setup() {
  const required = ["API_BASE_URL", "TRAIN_RUN_ID", "SEAT_ID", "ORIGIN_STATION_ID", "DESTINATION_STATION_ID", "FARE_QUOTE_ID"];
  for (const name of required) {
    if (!__ENV[name]) throw new Error(`${name} is required`);
  }
  return { baseURL: __ENV.API_BASE_URL.replace(/\/$/, "") };
}

export default function (data) {
  const payload = JSON.stringify({
    fareQuoteId: __ENV.FARE_QUOTE_ID,
    trainRunId: __ENV.TRAIN_RUN_ID,
    seatId: __ENV.SEAT_ID,
    originStationId: __ENV.ORIGIN_STATION_ID,
    destinationStationId: __ENV.DESTINATION_STATION_ID,
    passenger: {
      fullName: `Load Contender ${__VU}`,
      email: `contender-${__VU}-${__ITER}@example.test`,
      phone: "+94770000000",
    },
  });
  const response = http.post(`${data.baseURL}/api/v1/bookings`, payload, {
    headers: {
      "Content-Type": "application/json",
      "Idempotency-Key": `load-${Date.now()}-${__VU}-${__ITER}`,
    },
  });

  const expected = check(response, { "returns 201 or 409": (result) => result.status === 201 || result.status === 409 });
  unexpectedResponses.add(!expected);
  if (response.status === 201) bookingSuccesses.add(1);
}
