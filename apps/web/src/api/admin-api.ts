import { api } from "./api-instance";
import type { AdminDashboard } from "@/types";

export const adminApi = {
  getDashboard: async (trainRunId: string, adminKey: string) => {
    const response = await api.get<AdminDashboard>(
      `/api/v1/admin/train-runs/${trainRunId}/dashboard`,
      { headers: { Authorization: `Bearer ${adminKey}` } },
    );
    return response.data;
  },
};

export function getAdminDashboard(trainRunId: string, adminKey: string) {
  return adminApi.getDashboard(trainRunId, adminKey);
}
