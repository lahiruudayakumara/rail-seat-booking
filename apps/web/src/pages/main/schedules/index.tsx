import { Helmet } from "react-helmet-pro";
import { SchedulesView } from "@/sections/main/schedules/view";

export function SchedulesRoutePage() {
  return (
    <>
      <Helmet>
        <title>Timetable & Stations | Lanka Rail Reserve</title>
        <meta
          name="description"
          content="View Sri Lanka Main Line station distance map and daily train schedules."
        />
      </Helmet>
      <SchedulesView />
    </>
  );
}

export default SchedulesRoutePage;
