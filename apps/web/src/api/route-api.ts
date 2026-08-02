import { api } from "./api-instance";
import type { Route, Station } from "@/types";

export const routeApi = {
  getRoutes: async () => {
    const res = await api.get<{ items: Route[] }>("/api/v1/routes");
    return res.data;
  },

  getRouteStations: async (routeId: string) => {
    const res = await api.get<{ items: Station[] }>(`/api/v1/routes/${routeId}/stations`);
    return res.data;
  },
};

export function getRoutes() {
  return routeApi.getRoutes();
}

export function getRouteStations(routeId: string) {
  return routeApi.getRouteStations(routeId);
}

export default routeApi;
