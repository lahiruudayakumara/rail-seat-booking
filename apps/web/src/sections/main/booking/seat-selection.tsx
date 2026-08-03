import { useTranslation } from "react-i18next";
import { useEffect, useState } from "react";
import type { Seat } from "@/types";
import { Button, EmptyState, InlineError, LoadingSpinner, RefreshCw, SectionCard, Ticket } from "@/components";
import { useJourneySearch } from "@/hooks/use-journey-search";
import { useSeatSelection } from "@/hooks/use-seat-selection";
import { WaitlistForm } from "./waitlist-form";

export function SeatSelection() {
  const { t } = useTranslation();
  const { origin, destination, originId, destinationId } = useJourneySearch();
  const { runId, selectedSeats, seatsQuery, quoteMutation, groupedSeats, preferences, handleChooseSeat } =
        
    useSeatSelection();

  const typedGroupedSeats = groupedSeats as [string, Seat[]][];
  const seats = seatsQuery.data?.items ?? [];
  const fullyBooked = seats.length > 0 && seats.every((seat) => seat.availabilityStatus === "BOOKED");
  const preferredCoachClass = preferences?.preferredCoachClass;
  const [activeCoach, setActiveCoach] = useState("");
  useEffect(() => {
    if (!typedGroupedSeats.some(([coach]) => coach === activeCoach)) {
      setActiveCoach(typedGroupedSeats[0]?.[0] ?? "");
    }
  }, [activeCoach, typedGroupedSeats]);
  useEffect(() => {
    if (!preferredCoachClass || preferredCoachClass === "ANY") return;
    const preferredCoach = typedGroupedSeats.find(([, seats]) => seats[0]?.coachClass === preferredCoachClass)?.[0];
    if (preferredCoach) setActiveCoach(preferredCoach);
  }, [preferredCoachClass, typedGroupedSeats]);

  if (!runId) return null;

  const originName = origin ? t(`stations.${origin.name}`, origin.name) : "";
  const destinationName = destination ? t(`stations.${destination.name}`, destination.name) : "";

  return (
    <div id="seat-results" className="scroll-mt-5">
    <SectionCard title={t("seats.title")} icon={<Ticket size={22} />}>
      <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
        <div><p className="text-sm text-stone-500">{t("seats.sub", { origin: originName, destination: destinationName })}</p><p className="mt-1 text-xs font-semibold text-[#6b1724]">Select up to 6 seats. Choose one seat for an individual booking or 2–6 for a group.</p></div>
        <div className="flex items-center gap-3 text-xs font-semibold text-stone-500" aria-label="Seat map legend">
          <span className="flex items-center gap-1.5"><i className="h-3 w-3 rounded bg-stone-100 ring-1 ring-stone-300" />{t("seats.available")}</span>
          <span className="flex items-center gap-1.5"><i className="h-3 w-3 rounded bg-[#851e2e]" />{t("seats.selected")}</span>
          <span className="flex items-center gap-1.5"><i className="h-3 w-3 rounded bg-stone-300" />{t("seats.booked")}</span>
        </div>
      </div>
      {preferences && (preferences.preferredCoachClass !== "ANY" || preferences.preferredSeatType !== "ANY") && <p className="mb-4 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs font-semibold text-amber-900">Your saved preferences are applied. Matching coach classes open first and preferred seats are highlighted.</p>}
      {selectedSeats.length > 0 && <div className="mb-4 flex flex-wrap items-center justify-between gap-2 rounded-xl border border-[#6b1724]/20 bg-[#6b1724]/5 px-4 py-3"><strong className="text-sm text-[#6b1724]">{selectedSeats.length} of 6 seats selected</strong><span className="text-xs font-semibold text-stone-600">{selectedSeats.map((item) => `${item.seat.coachCode} · ${item.seat.label}`).join("  •  ")}</span></div>}

      {seatsQuery.isLoading ? (
        <LoadingSpinner label={t("seats.loading")} />
      ) : seatsQuery.isError ? (
        <div className="grid justify-items-start gap-3">
          <InlineError message={t("seats.loadError")} />
          <Button variant="secondary" onClick={() => void seatsQuery.refetch()}><RefreshCw size={16} /> {t("common.tryAgain")}</Button>
        </div>
      ) : fullyBooked ? (
        <WaitlistForm
          trainRunId={runId}
          originStationId={originId}
          destinationStationId={destinationId}
          initialCoachClass={preferences?.preferredCoachClass ?? "ANY"}
        />
      ) : typedGroupedSeats.length ? (
        <>
          <div className="mb-4 flex gap-2 overflow-x-auto pb-1" role="tablist" aria-label="Reserved coaches">
            {typedGroupedSeats.map(([coach, items]) => {
              const available = items.filter((item) => item.availabilityStatus !== "BOOKED").length;
              return (
                <button key={coach} type="button" role="tab" aria-selected={activeCoach === coach} onClick={() => setActiveCoach(coach)} className={`whitespace-nowrap rounded-lg border px-4 py-2 text-sm font-bold ${activeCoach === coach ? "border-[#851e2e] bg-[#851e2e] text-white" : "border-stone-300 bg-white text-stone-700 hover:border-[#851e2e]"}`}>
                  {t("seats.coach", { code: coach })} <span className="ml-1 opacity-75">· {available}</span>
                </button>
              );
            })}
          </div>
          {typedGroupedSeats.filter(([coach]) => coach === activeCoach).map(([coach, items]) => (
          <div className="coach" key={coach}>
            <div className="coach-title">
              <strong>{t("seats.coach", { code: coach })}</strong>
              <span>
                {t("seats.classAvailable", {
                  coachClass: items[0].coachClass,
                  count: items.filter((item) => item.availabilityStatus !== "BOOKED").length,
                })}
              </span>
            </div>

            <div className="seat-grid">
              {items.map((item) => {
                const isSelected = selectedSeats.some((selection) => selection.seat.id === item.id);
                const isBooked = item.availabilityStatus === "BOOKED";
                const isWindow = item.attributes.includes("WINDOW");
                const isPreferred = preferences?.preferredSeatType !== "ANY" && item.attributes.includes(preferences?.preferredSeatType ?? "");

                return (
                  <button
                    key={item.id}
                    aria-label={`Seat ${item.label}, ${item.coachClass} class, ${isSelected ? t("seats.selected") : isBooked ? t("seats.booked") : item.attributes.join(" ")}`}
                    aria-pressed={isSelected}
                    className={`seat ${isSelected ? "selected" : isBooked ? "booked" : isPreferred ? "ring-2 ring-amber-400 ring-offset-2" : ""}`}
                    disabled={(isBooked && !isSelected) || quoteMutation.isPending}
                    onClick={() => handleChooseSeat(item)}
                  >
                    <span>{item.label}</span>
                    <small>{isSelected ? "✓" : isBooked ? t("seats.booked") : isWindow ? "◫" : "·"}</small>
                  </button>
                );
              })}
            </div>
          </div>
          ))}
        </>
      ) : (
        <EmptyState message={t("seats.empty")} />
      )}
    </SectionCard>
    </div>
  );
}
