import { useTranslation } from "react-i18next";
import { useEffect, useState } from "react";
import type { Seat } from "@/types";
import { Button, EmptyState, InlineError, LoadingSpinner, RefreshCw, SectionCard, Ticket } from "@/components";
import { useJourneySearch } from "@/hooks/use-journey-search";
import { useSeatSelection } from "@/hooks/use-seat-selection";

export function SeatSelection() {
  const { t } = useTranslation();
  const { origin, destination } = useJourneySearch();
  const { runId, selectedSeat, seatsQuery, groupedSeats, handleChooseSeat } =
    useSeatSelection();

  const typedGroupedSeats = groupedSeats as [string, Seat[]][];
  const [activeCoach, setActiveCoach] = useState("");
  useEffect(() => {
    if (!typedGroupedSeats.some(([coach]) => coach === activeCoach)) {
      setActiveCoach(typedGroupedSeats[0]?.[0] ?? "");
    }
  }, [activeCoach, typedGroupedSeats]);

  if (!runId) return null;

  const originName = origin ? t(`stations.${origin.name}`, origin.name) : "";
  const destinationName = destination ? t(`stations.${destination.name}`, destination.name) : "";

  return (
    <div id="seat-results" className="scroll-mt-5">
    <SectionCard title={t("seats.title")} icon={<Ticket size={22} />}>
      <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
        <p className="text-sm text-stone-500">{t("seats.sub", { origin: originName, destination: destinationName })}</p>
        <div className="flex items-center gap-3 text-xs font-semibold text-stone-500" aria-label="Seat map legend">
          <span className="flex items-center gap-1.5"><i className="h-3 w-3 rounded bg-stone-100 ring-1 ring-stone-300" />{t("seats.available")}</span>
          <span className="flex items-center gap-1.5"><i className="h-3 w-3 rounded bg-[#851e2e]" />{t("seats.selected")}</span>
          <span className="flex items-center gap-1.5"><i className="h-3 w-3 rounded bg-stone-300" />{t("seats.booked")}</span>
        </div>
      </div>

      {seatsQuery.isLoading ? (
        <LoadingSpinner label={t("seats.loading")} />
      ) : seatsQuery.isError ? (
        <div className="grid justify-items-start gap-3">
          <InlineError message={t("seats.loadError")} />
          <Button variant="secondary" onClick={() => void seatsQuery.refetch()}><RefreshCw size={16} /> {t("common.tryAgain")}</Button>
        </div>
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
                const isSelected = selectedSeat?.id === item.id;
                const isBooked = item.availabilityStatus === "BOOKED";
                const isWindow = item.attributes.includes("WINDOW");

                return (
                  <button
                    key={item.id}
                    aria-label={`Seat ${item.label}, ${item.coachClass} class, ${isSelected ? t("seats.selected") : isBooked ? t("seats.booked") : item.attributes.join(" ")}`}
                    aria-pressed={isSelected}
                    className={`seat ${isSelected ? "selected" : isBooked ? "booked" : ""}`}
                    disabled={isBooked}
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
