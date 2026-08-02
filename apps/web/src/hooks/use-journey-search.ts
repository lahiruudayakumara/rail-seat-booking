import { useQuery } from "@tanstack/react-query";
import type { Station } from "@/types";
import { getRouteStations } from "../api";
import { useAppDispatch, useAppSelector } from "../store";
import {
  setDate,
  setDestinationId,
  setOriginId,
  setRouteId,
  setSearched,
} from "../store/slices/search-slice";
import { setNotice } from "../store/slices/ui-slice";
import { resetBookingSelection } from "../store/slices/booking-slice";

export function useJourneySearch() {
  const dispatch = useAppDispatch();
  const { routeId, originId, destinationId, date, searched } = useAppSelector(
    (state) => state.search,
  );

  const stationsQuery = useQuery({
    queryKey: ["stations", routeId],
    queryFn: () => getRouteStations(routeId),
    enabled: Boolean(routeId),
  });

  const stationItems: Station[] = stationsQuery.data?.items ?? [];

  const origin = stationItems.find((x: Station) => x.id === originId);
  const destination = stationItems.find((x: Station) => x.id === destinationId);

  const handleSearch = () => {
    if (!origin || !destination) return;

    if (
      origin.position != null &&
      destination.position != null &&
      origin.position >= destination.position
    ) {
      dispatch(
        setNotice(
          "Choose a destination that comes after the origin on this route.",
        ),
      );
      dispatch(setSearched(false));
      return;
    }

    dispatch(setNotice(""));
    dispatch(setSearched(true));
  };

  return {
    routeId,
    originId,
    destinationId,
    date,
    searched,
    stationItems,
    stationsQuery,
    origin,
    destination,
    setRouteId: (id: string) => {
      dispatch(setRouteId(id));
      dispatch(resetBookingSelection());
    },
    setOriginId: (id: string) => {
      dispatch(setOriginId(id));
      dispatch(resetBookingSelection());
    },
    setDestinationId: (id: string) => {
      dispatch(setDestinationId(id));
      dispatch(resetBookingSelection());
    },
    setDate: (val: string) => {
      dispatch(setDate(val));
      dispatch(resetBookingSelection());
    },
    handleSearch,
  };
}
