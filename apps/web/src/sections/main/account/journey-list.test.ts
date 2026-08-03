import { expect, test } from "vitest";
import type { Booking, TrainRun } from "@/types";
import { filterAndSortJourneys, journeyCategory } from "./journey-list";

const booking = (id: string, status: string, createdAt: string): Booking => ({
  id,
  reference: `BK-${id}`,
  status,
  trainRunId: `run-${id}`,
  seat: { id: `seat-${id}`, label: "1A", coachId: "coach-1", coachCode: "R1", coachClass: "FIRST", attributes: ["WINDOW"] },
  originStationId: "fort",
  destinationStationId: "kandy",
  fare: { amountMinor: 46000, currency: "LKR", currencyScale: 2 },
  createdAt,
});

const run = (id: string, arrivalAt: string): TrainRun => ({
  id: `run-${id}`,
  routeId: "route-1",
  trainId: "train-1",
  serviceDate: arrivalAt.slice(0, 10),
  departureAt: arrivalAt,
  arrivalAt,
  status: "SCHEDULED",
});

test("categorizes account journeys using booking and service state", () => {
  const now = new Date("2026-08-04T12:00:00Z");
  expect(journeyCategory(booking("1", "CONFIRMED", now.toISOString()), run("1", "2026-08-05T12:00:00Z"), now)).toBe("UPCOMING");
  expect(journeyCategory(booking("2", "CONFIRMED", now.toISOString()), run("2", "2026-08-03T12:00:00Z"), now)).toBe("PAST");
  expect(journeyCategory(booking("3", "CANCELLED", now.toISOString()), undefined, now)).toBe("CANCELLED");
});

test("filters by journey details and sorts bookings", () => {
  const older = booking("OLD", "CONFIRMED", "2026-08-01T00:00:00Z");
  const newer = booking("NEW", "CANCELLED", "2026-08-03T00:00:00Z");
  const details = {
    stations: {
      fort: { id: "fort", code: "FOT", name: "Colombo Fort" },
      kandy: { id: "kandy", code: "KDY", name: "Kandy" },
    },
    runs: {},
  };

  expect(filterAndSortJourneys([older, newer], details, "ALL", "kandy", "BOOKED_DESC").map((item) => item.id)).toEqual(["NEW", "OLD"]);
  expect(filterAndSortJourneys([older, newer], details, "CANCELLED", "", "BOOKED_DESC").map((item) => item.id)).toEqual(["NEW"]);
});
