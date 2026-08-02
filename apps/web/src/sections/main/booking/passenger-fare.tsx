import { useTranslation } from "react-i18next";
import { useEffect, useState } from "react";
import {
  ArrowRight,
  Button,
  Check,
  Field,
  InlineError,
  LoadingSpinner,
  RefreshCw,
  SectionCard,
  Users,
} from "@/components";
import { useBookingFlow } from "@/hooks/use-booking-flow";
import { useJourneySearch } from "@/hooks/use-journey-search";
import { useSeatSelection } from "@/hooks/use-seat-selection";
import { formatMoney } from "@/utils";

export function PassengerFare() {
  const { t } = useTranslation();
  const { origin, destination } = useJourneySearch();
  const { selectedSeat, quote, hold, quoteMutation } = useSeatSelection();
  const { form, bookingMutation, submitPassenger } = useBookingFlow();

  const [secondsLeft, setSecondsLeft] = useState(0);
  useEffect(() => {
    const update = () => setSecondsLeft(Math.max(0, Math.floor((new Date(hold?.expiresAt ?? 0).getTime() - Date.now()) / 1000)));
    update();
    const timer = window.setInterval(update, 1000);
    return () => window.clearInterval(timer);
  }, [hold?.expiresAt]);

  if (!selectedSeat) return null;

  const originName = origin ? t(`stations.${origin.name}`, origin.name) : "";
  const destinationName = destination ? t(`stations.${destination.name}`, destination.name) : "";
  const holdExpired = Boolean(hold) && secondsLeft === 0;
  const minutes = Math.floor(secondsLeft / 60);
  const seconds = String(secondsLeft % 60).padStart(2, "0");

  return (
    <SectionCard title={t("passenger.title")} icon={<Users size={22} />}>
      <div className="grid gap-8 md:grid-cols-[1fr_320px]">
        <form className="grid gap-4" onSubmit={form.handleSubmit(submitPassenger)}>
          <Field label={t("passenger.fullName")} error={form.formState.errors.fullName?.message}>
            <input
              placeholder={t("passenger.fullNamePlaceholder")}
              autoComplete="name"
              {...form.register("fullName")}
            />
          </Field>

          <div className="grid gap-4 md:grid-cols-2">
            <Field label={t("passenger.email")} error={form.formState.errors.email?.message}>
              <input type="email" autoComplete="email" {...form.register("email")} />
            </Field>

            <Field label={t("passenger.phone")} error={form.formState.errors.phone?.message}>
              <input
                placeholder="+94770000000"
                autoComplete="tel"
                {...form.register("phone")}
              />
            </Field>
          </div>

          <p className="text-xs leading-5 text-stone-500">{t("passenger.contactHint")}</p>

          <Button
            variant="primary"
            className="justify-center"
            disabled={!quote || !hold || holdExpired || bookingMutation.isPending}
          >
            {bookingMutation.isPending ? t("passenger.confirming") : quote ? t("passenger.payButton", { amount: formatMoney(quote) }) : t("passenger.confirmButton")}
            <Check size={16} />
          </Button>
        </form>

        <aside className="fare-card">
          <p className="section-kicker">{t("passenger.fareKicker")}</p>

          <div className="fare-row">
            <span>{originName}</span>
            <ArrowRight size={16} />
            <span>{destinationName}</span>
          </div>

          <div className="fare-row">
            <span>{t("passenger.coachSeat")}</span>
            <strong>
              {selectedSeat.coachCode} · {selectedSeat.label}
            </strong>
          </div>

          {quoteMutation.isPending ? (
            <LoadingSpinner label={t("passenger.calculating")} />
          ) : quoteMutation.isError ? (
            <div className="grid gap-3">
              <InlineError message={t("passenger.quoteError")} />
              <Button variant="secondary" onClick={() => quoteMutation.mutate(selectedSeat.id)}><RefreshCw size={15} /> {t("common.tryAgain")}</Button>
            </div>
          ) : quote ? (
            <>
              <div className="fare-total">
                <span>{t("passenger.total")}</span>
                <strong>{formatMoney(quote)}</strong>
              </div>
              <p className="mt-3 text-xs text-white/70">
                {t("passenger.distanceTimer", {
                  distance: quote.distanceKm,
                  expires: new Date(hold?.expiresAt ?? quote.expiresAt).toLocaleTimeString([], {
                    hour: "2-digit",
                    minute: "2-digit",
                  }),
                })}
              </p>
              <p className={`mt-2 font-mono text-xs font-bold ${holdExpired ? "text-amber-300" : "text-white"}`} aria-live="polite">
                {holdExpired ? t("passenger.holdExpired") : t("passenger.holdCountdown", { minutes, seconds })}
              </p>
              {holdExpired && (
                <Button className="mt-3 w-full" variant="secondary" onClick={() => quoteMutation.mutate(selectedSeat.id)}>
                  <RefreshCw size={15} /> {t("passenger.renewHold")}
                </Button>
              )}
            </>
          ) : (
            <InlineError message={t("passenger.quoteError")} />
          )}
        </aside>
      </div>
    </SectionCard>
  );
}
