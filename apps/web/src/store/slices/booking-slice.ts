import { createSlice, type PayloadAction } from "@reduxjs/toolkit";
import type { Booking, BookingGroup, BookingHold, FareQuote, Seat, SeatSelectionItem, Ticket } from "@/types";

export interface BookingState {
  runId: string;
  selectedSeat?: Seat;
  selectedSeats: SeatSelectionItem[];
  quote?: FareQuote;
  hold?: BookingHold;
  booking?: Booking;
  ticket?: Ticket;
  group?: BookingGroup;
  tickets: Ticket[];
}

const initialState: BookingState = {
  runId: "",
  selectedSeat: undefined,
  selectedSeats: [],
  quote: undefined,
  hold: undefined,
  booking: undefined,
  ticket: undefined,
  group: undefined,
  tickets: [],
};

export const bookingSlice = createSlice({
  name: "booking",
  initialState,
  reducers: {
    setRunId: (state, action: PayloadAction<string>) => {
      state.runId = action.payload;
    },
    setSelectedSeat: (state, action: PayloadAction<Seat | undefined>) => {
      state.selectedSeat = action.payload;
      if (!action.payload) state.selectedSeats = [];
    },
    addSeatSelection: (state, action: PayloadAction<SeatSelectionItem>) => {
      const index = state.selectedSeats.findIndex((item) => item.seat.id === action.payload.seat.id);
      if (index >= 0) state.selectedSeats[index] = action.payload;
      else if (state.selectedSeats.length < 6) state.selectedSeats.push(action.payload);
      state.selectedSeat = state.selectedSeats[0]?.seat;
      state.quote = state.selectedSeats[0]?.quote;
      state.hold = state.selectedSeats[0]?.hold;
    },
    removeSeatSelection: (state, action: PayloadAction<string>) => {
      state.selectedSeats = state.selectedSeats.filter((item) => item.seat.id !== action.payload);
      state.selectedSeat = state.selectedSeats[0]?.seat;
      state.quote = state.selectedSeats[0]?.quote;
      state.hold = state.selectedSeats[0]?.hold;
    },
    setQuote: (state, action: PayloadAction<FareQuote | undefined>) => {
      state.quote = action.payload;
    },
    setHold: (state, action: PayloadAction<BookingHold | undefined>) => {
      state.hold = action.payload;
    },
    setBooking: (state, action: PayloadAction<Booking | undefined>) => {
      state.booking = action.payload;
    },
    setTicket: (state, action: PayloadAction<Ticket | undefined>) => {
      state.ticket = action.payload;
    },
    setBookingGroup: (state, action: PayloadAction<BookingGroup | undefined>) => {
      state.group = action.payload;
    },
    setGroupTickets: (state, action: PayloadAction<Ticket[]>) => {
      state.tickets = action.payload;
    },
    resetBookingSelection: (state) => {
      state.runId = "";
      state.selectedSeat = undefined;
      state.selectedSeats = [];
      state.quote = undefined;
      state.hold = undefined;
      state.group = undefined;
      state.tickets = [];
    },
    clearSelectedSeatAndQuote: (state) => {
      state.selectedSeat = undefined;
      state.selectedSeats = [];
      state.quote = undefined;
      state.hold = undefined;
    },
  },
});

export const {
  setRunId,
  setSelectedSeat,
  addSeatSelection,
  removeSeatSelection,
  setQuote,
  setHold,
  setBooking,
  setTicket,
  setBookingGroup,
  setGroupTickets,
  resetBookingSelection,
  clearSelectedSeatAndQuote,
} = bookingSlice.actions;

export default bookingSlice.reducer;
