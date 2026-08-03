import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import axios from "axios";
import { useForm } from "react-hook-form";
import { useEffect } from "react";
import { z } from "zod";
import type { ApiError, CreateBookingRequest } from "@/types";
import { cancelBooking, checkoutSandbox, checkoutSandboxGroup, createBooking, createBookingGroup, paymentProvider, redirectToPayHere, startPayHereCheckout, startPayHereGroupCheckout } from "../api";
import { useAppDispatch, useAppSelector } from "../store";
import {
  setBooking,
  setBookingGroup,
  setGroupTickets,
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
import { usePassengerAuth } from "@/auth/use-passenger-auth";

const passengerSchema = z.object({
  fullName: z.string().trim().min(2, "Full name is required"),
  email: z.string().trim().refine((value) => !value || z.email().safeParse(value).success, "Enter a valid email address"),
  phone: z.string().trim().refine((value) => !value || /^\+[1-9]\d{7,14}$/.test(value), "Use international format, for example +94770000000"),
  billingAddress: z.string().trim(),
  city: z.string().trim(),
}).refine((value) => Boolean(value.email || value.phone), {
  message: "Enter an email address or phone number",
  path: ["email"],
}).refine((value) => paymentProvider !== "payhere" || Boolean(value.email && value.phone), {
  message: "PayHere requires both an email address and phone number",
  path: ["phone"],
}).refine((value) => paymentProvider !== "payhere" || value.billingAddress.length >= 3, {
  message: "Billing address is required for PayHere",
  path: ["billingAddress"],
}).refine((value) => paymentProvider !== "payhere" || value.city.length >= 2, {
  message: "City is required for PayHere",
  path: ["city"],
});

export type PassengerFormValues = z.infer<typeof passengerSchema>;

export function useBookingFlow() {
  const dispatch = useAppDispatch();
  const { selectedSeat, selectedSeats, quote, hold, booking, ticket, group, tickets } = useAppSelector((state) => state.booking);
  const { notice } = useAppSelector((state) => state.ui);
  const { runId, seatsQuery } = useSeatSelection();
  const { originId, destinationId } = useJourneySearch();
  const { account } = usePassengerAuth();

  const form = useForm<PassengerFormValues>({
    resolver: zodResolver(passengerSchema),
    defaultValues: { fullName: "", email: "", phone: "", billingAddress: "", city: "" },
  });

  useEffect(() => {
    if (!account) return;
    form.setValue("fullName", account.fullName);
    form.setValue("email", account.email);
    form.setValue("phone", account.phone ?? "");
  }, [account, form]);

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
        if (paymentProvider === "payhere") {
          const session = await startPayHereCheckout(heldBooking.id, heldBooking.managementToken, values.billingAddress, values.city);
          return { kind: "payhere" as const, session, bookingToken: heldBooking.managementToken };
        }
        return { kind: "confirmed" as const, result: await checkoutSandbox(heldBooking.id, heldBooking.managementToken) };
      });
    },
    onSuccess: (data) => {
      if (data.kind === "payhere") {
        dispatch(setNotice("Redirecting securely to PayHere Sandbox…"));
        redirectToPayHere(data.session, data.bookingToken);
        return;
      }
      dispatch(setBooking(data.result.booking));
      dispatch(setTicket(data.result.ticket));
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

  const groupMutation = useMutation({
    mutationFn: async ({ values, passengers }: { values: PassengerFormValues; passengers: Record<string, CreateBookingRequest["passenger"]> }) => {
      if (!runId || selectedSeats.length < 2) throw new Error("Select at least two seats for a group booking");
      const heldGroup = await createBookingGroup({
        members: selectedSeats.map(({ seat, quote: seatQuote, hold: seatHold }) => ({
          holdId: seatHold.id,
          holdToken: seatHold.managementToken,
          fareQuoteId: seatQuote.id,
          trainRunId: runId,
          seatId: seat.id,
          originStationId: originId,
          destinationStationId: destinationId,
          passenger: passengers[seat.id],
        })),
      });
      if (!heldGroup.managementToken) throw new Error("Group payment token missing");
      if (paymentProvider === "payhere") {
        const session = await startPayHereGroupCheckout(heldGroup.id, heldGroup.managementToken, values.billingAddress, values.city);
        return { kind: "group-payhere" as const, session, bookingToken: heldGroup.managementToken };
      }
      return { kind: "group-confirmed" as const, result: await checkoutSandboxGroup(heldGroup.id, heldGroup.managementToken) };
    },
    onSuccess: (data) => {
      if (data.kind === "group-payhere") {
        dispatch(setNotice("Redirecting your group securely to PayHere Sandbox…"));
        redirectToPayHere(data.session, data.bookingToken);
        return;
      }
      dispatch(setBookingGroup(data.result.group));
      dispatch(setGroupTickets(data.result.tickets));
      dispatch(setNotice("Group booking confirmed with one payment and reference."));
      void seatsQuery.refetch();
    },
    onError: (err: Error) => {
      const message = axios.isAxiosError<ApiError>(err) ? err.response?.data?.message : err.message;
      dispatch(setNotice(message || "Group booking could not be completed."));
      if (axios.isAxiosError(err) && err.response?.status === 409) void seatsQuery.refetch();
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

  const submitGroup = (values: PassengerFormValues, passengers: Record<string, CreateBookingRequest["passenger"]>) => {
    groupMutation.mutate({ values, passengers });
  };

  const startOver = () => {
    dispatch(setBooking(undefined));
    dispatch(setTicket(undefined));
    dispatch(setBookingGroup(undefined));
    dispatch(setGroupTickets([]));
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
    group,
    tickets,
    selectedSeat,
    quote,
    bookingMutation,
    groupMutation,
    cancelMutation,
    submitPassenger,
    submitGroup,
    startOver,
  };
}
