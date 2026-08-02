import { useTranslation } from "react-i18next";
import { useState } from "react";
import { Button, Check } from "@/components";
import { useBookingFlow } from "@/hooks/use-booking-flow";
import { formatMoney } from "@/utils";

export function BookingConfirmation() {
  const { t } = useTranslation();
  const { booking, ticket, cancelMutation, startOver } = useBookingFlow();
  const [copied, setCopied] = useState(false);
  const [confirmCancel, setConfirmCancel] = useState(false);

  if (!booking) return null;

  return (
    <section className="confirmation" aria-live="polite">
      <div className="confirmation-icon">
        <Check size={28} />
      </div>

      <p className="section-kicker">{t("confirmation.kicker", { status: booking.status })}</p>
      <h2 className="font-heading text-3xl font-extrabold text-stone-900">
        {t("confirmation.title")}
      </h2>
      <p className="mt-2 text-sm text-stone-600">{t("confirmation.sub")}</p>
      <strong className="reference">{booking.reference}</strong>
      <button
        type="button"
        className="mt-3 text-sm font-bold text-[#6b1724] underline decoration-[#6b1724]/30 underline-offset-4"
        onClick={async () => {
          await navigator.clipboard.writeText(booking.reference);
          setCopied(true);
          window.setTimeout(() => setCopied(false), 2000);
        }}
      >
        {copied ? t("common.copied") : t("common.copyReference")}
      </button>
      {ticket && (
        <div className="mt-4 rounded-lg border border-stone-300 bg-white p-3">
          <span className="block text-xs font-bold uppercase text-stone-500">
            {t("confirmation.ticketCode")}
          </span>
          <code className="break-all text-xs text-stone-800">{ticket.verificationCode}</code>
        </div>
      )}

      <div className="mt-6 flex flex-wrap justify-center gap-3">
        <span className="status-pill">
          Seat {booking.seat.coachCode} · {booking.seat.label}
        </span>
        <span className="status-pill">{formatMoney(booking.fare)}</span>
      </div>

      {booking.refund && (
        <p className="mt-4 text-sm font-semibold text-emerald-700">
          {t("confirmation.refunded", {
            amount: formatMoney({
              amountMinor: booking.refund.amountMinor,
              currency: booking.refund.currency,
              currencyScale: booking.fare.currencyScale,
            }),
          })}
        </p>
      )}

      {booking.status === "CONFIRMED" && (
        <div className="mt-6 flex flex-col items-center gap-3">
          <div className="flex flex-wrap justify-center gap-3">
            <Button onClick={startOver}>{t("common.bookAnother")}</Button>
            {!confirmCancel && <Button variant="secondary" onClick={() => setConfirmCancel(true)}>{t("confirmation.cancelButton")}</Button>}
          </div>
          {confirmCancel && (
            <div className="w-full max-w-lg rounded-xl border border-amber-200 bg-amber-50 p-4 text-left" role="alert">
              <strong className="text-sm text-stone-900">{t("confirmation.cancelConfirmTitle")}</strong>
              <p className="mt-1 text-xs leading-5 text-stone-600">{t("confirmation.cancelConfirmBody")}</p>
              <div className="mt-3 flex flex-wrap gap-2">
                <Button variant="secondary" onClick={() => setConfirmCancel(false)}>{t("confirmation.keepBooking")}</Button>
                <Button onClick={() => cancelMutation.mutate(booking.id)} disabled={cancelMutation.isPending}>
                  {cancelMutation.isPending ? t("confirmation.cancelling") : t("confirmation.confirmCancellation")}
                </Button>
              </div>
            </div>
          )}
        </div>
      )}
    </section>
  );
}
