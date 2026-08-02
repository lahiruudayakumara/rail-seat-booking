import { Helmet } from "react-helmet-pro";
import { BookingView } from "@/sections/main/booking/view";

export function HomePage() {
  return (
    <>
      <Helmet>
        <title>Book Train Seats | Lanka Rail Reserve</title>
        <meta
          name="description"
          content="Reserve Sri Lanka Hill Country railway journeys by segment. Fast, easy seat booking."
        />
      </Helmet>
      <BookingView />
    </>
  );
}

export default HomePage;
