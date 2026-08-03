import type { Booking, Station, TrainRun } from "@/types";

export type JourneyFilter = "ALL" | "UPCOMING" | "PAST" | "CANCELLED";
export type JourneySort = "BOOKED_DESC" | "BOOKED_ASC" | "TRAVEL_ASC";

export type JourneyDetails = {
  stations: Record<string, Station>;
  runs: Record<string, TrainRun>;
};

export function journeyCategory(
  booking: Booking,
  run?: TrainRun,
  now = new Date(),
): Exclude<JourneyFilter, "ALL"> {
  if (booking.status === "CANCELLED") return "CANCELLED";
  if (booking.status === "COMPLETED" || booking.status === "EXPIRED") return "PAST";
  if (run && new Date(run.arrivalAt).getTime() < now.getTime()) return "PAST";
  return "UPCOMING";
}

export function filterAndSortJourneys(
  bookings: Booking[],
  details: JourneyDetails | undefined,
  filter: JourneyFilter,
  search: string,
  sort: JourneySort,
  now = new Date(),
) {
  const query = search.trim().toLocaleLowerCase();
  const items = bookings.filter((booking) => {
    const run = details?.runs[booking.trainRunId];
    if (filter !== "ALL" && journeyCategory(booking, run, now) !== filter) return false;
    if (!query) return true;

    const origin = details?.stations[booking.originStationId];
    const destination = details?.stations[booking.destinationStationId];
    return [
      booking.reference,
      booking.status,
      booking.seat.coachCode,
      booking.seat.label,
      origin?.code,
      origin?.name,
      destination?.code,
      destination?.name,
    ].some((value) => value?.toLocaleLowerCase().includes(query));
  });

  return items.sort((left, right) => {
    if (sort === "BOOKED_ASC") {
      return new Date(left.createdAt).getTime() - new Date(right.createdAt).getTime();
    }
    if (sort === "TRAVEL_ASC") {
      const leftDeparture = details?.runs[left.trainRunId]?.departureAt;
      const rightDeparture = details?.runs[right.trainRunId]?.departureAt;
      return (leftDeparture ? new Date(leftDeparture).getTime() : Number.MAX_SAFE_INTEGER)
        - (rightDeparture ? new Date(rightDeparture).getTime() : Number.MAX_SAFE_INTEGER);
    }
    return new Date(right.createdAt).getTime() - new Date(left.createdAt).getTime();
  });
}
