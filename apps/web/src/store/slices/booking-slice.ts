import { createSlice, type PayloadAction } from "@reduxjs/toolkit";
import type { Booking, BookingHold, FareQuote, Seat, Ticket } from "@/types";

export interface BookingState {
  runId: string;
  selectedSeat?: Seat;
  quote?: FareQuote;
  hold?: BookingHold;
  booking?: Booking;
  ticket?: Ticket;
}

const initialState: BookingState = {
  runId: "",
  selectedSeat: undefined,
  quote: undefined,
  hold: undefined,
  booking: undefined,
  ticket: undefined,
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
    resetBookingSelection: (state) => {
      state.runId = "";
      state.selectedSeat = undefined;
      state.quote = undefined;
      state.hold = undefined;
    },
    clearSelectedSeatAndQuote: (state) => {
      state.selectedSeat = undefined;
      state.quote = undefined;
      state.hold = undefined;
    },
  },
});

export const {
  setRunId,
  setSelectedSeat,
  setQuote,
  setHold,
  setBooking,
  setTicket,
  resetBookingSelection,
  clearSelectedSeatAndQuote,
} = bookingSlice.actions;

export default bookingSlice.reducer;
