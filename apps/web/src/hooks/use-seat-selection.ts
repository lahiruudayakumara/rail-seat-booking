import { useMutation, useQuery } from "@tanstack/react-query";
import type { Seat } from "@/types";
import { createHold, getQuote, getSeatMap } from "../api";
import { useAppDispatch, useAppSelector } from "../store";
import { setHold, setQuote, setSelectedSeat } from "../store/slices/booking-slice";
import { useJourneySearch } from "./use-journey-search";
import { useTrainSelection } from "./use-train-selection";

export function useSeatSelection() {
  const dispatch = useAppDispatch();
  const { selectedSeat, quote, hold } = useAppSelector((state) => state.booking);
  const { originId, destinationId } = useJourneySearch();
  const { runId } = useTrainSelection();

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

  const groupedSeats = Object.entries(
    (seatsQuery.data?.items ?? []).reduce<Record<string, Seat[]>>((acc, seat: Seat) => {
      acc[seat.coachCode] = acc[seat.coachCode] ?? [];
      acc[seat.coachCode].push(seat);
      return acc;
    }, {}),
  );

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
    handleChooseSeat,
  };
}
