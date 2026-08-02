import { useQuery } from "@tanstack/react-query";
import { getTrainRuns } from "../api";
import { useAppDispatch, useAppSelector } from "../store";
import { setRunId, setSelectedSeat } from "../store/slices/booking-slice";
import { useJourneySearch } from "./use-journey-search";

export function useTrainSelection() {
  const dispatch = useAppDispatch();
  const { runId } = useAppSelector((state) => state.booking);
  const { routeId, date, searched } = useJourneySearch();

  const trainRunsQuery = useQuery({
    queryKey: ["train-runs", routeId, date],
    queryFn: () => getTrainRuns(routeId, date),
    enabled: Boolean(routeId && date && searched),
  });

  const handleChooseRun = (nextRunId: string) => {
    dispatch(setRunId(nextRunId));
    dispatch(setSelectedSeat(undefined));
  };

  return {
    runId,
    searched,
    trainRunsQuery,
    handleChooseRun,
  };
}
