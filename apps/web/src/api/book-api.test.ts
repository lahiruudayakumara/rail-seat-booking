import { beforeEach, expect, test, vi } from "vitest";
import { api } from "./api-instance";
import { getBookingByReference } from "./book-api";

vi.mock("./api-instance", () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

beforeEach(() => {
  vi.mocked(api.post).mockReset();
  vi.mocked(api.post).mockResolvedValue({ data: { reference: "BK-TEST1234" } });
});

test.each([
  [" Passenger@Example.COM ", "passenger@example.com"],
  ["077 000 0123", "+94770000123"],
])("posts a booking lookup using normalized contact %s", async (contact, expected) => {
  await getBookingByReference(" bk-test1234 ", contact);

  expect(api.post).toHaveBeenCalledWith("/api/v1/bookings/access", {
    reference: "BK-TEST1234",
    contact: expected,
  });
});

test("routes group references to the group access endpoint", async () => {
  await getBookingByReference(" gr-testgroup12 ", "077 000 0123");

  expect(api.post).toHaveBeenCalledWith("/api/v1/booking-groups/access", {
    reference: "GR-TESTGROUP12",
    contact: "+94770000123",
  });
});
