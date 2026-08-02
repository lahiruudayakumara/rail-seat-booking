import { RouterProvider } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { HelmetProvider } from "react-helmet-pro";
import { Provider as ReduxProvider } from "react-redux";
import { router } from "@/routes/path";
import { store } from "@/store";
import { PassengerAuthProvider } from "@/auth/passenger-auth-context";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
});

export function AppRouter() {
  return (
    <ReduxProvider store={store}>
      <QueryClientProvider client={queryClient}>
        <PassengerAuthProvider>
          <HelmetProvider>
            <RouterProvider router={router} />
          </HelmetProvider>
        </PassengerAuthProvider>
      </QueryClientProvider>
    </ReduxProvider>
  );
}

export default AppRouter;
