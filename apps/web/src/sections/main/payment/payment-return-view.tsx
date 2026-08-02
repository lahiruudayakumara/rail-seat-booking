import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { useNavigate, useSearchParams } from "react-router-dom";
import { clearStoredPayHereAccess, getPayHerePayment, getStoredPayHereAccess } from "@/api";
import { Button, Check, LoadingSpinner, RefreshCw } from "@/components";
import { BookingConfirmation } from "@/sections/main/booking/booking-confirmation";
import { setBooking, setTicket } from "@/store/slices/booking-slice";
import { useAppDispatch, useAppSelector } from "@/store";

export function PaymentReturnView({ cancelled = false }: { cancelled?: boolean }) {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const confirmedBooking = useAppSelector((state) => state.booking.booking);
  const paymentId = params.get("paymentId") ?? "";
  const access = paymentId ? getStoredPayHereAccess(paymentId) : undefined;
  const paymentQuery = useQuery({
    queryKey: ["payhere-payment", paymentId],
    queryFn: () => getPayHerePayment(paymentId, access?.bookingToken ?? ""),
    enabled: Boolean(paymentId && access?.bookingToken),
    retry: 2,
    refetchInterval: (query) => query.state.data?.status === "PENDING" ? 2_000 : false,
  });

  useEffect(() => {
    if (paymentQuery.data?.status !== "PAID" || !paymentQuery.data.ticket) return;
    dispatch(setBooking(paymentQuery.data.booking));
    dispatch(setTicket(paymentQuery.data.ticket));
    clearStoredPayHereAccess(paymentId);
  }, [dispatch, paymentId, paymentQuery.data]);

  if (confirmedBooking && paymentQuery.data?.status === "PAID") return <BookingConfirmation />;

  if (!paymentId || !access) {
    return <PaymentState title="Payment session unavailable" body="This payment session is no longer available in this browser. Use your booking reference to check the reservation." action={() => navigate("/lookup")} actionLabel="Lookup booking" />;
  }

  if (paymentQuery.isError) {
    return <PaymentState title="We could not check the payment" body="Your reservation has not been confirmed on this screen. Retry safely; this does not charge you again." action={() => paymentQuery.refetch()} actionLabel="Check again" />;
  }

  if (paymentQuery.data?.status === "FAILED") {
    clearStoredPayHereAccess(paymentId);
    return <PaymentState title="Payment was not completed" body="The payment was cancelled or failed. No ticket was issued." action={() => navigate("/")} actionLabel="Start another booking" />;
  }

  if (paymentQuery.data?.status === "DISPUTED") {
    return <PaymentState title="Payment requires review" body="Do not pay again. Keep your booking reference and contact support so the payment can be reconciled or refunded." action={() => navigate("/lookup")} actionLabel="View booking lookup" />;
  }

  return (
    <section className="mx-auto max-w-xl rounded-2xl border border-stone-200 bg-white p-8 text-center shadow-xl" aria-live="polite">
      <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-amber-100 text-[#6b1724]"><RefreshCw className="animate-spin" size={26} /></div>
      <p className="section-kicker mt-5">PAYHERE SANDBOX</p>
      <h2 className="font-heading text-3xl font-extrabold text-stone-900">{cancelled ? "Checking the cancelled payment" : "Confirming your payment"}</h2>
      <p className="mt-3 text-sm leading-6 text-stone-600">PayHere sends confirmation directly to our server. Keep this page open while we verify it and issue your ticket.</p>
      <div className="mt-6"><LoadingSpinner label="Waiting for verified payment notification" /></div>
      <button type="button" className="mt-5 text-sm font-bold text-[#6b1724] underline underline-offset-4" onClick={() => paymentQuery.refetch()}>Check now</button>
    </section>
  );
}

function PaymentState({ title, body, action, actionLabel }: { title: string; body: string; action: () => void; actionLabel: string }) {
  return <section className="mx-auto max-w-xl rounded-2xl border border-stone-200 bg-white p-8 text-center shadow-xl"><div className="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-stone-100 text-[#6b1724]"><Check size={26} /></div><h2 className="mt-5 font-heading text-3xl font-extrabold text-stone-900">{title}</h2><p className="mt-3 text-sm leading-6 text-stone-600">{body}</p><Button className="mt-6" onClick={action}>{actionLabel}</Button></section>;
}
