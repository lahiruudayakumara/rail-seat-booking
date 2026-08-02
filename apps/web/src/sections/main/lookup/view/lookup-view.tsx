import { useState } from "react";
import { useTranslation } from "react-i18next";
import { Button, Field, InlineError, TrainFront } from "@/components";
import { useBookingFlow } from "@/hooks/use-booking-flow";
import { formatMoney } from "@/utils";

const LookupView = () => {
  const { t } = useTranslation();
  const { booking, cancelMutation } = useBookingFlow();
  const [refInput, setRefInput] = useState("");
  const [searched, setSearched] = useState(false);
  const [errorMsg, setErrorMsg] = useState("");

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (!refInput.trim()) return;
    setSearched(true);
    if (booking && booking.reference.toLowerCase() === refInput.trim().toLowerCase()) {
      setErrorMsg("");
    } else {
      setErrorMsg(t("lookup.notFound"));
    }
  };

  return (
    <section className="panel p-6 md:p-10 max-w-3xl mx-auto my-10">
      <div className="border-b border-stone-200 pb-5 mb-6">
        <p className="section-kicker">{t("nav.lookup")}</p>
        <h2 className="font-heading text-3xl font-extrabold text-stone-900">
          {t("lookup.title")}
        </h2>
        <p className="mt-2 text-sm text-stone-600">
          {t("lookup.sub")}
        </p>
      </div>

      <form onSubmit={handleSearch} className="flex flex-col sm:flex-row gap-4 items-end">
        <div className="flex-1 w-full">
          <Field label={t("lookup.inputLabel")}>
            <input
              placeholder={t("lookup.placeholder")}
              value={refInput}
              onChange={(e) => setRefInput(e.target.value)}
              className="w-full"
            />
          </Field>
        </div>
        <Button variant="primary" type="submit" className="w-full sm:w-auto">
          {t("lookup.button")}
        </Button>
      </form>

      {searched && errorMsg && (
        <InlineError message={errorMsg} />
      )}

      {/* Matching Booking Found Result */}
      {booking && booking.reference.toLowerCase() === refInput.trim().toLowerCase() && (
        <div className="mt-8 border-t border-stone-200 pt-6">
          <div className="ticket-card">
            <div className="ticket-header">
              <div className="flex items-center gap-2">
                <TrainFront size={18} />
                <span className="font-mono text-xs font-bold tracking-widest uppercase">
                  {t("brand.title")}
                </span>
              </div>
              <span className="font-mono text-xs font-bold tracking-widest">
                {booking.reference}
              </span>
            </div>

            <div className="ticket-body">
              <div className="ticket-main">
                <div className="ticket-info">
                  <div className="flex items-center justify-between gap-4">
                    <div>
                      <span className="text-xs font-bold text-stone-400 block uppercase">
                        Reserved Seat
                      </span>
                      <strong className="text-base font-bold text-stone-900">
                        Coach {booking.seat.coachCode} · Seat {booking.seat.label}
                      </strong>
                    </div>

                    <div className="text-right">
                      <span className="text-xs font-bold text-stone-400 block uppercase">
                        Fare Paid
                      </span>
                      <strong className="text-base font-bold text-maroon-900">
                        {formatMoney(booking.fare)}
                      </strong>
                    </div>
                  </div>

                  <div className="mt-3 flex items-center justify-between border-t border-stone-200 pt-2 text-xs">
                    <span className="font-bold text-stone-500">STATUS</span>
                    <span className="status-pill">{booking.status}</span>
                  </div>
                </div>
              </div>

              <div className="ticket-stub">
                {booking.status === "CONFIRMED" && (
                  <Button
                    variant="secondary"
                    onClick={() => cancelMutation.mutate(booking.id)}
                    disabled={cancelMutation.isPending}
                    className="w-full text-xs"
                  >
                    {cancelMutation.isPending ? t("confirmation.cancelling") : t("confirmation.cancelButton")}
                  </Button>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </section>
  );
}

export default LookupView;
