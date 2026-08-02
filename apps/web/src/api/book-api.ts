import { api } from "./api-instance";
import type { Booking, CreateBookingRequest, FareQuote, Seat } from "@/types";

export const bookApi = {
  getAvailableSeats: async (runId: string, originId: string, destinationId: string) => {
    const res = await api.get<{ items: Seat[] }>(
      `/api/v1/train-runs/${runId}/available-seats?originStationId=${originId}&destinationStationId=${destinationId}`,
    );
    return res.data;
  },

  getSeatMap: async (runId: string, originId: string, destinationId: string) => {
    const res = await api.get<{ items: Seat[] }>(
      `/api/v1/train-runs/${runId}/seat-map?originStationId=${originId}&destinationStationId=${destinationId}`,
    );
    return res.data;
  },

  getQuote: async (params: {
    runId: string;
    originStationId: string;
    destinationStationId: string;
    seatId: string;
  }) => {
    const res = await api.post<FareQuote>("/api/v1/fare-quotes", {
      trainRunId: params.runId,
      seatId: params.seatId,
      originStationId: params.originStationId,
      destinationStationId: params.destinationStationId,
    });
    return res.data;
  },

  createBooking: async (body: CreateBookingRequest) => {
    const idempotencyKey = crypto.randomUUID();
    const res = await api.post<Booking>("/api/v1/bookings", body, {
      headers: { "Idempotency-Key": idempotencyKey },
    });
    return res.data;
  },

  cancelBooking: async (bookingId: string) => {
    const res = await api.post<Booking>(`/api/v1/bookings/${bookingId}/cancel`, {});
    return res.data;
  },

  getBookingByReference: async (reference: string) => {
    const res = await api.get<Booking>(
      `/api/v1/bookings/reference/${encodeURIComponent(reference.trim())}`,
    );
    return res.data;
  },
};

export function getAvailableSeats(runId: string, originId: string, destinationId: string) {
  return bookApi.getAvailableSeats(runId, originId, destinationId);
}

export function getSeatMap(runId: string, originId: string, destinationId: string) {
  return bookApi.getSeatMap(runId, originId, destinationId);
}

export function getQuote(params: {
  runId: string;
  originStationId: string;
  destinationStationId: string;
  seatId: string;
}) {
  return bookApi.getQuote(params);
}

export function createBooking(body: CreateBookingRequest) {
  return bookApi.createBooking(body);
}

export function cancelBooking(bookingId: string) {
  return bookApi.cancelBooking(bookingId);
}

export function getBookingByReference(reference: string) {
  return bookApi.getBookingByReference(reference);
}

export default bookApi;
