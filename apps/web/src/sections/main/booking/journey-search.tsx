import { useEffect, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { useTranslation } from "react-i18next";
import type { Route, Station } from "@/types";
import { getRoutes } from "@/api";
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
    setRouteId,
    setOriginId,
    setDestinationId,
    setDate,
    handleSearch,
  } = useJourneySearch();

  const routesQuery = useQuery({
    queryKey: ["routes"],
    queryFn: getRoutes,
  });
  const routeItems: Route[] = useMemo(
    () => routesQuery.data?.items ?? [],
    [routesQuery.data?.items],
  );

  useEffect(() => {
    if (!routeId && routeItems.length === 1) {
      setRouteId(routeItems[0].id);
    }
  }, [routeId, routeItems, setRouteId]);

  const formatStationName = (name: string) =>
    t(`stations.${name}`, name);

  const originPosition = stationItems.find((station) => station.id === originId)?.position;
  const destinationItems = stationItems.filter(
    (station) => originPosition != null && station.position != null && station.position > originPosition,
  );
  const canSearch = Boolean(routeId && originId && destinationId && date);

  const search = () => {
    handleSearch();
    window.setTimeout(() => document.getElementById("train-results")?.scrollIntoView({ behavior: "smooth", block: "start" }), 50);
  };

  return (
    <section className="panel p-6 md:p-8" aria-labelledby="journey-heading">
      <div className="mb-5 border-b border-stone-200 pb-4">
        <p className="section-kicker">{t("search.kicker")}</p>
        <h2 id="journey-heading" className="font-heading text-2xl font-bold text-stone-900">
          {t("search.title")}
        </h2>
      </div>

      <div className="grid gap-4 md:grid-cols-5">
        <Field label={t("search.routeLabel")}>
          <select
            value={routeId}
            onChange={(event) => setRouteId(event.target.value)}
            disabled={routesQuery.isLoading}
          >
            <option value="">{t("search.routePlaceholder")}</option>
            {routeItems.map((route) => (
              <option key={route.id} value={route.id}>
                {route.name}
              </option>
            ))}
          </select>
        </Field>

        <Field label={t("search.originLabel")}>
          <select
            value={originId}
            onChange={(e) => setOriginId(e.target.value)}
            disabled={!routeId || stationsQuery.isLoading}
          >
            <option value="">{t("search.originPlaceholder")}</option>
            {stationItems.slice(0, -1).map((x: Station) => (
              <option key={x.id} value={x.id}>
                {formatStationName(x.name)}
              </option>
            ))}
          </select>
        </Field>

        <Field label={t("search.destinationLabel")}>
          <select
            value={destinationId}
            onChange={(e) => setDestinationId(e.target.value)}
            disabled={!originId || stationsQuery.isLoading}
          >
            <option value="">{t("search.destinationPlaceholder")}</option>
            {destinationItems.map((x: Station) => (
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
            onClick={search}
            disabled={!canSearch || routesQuery.isLoading || stationsQuery.isLoading}
            className="w-full"
          >
            <span>{t("search.submitButton")}</span>
            <ArrowRight size={16} />
          </Button>
        </div>
      </div>

      <p className="mt-4 text-xs text-stone-500">
        {originId ? t("search.destinationHint") : t("search.originHint")}
      </p>

      {(routesQuery.isError || stationsQuery.isError) && (
        <InlineError message={t("search.loadError")} />
      )}
    </section>
  );
}
