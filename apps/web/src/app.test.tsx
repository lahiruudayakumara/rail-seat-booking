import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Provider as ReduxProvider } from "react-redux";
import { MemoryRouter } from "react-router-dom";
import { vi } from "vitest";
import { createWaitlistEntry, getBookingByReference, registerPassenger } from "@/api";
import { AppRouter } from "@/routes/router";
import LookupView from "@/sections/main/lookup/view/lookup-view";
import { AdminDashboardView } from "@/sections/admin/dashboard-view";
import { store } from "@/store";
import { PassengerAuthProvider } from "@/auth/passenger-auth-context";
import { AccountView } from "@/sections/main/account/account-view";
import { WaitlistForm } from "@/sections/main/booking/waitlist-form";

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
  getStations: vi.fn().mockResolvedValue({
    items: [
      { id: "station-1", code: "FOT", name: "Colombo Fort" },
      { id: "station-2", code: "KDY", name: "Kandy" },
    ],
  }),
  getTrainRuns: vi.fn().mockResolvedValue({ items: [] }),
  getTrainRun: vi.fn().mockResolvedValue({
    id: "run-1",
    routeId: "route-1",
    trainId: "train-1",
    serviceDate: "2026-08-03",
    departureAt: "2026-08-03T00:00:00Z",
    arrivalAt: "2026-08-03T03:00:00Z",
    status: "SCHEDULED",
  }),
        
  getAvailableSeats: vi.fn().mockResolvedValue({ items: [] }),
  getSeatMap: vi.fn().mockResolvedValue({ items: [] }),
  getQuote: vi.fn(),
  createHold: vi.fn(),
  releaseHold: vi.fn(),
  createBooking: vi.fn(),
  createBookingGroup: vi.fn(),
  checkoutSandbox: vi.fn(),
  checkoutSandboxGroup: vi.fn(),
  paymentProvider: "sandbox",
  startPayHereCheckout: vi.fn(),
  startPayHereGroupCheckout: vi.fn(),
  getPayHerePayment: vi.fn(),
  redirectToPayHere: vi.fn(),
  getStoredPayHereAccess: vi.fn(),
  clearStoredPayHereAccess: vi.fn(),
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
  createWaitlistEntry: vi.fn(),
  accessWaitlistEntry: vi.fn(),
  cancelWaitlistEntry: vi.fn(),
  getAdminDashboard: vi.fn(),
  getCurrentPassenger: vi.fn().mockResolvedValue(null),
  loginPassenger: vi.fn(),
  registerPassenger: vi.fn().mockResolvedValue({
    id: "account-1",
    fullName: "Anura Perera",
    email: "anura@example.com",
    phone: "+94770000000",
    createdAt: "2026-08-03T00:00:00Z",
  }),
  logoutPassenger: vi.fn(),
  getPassengerBookings: vi.fn().mockResolvedValue([]),
  getSavedTravellers: vi.fn().mockResolvedValue([]),
  createSavedTraveller: vi.fn(),
  updateSavedTraveller: vi.fn(),
  deleteSavedTraveller: vi.fn(),
  getPassengerPreferences: vi.fn().mockResolvedValue({ preferredCoachClass: "ANY", preferredSeatType: "ANY", language: "en" }),
  updatePassengerPreferences: vi.fn(),
}));

test("renders the journey search page", async () => {
  const user = userEvent.setup();
  render(<AppRouter />);
  expect(screen.getByRole("heading", { name: /where are you travelling/i })).toBeInTheDocument();
  expect(screen.getByRole("button", { name: /find trains/i })).toBeInTheDocument();
  expect(await screen.findByRole("option", { name: "Colombo Fort" })).toBeInTheDocument();
  await user.selectOptions(screen.getByRole("combobox", { name: "Origin Station" }), "station-1");
  expect(screen.getByRole("option", { name: "Kandy" })).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "සිංහල" }));
  expect(
    screen.getByRole("option", { name: "කොළඹ කොටුව–බදුල්ල" }),
  ).toBeInTheDocument();
  await user.click(screen.getByRole("button", { name: "English" }));
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
        <PassengerAuthProvider>
          <LookupView />
        </PassengerAuthProvider>
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

test("creates an optional passenger account", async () => {
  const user = userEvent.setup();
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <MemoryRouter>
      <QueryClientProvider client={queryClient}>
        <PassengerAuthProvider>
          <AccountView />
        </PassengerAuthProvider>
      </QueryClientProvider>
    </MemoryRouter>,
  );

  await user.click(await screen.findByRole("tab", { name: "Create account" }));
  await user.type(screen.getByLabelText("Full name"), "Anura Perera");
  await user.type(screen.getByLabelText("Email address"), "anura@example.com");
  await user.type(screen.getByLabelText(/Phone/), "+94770000000");
  await user.type(screen.getByLabelText("Password"), "ScenicRail2026");
  await user.type(screen.getByLabelText("Confirm password"), "ScenicRail2026");
  await user.click(screen.getByRole("button", { name: "Create passenger account" }));

  expect(registerPassenger).toHaveBeenCalledWith({
    fullName: "Anura Perera",
    email: "anura@example.com",
    phone: "+94770000000",
    password: "ScenicRail2026",
  });
  expect(await screen.findByText(/Welcome, Anura Perera/)).toBeInTheDocument();
});

test("joins the waitlist for a sold-out journey segment", async () => {
  vi.mocked(createWaitlistEntry).mockResolvedValueOnce({
    id: "waitlist-1",
    reference: "WL-TESTENTRY12",
    trainRunId: "run-1",
    originStationId: "station-1",
    destinationStationId: "station-2",
    fullName: "Anura Perera",
    email: "anura@example.com",
    preferredCoachClass: "ANY",
    status: "WAITING",
    createdAt: "2026-08-03T00:00:00Z",
    managementToken: "waitlist-token",
  });
  const user = userEvent.setup();
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } });
  render(
    <QueryClientProvider client={queryClient}>
      <PassengerAuthProvider>
        <WaitlistForm trainRunId="run-1" originStationId="station-1" destinationStationId="station-2" />
      </PassengerAuthProvider>
    </QueryClientProvider>,
  );

  await user.type(screen.getByRole("textbox", { name: "Full name" }), "Anura Perera");
  await user.type(screen.getByRole("textbox", { name: "Email (optional)" }), "anura@example.com");
  await user.click(screen.getByRole("button", { name: "Join waitlist" }));

  expect(createWaitlistEntry).toHaveBeenCalledWith({
    trainRunId: "run-1",
    originStationId: "station-1",
    destinationStationId: "station-2",
    fullName: "Anura Perera",
    email: "anura@example.com",
    phone: "",
    preferredCoachClass: "ANY",
  });
  expect(await screen.findByText("WL-TESTENTRY12")).toBeInTheDocument();
});
