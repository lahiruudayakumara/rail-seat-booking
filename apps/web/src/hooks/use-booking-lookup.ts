import { useMutation } from "@tanstack/react-query";
import { getBookingByReference } from "@/api";
import { useAppDispatch } from "@/store";
import { setBooking } from "@/store/slices/booking-slice";

export function useBookingLookup() {
  const dispatch = useAppDispatch();

  return useMutation({
    mutationFn: ({ reference, contact }: { reference: string; contact: string }) =>
      getBookingByReference(reference, contact),
    onMutate: () => dispatch(setBooking(undefined)),
    onSuccess: (booking) => dispatch(setBooking(booking)),
  });
}
