import { useTranslation } from "react-i18next";
import type { TrainRun } from "@/types";
import { Button, Calendar, Check, Clock, EmptyState, InlineError, LoadingSpinner, RefreshCw, SectionCard, TrainFront } from "@/components";
import { useJourneySearch } from "@/hooks/use-journey-search";
import { useTrainSelection } from "@/hooks/use-train-selection";

function getStationCode(name: string): string {
  if (!name) return "STN";
  const words = name.split(" ");
  if (words.length >= 2) return (words[0][0] + words[1][0] + (words[1][1] || "")).toUpperCase();
  return name.slice(0, 3).toUpperCase();
}

function formatDuration(departure: Date, arrival: Date) {
  const minutes = Math.max(0, Math.round((arrival.getTime() - departure.getTime()) / 60_000));
  const hours = Math.floor(minutes / 60);
  const remainder = minutes % 60;
  return `${hours} hr${remainder ? ` ${remainder} min` : ""}`;
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
    <div id="train-results" className="scroll-mt-5">
      <SectionCard title={t("trains.title")} icon={<TrainFront size={22} />}>
        {trainRunsQuery.isLoading ? (
          <LoadingSpinner label={t("trains.loading")} />
        ) : trainRunsQuery.isError ? (
          <div className="grid justify-items-start gap-3">
            <InlineError message={t("trains.loadError")} />
            <Button variant="secondary" onClick={() => void trainRunsQuery.refetch()}><RefreshCw size={16} /> {t("common.tryAgain")}</Button>
          </div>
        ) : trainRunsQuery.data?.items.length ? (
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            {trainRunsQuery.data.items.map((run: TrainRun) => {
              const isSelected = runId === run.id;
              const departure = new Date(run.departureAt);
              const arrival = new Date(run.arrivalAt);
              const timeFormat: Intl.DateTimeFormatOptions = { hour: "2-digit", minute: "2-digit" };
              const departureTime = departure.toLocaleTimeString("en-LK", timeFormat);
              const arrivalTime = arrival.toLocaleTimeString("en-LK", timeFormat);
              const duration = formatDuration(departure, arrival);
              const serviceNumber = run.id.slice(0, 6).toUpperCase();
              const serviceDate = new Date(`${run.serviceDate}T00:00:00`).toLocaleDateString("en-LK", {
                weekday: "short",
                day: "numeric",
                month: "short",
                year: "numeric",
              });

              return (
                <button
                  type="button"
                  key={run.id}
                  className={`group w-full overflow-hidden rounded-2xl border bg-white text-left transition duration-200 focus-visible:outline-none focus-visible:ring-4 focus-visible:ring-[#6b1724]/20 ${isSelected ? "border-[#6b1724] shadow-[0_12px_30px_rgba(107,23,36,0.15)] ring-2 ring-[#6b1724]" : "border-stone-200 shadow-[0_6px_22px_rgba(54,35,28,0.07)] hover:-translate-y-0.5 hover:border-[#6b1724]/60 hover:shadow-[0_12px_30px_rgba(107,23,36,0.12)]"}`}
                  aria-pressed={isSelected}
                  onClick={() => {
                    handleChooseRun(run.id);
                    window.setTimeout(() => document.getElementById("seat-results")?.scrollIntoView({ behavior: "smooth", block: "start" }), 50);
                  }}
                >
                  <div className={`flex items-center justify-between gap-2 border-b px-3 py-2 ${isSelected ? "border-[#6b1724]/15 bg-[#fff8f6]" : "border-stone-100 bg-[#faf9f7]"}`}>
                    <div className="flex min-w-0 items-center gap-2">
                      <span className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-lg ${isSelected ? "bg-[#6b1724] text-white" : "bg-[#6b1724]/10 text-[#6b1724]"}`}><TrainFront size={15} /></span>
                      <div className="min-w-0"><span className="block truncate text-[7px] font-bold uppercase leading-none tracking-[0.08em] text-stone-400">Reserved service</span><strong className="mt-0.5 block truncate font-mono text-xs font-extrabold leading-none text-stone-900">Service {serviceNumber}</strong></div>
                    </div>
                    <span className="inline-flex shrink-0 items-center gap-1 rounded-full bg-emerald-100 px-1.5 py-0.5 text-[7px] font-bold uppercase leading-none tracking-normal text-emerald-800"><span className="h-1 w-1 rounded-full bg-emerald-500" />{run.status}</span>
                  </div>

                  <div className="px-3 py-3">
                    <div className="grid grid-cols-[minmax(0,1fr)_64px_minmax(0,1fr)] items-center gap-1.5">
                      <ServiceStop time={departureTime} code={originCode} name={originName} />
                      <div className="flex flex-col items-center px-1">
                        <span className="inline-flex items-center gap-1 rounded-full bg-stone-100 px-2 py-1 text-[9px] font-extrabold text-stone-600"><Clock size={10} />{duration}</span>
                        <div className="mt-1.5 flex w-full items-center" aria-hidden="true"><span className="h-1.5 w-1.5 rounded-full border-2 border-[#6b1724] bg-white" /><span className="h-px flex-1 bg-stone-300" /><span className="flex h-5 w-5 items-center justify-center rounded-full bg-[#6b1724] text-white shadow-sm"><TrainFront size={10} /></span><span className="h-px flex-1 bg-stone-300" /><span className="h-1.5 w-1.5 rounded-full border-2 border-[#6b1724] bg-white" /></div>
                      </div>
                      <ServiceStop time={arrivalTime} code={destinationCode} name={destinationName} align="right" />
                    </div>

                    <div className="mt-3 flex items-center justify-between gap-2 border-t border-stone-100 pt-2.5">
                      <div className="flex min-w-0 items-center gap-1.5 text-[10px] font-bold text-stone-600"><Calendar size={13} className="shrink-0 text-[#6b1724]" /><span className="truncate">{serviceDate}</span></div>
                      <span className={`inline-flex shrink-0 items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-[10px] font-extrabold transition ${isSelected ? "bg-[#6b1724] text-white" : "bg-[#6b1724]/8 text-[#6b1724] group-hover:bg-[#6b1724] group-hover:text-white"}`}>{isSelected ? <><Check size={13} />Selected</> : "Choose"}</span>
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
    </div>
  );
}

function ServiceStop({ time, code, name, align = "left" }: { time: string; code: string; name: string; align?: "left" | "right" }) {
  return <div className={align === "right" ? "min-w-0 text-right" : "min-w-0 text-left"}><span className="font-mono text-base font-black tracking-tight text-stone-900 lg:text-lg">{time}</span><div className="mt-1"><span className="block font-mono text-xs font-black text-[#6b1724]">{code}</span><span className="block truncate text-[10px] font-bold text-stone-500">{name}</span></div></div>;
}
