import { useTranslation } from "react-i18next";
import type { Station } from "@/types";
import { ArrowRight, Button, Field, InlineError } from "@/components";
import { useJourneySearch } from "@/hooks/use-journey-search";
import { getTomorrowDate } from "@/store/slices/search-slice";

export function JourneySearch() {
  const { t } = useTranslation();
  const {
    routeId,
    originId,
    destinationId,
    date,
    stationItems,
    stationsQuery,
    setOriginId,
    setDestinationId,
    setDate,
    handleSearch,
  } = useJourneySearch();

  const formatStationName = (name: string) =>
    t(`stations.${name}`, name);

  return (
    <section className="panel p-6 md:p-8" aria-labelledby="journey-heading">
      <div className="mb-5 border-b border-stone-200 pb-4">
        <p className="section-kicker">{t("search.kicker")}</p>
        <h2 id="journey-heading" className="font-heading text-2xl font-bold text-stone-900">
          {t("search.title")}
        </h2>
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        <Field label={t("search.originLabel")}>
          <select value={originId} onChange={(e) => setOriginId(e.target.value)}>
            <option value="">{t("search.originPlaceholder")}</option>
            {stationItems.slice(0, -1).map((x: Station) => (
              <option key={x.id} value={x.id}>
                {formatStationName(x.name)}
              </option>
            ))}
          </select>
        </Field>

        <Field label={t("search.destinationLabel")}>
          <select value={destinationId} onChange={(e) => setDestinationId(e.target.value)}>
            <option value="">{t("search.destinationPlaceholder")}</option>
            {stationItems.slice(1).map((x: Station) => (
              <option key={x.id} value={x.id}>
                {formatStationName(x.name)}
              </option>
            ))}
          </select>
        </Field>

        <Field label={t("search.dateLabel")}>
          <input
            type="date"
            min={getTomorrowDate()}
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
        </Field>

        <div className="flex flex-col justify-end">
          <Button
            variant="primary"
            onClick={handleSearch}
            disabled={!routeId || stationsQuery.isLoading}
            className="w-full"
          >
            <span>{t("search.submitButton")}</span>
            <ArrowRight size={16} />
          </Button>
        </div>
      </div>

      {stationsQuery.isError && (
        <InlineError message={t("search.loadError")} />
      )}
    </section>
  );
}
