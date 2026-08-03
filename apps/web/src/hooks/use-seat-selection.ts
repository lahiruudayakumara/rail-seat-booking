import { useMutation, useQuery } from "@tanstack/react-query";
import { useMemo } from "react";
import type { Seat } from "@/types";
import { createHold, getPassengerPreferences, getQuote, getSeatMap, releaseHold } from "../api";
import { usePassengerAuth } from "@/auth/use-passenger-auth";
import { useAppDispatch, useAppSelector } from "../store";
import { addSeatSelection, removeSeatSelection } from "../store/slices/booking-slice";
import { useJourneySearch } from "./use-journey-search";
import { useTrainSelection } from "./use-train-selection";

export function useSeatSelection() {
  const dispatch = useAppDispatch();
  const { selectedSeat, selectedSeats, quote, hold } = useAppSelector((state) => state.booking);
  const { originId, destinationId } = useJourneySearch();
  const { runId } = useTrainSelection();
  const { account } = usePassengerAuth();

  const preferencesQuery = useQuery({
    queryKey: ["passenger-preferences"],
    queryFn: getPassengerPreferences,
    enabled: Boolean(account),
    staleTime: 60_000,
  });

  const seatsQuery = useQuery({
    queryKey: ["seat-map", runId, originId, destinationId],
    queryFn: () => getSeatMap(runId, originId, destinationId),
    enabled: Boolean(runId && originId && destinationId),
  });

  const quoteMutation = useMutation({
    mutationFn: async (seat: Seat) => {
      const nextQuote = await getQuote({ runId, originStationId: originId, destinationStationId: destinationId, seatId: seat.id });
      const nextHold = await createHold(nextQuote.id);
      return { seat, quote: nextQuote, hold: nextHold };
    },
    onSuccess: (data) => {
      dispatch(addSeatSelection(data));
      void seatsQuery.refetch();
    },
  });

  const releaseMutation = useMutation({
    mutationFn: ({ holdId, holdToken }: { holdId: string; holdToken: string }) =>
      releaseHold(holdId, holdToken),
    onSettled: () => void seatsQuery.refetch(),
  });

  const groupedSeats = useMemo(() => {
    const preferredClass = preferencesQuery.data?.preferredCoachClass;
    const entries = Object.entries(
      (seatsQuery.data?.items ?? []).reduce<Record<string, Seat[]>>((acc, seat: Seat) => {
      acc[seat.coachCode] = acc[seat.coachCode] ?? [];
      acc[seat.coachCode].push(seat);
      return acc;
      }, {}),
    );
    entries.forEach(([, seats]) => seats.sort((left, right) => left.label.localeCompare(right.label, undefined, { numeric: true })));
    return entries.sort(([, left], [, right]) => {
      if (!preferredClass || preferredClass === "ANY") return 0;
      return (left[0]?.coachClass === preferredClass ? 0 : 1) - (right[0]?.coachClass === preferredClass ? 0 : 1);
    });
  }, [preferencesQuery.data?.preferredCoachClass, seatsQuery.data?.items]);

  const handleChooseSeat = (seat: Seat) => {
    const selected = selectedSeats.find((item) => item.seat.id === seat.id);
    if (selected) {
      dispatch(removeSeatSelection(seat.id));
      releaseMutation.mutate({
        holdId: selected.hold.id,
        holdToken: selected.hold.managementToken,
      });
      return;
    }
    if (seat.availabilityStatus === "BOOKED" || selectedSeats.length >= 6) return;
    quoteMutation.mutate(seat);
  };

  return {
    runId,
    selectedSeat,
    selectedSeats,
    quote,
    hold,
    seatsQuery,
    quoteMutation,
    releaseMutation,
    groupedSeats,
    preferences: preferencesQuery.data,
    handleChooseSeat,
  };
}
