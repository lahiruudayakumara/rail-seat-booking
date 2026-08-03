import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { QRCodeSVG } from "qrcode.react";
import { Button, Field, InlineError, TrainFront } from "@/components";
import { useBookingFlow } from "@/hooks/use-booking-flow";
import { useBookingLookup } from "@/hooks/use-booking-lookup";
import { formatMoney } from "@/utils";
import { usePassengerAuth } from "@/auth/use-passenger-auth";

const LookupView = () => {
  const { t } = useTranslation();
  const { booking, cancelMutation } = useBookingFlow();
  const lookupMutation = useBookingLookup();
  const { account } = usePassengerAuth();
  const [refInput, setRefInput] = useState("");
  const [contactInput, setContactInput] = useState("");
  const [contactEdited, setContactEdited] = useState(false);
  const [searched, setSearched] = useState(false);
  const [submittedReference, setSubmittedReference] = useState("");
  const [formError, setFormError] = useState("");

  useEffect(() => {
    if (!contactEdited && !contactInput && account?.email) setContactInput(account.email);
  }, [account?.email, contactEdited, contactInput]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    const reference = refInput.trim().toLocaleUpperCase();
    const contact = contactInput.trim();
    if (!reference || !contact) {
      setSearched(false);
      setFormError("Enter both the booking reference and booking email or phone.");
      return;
    }
    setFormError("");
    setSubmittedReference(reference);
    setSearched(true);
    lookupMutation.mutate({ reference, contact });
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

      <form onSubmit={handleSearch} className="grid gap-4 sm:grid-cols-[1fr_1fr_auto] items-end">
        <div className="flex-1 w-full">
          <Field label={t("lookup.inputLabel")}>
            <input
              placeholder={t("lookup.placeholder")}
              value={refInput}
              autoCapitalize="characters"
              autoComplete="off"
              onChange={(e) => {
                setRefInput(e.target.value);
                setSearched(false);
                setFormError("");
              }}
              className="w-full"
            />
          </Field>
        </div>
        <div className="w-full">
          <Field label={t("lookup.contactLabel")}>
            <input
              placeholder={t("lookup.contactPlaceholder")}
              value={contactInput}
              autoCapitalize="none"
              autoComplete="email"
              onChange={(e) => {
                setContactInput(e.target.value);
                setContactEdited(true);
                setSearched(false);
                setFormError("");
              }}
              className="w-full"
            />
          </Field>
        </div>
        <Button
          variant="primary"
          type="submit"
          disabled={lookupMutation.isPending}
          className="w-full sm:w-auto"
        >
          {lookupMutation.isPending ? t("lookup.searching") : t("lookup.button")}
        </Button>
      </form>

      {formError && <InlineError message={formError} />}

      <p className="mt-3 text-xs leading-5 text-stone-500">
        Enter the email or phone used for this booking. Sri Lankan phone numbers can use <span className="font-semibold text-stone-700">077…</span> or <span className="font-semibold text-stone-700">+9477…</span> format.
      </p>

      {searched && lookupMutation.isError && (
        <InlineError message={t("lookup.notFound")} />
      )}

      {/* Matching Booking Found Result */}
      {searched && !lookupMutation.isPending && booking && booking.reference.toLocaleUpperCase() === submittedReference && (
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

              <div className="ticket-stub flex flex-col items-center justify-between gap-3">
                <div className="rounded bg-white p-1.5 shadow-sm border border-stone-200">
                  <QRCodeSVG
                    value={booking.reference}
                    size={64}
                    bgColor="#ffffff"
                    fgColor="#6b1724"
                    level="M"
                  />
                </div>
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
