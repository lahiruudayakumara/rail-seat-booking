import { JourneySearch } from "@/sections/main/booking/journey-search";
import { TrainSelection } from "@/sections/main/booking/train-selection";
import { SeatSelection } from "@/sections/main/booking/seat-selection";
import { PassengerFare } from "@/sections/main/booking/passenger-fare";
import { BookingConfirmation } from "@/sections/main/booking/booking-confirmation";


const BookingView = () => {
  return (
    <div className="flex flex-col gap-6">
      <JourneySearch />
      <TrainSelection />
      <SeatSelection />
      <PassengerFare />
      <BookingConfirmation />
    </div>
  );
}

export default BookingView;
