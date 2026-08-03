import { beforeEach, expect, test, vi } from "vitest";
import { api } from "./api-instance";
import { accessWaitlistEntry, cancelWaitlistEntry, createWaitlistEntry } from "./waitlist-api";

vi.mock("./api-instance", () => ({
  api: { post: vi.fn() },
}));

beforeEach(() => {
  vi.mocked(api.post).mockReset();
  vi.mocked(api.post).mockResolvedValue({ data: { id: "entry-1", reference: "WL-TESTENTRY12" } });
});

test("creates a segment-specific waitlist entry", async () => {
  const request = {
    trainRunId: "run-1",
    originStationId: "station-1",
    destinationStationId: "station-2",
    fullName: "Anura Perera",
    email: "anura@example.com",
    phone: "",
    preferredCoachClass: "FIRST" as const,
  };

  await createWaitlistEntry(request);

  expect(api.post).toHaveBeenCalledWith("/api/v1/waitlist-entries", request);
});

test("normalizes contact when recovering waitlist access", async () => {
  await accessWaitlistEntry(" wl-testentry12 ", "077 000 0123");

  expect(api.post).toHaveBeenCalledWith("/api/v1/waitlist-entries/access", {
    reference: "WL-TESTENTRY12",
    contact: "+94770000123",
  });
});

test("uses the management token when leaving the waitlist", async () => {
  await cancelWaitlistEntry("entry-1", "signed-token");

  expect(api.post).toHaveBeenCalledWith(
    "/api/v1/waitlist-entries/entry-1/cancel",
    {},
    { headers: { Authorization: "Bearer signed-token" } },
  );
});
