import { useTranslation } from "react-i18next";
import type { Seat } from "@/types";
import { EmptyState, LoadingSpinner, SectionCard, Ticket } from "@/components";
import { useJourneySearch } from "@/hooks/use-journey-search";
import { useSeatSelection } from "@/hooks/use-seat-selection";

export function SeatSelection() {
  const { t } = useTranslation();
  const { origin, destination } = useJourneySearch();
  const { runId, selectedSeat, seatsQuery, groupedSeats, handleChooseSeat } =
    useSeatSelection();

  if (!runId) return null;

  const originName = origin ? t(`stations.${origin.name}`, origin.name) : "";
  const destinationName = destination ? t(`stations.${destination.name}`, destination.name) : "";

  const typedGroupedSeats = groupedSeats as [string, Seat[]][];

  return (
    <SectionCard title={t("seats.title")} icon={<Ticket size={22} />}>
      <p className="mb-4 text-sm text-stone-500">
        {t("seats.sub", { origin: originName, destination: destinationName })}
      </p>

      {seatsQuery.isFetching ? (
        <LoadingSpinner label={t("seats.loading")} />
      ) : typedGroupedSeats.length ? (
        typedGroupedSeats.map(([coach, items]) => (
          <div className="coach" key={coach}>
            <div className="coach-title">
              <strong>{t("seats.coach", { code: coach })}</strong>
              <span>
                {t("seats.classAvailable", {
                  coachClass: items[0].coachClass,
                  count: items.length,
                })}
              </span>
            </div>

            <div className="seat-grid">
              {items.map((item) => {
                const isSelected = selectedSeat?.id === item.id;
                const isWindow = item.attributes.includes("WINDOW");

                return (
                  <button
                    key={item.id}
                    aria-label={`Seat ${item.label}, ${item.coachClass} class, ${item.attributes.join(" ")}`}
                    aria-pressed={isSelected}
                    className={`seat ${isSelected ? "selected" : ""}`}
                    onClick={() => handleChooseSeat(item)}
                  >
                    <span>{item.label}</span>
                    <small>{isWindow ? "◫" : "·"}</small>
                  </button>
                );
              })}
            </div>
          </div>
        ))
      ) : (
        <EmptyState message={t("seats.empty")} />
      )}
    </SectionCard>
  );
}
