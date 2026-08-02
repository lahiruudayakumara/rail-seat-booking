import { api } from "./api-instance";
import type { Booking, BookingHold, CheckoutResult, CreateBookingRequest, FareQuote, PayHereCheckoutSession, PayHerePaymentStatus, Seat } from "@/types";

export const paymentProvider = import.meta.env.VITE_PAYMENT_PROVIDER === "payhere" ? "payhere" : "sandbox";

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

  createHold: async (fareQuoteId: string) => {
    const res = await api.post<BookingHold>("/api/v1/booking-holds", { fareQuoteId });
    return res.data;
  },

  createBooking: async (body: CreateBookingRequest) => {
    const idempotencyKey = crypto.randomUUID();
    const res = await api.post<Booking>("/api/v1/bookings", body, {
      headers: { "Idempotency-Key": idempotencyKey },
    });
    return res.data;
  },

  checkoutSandbox: async (bookingId: string, bookingToken: string) => {
    const res = await api.post<CheckoutResult>(
      "/api/v1/payments/sandbox",
      { bookingId, bookingToken },
      { headers: { "Idempotency-Key": crypto.randomUUID() } },
    );
    return res.data;
  },

  startPayHereCheckout: async (bookingId: string, bookingToken: string, billingAddress: string, city: string) => {
    const res = await api.post<PayHereCheckoutSession>(
      "/api/v1/payments/payhere",
      { bookingId, bookingToken, billingAddress, city },
      { headers: { "Idempotency-Key": crypto.randomUUID() } },
    );
    return res.data;
  },

  getPayHerePayment: async (paymentId: string, bookingToken: string) => {
    const res = await api.get<PayHerePaymentStatus>(`/api/v1/payments/${paymentId}`, {
      headers: { Authorization: `Bearer ${bookingToken}` },
    });
    return res.data;
  },

  cancelBooking: async (bookingId: string, managementToken?: string) => {
    const res = await api.post<Booking>(`/api/v1/bookings/${bookingId}/cancel`, {}, {
      headers: managementToken ? { Authorization: `Bearer ${managementToken}` } : undefined,
    });
    return res.data;
  },

  getBookingByReference: async (reference: string, contact: string) => {
    const res = await api.post<Booking>("/api/v1/bookings/access", {
      reference: reference.trim(),
      contact: contact.trim(),
    });
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

export function createHold(fareQuoteId: string) {
  return bookApi.createHold(fareQuoteId);
}

export function createBooking(body: CreateBookingRequest) {
  return bookApi.createBooking(body);
}

export function checkoutSandbox(bookingId: string, bookingToken: string) {
  return bookApi.checkoutSandbox(bookingId, bookingToken);
}

export function startPayHereCheckout(bookingId: string, bookingToken: string, billingAddress: string, city: string) {
  return bookApi.startPayHereCheckout(bookingId, bookingToken, billingAddress, city);
}

export function getPayHerePayment(paymentId: string, bookingToken: string) {
  return bookApi.getPayHerePayment(paymentId, bookingToken);
}

export function redirectToPayHere(session: PayHereCheckoutSession, bookingToken: string) {
  sessionStorage.setItem(`rail-payhere-${session.paymentId}`, JSON.stringify({ bookingId: session.bookingId, bookingToken }));
  const form = document.createElement("form");
  form.method = "POST";
  form.action = session.actionUrl;
  for (const [name, value] of Object.entries(session.fields)) {
    const input = document.createElement("input");
    input.type = "hidden";
    input.name = name;
    input.value = value;
    form.appendChild(input);
  }
  document.body.appendChild(form);
  form.submit();
}

export function getStoredPayHereAccess(paymentId: string) {
  const raw = sessionStorage.getItem(`rail-payhere-${paymentId}`);
  if (!raw) return undefined;
  try {
    return JSON.parse(raw) as { bookingId: string; bookingToken: string };
  } catch {
    return undefined;
  }
}

export function clearStoredPayHereAccess(paymentId: string) {
  sessionStorage.removeItem(`rail-payhere-${paymentId}`);
}

export function cancelBooking(bookingId: string, managementToken?: string) {
  return bookApi.cancelBooking(bookingId, managementToken);
}

export function getBookingByReference(reference: string, contact: string) {
  return bookApi.getBookingByReference(reference, contact);
}

export default bookApi;
