import axios from "axios";
import { useMutation } from "@tanstack/react-query";
import { useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { usePassengerAuth } from "@/auth/use-passenger-auth";
import { Button, Check, Field, InlineError } from "@/components";
import { cancelWaitlistEntry, createWaitlistEntry } from "@/api";
import type { ApiError, WaitlistEntry } from "@/types";

type Props = {
  trainRunId: string;
  originStationId: string;
  destinationStationId: string;
  initialCoachClass?: "ANY" | "FIRST" | "SECOND";
};

function errorMessage(error: unknown, fallback: string) {
  if (axios.isAxiosError<ApiError>(error)) return error.response?.data?.message ?? fallback;
  return fallback;
}

export function WaitlistForm({ trainRunId, originStationId, destinationStationId, initialCoachClass = "ANY" }: Props) {
  const { t } = useTranslation();
  const { account } = usePassengerAuth();
  const [fullName, setFullName] = useState(account?.fullName ?? "");
  const [email, setEmail] = useState(account?.email ?? "");
  const [phone, setPhone] = useState(account?.phone ?? "");
  const [coachClass, setCoachClass] = useState(initialCoachClass);
  const [entry, setEntry] = useState<WaitlistEntry>();
  const [validationError, setValidationError] = useState("");

  const createMutation = useMutation({
    mutationFn: (request: Parameters<typeof createWaitlistEntry>[0]) => createWaitlistEntry(request),
    onSuccess: setEntry,
  });
  const cancelMutation = useMutation({
    mutationFn: ({ id, token }: { id: string; token: string }) => cancelWaitlistEntry(id, token),
    onSuccess: setEntry,
  });

  const submit = (event: FormEvent) => {
    event.preventDefault();
    setValidationError("");
    if (fullName.trim().length < 2) {
      setValidationError(t("waitlist.nameError"));
      return;
    }
    if (!email.trim() && !phone.trim()) {
      setValidationError(t("waitlist.contactError"));
      return;
    }
    createMutation.mutate({
      trainRunId,
      originStationId,
      destinationStationId,
      fullName: fullName.trim(),
      email: email.trim(),
      phone: phone.trim(),
      preferredCoachClass: coachClass,
    });
  };

  if (entry) {
    return (
      <div className="rounded-2xl border border-emerald-200 bg-emerald-50 p-5" aria-live="polite">
        <div className="flex items-start gap-3">
          <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-emerald-700 text-white"><Check size={18} /></span>
          <div>
            <p className="text-xs font-extrabold uppercase tracking-widest text-emerald-800">{t("waitlist.joinedKicker")}</p>
            <h3 className="mt-1 font-heading text-xl font-extrabold text-stone-900">{t("waitlist.joinedTitle")}</h3>
            <p className="mt-2 text-sm leading-6 text-stone-600">{t("waitlist.joinedBody")}</p>
          </div>
        </div>
        <div className="mt-4 rounded-xl border border-emerald-200 bg-white px-4 py-3">
          <span className="block text-xs font-bold uppercase tracking-wide text-stone-500">{t("waitlist.reference")}</span>
          <strong className="mt-1 block font-mono text-lg text-[#6b1724]">{entry.reference}</strong>
        </div>
        {entry.status === "WAITING" && entry.managementToken && (
          <Button
            type="button"
            variant="secondary"
            className="mt-4"
            disabled={cancelMutation.isPending}
            onClick={() => cancelMutation.mutate({ id: entry.id, token: entry.managementToken! })}
          >
            {cancelMutation.isPending ? t("waitlist.cancelling") : t("waitlist.cancel")}
          </Button>
        )}
        {entry.status === "CANCELLED" && <p className="mt-4 text-sm font-bold text-stone-700">{t("waitlist.cancelled")}</p>}
        {cancelMutation.isError && <div className="mt-4"><InlineError message={errorMessage(cancelMutation.error, t("waitlist.cancelError"))} /></div>}
      </div>
    );
  }

  return (
    <div className="rounded-2xl border border-amber-200 bg-amber-50 p-5 md:p-6">
      <p className="text-xs font-extrabold uppercase tracking-widest text-amber-900">{t("waitlist.kicker")}</p>
      <h3 className="mt-1 font-heading text-2xl font-extrabold text-stone-900">{t("waitlist.title")}</h3>
      <p className="mt-2 max-w-2xl text-sm leading-6 text-stone-600">{t("waitlist.body")}</p>
      <form onSubmit={submit} className="mt-5 grid gap-4 md:grid-cols-2">
        <Field label={t("waitlist.fullName")}>
          <input value={fullName} autoComplete="name" onChange={(event) => setFullName(event.target.value)} />
        </Field>
        <Field label={t("waitlist.coachClass")}>
          <select value={coachClass} onChange={(event) => setCoachClass(event.target.value as typeof coachClass)}>
            <option value="ANY">{t("waitlist.anyClass")}</option>
            <option value="FIRST">{t("waitlist.firstClass")}</option>
            <option value="SECOND">{t("waitlist.secondClass")}</option>
          </select>
        </Field>
        <Field label={t("waitlist.email")}>
          <input type="email" value={email} autoComplete="email" onChange={(event) => setEmail(event.target.value)} />
        </Field>
        <Field label={t("waitlist.phone")}>
          <input value={phone} inputMode="tel" autoComplete="tel" onChange={(event) => setPhone(event.target.value)} />
        </Field>
        <div className="md:col-span-2">
          <p className="mb-3 text-xs leading-5 text-stone-600">{t("waitlist.contactHint")}</p>
          {validationError && <div className="mb-3"><InlineError message={validationError} /></div>}
          {createMutation.isError && <div className="mb-3"><InlineError message={errorMessage(createMutation.error, t("waitlist.joinError"))} /></div>}
          <Button type="submit" disabled={createMutation.isPending}>
            {createMutation.isPending ? t("waitlist.joining") : t("waitlist.join")}
          </Button>
        </div>
      </form>
    </div>
  );
}
