import { beforeEach, expect, test, vi } from "vitest";
import { api } from "./api-instance";
import { updatePassengerPreferences } from "./passenger-account-api";

vi.mock("./api-instance", () => ({
  api: {
    put: vi.fn(),
  },
}));

beforeEach(() => {
  vi.mocked(api.put).mockReset();
});

test("does not send read-only preference response fields when saving", async () => {
  const saved = {
    preferredCoachClass: "FIRST" as const,
    preferredSeatType: "WINDOW" as const,
    language: "en" as const,
    updatedAt: "2026-08-04T01:00:00Z",
  };
  vi.mocked(api.put).mockResolvedValue({ data: saved });

  await updatePassengerPreferences(saved);

  expect(api.put).toHaveBeenCalledWith("/api/v1/passenger/preferences", {
    preferredCoachClass: "FIRST",
    preferredSeatType: "WINDOW",
    language: "en",
  });
});
