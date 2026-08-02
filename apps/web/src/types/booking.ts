import { z } from "zod";

export const passengerSchema = z.object({
  fullName: z.string().min(2, "Full name is required"),
  email: z.string().email("Valid email is required"),
  phone: z.string().min(8, "Phone number is required"),
});

export type PassengerFormValues = z.infer<typeof passengerSchema>;
