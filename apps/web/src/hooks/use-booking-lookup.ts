import { useMutation } from "@tanstack/react-query";
import { getBookingByReference } from "@/api";
import { useAppDispatch } from "@/store";
import { setBooking } from "@/store/slices/booking-slice";

export function useBookingLookup() {
  const dispatch = useAppDispatch();

  return useMutation({
    mutationFn: (reference: string) => getBookingByReference(reference),
    onMutate: () => dispatch(setBooking(undefined)),
    onSuccess: (booking) => dispatch(setBooking(booking)),
  });
}
