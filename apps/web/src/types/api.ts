export type Route = {
  id: string;
  code: string;
  name: string;
  direction: string;
  timezone: string;
};

export type Station = {
  id: string;
  code: string;
  name: string;
  position?: number;
  cumulativeDistanceKm?: string;
};

export type TrainRun = {
  id: string;
  routeId: string;
  trainId: string;
  serviceDate: string;
  departureAt: string;
  arrivalAt: string;
  status: string;
};

export type Seat = {
  id: string;
  label: string;
  coachId: string;
  coachCode: string;
  coachClass: string;
  attributes: string[];
  availabilityStatus?: "AVAILABLE" | "BOOKED";
};

export type FareQuote = {
  id: string;
  distanceKm: string;
  amountMinor: number;
  currency: string;
  currencyScale: number;
  breakdown: Record<string, number>;
  expiresAt: string;
};

export type BookingHold = {
  id: string;
  status: "HELD";
  expiresAt: string;
  managementToken: string;
};

export type Ticket = {
  id: string;
  verificationCode: string;
  status: "ACTIVE" | "CANCELLED" | "USED";
};

export type CheckoutResult = {
  payment: { id: string; status: string; provider: string; providerReference: string };
  ticket: Ticket;
  booking: Booking;
};

export type Money = {
  amountMinor: number;
  currency: string;
  currencyScale: number;
};

export type Booking = {
  id: string;
  reference: string;
  status: string;
  trainRunId: string;
  seat: Seat;
  originStationId: string;
  destinationStationId: string;
  fare: Money;
  createdAt: string;
  confirmedAt?: string;
  cancelledAt?: string;
  managementToken?: string;
  refund?: { id: string; status: string; amountMinor: number; currency: string };
};

export type CreateBookingRequest = {
  holdId: string;
  holdToken: string;
  fareQuoteId: string;
  trainRunId: string;
  seatId: string;
  originStationId: string;
  destinationStationId: string;
  passenger: {
    fullName: string;
    email: string;
    phone: string;
  };
};

export type ApiError = {
  code: string;
  message: string;
  details?: unknown;
  requestId?: string;
};

export type AdminDashboard = {
  trainRunId: string;
  sellableSeatSegments: number;
  occupiedSeatSegments: number;
  segmentUtilizationPercent: number;
  confirmedBookings: number;
  cancelledBookings: number;
  heldBookings: number;
  grossRevenueMinor: number;
  refundedMinor: number;
  netRevenueMinor: number;
  currency: string;
  pendingDeliveries: number;
};
