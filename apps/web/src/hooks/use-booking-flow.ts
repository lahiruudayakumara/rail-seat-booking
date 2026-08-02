import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import axios from "axios";
import { useForm } from "react-hook-form";
import { z } from "zod";
import type { ApiError } from "@/types";
import { cancelBooking, createBooking } from "../api";
import { useAppDispatch, useAppSelector } from "../store";
import {
  setBooking,
  setHold,
  setQuote,
  setRunId,
  setSelectedSeat,
} from "../store/slices/booking-slice";
import { setSearched } from "../store/slices/search-slice";
import { setNotice } from "../store/slices/ui-slice";
import { useJourneySearch } from "./use-journey-search";
import { useSeatSelection } from "./use-seat-selection";

const passengerSchema = z.object({
  fullName: z.string().min(2, "Full name is required"),
  email: z.string().email("Valid email is required"),
  phone: z.string().min(8, "Phone number is required"),
});

export type PassengerFormValues = z.infer<typeof passengerSchema>;

export function useBookingFlow() {
  const dispatch = useAppDispatch();
  const { selectedSeat, quote, hold, booking } = useAppSelector((state) => state.booking);
  const { notice } = useAppSelector((state) => state.ui);
  const { runId, seatsQuery } = useSeatSelection();
  const { originId, destinationId } = useJourneySearch();

  const form = useForm<PassengerFormValues>({
    resolver: zodResolver(passengerSchema),
    defaultValues: { fullName: "", email: "", phone: "" },
  });

  const bookingMutation = useMutation({
    mutationFn: (values: PassengerFormValues) => {
      if (!runId || !selectedSeat || !quote || !hold) throw new Error("Booking selection missing");
      return createBooking({
        holdId: hold.id,
        holdToken: hold.managementToken,
        fareQuoteId: quote.id,
        trainRunId: runId,
        seatId: selectedSeat.id,
        originStationId: originId,
        destinationStationId: destinationId,
        passenger: {
          fullName: values.fullName,
          email: values.email,
          phone: values.phone,
        },
      });
    },
    onSuccess: (data) => {
      dispatch(setBooking(data));
      dispatch(setNotice("Booking confirmed safely. Hold your reference tight."));
      void seatsQuery.refetch();
    },
    onError: (err: Error) => {
      if (axios.isAxiosError<ApiError>(err)) {
        const apiError = err.response?.data;
        if (err.response?.status === 409) {
          dispatch(setSelectedSeat(undefined));
          dispatch(setQuote(undefined));
          dispatch(setHold(undefined));
          void seatsQuery.refetch();
        }
        dispatch(setNotice(apiError?.message || "Failed to create booking"));
        return;
      }
      dispatch(setNotice(err.message || "Failed to create booking"));
    },
  });

  const cancelMutation = useMutation({
    mutationFn: (bookingId: string) => {
      if (!booking?.managementToken) throw new Error("Booking access verification is required");
      return cancelBooking(bookingId, booking.managementToken);
    },
    onSuccess: () => {
      dispatch(setBooking(undefined));
      dispatch(setSelectedSeat(undefined));
      dispatch(setQuote(undefined));
      dispatch(setHold(undefined));
      dispatch(setRunId(""));
      dispatch(setSearched(false));
      dispatch(setNotice("Booking cancelled. You can search again whenever ready."));
    },
  });

  const submitPassenger = (values: PassengerFormValues) => {
    bookingMutation.mutate(values);
  };

  return {
    form,
    notice,
    booking,
    selectedSeat,
    quote,
    bookingMutation,
    cancelMutation,
    submitPassenger,
  };
}
