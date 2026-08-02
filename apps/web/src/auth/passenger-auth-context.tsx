import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import axios from "axios";
import type { ReactNode } from "react";
import {
  getCurrentPassenger,
  loginPassenger,
  logoutPassenger,
  registerPassenger,
} from "@/api";
import { PassengerAuthContext } from "./use-passenger-auth";
import type { RegisterPassengerRequest } from "@/types";

export function PassengerAuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const accountQuery = useQuery({
    queryKey: ["passenger-account"],
    queryFn: async () => {
      try {
        return await getCurrentPassenger();
      } catch (error) {
        if (axios.isAxiosError(error) && error.response?.status === 401) return null;
        throw error;
      }
    },
    retry: false,
    staleTime: 60_000,
  });

  const loginMutation = useMutation({
    mutationFn: (values: { email: string; password: string }) => loginPassenger(values),
    onSuccess: (account) => queryClient.setQueryData(["passenger-account"], account),
  });
  const registerMutation = useMutation({
    mutationFn: (values: RegisterPassengerRequest) => registerPassenger(values),
    onSuccess: (account) => queryClient.setQueryData(["passenger-account"], account),
  });
  const logoutMutation = useMutation({
    mutationFn: logoutPassenger,
    onSuccess: () => {
      queryClient.setQueryData(["passenger-account"], null);
      queryClient.removeQueries({ queryKey: ["passenger-bookings"] });
    },
  });

  return (
    <PassengerAuthContext.Provider
      value={{
        account: accountQuery.data ?? null,
        isLoading: accountQuery.isLoading,
        login: loginMutation.mutateAsync,
        register: registerMutation.mutateAsync,
        logout: logoutMutation.mutateAsync,
      }}
    >
      {children}
    </PassengerAuthContext.Provider>
  );
}
