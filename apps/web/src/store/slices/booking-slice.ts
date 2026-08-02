import { createSlice, type PayloadAction } from "@reduxjs/toolkit";
import type { Booking, FareQuote, Seat } from "@/types";

export interface BookingState {
  runId: string;
  selectedSeat?: Seat;
  quote?: FareQuote;
  booking?: Booking;
}

const initialState: BookingState = {
  runId: "",
  selectedSeat: undefined,
  quote: undefined,
  booking: undefined,
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
    setBooking: (state, action: PayloadAction<Booking | undefined>) => {
      state.booking = action.payload;
    },
    resetBookingSelection: (state) => {
      state.runId = "";
      state.selectedSeat = undefined;
      state.quote = undefined;
    },
    clearSelectedSeatAndQuote: (state) => {
      state.selectedSeat = undefined;
      state.quote = undefined;
    },
  },
});

export const {
  setRunId,
  setSelectedSeat,
  setQuote,
  setBooking,
  resetBookingSelection,
  clearSelectedSeatAndQuote,
} = bookingSlice.actions;

export default bookingSlice.reducer;
