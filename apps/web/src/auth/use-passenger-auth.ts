import { createContext, useContext } from "react";
import type { PassengerAccount, RegisterPassengerRequest } from "@/types";

export type PassengerAuthContextValue = {
  account: PassengerAccount | null;
  isLoading: boolean;
  login: (values: { email: string; password: string }) => Promise<PassengerAccount>;
  register: (values: RegisterPassengerRequest) => Promise<PassengerAccount>;
  logout: () => Promise<void>;
};

export const PassengerAuthContext = createContext<PassengerAuthContextValue | null>(null);

export function usePassengerAuth() {
  const value = useContext(PassengerAuthContext);
  if (!value) throw new Error("usePassengerAuth must be used inside PassengerAuthProvider");
  return value;
}
