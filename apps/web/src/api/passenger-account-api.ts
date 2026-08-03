import type { Booking, PassengerAccount, PassengerPreferences, PassengerPreferencesInput, RegisterPassengerRequest, SavedTraveller, SavedTravellerInput } from "@/types";
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

export async function getSavedTravellers() {
  const response = await api.get<{ items: SavedTraveller[] }>("/api/v1/passenger/travellers");
  return response.data.items;
}

export async function createSavedTraveller(body: SavedTravellerInput) {
  const response = await api.post<SavedTraveller>("/api/v1/passenger/travellers", body);
  return response.data;
}

export async function updateSavedTraveller(id: string, body: SavedTravellerInput) {
  const response = await api.put<SavedTraveller>(`/api/v1/passenger/travellers/${id}`, body);
  return response.data;
}

export async function deleteSavedTraveller(id: string) {
  await api.delete(`/api/v1/passenger/travellers/${id}`);
}

export async function getPassengerPreferences() {
  const response = await api.get<PassengerPreferences>("/api/v1/passenger/preferences");
  return response.data;
}

export async function updatePassengerPreferences(body: PassengerPreferencesInput) {
  const { preferredCoachClass, preferredSeatType, language } = body;
  const response = await api.put<PassengerPreferences>("/api/v1/passenger/preferences", {
    preferredCoachClass,
    preferredSeatType,
    language,
  });
  return response.data;
}
