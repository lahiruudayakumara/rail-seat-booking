import { useState } from "react";
import { useTranslation } from "react-i18next";
import { QRCodeSVG } from "qrcode.react";
import { Button, Check, ConfirmationModal, Copy, Download, TrainFront } from "@/components";
import { useBookingFlow } from "@/hooks/use-booking-flow";
import { formatMoney } from "@/utils";

export function BookingConfirmation() {
  const { t } = useTranslation();
  const { booking, ticket, cancelMutation, startOver } = useBookingFlow();

  const [copiedRef, setCopiedRef] = useState(false);
  const [copiedCode, setCopiedCode] = useState(false);
  const [confirmCancel, setConfirmCancel] = useState(false);

  if (!booking) return null;

  const handleCopyReference = async () => {
    try {
      await navigator.clipboard.writeText(booking.reference);
      setCopiedRef(true);
      setTimeout(() => setCopiedRef(false), 2000);
    } catch {
      // fallback
    }
  };

  const handleCopyCode = async (code: string) => {
    try {
      await navigator.clipboard.writeText(code);
      setCopiedCode(true);
      setTimeout(() => setCopiedCode(false), 2000);
    } catch {
      // fallback
    }
  };

  const handleDownloadTicket = () => {
    window.print();
  };

  const verificationCode = ticket?.verificationCode;

  // Formulate real QR code payload for verification
  const qrPayload = verificationCode ? JSON.stringify({
    ref: booking.reference,
    code: verificationCode,
    seat: `${booking.seat.coachCode}-${booking.seat.label}`,
    status: booking.status,
  }) : "";

  return (
    <section className="mx-auto mt-8 max-w-3xl" aria-live="polite">
      {/* Boarding Pass Ticket Container with id for print export */}
      <div
        id="printable-ticket"
        className="overflow-hidden rounded-2xl border border-stone-200 bg-white shadow-xl"
      >
        {/* Maroon Header Banner */}
        <div className="bg-[#6b1724] px-6 py-6 text-white md:px-8">
          <div className="flex flex-col items-center text-center">
            <div className="mb-3 flex h-14 w-14 items-center justify-center rounded-full bg-white/10 text-white backdrop-blur-md">
              <Check size={32} />
            </div>
            <p className="font-mono text-xs font-bold tracking-widest text-amber-300 uppercase">
              BOOKING {booking.status || "CONFIRMED"}
            </p>
            <h2 className="mt-1 font-heading text-3xl font-extrabold text-white md:text-4xl">
              {t("confirmation.title")}
            </h2>
            <p className="mt-1 text-sm text-stone-200">
              {t("confirmation.sub")}
            </p>
          </div>
        </div>

        {/* Ticket Main Content */}
        <div className="p-6 md:p-8">
          {/* Main Info Grid */}
          <div className="grid gap-6 md:grid-cols-[1fr_auto]">
            {/* Left Info Column */}
            <div className="flex flex-col gap-5">
              {/* Booking Reference Box */}
              {verificationCode && <div className="rounded-xl border border-stone-200/80 bg-stone-50/80 p-4">
                <span className="block font-mono text-[11px] font-bold uppercase tracking-wider text-stone-400">
                  {t("confirmation.referenceLabel")}
                </span>
                <div className="mt-1 flex flex-wrap items-center gap-3">
                  <strong className="font-mono text-2xl font-black text-[#6b1724] tracking-wider">
                    {booking.reference}
                  </strong>
                  <button
                    type="button"
                    onClick={handleCopyReference}
                    className="no-print inline-flex items-center gap-1.5 rounded-lg border border-stone-300 bg-white px-3 py-1.5 text-xs font-bold text-stone-700 shadow-sm transition hover:bg-stone-100 active:scale-95 cursor-pointer"
                  >
                    {copiedRef ? (
                      <>
                        <Check size={14} className="text-emerald-600" />
                        <span className="text-emerald-700">{t("common.copied")}</span>
                      </>
                    ) : (
                      <>
                        <Copy size={14} />
                        <span>{t("common.copyReference")}</span>
                      </>
                    )}
                  </button>
                </div>
              </div>}

              {/* Ticket Verification Code Box */}
              <div className="rounded-xl border border-stone-200/80 bg-stone-50/80 p-4">
                <span className="block font-mono text-[11px] font-bold uppercase tracking-wider text-stone-400">
                  {t("confirmation.ticketCode")}
                </span>
                <div className="mt-2 flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                  <code className="break-all rounded bg-stone-200/60 px-2.5 py-1.5 font-mono text-xs text-stone-800">
                    {verificationCode}
                  </code>
                  <button
                    type="button"
                    onClick={() => verificationCode && handleCopyCode(verificationCode)}
                    className="no-print inline-flex shrink-0 items-center justify-center gap-1.5 rounded-lg border border-stone-300 bg-white px-3 py-1.5 text-xs font-bold text-stone-700 shadow-sm transition hover:bg-stone-100 active:scale-95 cursor-pointer"
                  >
                    {copiedCode ? (
                      <>
                        <Check size={14} className="text-emerald-600" />
                        <span className="text-emerald-700">{t("common.copied")}</span>
                      </>
                    ) : (
                      <>
                        <Copy size={14} />
                        <span>Copy Code</span>
                      </>
                    )}
                  </button>
                </div>
              </div>

              {/* Seat & Fare Badges Grid */}
              <div className="grid grid-cols-2 gap-3">
                <div className="rounded-xl bg-[#6b1724]/5 border border-[#6b1724]/15 p-3.5">
                  <span className="block text-[10px] font-bold uppercase tracking-wider text-stone-500">
                    SEAT LOCATION
                  </span>
                  <strong className="mt-0.5 block font-mono text-lg font-bold text-[#6b1724]">
                    Seat {booking.seat.coachCode} · {booking.seat.label}
                  </strong>
                </div>

                <div className="rounded-xl bg-[#6b1724]/5 border border-[#6b1724]/15 p-3.5">
                  <span className="block text-[10px] font-bold uppercase tracking-wider text-stone-500">
                    TOTAL FARE
                  </span>
                  <strong className="mt-0.5 block font-mono text-lg font-bold text-[#6b1724]">
                    {formatMoney(booking.fare)}
                  </strong>
                </div>
              </div>
            </div>

            {/* Right Stub & Real Dynamic SVG QR Code Box */}
            {verificationCode && <div className="flex flex-col items-center justify-center rounded-xl border border-stone-200 bg-stone-50 p-5 text-center">
              <div className="rounded-xl bg-white p-3 shadow-md border border-stone-200">
                <QRCodeSVG
                  value={qrPayload}
                  size={120}
                  bgColor="#ffffff"
                  fgColor="#6b1724"
                  level="M"
                  includeMargin={false}
                />
              </div>
              <span className="mt-3 font-mono text-[10px] font-bold tracking-widest text-stone-400 uppercase">
                DIGITAL VERIFICATION
              </span>
              <div className="mt-1 flex items-center gap-1 text-xs font-bold text-stone-700">
                <TrainFront size={14} />
                <span>Lanka Rail Reserve</span>
              </div>
            </div>}
          </div>

          {/* Refund Notice if exists */}
          {booking.refund && (
            <div className="mt-6 rounded-xl border border-emerald-200 bg-emerald-50 p-4 text-emerald-800">
              <p className="text-sm font-semibold">
                {t("confirmation.refunded", {
                  amount: formatMoney({
                    amountMinor: booking.refund.amountMinor,
                    currency: booking.refund.currency,
                    currencyScale: booking.fare.currencyScale,
                  }),
                })}
              </p>
            </div>
          )}

          {/* Action Buttons Section */}
          {booking.status === "CONFIRMED" && (
            <div className="no-print mt-8 border-t border-stone-200 pt-6">
              <div className="flex flex-wrap items-center justify-center gap-4">
                <Button variant="primary" onClick={handleDownloadTicket}>
                  <Download size={18} />
                  <span>Download / Print Ticket</span>
                </Button>

                <Button variant="secondary" onClick={startOver}>
                  {t("common.bookAnother")}
                </Button>

                <Button
                  variant="secondary"
                  onClick={() => setConfirmCancel(true)}
                >
                  {t("confirmation.cancelButton")}
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>
      <ConfirmationModal
        open={confirmCancel && booking.status === "CONFIRMED"}
        title={t("confirmation.cancelConfirmTitle")}
        description={t("confirmation.cancelConfirmBody")}
        confirmLabel={t("confirmation.confirmCancellation")}
        cancelLabel={t("confirmation.keepBooking")}
        tone="danger"
        isPending={cancelMutation.isPending}
        onClose={() => setConfirmCancel(false)}
        onConfirm={() => cancelMutation.mutate(booking.id)}
      />
    </section>
  );
}
