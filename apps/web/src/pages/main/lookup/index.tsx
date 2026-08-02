import { Helmet } from "react-helmet-pro";
import { LookupView } from "@/sections/main/lookup/view";

export function LookupRoutePage() {
  return (
    <>
      <Helmet>
        <title>Lookup Booking | Lanka Rail Reserve</title>
        <meta
          name="description"
          content="Search and manage your active railway seat reservation."
        />
      </Helmet>
      <LookupView />
    </>
  );
}

export default LookupRoutePage;
