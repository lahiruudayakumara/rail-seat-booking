import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import axios from "axios";
import { useForm } from "react-hook-form";
import { z } from "zod";
import type { ApiError } from "@/types";
import { cancelBooking, checkoutSandbox, createBooking } from "../api";
import { useAppDispatch, useAppSelector } from "../store";
import {
  setBooking,
  setHold,
  setQuote,
  setRunId,
  setSelectedSeat,
  setTicket,
} from "../store/slices/booking-slice";
import { setSearched } from "../store/slices/search-slice";
import { setNotice } from "../store/slices/ui-slice";
import { useJourneySearch } from "./use-journey-search";
import { useSeatSelection } from "./use-seat-selection";

const passengerSchema = z.object({
  fullName: z.string().trim().min(2, "Full name is required"),
  email: z.string().trim().refine((value) => !value || z.email().safeParse(value).success, "Enter a valid email address"),
  phone: z.string().trim().refine((value) => !value || /^\+[1-9]\d{7,14}$/.test(value), "Use international format, for example +94770000000"),
}).refine((value) => Boolean(value.email || value.phone), {
  message: "Enter an email address or phone number",
  path: ["email"],
});

export type PassengerFormValues = z.infer<typeof passengerSchema>;

export function useBookingFlow() {
  const dispatch = useAppDispatch();
  const { selectedSeat, quote, hold, booking, ticket } = useAppSelector((state) => state.booking);
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
      }).then(async (heldBooking) => {
        if (!heldBooking.managementToken) throw new Error("Booking payment token missing");
        return checkoutSandbox(heldBooking.id, heldBooking.managementToken);
      });
    },
    onSuccess: (data) => {
      dispatch(setBooking(data.booking));
      dispatch(setTicket(data.ticket));
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
      dispatch(setTicket(undefined));
      dispatch(setRunId(""));
      dispatch(setSearched(false));
      dispatch(setNotice("Booking cancelled. You can search again whenever ready."));
    },
    onError: (err: Error) => {
      const message = axios.isAxiosError<ApiError>(err) ? err.response?.data?.message : err.message;
      dispatch(setNotice(message || "Booking could not be cancelled. Please try again."));
    },
  });

  const submitPassenger = (values: PassengerFormValues) => {
    bookingMutation.mutate(values);
  };

  const startOver = () => {
    dispatch(setBooking(undefined));
    dispatch(setTicket(undefined));
    dispatch(setSelectedSeat(undefined));
    dispatch(setQuote(undefined));
    dispatch(setHold(undefined));
    dispatch(setRunId(""));
    dispatch(setSearched(false));
    dispatch(setNotice(""));
    form.reset();
  };

  return {
    form,
    notice,
    booking,
    ticket,
    selectedSeat,
    quote,
    bookingMutation,
    cancelMutation,
    submitPassenger,
    startOver,
  };
}
