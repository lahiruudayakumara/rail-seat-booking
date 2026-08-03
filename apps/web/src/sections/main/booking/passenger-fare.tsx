import { useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { getSavedTravellers, paymentProvider } from "@/api";
import { usePassengerAuth } from "@/auth/use-passenger-auth";
import { ArrowRight, Button, Check, Field, InlineError, LoadingSpinner, RefreshCw, SectionCard, Users } from "@/components";
import { useBookingFlow, type PassengerFormValues } from "@/hooks/use-booking-flow";
import { useJourneySearch } from "@/hooks/use-journey-search";
import { useSeatSelection } from "@/hooks/use-seat-selection";
import type { CreateBookingRequest } from "@/types";
import { formatMoney } from "@/utils";

type PassengerInput = CreateBookingRequest["passenger"];

export function PassengerFare() {
  const { t } = useTranslation();
  const { origin, destination } = useJourneySearch();
  const { selectedSeat, selectedSeats, quote, hold, quoteMutation } = useSeatSelection();
  const { form, bookingMutation, groupMutation, submitPassenger, submitGroup } = useBookingFlow();
  const { account } = usePassengerAuth();
  const navigate = useNavigate();
  const travellersQuery = useQuery({ queryKey: ["saved-travellers"], queryFn: getSavedTravellers, enabled: Boolean(account), staleTime: 60_000 });
  const [secondsLeft, setSecondsLeft] = useState(0);
  const [travellerId, setTravellerId] = useState("account");
  const [groupPassengers, setGroupPassengers] = useState<Record<string, PassengerInput>>({});
  const [assignmentSources, setAssignmentSources] = useState<Record<string, string>>({});
  const [groupError, setGroupError] = useState("");
  const isGroup = selectedSeats.length > 1;

  const holdExpiresAt = useMemo(() => selectedSeats.reduce<string | undefined>((earliest, item) => {
    if (!earliest) return item.hold.expiresAt;
    return new Date(item.hold.expiresAt).getTime() < new Date(earliest).getTime() ? item.hold.expiresAt : earliest;
  }, undefined) ?? hold?.expiresAt, [hold?.expiresAt, selectedSeats]);

  useEffect(() => {
    const update = () => setSecondsLeft(Math.max(0, Math.floor((new Date(holdExpiresAt ?? 0).getTime() - Date.now()) / 1000)));
    update();
    const timer = window.setInterval(update, 1000);
    return () => window.clearInterval(timer);
  }, [holdExpiresAt]);

  useEffect(() => {
    if (!isGroup) return;
    setGroupPassengers((current) => {
      const next = { ...current };
      selectedSeats.forEach((selection, index) => {
        if (next[selection.seat.id]) return;
        const saved = index === 0 ? account : travellersQuery.data?.[index - 1];
        next[selection.seat.id] = saved
          ? { fullName: saved.fullName, email: saved.email ?? account?.email ?? "", phone: saved.phone ?? account?.phone ?? "" }
          : { fullName: "", email: "", phone: "" };
      });
      return Object.fromEntries(Object.entries(next).filter(([seatId]) => selectedSeats.some((selection) => selection.seat.id === seatId)));
    });
    setAssignmentSources((current) => {
      const next = { ...current };
      selectedSeats.forEach((selection, index) => {
        if (!next[selection.seat.id]) next[selection.seat.id] = index === 0 && account ? "account" : travellersQuery.data?.[index - 1]?.id ?? "manual";
      });
      return next;
    });
  }, [account, isGroup, selectedSeats, travellersQuery.data]);

  useEffect(() => {
    if (!isGroup) return;
    const lead = groupPassengers[selectedSeats[0]?.seat.id ?? ""];
    if (!lead) return;
    form.setValue("fullName", lead.fullName);
    form.setValue("email", lead.email);
    form.setValue("phone", lead.phone);
  }, [form, groupPassengers, isGroup, selectedSeats]);

  if (!selectedSeat || !quote || !hold) return null;

  const originName = origin ? t(`stations.${origin.name}`, origin.name) : "";
  const destinationName = destination ? t(`stations.${destination.name}`, destination.name) : "";
  const holdExpired = Boolean(holdExpiresAt) && secondsLeft === 0;
  const minutes = Math.floor(secondsLeft / 60);
  const seconds = String(secondsLeft % 60).padStart(2, "0");
  const totalFare = { ...quote, amountMinor: selectedSeats.reduce((sum, item) => sum + item.quote.amountMinor, 0) };

  const submit = (values: PassengerFormValues) => {
    if (!isGroup) {
      submitPassenger(values);
      return;
    }
    const invalid = selectedSeats.some(({ seat }) => {
      const passenger = groupPassengers[seat.id];
      return !passenger?.fullName.trim() || (!passenger.email.trim() && !passenger.phone.trim());
    });
    if (invalid) {
      setGroupError("Every selected seat needs a passenger name and email or phone.");
      return;
    }
    setGroupError("");
    submitGroup(values, groupPassengers);
  };

  return (
    <SectionCard title={t("passenger.title")} icon={<Users size={22} />}>
      <div className={`mb-5 rounded-xl border p-4 text-sm ${account ? "border-emerald-200 bg-emerald-50 text-emerald-800" : "border-stone-200 bg-stone-50 text-stone-600"}`}>
        {account ? <><strong>Booking as {account.fullName}</strong><p className="mt-1 text-xs">{isGroup ? "Assign your saved travellers below. The group will be saved in your account." : "This journey will be saved in your passenger account."}</p></> : <div className="flex flex-wrap items-center justify-between gap-3"><div><strong className="text-stone-800">Continue as guest</strong><p className="mt-1 text-xs">Or sign in to assign saved travellers and speed up checkout.</p></div><Button type="button" variant="secondary" onClick={() => navigate("/account")}>Sign in</Button></div>}
      </div>

      <div className="grid gap-8 md:grid-cols-[1fr_320px]">
        <form className="grid gap-4" onSubmit={form.handleSubmit(submit)}>
          {!isGroup && account && Boolean(travellersQuery.data?.length) && <TravellerSelector value={travellerId} account={account} travellers={travellersQuery.data ?? []} onChange={(nextId) => {
            setTravellerId(nextId);
            const traveller = nextId === "account" ? account : travellersQuery.data?.find((item) => item.id === nextId);
            if (!traveller) return;
            form.setValue("fullName", traveller.fullName, { shouldValidate: true });
            form.setValue("email", traveller.email ?? account.email, { shouldValidate: true });
            form.setValue("phone", traveller.phone ?? account.phone ?? "", { shouldValidate: true });
          }} />}

          {!isGroup ? <IndividualPassengerFields form={form} t={t} /> : <div className="grid gap-3">
            <div><p className="section-kicker">GROUP PASSENGERS</p><h3 className="font-heading text-xl font-extrabold text-stone-900">Assign one traveller to each seat</h3><p className="mt-1 text-xs leading-5 text-stone-500">All seats share one payment and one group reference. Each passenger receives a separate ticket.</p></div>
            {selectedSeats.map(({ seat }, index) => {
              const passenger = groupPassengers[seat.id] ?? { fullName: "", email: "", phone: "" };
              const source = assignmentSources[seat.id] ?? "manual";
              const updatePassenger = (field: keyof PassengerInput, value: string) => setGroupPassengers((current) => ({ ...current, [seat.id]: { ...passenger, [field]: value } }));
              return <div key={seat.id} className="rounded-xl border border-stone-200 bg-stone-50 p-4">
                <div className="flex items-center justify-between gap-3"><div><span className="text-[10px] font-extrabold uppercase tracking-wider text-stone-400">Passenger {index + 1}</span><strong className="block text-sm text-[#6b1724]">Coach {seat.coachCode} · Seat {seat.label}</strong></div><span className="rounded-full bg-white px-2.5 py-1 text-[10px] font-bold text-stone-500 ring-1 ring-stone-200">{seat.coachClass}</span></div>
                {account && <label className="field mt-3"><span>Saved traveller</span><select value={source} onChange={(event) => {
                  const nextSource = event.target.value;
                  setAssignmentSources((current) => ({ ...current, [seat.id]: nextSource }));
                  const saved = nextSource === "account" ? account : travellersQuery.data?.find((item) => item.id === nextSource);
                  setGroupPassengers((current) => ({ ...current, [seat.id]: saved ? { fullName: saved.fullName, email: saved.email ?? account.email, phone: saved.phone ?? account.phone ?? "" } : { fullName: "", email: "", phone: "" } }));
                }}><option value="account">{account.fullName} (myself)</option>{travellersQuery.data?.map((traveller) => <option key={traveller.id} value={traveller.id}>{traveller.fullName}</option>)}<option value="manual">Enter manually</option></select></label>}
                <div className="mt-3 grid gap-3"><label className="field"><span>Full name</span><input value={passenger.fullName} onChange={(event) => updatePassenger("fullName", event.target.value)} minLength={2} maxLength={120} required /></label><div className="grid gap-3 sm:grid-cols-2"><label className="field"><span>Email</span><input type="email" value={passenger.email} onChange={(event) => updatePassenger("email", event.target.value)} /></label><label className="field"><span>Phone</span><input value={passenger.phone} onChange={(event) => updatePassenger("phone", event.target.value)} placeholder="+94770000000" pattern="\+[1-9][0-9]{7,14}" /></label></div></div>
              </div>;
            })}
            {groupError && <p className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-xs font-semibold text-red-700" role="alert">{groupError}</p>}
          </div>}

          {paymentProvider === "payhere" && <div className="grid gap-4 md:grid-cols-2"><Field label="Billing address" error={form.formState.errors.billingAddress?.message}><input autoComplete="street-address" placeholder="No. 1, Main Street" {...form.register("billingAddress")} /></Field><Field label="City" error={form.formState.errors.city?.message}><input autoComplete="address-level2" placeholder="Colombo" {...form.register("city")} /></Field></div>}
          <p className="text-xs leading-5 text-stone-500">{t("passenger.contactHint")}</p>
          <Button variant="primary" className="justify-center" disabled={holdExpired || bookingMutation.isPending || groupMutation.isPending}>
            {bookingMutation.isPending || groupMutation.isPending ? t("passenger.confirming") : isGroup ? `${paymentProvider === "payhere" ? "Continue to PayHere" : "Pay once & confirm group"} · ${formatMoney(totalFare)}` : paymentProvider === "payhere" ? `Continue to PayHere · ${formatMoney(quote)}` : t("passenger.payButton", { amount: formatMoney(quote) })}<Check size={16} />
          </Button>
        </form>

        <aside className="fare-card">
          <p className="section-kicker">{isGroup ? "GROUP FARE" : t("passenger.fareKicker")}</p>
          <div className="fare-row"><span>{originName}</span><ArrowRight size={16} /><span>{destinationName}</span></div>
          <div className="fare-row"><span>{t("passenger.coachSeat")}</span><strong>{isGroup ? `${selectedSeats.length} seats` : `${selectedSeat.coachCode} · ${selectedSeat.label}`}</strong></div>
          {isGroup && <div className="my-3 grid gap-1 border-y border-white/10 py-3 text-xs text-white/75">{selectedSeats.map(({ seat, quote: seatQuote }) => <div key={seat.id} className="flex justify-between gap-3"><span>{seat.coachCode} · {seat.label}</span><strong>{formatMoney(seatQuote)}</strong></div>)}</div>}
          {quoteMutation.isPending ? <LoadingSpinner label={t("passenger.calculating")} /> : quoteMutation.isError ? <div className="grid gap-3"><InlineError message={t("passenger.quoteError")} /><Button variant="secondary" onClick={() => quoteMutation.mutate(selectedSeat)}><RefreshCw size={15} /> {t("common.tryAgain")}</Button></div> : <><div className="fare-total"><span>{t("passenger.total")}</span><strong>{formatMoney(totalFare)}</strong></div><p className="mt-3 text-xs text-white/70">{t("passenger.distanceTimer", { distance: quote.distanceKm, expires: new Date(holdExpiresAt ?? quote.expiresAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }) })}</p><p className={`mt-2 font-mono text-xs font-bold ${holdExpired ? "text-amber-300" : "text-white"}`} aria-live="polite">{holdExpired ? t("passenger.holdExpired") : t("passenger.holdCountdown", { minutes, seconds })}</p>{holdExpired && <p className="mt-3 text-xs font-semibold text-amber-200">Return to the seat map and select the group again to renew every hold safely.</p>}</>}
        </aside>
      </div>
    </SectionCard>
  );
}

function IndividualPassengerFields({ form, t }: { form: ReturnType<typeof useBookingFlow>["form"]; t: (key: string) => string }) {
  return <><Field label={t("passenger.fullName")} error={form.formState.errors.fullName?.message}><input placeholder={t("passenger.fullNamePlaceholder")} autoComplete="name" {...form.register("fullName")} /></Field><div className="grid gap-4 md:grid-cols-2"><Field label={t("passenger.email")} error={form.formState.errors.email?.message}><input type="email" autoComplete="email" {...form.register("email")} /></Field><Field label={t("passenger.phone")} error={form.formState.errors.phone?.message}><input placeholder="+94770000000" autoComplete="tel" {...form.register("phone")} /></Field></div></>;
}

function TravellerSelector({ value, account, travellers, onChange }: { value: string; account: { fullName: string }; travellers: Array<{ id: string; fullName: string }>; onChange: (value: string) => void }) {
  return <label className="field"><span>Who is travelling?</span><select value={value} onChange={(event) => onChange(event.target.value)}><option value="account">{account.fullName} (myself)</option>{travellers.map((traveller) => <option key={traveller.id} value={traveller.id}>{traveller.fullName}</option>)}</select><small className="!text-stone-500">Selecting a saved passenger fills the details below.</small></label>;
}
