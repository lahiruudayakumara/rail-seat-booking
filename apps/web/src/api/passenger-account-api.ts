import type { Booking, PassengerAccount, RegisterPassengerRequest } from "@/types";
import { api } from "./api-instance";

export async function registerPassenger(body: RegisterPassengerRequest) {
  const response = await api.post<PassengerAccount>("/api/v1/passenger/register", body);
  return response.data;
}

export async function loginPassenger(body: { email: string; password: string }) {
  const response = await api.post<PassengerAccount>("/api/v1/passenger/login", body);
  return response.data;
}

export async function logoutPassenger() {
  await api.post("/api/v1/passenger/logout");
}

export async function getCurrentPassenger() {
  const response = await api.get<PassengerAccount>("/api/v1/passenger/me");
  return response.data;
}

export async function getPassengerBookings() {
  const response = await api.get<{ items: Booking[] }>("/api/v1/passenger/bookings");
  return response.data.items;
}
