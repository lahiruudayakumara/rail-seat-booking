import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { QRCodeSVG } from "qrcode.react";
import { useEffect, useId, useMemo, useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { cancelBooking, getPassengerBookings, getStations, getTrainRun } from "@/api";
import { usePassengerAuth } from "@/auth/use-passenger-auth";
import { ArrowRight, Button, Calendar, Clock, ConfirmationModal, Eye, EyeOff, LoadingSpinner, Ticket, TrainFront, Users } from "@/components";
import type { ApiError, Booking, Station, TrainRun } from "@/types";
import { formatMoney } from "@/utils";
import { SavedPassengersPreferences } from "./saved-passengers-preferences";
import { filterAndSortJourneys, journeyCategory, type JourneyFilter, type JourneySort } from "./journey-list";

type Mode = "login" | "register";

function errorMessage(error: unknown) {
  if (axios.isAxiosError<ApiError>(error)) return error.response?.data?.message ?? "Request failed.";
  return error instanceof Error ? error.message : "Request failed.";
}

function PasswordField({ label, name, autoComplete }: { label: string; name: string; autoComplete: string }) {
  const [visible, setVisible] = useState(false);
  const inputId = useId();

  return (
    <div className="grid min-w-0 gap-1.5 text-sm font-bold text-stone-700">
      <label htmlFor={inputId}>{label}</label>
      <div className="relative">
        <input
          id={inputId}
          className="form-input pr-11"
          type={visible ? "text" : "password"}
          name={name}
          autoComplete={autoComplete}
          minLength={10}
          maxLength={72}
          required
        />
        <button
          type="button"
          className="absolute inset-y-0 right-0 flex w-11 items-center justify-center rounded-r-[10px] text-stone-400 transition hover:bg-stone-50 hover:text-[#6b1724] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-[#6b1724]"
          aria-label={`${visible ? "Hide" : "Show"} ${label.toLowerCase()}`}
          aria-pressed={visible}
          onClick={() => setVisible((current) => !current)}
        >
          {visible ? <EyeOff size={18} aria-hidden="true" /> : <Eye size={18} aria-hidden="true" />}
        </button>
      </div>
    </div>
  );
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
      <section className="mx-auto max-w-2xl overflow-hidden rounded-2xl border border-stone-200 bg-white shadow-xl">
        <div className="bg-[#6b1724] px-5 py-6 text-white sm:px-7 sm:py-7">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-white/10"><Users size={22} /></div>
          <h2 className="mt-4 font-heading text-3xl font-extrabold">Passenger account</h2>
          <p className="mt-2 text-sm leading-6 text-white/75">Sign in for faster checkout and private access to all your journeys. Guest booking remains available.</p>
        </div>
        <div className="p-5 sm:p-6 md:p-8">
          <div className="grid grid-cols-2 rounded-xl bg-stone-100 p-1" role="tablist" aria-label="Passenger account access">
            {(["login", "register"] as const).map((item) => (
              <button key={item} type="button" role="tab" aria-selected={mode === item} onClick={() => { setMode(item); setError(""); }} className={`rounded-lg px-4 py-2.5 text-sm font-bold transition ${mode === item ? "bg-white text-[#6b1724] shadow-sm" : "text-stone-500"}`}>
                {item === "login" ? "Sign in" : "Create account"}
              </button>
            ))}
          </div>
          <form className={`mt-6 grid gap-4 ${mode === "register" ? "sm:grid-cols-2" : ""}`} onSubmit={submit}>
            {mode === "register" && <label className="grid min-w-0 gap-1.5 text-sm font-bold text-stone-700"><span>Full name</span><input className="form-input" name="fullName" autoComplete="name" minLength={2} maxLength={120} required /></label>}
            {mode === "register" && <label className="grid min-w-0 gap-1.5 text-sm font-bold text-stone-700"><span>Phone <span className="font-normal text-stone-400">(optional)</span></span><input className="form-input" type="tel" name="phone" autoComplete="tel" placeholder="+94770000000" pattern="\+[1-9][0-9]{7,14}" /></label>}
            <label className={`grid min-w-0 gap-1.5 text-sm font-bold text-stone-700 ${mode === "register" ? "sm:col-span-2" : ""}`}><span>Email address</span><input className="form-input" type="email" name="email" autoComplete="email" required /></label>
            <PasswordField label="Password" name="password" autoComplete={mode === "login" ? "current-password" : "new-password"} />
            {mode === "register" && <>
              <PasswordField label="Confirm password" name="confirmPassword" autoComplete="new-password" />
              <p className="-mt-1 text-xs leading-5 text-stone-500 sm:col-span-2">Use 10 or more characters with upper-case, lower-case, and a number.</p>
            </>}
            {error && <p className={`rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 ${mode === "register" ? "sm:col-span-2" : ""}`} role="alert">{error}</p>}
            <Button className={`w-full ${mode === "register" ? "sm:col-span-2" : ""}`} type="submit" disabled={submitting}>{submitting ? "Please wait…" : mode === "login" ? "Sign in securely" : "Create passenger account"}</Button>
          </form>
        </div>
      </section>
    );
  }

  return <SignedInAccount account={account} onLogout={async () => { await logout(); navigate("/"); }} />;
}

function SignedInAccount({ account, onLogout }: { account: { fullName: string; email: string; phone?: string }; onLogout: () => Promise<void> }) {
  const pageSize = 4;
  const queryClient = useQueryClient();
  const [cancelId, setCancelId] = useState<string>();
  const [journeyFilter, setJourneyFilter] = useState<JourneyFilter>("ALL");
  const [journeySearch, setJourneySearch] = useState("");
  const [journeySort, setJourneySort] = useState<JourneySort>("BOOKED_DESC");
  const [journeyPage, setJourneyPage] = useState(1);
  const bookingsQuery = useQuery({ queryKey: ["passenger-bookings"], queryFn: getPassengerBookings });
  const bookingRunIds = [...new Set(bookingsQuery.data?.map((booking) => booking.trainRunId) ?? [])];
  const journeyDetailsQuery = useQuery({
    queryKey: ["passenger-journey-details", bookingRunIds],
    enabled: bookingRunIds.length > 0,
    queryFn: async () => {
      const [stationPage, runs] = await Promise.all([
        getStations(),
        Promise.all(bookingRunIds.map((runId) => getTrainRun(runId))),
      ]);
      return {
        stations: Object.fromEntries(stationPage.items.map((station) => [station.id, station])),
        runs: Object.fromEntries(runs.map((run) => [run.id, run])),
      } as { stations: Record<string, Station>; runs: Record<string, TrainRun> };
    },
  });
    
  const cancelMutation = useMutation({
    mutationFn: (bookingId: string) => cancelBooking(bookingId, ""),
    onSuccess: () => { setCancelId(undefined); void queryClient.invalidateQueries({ queryKey: ["passenger-bookings"] }); },
  });
  const cancellingBooking = bookingsQuery.data?.find((booking) => booking.id === cancelId);
  const bookings = useMemo(() => bookingsQuery.data ?? [], [bookingsQuery.data]);
  const filteredBookings = useMemo(
    () => filterAndSortJourneys(bookings, journeyDetailsQuery.data, journeyFilter, journeySearch, journeySort),
    [bookings, journeyDetailsQuery.data, journeyFilter, journeySearch, journeySort],
  );
  const journeyCounts = useMemo(() => {
    const counts: Record<JourneyFilter, number> = { ALL: bookings.length, UPCOMING: 0, PAST: 0, CANCELLED: 0 };
    bookings.forEach((booking) => { counts[journeyCategory(booking, journeyDetailsQuery.data?.runs[booking.trainRunId])] += 1; });
    return counts;
  }, [bookings, journeyDetailsQuery.data]);
  const pageCount = Math.max(1, Math.ceil(filteredBookings.length / pageSize));
  const visibleBookings = filteredBookings.slice((journeyPage - 1) * pageSize, journeyPage * pageSize);

  useEffect(() => { setJourneyPage(1); }, [journeyFilter, journeySearch, journeySort]);
  useEffect(() => { setJourneyPage((current) => Math.min(current, pageCount)); }, [pageCount]);

  return (
    <section className="grid gap-6">
      <div className="flex flex-col gap-5 rounded-2xl border border-stone-200 bg-white p-6 shadow-sm sm:flex-row sm:items-center sm:justify-between">
        <div><p className="section-kicker">PASSENGER ACCOUNT</p><h2 className="font-heading text-3xl font-extrabold text-stone-900">Welcome, {account.fullName}</h2><p className="mt-1 text-sm text-stone-500">{account.email}{account.phone ? ` · ${account.phone}` : ""}</p></div>
        <Button variant="secondary" onClick={onLogout}>Sign out</Button>
      </div>
      <SavedPassengersPreferences />
      <div className="rounded-2xl border border-stone-200 bg-white p-6 shadow-sm md:p-8">
        <div className="flex flex-wrap items-end justify-between gap-3"><div><p className="section-kicker">MY JOURNEYS</p><h3 className="font-heading text-2xl font-extrabold text-stone-900">Your bookings</h3></div><Button onClick={() => window.location.assign("/")}>Book a new journey</Button></div>
        {bookingsQuery.isLoading && <div className="py-10"><LoadingSpinner label="Loading your bookings" /></div>}
        {bookingsQuery.isError && <p className="mt-6 rounded-lg bg-red-50 p-4 text-sm text-red-700">{errorMessage(bookingsQuery.error)}</p>}
        {bookingsQuery.data?.length === 0 && <div className="mt-6 rounded-xl border border-dashed border-stone-300 p-8 text-center"><Ticket className="mx-auto text-stone-400" size={30} /><p className="mt-3 font-bold text-stone-800">No account bookings yet</p><p className="mt-1 text-sm text-stone-500">Your next signed-in reservation will appear here.</p></div>}
        {bookings.length > 0 && (
          <div className="mt-6 grid gap-4 border-y border-stone-200 py-4">
            <div className="flex gap-2 overflow-x-auto pb-1" role="tablist" aria-label="Filter your journeys">
              {(["ALL", "UPCOMING", "PAST", "CANCELLED"] as JourneyFilter[]).map((filter) => (
                <button
                  key={filter}
                  type="button"
                  role="tab"
                  aria-selected={journeyFilter === filter}
                  onClick={() => setJourneyFilter(filter)}
                  className={`shrink-0 rounded-full px-3.5 py-2 text-xs font-extrabold transition ${journeyFilter === filter ? "bg-[#6b1724] text-white shadow-sm" : "bg-stone-100 text-stone-600 hover:bg-stone-200"}`}
                >
                  {filter === "ALL" ? "All" : filter === "UPCOMING" ? "Upcoming" : filter === "PAST" ? "Past" : "Cancelled"}
                  <span className={`ml-2 rounded-full px-1.5 py-0.5 text-[10px] ${journeyFilter === filter ? "bg-white/20" : "bg-white"}`}>{journeyCounts[filter]}</span>
                </button>
              ))}
            </div>
            <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_auto]">
              <label className="relative min-w-0">
                <span className="sr-only">Search bookings</span>
                <input
                  type="search"
                  className="form-input"
                  value={journeySearch}
                  onChange={(event) => setJourneySearch(event.target.value)}
                  placeholder="Search reference, station, coach or seat"
                />
              </label>
              <label className="grid grid-cols-[auto_minmax(0,1fr)] items-center gap-2 text-xs font-bold text-stone-500">
                <span>Sort</span>
                <select className="form-input min-w-40" value={journeySort} onChange={(event) => setJourneySort(event.target.value as JourneySort)}>
                  <option value="BOOKED_DESC">Recently booked</option>
                  <option value="BOOKED_ASC">Oldest booked</option>
                  <option value="TRAVEL_ASC">Travel date</option>
                </select>
              </label>
            </div>
            <p className="text-xs font-semibold text-stone-500" role="status">Showing {filteredBookings.length} of {bookings.length} bookings</p>
          </div>
        )}
        <div className="mt-6 grid gap-5">
          {visibleBookings.map((booking) => (
            <JourneyTicket
              key={booking.id}
              booking={booking}
              origin={journeyDetailsQuery.data?.stations[booking.originStationId]}
              destination={journeyDetailsQuery.data?.stations[booking.destinationStationId]}
              run={journeyDetailsQuery.data?.runs[booking.trainRunId]}
              onCancel={() => { cancelMutation.reset(); setCancelId(booking.id); }}
            />
          ))}
          {bookings.length > 0 && filteredBookings.length === 0 && (
            <div className="rounded-xl border border-dashed border-stone-300 px-5 py-10 text-center"><Ticket className="mx-auto text-stone-400" size={28} /><p className="mt-3 font-bold text-stone-800">No matching journeys</p><p className="mt-1 text-sm text-stone-500">Try another filter or clear your search.</p><button type="button" className="mt-4 text-sm font-bold text-[#6b1724] underline underline-offset-4" onClick={() => { setJourneyFilter("ALL"); setJourneySearch(""); }}>Clear filters</button></div>
          )}
        </div>
        {filteredBookings.length > pageSize && (
          <nav className="mt-6 flex flex-col items-center justify-between gap-3 border-t border-stone-200 pt-5 sm:flex-row" aria-label="Bookings pagination">
            <p className="text-xs font-semibold text-stone-500">Page {journeyPage} of {pageCount}</p>
            <div className="flex items-center gap-1.5">
              <button type="button" className="rounded-lg border border-stone-200 px-3 py-2 text-xs font-bold text-stone-700 disabled:cursor-not-allowed disabled:opacity-40" disabled={journeyPage === 1} onClick={() => setJourneyPage((page) => page - 1)}>Previous</button>
              {Array.from({ length: pageCount }, (_, index) => index + 1).map((page) => (
                <button key={page} type="button" aria-label={`Go to bookings page ${page}`} aria-current={journeyPage === page ? "page" : undefined} className={`h-9 min-w-9 rounded-lg text-xs font-extrabold ${journeyPage === page ? "bg-[#6b1724] text-white" : "border border-stone-200 text-stone-600 hover:bg-stone-50"}`} onClick={() => setJourneyPage(page)}>{page}</button>
              ))}
              <button type="button" className="rounded-lg border border-stone-200 px-3 py-2 text-xs font-bold text-stone-700 disabled:cursor-not-allowed disabled:opacity-40" disabled={journeyPage === pageCount} onClick={() => setJourneyPage((page) => page + 1)}>Next</button>
            </div>
          </nav>
        )}
      </div>
      <ConfirmationModal
        open={Boolean(cancellingBooking)}
        title="Cancel and refund this booking?"
        description={`Booking ${cancellingBooking?.reference ?? ""} will be cancelled and its seat released for this journey segment. This action cannot be undone.`}
        confirmLabel="Confirm cancellation"
        cancelLabel="Keep booking"
        tone="danger"
        isPending={cancelMutation.isPending}
        error={cancelMutation.isError ? errorMessage(cancelMutation.error) : undefined}
        onClose={() => { cancelMutation.reset(); setCancelId(undefined); }}
        onConfirm={() => { if (cancelId) cancelMutation.mutate(cancelId); }}
      />
    </section>
  );
}

function JourneyTicket({ booking, origin, destination, run, onCancel }: {
  booking: Booking;
  origin?: Station;
  destination?: Station;
  run?: TrainRun;
  onCancel: () => void;
}) {
  const departure = run ? new Date(run.departureAt) : undefined;
  const arrival = run ? new Date(run.arrivalAt) : undefined;
  const travelDate = run?.serviceDate
    ? new Date(`${run.serviceDate}T00:00:00`).toLocaleDateString("en-LK", { weekday: "short", day: "numeric", month: "short", year: "numeric" })
    : "Journey date unavailable";
  const time = (value?: Date) => value?.toLocaleTimeString("en-LK", { hour: "2-digit", minute: "2-digit" }) ?? "—";
  const statusClass = booking.status === "CONFIRMED"
    ? "bg-emerald-100 text-emerald-800"
    : booking.status === "CANCELLED"
      ? "bg-red-100 text-red-700"
      : "bg-stone-100 text-stone-700";

  return (
    <article className="overflow-hidden rounded-2xl border border-stone-200 bg-white shadow-[0_8px_28px_rgba(54,35,28,0.08)]" aria-label={`Booking ${booking.reference}`}>
      <header className="flex flex-wrap items-center justify-between gap-3 bg-[#6b1724] px-5 py-3 text-white md:px-6">
        <div className="flex items-center gap-2.5"><TrainFront size={19} /><span className="text-xs font-extrabold uppercase tracking-[0.18em]">Lanka Rail Reserve</span></div>
        <span className={`rounded-full px-3 py-1 text-[11px] font-extrabold tracking-wide ${statusClass}`}>{booking.status}</span>
      </header>

      <div className="ticket-body">
        <div className="ticket-main !items-stretch !p-0">
          <div className="w-full px-5 py-5 md:px-6 md:py-6">
            <div className="grid grid-cols-[1fr_auto_1fr] items-center gap-3">
              <StationPoint station={origin} fallback="Origin" time={time(departure)} />
              <div className="flex min-w-16 items-center gap-1 text-stone-300" aria-hidden="true"><span className="h-px flex-1 bg-stone-300" /><ArrowRight size={20} className="text-[#6b1724]" /></div>
              <StationPoint station={destination} fallback="Destination" time={time(arrival)} align="right" />
            </div>

            <div className="mt-5 flex items-center gap-2 border-t border-stone-100 pt-4 text-sm font-semibold text-stone-600"><Calendar size={16} className="text-[#6b1724]" /><span>{travelDate}</span></div>
            <dl className="mt-4 grid grid-cols-2 gap-x-5 gap-y-4 sm:grid-cols-4">
              <TicketDetail label="Coach" value={booking.seat.coachCode} />
              <TicketDetail label="Seat" value={booking.seat.label} />
              <TicketDetail label="Class" value={formatCoachClass(booking.seat.coachClass)} />
              <TicketDetail label="Fare" value={formatMoney(booking.fare)} accent />
            </dl>
          </div>
        </div>

        <div className="ticket-divider" aria-hidden="true"><span className="notch notch-top" /><span className="line" /><span className="notch notch-bottom" /></div>

        <aside className="ticket-stub !w-full !items-start !bg-[#faf7f2] md:!w-52 md:!px-5">
          <div className="w-full">
            <span className="block text-[10px] font-extrabold uppercase tracking-[0.16em] text-stone-400">Booking reference</span>
            <strong className="mt-1 block break-all font-mono text-base font-extrabold text-[#6b1724]">{booking.reference}</strong>
          </div>
          <div className="flex w-full items-end justify-between gap-4 md:mt-5 md:items-start">
            <div>
            <span className="block text-[10px] font-extrabold uppercase tracking-[0.16em] text-stone-400">Reserved on</span>
            <span className="mt-1 block text-xs font-bold text-stone-600">{new Date(booking.createdAt).toLocaleDateString("en-LK", { day: "numeric", month: "short", year: "numeric" })}</span>
            </div>
            <div className="shrink-0 rounded-lg border border-stone-200 bg-white p-1.5 shadow-sm" aria-label={`QR code for booking ${booking.reference}`}>
              <QRCodeSVG value={booking.reference} size={58} bgColor="#ffffff" fgColor="#6b1724" level="M" />
            </div>
          </div>
        </aside>
      </div>

      {booking.status === "CONFIRMED" && (
        <footer className="border-t border-stone-100 bg-white px-5 py-4 md:px-6">
          <div className="flex flex-wrap items-center justify-between gap-3"><p className="text-xs leading-5 text-stone-500">Plans changed? Review the refund before cancelling this journey.</p><button type="button" className="text-sm font-bold text-red-700 underline decoration-red-200 underline-offset-4 hover:text-red-800" onClick={onCancel}>Cancel booking</button></div>
        </footer>
      )}
    </article>
  );
}

function StationPoint({ station, fallback, time, align = "left" }: { station?: Station; fallback: string; time: string; align?: "left" | "right" }) {
  return <div className={align === "right" ? "text-right" : "text-left"}><span className="block font-mono text-2xl font-black tracking-tight text-stone-900">{station?.code ?? "—"}</span><span className="mt-0.5 block text-xs font-bold text-stone-500">{station?.name ?? fallback}</span><span className={`mt-2 flex items-center gap-1 text-xs font-extrabold text-[#6b1724] ${align === "right" ? "justify-end" : "justify-start"}`}><Clock size={13} />{time}</span></div>;
}

function TicketDetail({ label, value, accent = false }: { label: string; value: string; accent?: boolean }) {
  return <div><dt className="text-[10px] font-extrabold uppercase tracking-[0.14em] text-stone-400">{label}</dt><dd className={`mt-1 text-sm font-extrabold ${accent ? "text-[#6b1724]" : "text-stone-900"}`}>{value}</dd></div>;
}

function formatCoachClass(value: string) {
  return value.toLowerCase().replaceAll("_", " ").replace(/\b\w/g, (letter) => letter.toUpperCase());
}
