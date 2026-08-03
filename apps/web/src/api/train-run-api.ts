import { api } from "./api-instance";
import type { TrainRun } from "@/types";

export const trainRunApi = {
  getTrainRuns: async (routeId: string, date: string) => {
    const res = await api.get<{ items: TrainRun[] }>(
      `/api/v1/train-runs?travelDate=${date}&routeId=${routeId}`,
    );
    return res.data;
  },

  getTrainRun: async (trainRunId: string) => {
    const res = await api.get<TrainRun>(`/api/v1/train-runs/${trainRunId}`);
    return res.data;
  },
};

export function getTrainRuns(routeId: string, date: string) {
  return trainRunApi.getTrainRuns(routeId, date);
}

export function getTrainRun(trainRunId: string) {
  return trainRunApi.getTrainRun(trainRunId);
}

export default trainRunApi;
