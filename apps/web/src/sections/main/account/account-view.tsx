import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { cancelBooking, getPassengerBookings } from "@/api";
import { usePassengerAuth } from "@/auth/use-passenger-auth";
import { Button, LoadingSpinner, Ticket, Users } from "@/components";
import type { ApiError } from "@/types";
import { formatMoney } from "@/utils";

type Mode = "login" | "register";

function errorMessage(error: unknown) {
  if (axios.isAxiosError<ApiError>(error)) return error.response?.data?.message ?? "Request failed.";
  return error instanceof Error ? error.message : "Request failed.";
}

export function AccountView() {
  const { account, isLoading, login, register, logout } = usePassengerAuth();
  const [mode, setMode] = useState<Mode>("login");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const navigate = useNavigate();

  if (isLoading) {
    return <div className="rounded-2xl bg-white p-10 shadow-sm"><LoadingSpinner label="Checking your passenger account" /></div>;
  }

  if (!account) {
    const submit = async (event: FormEvent<HTMLFormElement>) => {
      event.preventDefault();
      setError("");
      setSubmitting(true);
      const data = new FormData(event.currentTarget);
      try {
        if (mode === "login") {
          await login({ email: String(data.get("email")), password: String(data.get("password")) });
        } else {
          const password = String(data.get("password"));
          if (password !== String(data.get("confirmPassword"))) throw new Error("Passwords do not match.");
          await register({
            fullName: String(data.get("fullName")),
            email: String(data.get("email")),
            phone: String(data.get("phone")),
            password,
          });
        }
      } catch (submitError) {
        setError(errorMessage(submitError));
      } finally {
        setSubmitting(false);
      }
    };

    return (
      <section className="mx-auto max-w-xl overflow-hidden rounded-2xl border border-stone-200 bg-white shadow-xl">
        <div className="bg-[#6b1724] px-7 py-7 text-white">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-white/10"><Users size={22} /></div>
          <h2 className="mt-4 font-heading text-3xl font-extrabold">Passenger account</h2>
          <p className="mt-2 text-sm leading-6 text-white/75">Sign in for faster checkout and private access to all your journeys. Guest booking remains available.</p>
        </div>
        <div className="p-6 md:p-8">
          <div className="grid grid-cols-2 rounded-xl bg-stone-100 p-1" role="tablist" aria-label="Passenger account access">
            {(["login", "register"] as const).map((item) => (
              <button key={item} type="button" role="tab" aria-selected={mode === item} onClick={() => { setMode(item); setError(""); }} className={`rounded-lg px-4 py-2.5 text-sm font-bold transition ${mode === item ? "bg-white text-[#6b1724] shadow-sm" : "text-stone-500"}`}>
                {item === "login" ? "Sign in" : "Create account"}
              </button>
            ))}
          </div>
          <form className="mt-6 grid gap-4" onSubmit={submit}>
            {mode === "register" && <label className="grid gap-1.5 text-sm font-bold text-stone-700">Full name<input className="form-input" name="fullName" autoComplete="name" minLength={2} maxLength={120} required /></label>}
            <label className="grid gap-1.5 text-sm font-bold text-stone-700">Email address<input className="form-input" type="email" name="email" autoComplete="email" required /></label>
            {mode === "register" && <label className="grid gap-1.5 text-sm font-bold text-stone-700">Phone <span className="font-normal text-stone-400">(optional)</span><input className="form-input" type="tel" name="phone" autoComplete="tel" placeholder="+94770000000" pattern="\+[1-9][0-9]{7,14}" /></label>}
            <label className="grid gap-1.5 text-sm font-bold text-stone-700">Password<input className="form-input" type="password" name="password" autoComplete={mode === "login" ? "current-password" : "new-password"} minLength={10} maxLength={72} required /></label>
            {mode === "register" && <>
              <label className="grid gap-1.5 text-sm font-bold text-stone-700">Confirm password<input className="form-input" type="password" name="confirmPassword" autoComplete="new-password" minLength={10} maxLength={72} required /></label>
              <p className="-mt-1 text-xs leading-5 text-stone-500">Use 10 or more characters with upper-case, lower-case, and a number.</p>
            </>}
            {error && <p className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700" role="alert">{error}</p>}
            <Button type="submit" disabled={submitting}>{submitting ? "Please wait…" : mode === "login" ? "Sign in securely" : "Create passenger account"}</Button>
          </form>
        </div>
      </section>
    );
  }

  return <SignedInAccount account={account} onLogout={async () => { await logout(); navigate("/"); }} />;
}

function SignedInAccount({ account, onLogout }: { account: { fullName: string; email: string; phone?: string }; onLogout: () => Promise<void> }) {
  const queryClient = useQueryClient();
  const [cancelId, setCancelId] = useState<string>();
  const bookingsQuery = useQuery({ queryKey: ["passenger-bookings"], queryFn: getPassengerBookings });
  const cancelMutation = useMutation({
    mutationFn: (bookingId: string) => cancelBooking(bookingId, ""),
    onSuccess: () => { setCancelId(undefined); void queryClient.invalidateQueries({ queryKey: ["passenger-bookings"] }); },
  });

  return (
    <section className="grid gap-6">
      <div className="flex flex-col gap-5 rounded-2xl border border-stone-200 bg-white p-6 shadow-sm sm:flex-row sm:items-center sm:justify-between">
        <div><p className="section-kicker">PASSENGER ACCOUNT</p><h2 className="font-heading text-3xl font-extrabold text-stone-900">Welcome, {account.fullName}</h2><p className="mt-1 text-sm text-stone-500">{account.email}{account.phone ? ` · ${account.phone}` : ""}</p></div>
        <Button variant="secondary" onClick={onLogout}>Sign out</Button>
      </div>
      <div className="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm md:p-8">
        <div className="flex flex-wrap items-end justify-between gap-3"><div><p className="section-kicker">MY JOURNEYS</p><h3 className="font-heading text-2xl font-extrabold text-stone-900">Your bookings</h3></div><Button onClick={() => window.location.assign("/")}>Book a new journey</Button></div>
        {bookingsQuery.isLoading && <div className="py-10"><LoadingSpinner label="Loading your bookings" /></div>}
        {bookingsQuery.isError && <p className="mt-6 rounded-lg bg-red-50 p-4 text-sm text-red-700">{errorMessage(bookingsQuery.error)}</p>}
        {bookingsQuery.data?.length === 0 && <div className="mt-6 rounded-xl border border-dashed border-stone-300 p-8 text-center"><Ticket className="mx-auto text-stone-400" size={30} /><p className="mt-3 font-bold text-stone-800">No account bookings yet</p><p className="mt-1 text-sm text-stone-500">Your next signed-in reservation will appear here.</p></div>}
        <div className="mt-6 grid gap-4">
          {bookingsQuery.data?.map((booking) => (
            <article key={booking.id} className="rounded-xl border border-stone-200 p-4 md:p-5">
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div><span className="text-xs font-bold uppercase tracking-wider text-stone-400">Booking reference</span><strong className="mt-1 block font-mono text-lg text-[#6b1724]">{booking.reference}</strong><p className="mt-2 text-sm text-stone-600">Coach {booking.seat.coachCode} · Seat {booking.seat.label} · {formatMoney(booking.fare)}</p><p className="mt-1 text-xs text-stone-400">Booked {new Date(booking.createdAt).toLocaleDateString()}</p></div>
                <span className={`rounded-full px-3 py-1 text-xs font-bold ${booking.status === "CONFIRMED" ? "bg-emerald-100 text-emerald-700" : "bg-stone-100 text-stone-600"}`}>{booking.status}</span>
              </div>
              {booking.status === "CONFIRMED" && <div className="mt-4 border-t border-stone-100 pt-4">
                {cancelId !== booking.id ? <button type="button" className="text-sm font-bold text-red-700 underline underline-offset-4" onClick={() => setCancelId(booking.id)}>Cancel booking</button> : <div className="rounded-lg border border-amber-200 bg-amber-50 p-3"><p className="text-sm font-bold text-stone-800">Cancel and refund this booking?</p><div className="mt-3 flex gap-2"><Button variant="secondary" onClick={() => setCancelId(undefined)}>Keep booking</Button><Button disabled={cancelMutation.isPending} onClick={() => cancelMutation.mutate(booking.id)}>{cancelMutation.isPending ? "Cancelling…" : "Confirm cancellation"}</Button></div>{cancelMutation.isError && <p className="mt-2 text-xs text-red-700">{errorMessage(cancelMutation.error)}</p>}</div>}
              </div>}
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
