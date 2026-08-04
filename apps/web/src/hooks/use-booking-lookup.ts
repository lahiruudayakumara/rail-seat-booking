import { useMutation } from "@tanstack/react-query";
import { getBookingByReference } from "@/api";
import { useAppDispatch } from "@/store";
import { setBooking } from "@/store/slices/booking-slice";
import { setBookingGroup } from "@/store/slices/booking-slice";

export function useBookingLookup() {
  const dispatch = useAppDispatch();

  return useMutation({
    mutationFn: ({ reference, contact }: { reference: string; contact: string }) =>
      getBookingByReference(reference, contact),
    onMutate: () => {
      dispatch(setBooking(undefined));
      dispatch(setBookingGroup(undefined));
    },
    onSuccess: (result) => {
      if ("members" in result) dispatch(setBookingGroup(result));
      else dispatch(setBooking(result));
    },
  });
}
