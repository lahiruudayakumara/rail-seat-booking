import { describe, expect, test } from "vitest";
import reducer, { addSeatSelection, removeSeatSelection } from "./booking-slice";
import type { SeatSelectionItem } from "@/types";

const selection = (index: number): SeatSelectionItem => ({
  seat: {
    id: `seat-${index}`,
    label: `${index}A`,
    coachId: "coach-1",
    coachCode: "R1",
    coachClass: "FIRST",
    attributes: ["WINDOW"],
  },
  quote: {
    id: `quote-${index}`,
    distanceKm: "120.000",
    amountMinor: 69000,
    currency: "LKR",
    currencyScale: 2,
    breakdown: {},
    expiresAt: "2026-08-04T12:00:00Z",
  },
  hold: {
    id: `hold-${index}`,
    status: "HELD",
    expiresAt: "2026-08-04T12:00:00Z",
    managementToken: `token-${index}`,
  },
});

describe("group seat selection", () => {
  test("keeps up to six distinct held seats and updates the legacy primary selection", () => {
    let state = reducer(undefined, { type: "init" });
    for (let index = 1; index <= 7; index += 1) {
      state = reducer(state, addSeatSelection(selection(index)));
    }

    expect(state.selectedSeats).toHaveLength(6);
    expect(state.selectedSeat?.id).toBe("seat-1");
    expect(state.hold?.id).toBe("hold-1");
  });

  test("promotes the next held seat when the first seat is removed", () => {
    let state = reducer(undefined, addSeatSelection(selection(1)));
    state = reducer(state, addSeatSelection(selection(2)));
    state = reducer(state, removeSeatSelection("seat-1"));

    expect(state.selectedSeats.map(({ seat }) => seat.id)).toEqual(["seat-2"]);
    expect(state.selectedSeat?.id).toBe("seat-2");
    expect(state.quote?.id).toBe("quote-2");
  });
});
