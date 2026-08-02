import { useTranslation } from "react-i18next";
import { Button, Check } from "@/components";
import { useBookingFlow } from "@/hooks/use-booking-flow";
import { formatMoney } from "@/utils";

export function BookingConfirmation() {
  const { t } = useTranslation();
  const { booking, cancelMutation } = useBookingFlow();

  if (!booking) return null;

  return (
    <section className="confirmation mt-8" aria-live="polite">
      <div className="confirmation-icon">
        <Check size={28} />
      </div>

      <p className="section-kicker">{t("confirmation.kicker", { status: booking.status })}</p>
      <h2 className="font-heading text-3xl font-extrabold text-stone-900">
        {t("confirmation.title")}
      </h2>
      <p className="mt-2 text-sm text-stone-600">{t("confirmation.sub")}</p>
      <strong className="reference">{booking.reference}</strong>

      <div className="mt-6 flex flex-wrap justify-center gap-3">
        <span className="status-pill">
          Seat {booking.seat.coachCode} · {booking.seat.label}
        </span>
        <span className="status-pill">{formatMoney(booking.fare)}</span>
      </div>

      {booking.status === "CONFIRMED" && (
        <div className="mt-6">
          <Button
            variant="secondary"
            onClick={() => cancelMutation.mutate(booking.id)}
            disabled={cancelMutation.isPending}
          >
            {cancelMutation.isPending ? t("confirmation.cancelling") : t("confirmation.cancelButton")}
          </Button>
        </div>
      )}
    </section>
  );
}
