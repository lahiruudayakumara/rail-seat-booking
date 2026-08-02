import { useTranslation } from "react-i18next";
import type { TrainRun } from "@/types";
import { EmptyState, LoadingSpinner, QrCode, SectionCard, TrainFront } from "@/components";
import { useJourneySearch } from "@/hooks/use-journey-search";
import { useTrainSelection } from "@/hooks/use-train-selection";

function getStationCode(name: string): string {
  if (!name) return "STN";
  const words = name.split(" ");
  if (words.length >= 2) {
    return (words[0][0] + words[1][0] + (words[1][1] || "")).toUpperCase();
  }
  return name.slice(0, 3).toUpperCase();
}

export function TrainSelection() {
  const { t } = useTranslation();
  const { origin, destination } = useJourneySearch();
  const { runId, searched, trainRunsQuery, handleChooseRun } = useTrainSelection();

  if (!searched) return null;

  const originName = origin ? t(`stations.${origin.name}`, origin.name) : "";
  const destinationName = destination ? t(`stations.${destination.name}`, destination.name) : "";

  const originCode = origin ? getStationCode(origin.name) : "ORG";
  const destinationCode = destination ? getStationCode(destination.name) : "DST";

  return (
    <SectionCard title={t("trains.title")} icon={<TrainFront size={22} />}>
      {trainRunsQuery.isLoading ? (
        <LoadingSpinner label={t("trains.loading")} />
      ) : trainRunsQuery.data?.items.length ? (
        <div className="grid gap-5">
          {trainRunsQuery.data.items.map((run: TrainRun) => {
            const isSelected = runId === run.id;

            const departureDate = new Date(run.departureAt);
            const arrivalDate = new Date(run.arrivalAt);

            const departureTime = departureDate.toLocaleTimeString("en-LK", {
              hour: "2-digit",
              minute: "2-digit",
            });

            const arrivalTime = arrivalDate.toLocaleTimeString("en-LK", {
              hour: "2-digit",
              minute: "2-digit",
            });

            // Calculate duration in hours & minutes
            const diffMs = Math.max(0, arrivalDate.getTime() - departureDate.getTime());
            const hours = Math.floor(diffMs / (1000 * 60 * 60));
            const mins = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60));
            const durationStr = `${hours}h ${mins > 0 ? `${mins}m` : ""}`;

            const ticketNo = `TK-${run.id.slice(0, 6).toUpperCase()}`;

            return (
              <button
                key={run.id}
                className={`ticket-card ${isSelected ? "selected" : ""}`}
                onClick={() => handleChooseRun(run.id)}
              >
                {/* Solid Maroon Top Boarding Pass Banner */}
                <div className="ticket-header">
                  <div className="flex items-center gap-2">
                    <TrainFront size={18} />
                    <span className="font-mono text-xs font-bold tracking-widest uppercase">
                      BOARDING PASS
                    </span>
                  </div>
                  <span className="font-mono text-xs font-bold tracking-widest opacity-90">
                    {ticketNo}
                  </span>
                </div>

                {/* Ticket Main Content */}
                <div className="ticket-body">
                  {/* Left Barcode & Main Journey Info */}
                  <div className="ticket-main">
                    {/* Vertical Barcode Graphic */}
                    <div className="ticket-barcode-vertical hidden md:flex" aria-hidden="true">
                      <span className="bar w-1"></span>
                      <span className="bar w-2"></span>
                      <span className="bar w-0.5"></span>
                      <span className="bar w-1.5"></span>
                      <span className="bar w-1"></span>
                      <span className="bar w-3"></span>
                      <span className="bar w-0.5"></span>
                      <span className="bar w-2"></span>
                    </div>

                    <div className="ticket-info">
                      {/* Top Stations Line */}
                      <div className="flex items-center justify-between gap-4">
                        {/* Origin Station */}
                        <div className="flex flex-col">
                          <span className="font-mono text-2xl font-extrabold text-maroon-900 leading-none">
                            {originCode}
                          </span>
                          <span className="mt-1 text-xs font-bold text-stone-600">
                            {originName}
                          </span>
                        </div>

                        {/* Train Line & Duration Center Graphic */}
                        <div className="flex flex-1 flex-col items-center justify-center px-4">
                          <span className="font-mono text-[10px] font-bold text-stone-500 uppercase tracking-wider mb-1">
                            {durationStr}
                          </span>
                          <div className="relative flex w-full items-center justify-center">
                            <div className="h-[2px] w-full border-t-2 border-dashed border-stone-300"></div>
                            <span className="absolute rounded-full bg-maroon-900 p-1 text-white shadow-sm">
                              <TrainFront size={14} />
                            </span>
                          </div>
                        </div>

                        {/* Destination Station */}
                        <div className="flex flex-col text-right">
                          <span className="font-mono text-2xl font-extrabold text-maroon-900 leading-none">
                            {destinationCode}
                          </span>
                          <span className="mt-1 text-xs font-bold text-stone-600">
                            {destinationName}
                          </span>
                        </div>
                      </div>

                      {/* Bottom Ticket Metadata Grid */}
                      <div className="mt-4 grid grid-cols-3 gap-3 border-t border-stone-200/80 pt-3 text-left">
                        <div>
                          <span className="block font-mono text-[10px] font-bold uppercase tracking-wider text-stone-400">
                            DEPARTURE
                          </span>
                          <span className="font-mono text-sm font-bold text-stone-900">
                            {departureTime}
                          </span>
                        </div>

                        <div>
                          <span className="block font-mono text-[10px] font-bold uppercase tracking-wider text-stone-400">
                            ARRIVAL
                          </span>
                          <span className="font-mono text-sm font-bold text-stone-900">
                            {arrivalTime}
                          </span>
                        </div>

                        <div>
                          <span className="block font-mono text-[10px] font-bold uppercase tracking-wider text-stone-400">
                            SERVICE DATE
                          </span>
                          <span className="font-mono text-sm font-bold text-stone-900">
                            {run.serviceDate}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Perforated Separator with Top/Bottom Notch Cutouts */}
                  <div className="ticket-divider" aria-hidden="true">
                    <span className="notch notch-top"></span>
                    <span className="line"></span>
                    <span className="notch notch-bottom"></span>
                  </div>

                  {/* Right Stub Section */}
                  <div className="ticket-stub">
                    <div className="mb-2 text-maroon-900 opacity-90">
                      <QrCode size={34} />
                    </div>

                    <span className="status-pill">{run.status}</span>

                    <div className="ticket-stub-barcode hidden sm:block" aria-hidden="true">
                      ||||||||||||
                    </div>
                  </div>
                </div>
              </button>
            );
          })}
        </div>
      ) : (
        <EmptyState message={t("trains.empty")} />
      )}
    </SectionCard>
  );
}
