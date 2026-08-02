import { JourneySearch } from "@/sections/main/booking/journey-search";
import { TrainSelection } from "@/sections/main/booking/train-selection";
import { SeatSelection } from "@/sections/main/booking/seat-selection";
import { PassengerFare } from "@/sections/main/booking/passenger-fare";
import { BookingConfirmation } from "@/sections/main/booking/booking-confirmation";
import { BookingProgress } from "@/sections/main/booking/booking-progress";
import { useAppSelector } from "@/store";


const BookingView = () => {
  const booking = useAppSelector((state) => state.booking.booking);
  return (
    <div className="flex flex-col gap-6">
      <BookingProgress />
      {booking ? (
        <BookingConfirmation />
      ) : (
        <>
          <JourneySearch />
          <TrainSelection />
          <SeatSelection />
          <PassengerFare />
        </>
      )}
    </div>
  );
}

export default BookingView;
