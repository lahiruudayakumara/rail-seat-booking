import { useTranslation } from "react-i18next";
import {
  ArrowRight,
  Button,
  Check,
  Field,
  InlineError,
  LoadingSpinner,
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
  const { selectedSeat, quote, quoteMutation } = useSeatSelection();
  const { form, bookingMutation, submitPassenger } = useBookingFlow();

  if (!selectedSeat) return null;

  const originName = origin ? t(`stations.${origin.name}`, origin.name) : "";
  const destinationName = destination ? t(`stations.${destination.name}`, destination.name) : "";

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

          <Button
            variant="primary"
            className="justify-center"
            disabled={!quote || bookingMutation.isPending}
          >
            {bookingMutation.isPending ? t("passenger.confirming") : t("passenger.confirmButton")}
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
          ) : quote ? (
            <>
              <div className="fare-total">
                <span>{t("passenger.total")}</span>
                <strong>{formatMoney(quote)}</strong>
              </div>
              <p className="mt-3 text-xs text-white/70">
                {t("passenger.distanceTimer", {
                  distance: quote.distanceKm,
                  expires: new Date(quote.expiresAt).toLocaleTimeString([], {
                    hour: "2-digit",
                    minute: "2-digit",
                  }),
                })}
              </p>
            </>
          ) : (
            <InlineError message={t("passenger.quoteError")} />
          )}
        </aside>
      </div>
    </SectionCard>
  );
}
