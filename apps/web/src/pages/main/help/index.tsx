import { Helmet } from "react-helmet-pro";
import { HelpView } from "@/sections/main/help/view";

export default function HelpPage() {
  return (
    <>
      <Helmet>
        <title>Help & FAQ | Lanka Rail Reserve</title>
        <meta
          name="description"
          content="Learn about segment seat booking, fare holds, and cancellation policy."
        />
      </Helmet>
      <HelpView />
    </>
  );
}