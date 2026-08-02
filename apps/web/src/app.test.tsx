import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Provider as ReduxProvider } from "react-redux";
import { MemoryRouter } from "react-router-dom";
import { vi } from "vitest";
import { getBookingByReference } from "@/api";
import { AppRouter } from "@/routes/router";
import LookupView from "@/sections/main/lookup/view/lookup-view";
import { AdminDashboardView } from "@/sections/admin/dashboard-view";
import { store } from "@/store";

vi.mock("@/api", () => ({
  getRoutes: vi.fn().mockResolvedValue({
    items: [
      {
        id: "route-1",
        code: "CF-BD-UP",
        name: "Colombo Fort–Badulla",
        direction: "UP_COUNTRY",
        timezone: "Asia/Colombo",
      },
    ],
  }),
  getRouteStations: vi.fn().mockResolvedValue({
    items: [
      { id: "station-1", code: "FOT", name: "Colombo Fort", position: 0 },
      { id: "station-2", code: "KDY", name: "Kandy", position: 1 },
    ],
  }),
  getTrainRuns: vi.fn().mockResolvedValue({ items: [] }),
  getAvailableSeats: vi.fn().mockResolvedValue({ items: [] }),
  getSeatMap: vi.fn().mockResolvedValue({ items: [] }),
  getQuote: vi.fn(),
  createHold: vi.fn(),
  createBooking: vi.fn(),
  checkoutSandbox: vi.fn(),
  cancelBooking: vi.fn(),
  getBookingByReference: vi.fn().mockResolvedValue({
    id: "booking-1",
    reference: "BK-TEST1234",
    status: "CONFIRMED",
    trainRunId: "run-1",
    seat: {
      id: "seat-1",
      label: "1A",
      coachId: "coach-1",
      coachCode: "R1",
      coachClass: "FIRST",
      attributes: ["WINDOW"],
    },
    originStationId: "station-1",
    destinationStationId: "station-2",
    fare: { amountMinor: 69000, currency: "LKR", currencyScale: 2 },
    createdAt: "2026-08-03T00:00:00Z",
    managementToken: "test-management-token",
  }),
  getAdminDashboard: vi.fn(),
}));

test("renders the journey search page", async () => {
  render(<AppRouter />);
  expect(screen.getByRole("heading", { name: /where are you travelling/i })).toBeInTheDocument();
  expect(screen.getByRole("button", { name: /find trains/i })).toBeInTheDocument();
  expect(await screen.findByRole("option", { name: "Colombo Fort" })).toBeInTheDocument();
  expect(screen.getByRole("option", { name: "Kandy" })).toBeInTheDocument();
});

test("keeps administrator credentials in the browser session", async () => {
  sessionStorage.clear();
  const user = userEvent.setup();
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>
        <AdminDashboardView />
      </QueryClientProvider>
    </MemoryRouter>,
  );

  await user.type(screen.getByLabelText("Administrator key"), "session-admin-key");
  await user.click(screen.getByRole("button", { name: "Open dashboard" }));

  expect(await screen.findByRole("heading", { name: "Operations control" })).toBeInTheDocument();
  expect(sessionStorage.getItem("rail-admin-session-key")).toBe("session-admin-key");
  await user.click(screen.getByRole("button", { name: "Sign out" }));
  expect(await screen.findByRole("heading", { name: "Operations dashboard" })).toBeInTheDocument();
  expect(sessionStorage.getItem("rail-admin-session-key")).toBeNull();
});

test("looks up a booking through the backend API", async () => {
  const user = userEvent.setup();
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  render(
    <ReduxProvider store={store}>
      <QueryClientProvider client={queryClient}>
        <LookupView />
      </QueryClientProvider>
    </ReduxProvider>,
  );

  await user.type(screen.getByRole("textbox", { name: "Booking Reference" }), "BK-TEST1234");
  await user.type(
    screen.getByRole("textbox", { name: "Booking email or phone" }),
    "passenger@example.com",
  );
  await user.click(screen.getByRole("button", { name: "Search Booking" }));

  expect(getBookingByReference).toHaveBeenCalledWith("BK-TEST1234", "passenger@example.com");
  expect(await screen.findByText("BK-TEST1234")).toBeInTheDocument();
});
