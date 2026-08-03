import { useMutation, useQuery } from "@tanstack/react-query";
import { useMemo } from "react";
import type { Seat } from "@/types";
import { createHold, getPassengerPreferences, getQuote, getSeatMap } from "../api";
import { usePassengerAuth } from "@/auth/use-passenger-auth";
import { useAppDispatch, useAppSelector } from "../store";
import { setHold, setQuote, setSelectedSeat } from "../store/slices/booking-slice";
import { useJourneySearch } from "./use-journey-search";
import { useTrainSelection } from "./use-train-selection";

export function useSeatSelection() {
  const dispatch = useAppDispatch();
  const { selectedSeat, quote, hold } = useAppSelector((state) => state.booking);
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
    mutationFn: async (seatId: string) => {
      const nextQuote = await getQuote({ runId, originStationId: originId, destinationStationId: destinationId, seatId });
      const nextHold = await createHold(nextQuote.id);
      return { quote: nextQuote, hold: nextHold };
    },
    onSuccess: (data) => {
      dispatch(setQuote(data.quote));
      dispatch(setHold(data.hold));
      void seatsQuery.refetch();
    },
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
    if (seat.availabilityStatus === "BOOKED") return;
    dispatch(setSelectedSeat(seat));
    dispatch(setQuote(undefined));
    dispatch(setHold(undefined));
    quoteMutation.mutate(seat.id);
  };

  return {
    runId,
    selectedSeat,
    quote,
    hold,
    seatsQuery,
    quoteMutation,
    groupedSeats,
    preferences: preferencesQuery.data,
    handleChooseSeat,
  };
}
