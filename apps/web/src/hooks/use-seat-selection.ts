import { useMutation, useQuery } from "@tanstack/react-query";
import type { Seat } from "@/types";
import { getQuote, getSeatMap } from "../api";
import { useAppDispatch, useAppSelector } from "../store";
import { setQuote, setSelectedSeat } from "../store/slices/booking-slice";
import { useJourneySearch } from "./use-journey-search";
import { useTrainSelection } from "./use-train-selection";

export function useSeatSelection() {
  const dispatch = useAppDispatch();
  const { selectedSeat, quote } = useAppSelector((state) => state.booking);
  const { originId, destinationId } = useJourneySearch();
  const { runId } = useTrainSelection();

  const seatsQuery = useQuery({
    queryKey: ["seat-map", runId, originId, destinationId],
    queryFn: () => getSeatMap(runId, originId, destinationId),
    enabled: Boolean(runId && originId && destinationId),
  });

  const quoteMutation = useMutation({
    mutationFn: (seatId: string) =>
      getQuote({ runId, originStationId: originId, destinationStationId: destinationId, seatId }),
    onSuccess: (data) => dispatch(setQuote(data)),
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
    quoteMutation.mutate(seat.id);
  };

  return {
    runId,
    selectedSeat,
    quote,
    seatsQuery,
    quoteMutation,
    groupedSeats,
    handleChooseSeat,
  };
}
