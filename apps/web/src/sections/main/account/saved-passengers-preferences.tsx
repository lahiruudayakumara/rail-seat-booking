import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import { useEffect, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { createSavedTraveller, deleteSavedTraveller, getPassengerPreferences, getSavedTravellers, updatePassengerPreferences, updateSavedTraveller } from "@/api";
import { Button, ConfirmationModal, LoadingSpinner, Users } from "@/components";
import type { ApiError, PassengerPreferences, SavedTraveller, SavedTravellerInput } from "@/types";

const defaultPreferences: PassengerPreferences = {
  preferredCoachClass: "ANY",
  preferredSeatType: "ANY",
  language: "en",
};

function requestError(error: unknown) {
  if (axios.isAxiosError<ApiError>(error)) return error.response?.data?.message ?? "Request failed.";
  return error instanceof Error ? error.message : "Request failed.";
}

export function SavedPassengersPreferences() {
  const queryClient = useQueryClient();
  const { i18n } = useTranslation();
  const [editing, setEditing] = useState<SavedTraveller>();
  const [deleteTarget, setDeleteTarget] = useState<SavedTraveller>();
  const [formError, setFormError] = useState("");
  const [preferences, setPreferences] = useState(defaultPreferences);

  const travellersQuery = useQuery({ queryKey: ["saved-travellers"], queryFn: getSavedTravellers });
  const preferencesQuery = useQuery({ queryKey: ["passenger-preferences"], queryFn: getPassengerPreferences });

  useEffect(() => {
    if (preferencesQuery.data) setPreferences(preferencesQuery.data);
  }, [preferencesQuery.data]);

  const travellerMutation = useMutation({
    mutationFn: ({ id, input }: { id?: string; input: SavedTravellerInput }) => id ? updateSavedTraveller(id, input) : createSavedTraveller(input),
    onSuccess: () => {
      setEditing(undefined);
      setFormError("");
      void queryClient.invalidateQueries({ queryKey: ["saved-travellers"] });
    },
    onError: (error) => setFormError(requestError(error)),
  });
  const deleteMutation = useMutation({
    mutationFn: deleteSavedTraveller,
    onSuccess: () => {
      setDeleteTarget(undefined);
      void queryClient.invalidateQueries({ queryKey: ["saved-travellers"] });
    },
  });
  const preferencesMutation = useMutation({
    mutationFn: updatePassengerPreferences,
    onSuccess: (saved) => {
      queryClient.setQueryData(["passenger-preferences"], saved);
      void i18n.changeLanguage(saved.language);
    },
  });

  const saveTraveller = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setFormError("");
    const data = new FormData(event.currentTarget);
    const input = {
      fullName: String(data.get("fullName") ?? "").trim(),
      email: String(data.get("email") ?? "").trim(),
      phone: String(data.get("phone") ?? "").trim(),
    };
    if (!input.email && !input.phone) {
      setFormError("Enter an email address or phone number.");
      return;
    }
    travellerMutation.mutate({ id: editing?.id, input });
  };

  return (
    <section className="rounded-2xl border border-stone-200 bg-white p-5 shadow-sm sm:p-6 md:p-8">
      <div className="flex items-start gap-3">
        <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-[#6b1724]/10 text-[#6b1724]"><Users size={20} /></span>
        <div><p className="section-kicker">QUICK BOOKING</p><h3 className="font-heading text-2xl font-extrabold text-stone-900">Saved passengers & preferences</h3><p className="mt-1 text-sm leading-6 text-stone-500">Save trusted traveller details and show your preferred seats first on future journeys.</p></div>
      </div>

      <div className="mt-6 grid gap-6 xl:grid-cols-[1.15fr_0.85fr]">
        <div className="rounded-xl border border-stone-200 bg-stone-50 p-4 sm:p-5">
          <div className="flex flex-wrap items-center justify-between gap-2"><div><h4 className="font-heading text-lg font-extrabold text-stone-900">Traveller profiles</h4><p className="mt-1 text-xs text-stone-500">Choose one during checkout to fill passenger details instantly.</p></div><span className="rounded-full bg-white px-2.5 py-1 text-xs font-bold text-stone-500 ring-1 ring-stone-200">{travellersQuery.data?.length ?? 0}/20</span></div>

          {travellersQuery.isLoading ? <div className="py-5"><LoadingSpinner label="Loading saved passengers" /></div> : (
            <div className="mt-4 grid gap-2">
              {travellersQuery.data?.map((traveller) => (
                <div key={traveller.id} className="flex flex-col gap-3 rounded-xl border border-stone-200 bg-white p-3 sm:flex-row sm:items-center sm:justify-between">
                  <div className="min-w-0"><strong className="block truncate text-sm text-stone-900">{traveller.fullName}</strong><span className="mt-0.5 block truncate text-xs text-stone-500">{traveller.email || traveller.phone}{traveller.email && traveller.phone ? ` · ${traveller.phone}` : ""}</span></div>
                  <div className="flex shrink-0 gap-2"><button type="button" className="rounded-lg px-2.5 py-1.5 text-xs font-bold text-[#6b1724] hover:bg-[#6b1724]/5" onClick={() => { setEditing(traveller); setFormError(""); }}>Edit</button><button type="button" className="rounded-lg px-2.5 py-1.5 text-xs font-bold text-red-700 hover:bg-red-50" onClick={() => { deleteMutation.reset(); setDeleteTarget(traveller); }}>Remove</button></div>
                </div>
              ))}
              {!travellersQuery.isLoading && !travellersQuery.data?.length && <p className="rounded-xl border border-dashed border-stone-300 bg-white p-4 text-center text-sm text-stone-500">No saved travellers yet.</p>}
            </div>
          )}

          <form key={editing?.id ?? "new"} className="mt-4 grid gap-3 border-t border-stone-200 pt-4" onSubmit={saveTraveller}>
            <div className="flex items-center justify-between"><h5 className="text-sm font-extrabold text-stone-800">{editing ? "Edit traveller" : "Add a traveller"}</h5>{editing && <button type="button" className="text-xs font-bold text-stone-500 hover:text-stone-800" onClick={() => { setEditing(undefined); setFormError(""); }}>Cancel edit</button>}</div>
            <label className="grid gap-1.5 text-xs font-bold text-stone-700"><span>Full name</span><input className="form-input" name="fullName" autoComplete="off" defaultValue={editing?.fullName} minLength={2} maxLength={120} required /></label>
            <div className="grid gap-3 sm:grid-cols-2">
              <label className="grid min-w-0 gap-1.5 text-xs font-bold text-stone-700"><span>Email</span><input className="form-input" type="email" name="email" autoComplete="off" defaultValue={editing?.email} /></label>
              <label className="grid min-w-0 gap-1.5 text-xs font-bold text-stone-700"><span>Phone</span><input className="form-input" type="tel" name="phone" autoComplete="off" placeholder="+94770000000" pattern="\+[1-9][0-9]{7,14}" defaultValue={editing?.phone} /></label>
            </div>
            <p className="text-[11px] leading-5 text-stone-500">At least one email address or phone number is required.</p>
            {(formError || travellerMutation.isError) && <p className="rounded-lg bg-red-50 px-3 py-2 text-xs font-semibold text-red-700" role="alert">{formError || requestError(travellerMutation.error)}</p>}
            <Button className="w-full" disabled={travellerMutation.isPending || (!editing && (travellersQuery.data?.length ?? 0) >= 20)}>{travellerMutation.isPending ? "Saving…" : editing ? "Update traveller" : "Save traveller"}</Button>
          </form>
        </div>

        <form className="h-fit rounded-xl border border-stone-200 p-4 sm:p-5" onSubmit={(event) => { event.preventDefault(); preferencesMutation.mutate(preferences); }}>
          <h4 className="font-heading text-lg font-extrabold text-stone-900">Booking preferences</h4><p className="mt-1 text-xs leading-5 text-stone-500">Preferences reorder available choices; they never reserve or hide other seats.</p>
          <div className="mt-4 grid gap-4">
            <label className="field"><span>Preferred class</span><select value={preferences.preferredCoachClass} onChange={(event) => setPreferences((current) => ({ ...current, preferredCoachClass: event.target.value as PassengerPreferences["preferredCoachClass"] }))}><option value="ANY">No preference</option><option value="FIRST">First class</option><option value="SECOND">Second class</option></select></label>
            <label className="field"><span>Preferred seat</span><select value={preferences.preferredSeatType} onChange={(event) => setPreferences((current) => ({ ...current, preferredSeatType: event.target.value as PassengerPreferences["preferredSeatType"] }))}><option value="ANY">No preference</option><option value="WINDOW">Window</option><option value="AISLE">Aisle</option></select></label>
            <label className="field"><span>Preferred language</span><select value={preferences.language} onChange={(event) => setPreferences((current) => ({ ...current, language: event.target.value as PassengerPreferences["language"] }))}><option value="en">English</option><option value="si">සිංහල</option><option value="ta">தமிழ்</option></select></label>
            {preferencesQuery.isError && <p className="text-xs font-semibold text-red-700">{requestError(preferencesQuery.error)}</p>}
            {preferencesMutation.isError && <p className="text-xs font-semibold text-red-700">{requestError(preferencesMutation.error)}</p>}
            {preferencesMutation.isSuccess && <p className="text-xs font-semibold text-emerald-700" role="status">Preferences saved and applied.</p>}
            <Button className="w-full" disabled={preferencesQuery.isLoading || preferencesMutation.isPending}>{preferencesMutation.isPending ? "Saving…" : "Save preferences"}</Button>
          </div>
        </form>
      </div>

      <ConfirmationModal open={Boolean(deleteTarget)} title="Remove saved traveller?" description={`${deleteTarget?.fullName ?? "This traveller"} will be removed from your saved profiles. Existing bookings are not affected.`} confirmLabel="Remove traveller" cancelLabel="Keep traveller" tone="danger" isPending={deleteMutation.isPending} error={deleteMutation.isError ? requestError(deleteMutation.error) : undefined} onClose={() => setDeleteTarget(undefined)} onConfirm={() => { if (deleteTarget) deleteMutation.mutate(deleteTarget.id); }} />
    </section>
  );
}
