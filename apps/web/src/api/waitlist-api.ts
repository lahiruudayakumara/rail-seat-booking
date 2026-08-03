import { api } from "./api-instance";
import { normalizeBookingLookupContact } from "./book-api";
import type { CreateWaitlistEntryRequest, WaitlistEntry } from "@/types";

export async function createWaitlistEntry(body: CreateWaitlistEntryRequest) {
  const response = await api.post<WaitlistEntry>("/api/v1/waitlist-entries", body);
  return response.data;
}

export async function accessWaitlistEntry(reference: string, contact: string) {
  const response = await api.post<WaitlistEntry>("/api/v1/waitlist-entries/access", {
    reference: reference.trim().toLocaleUpperCase(),
    contact: normalizeBookingLookupContact(contact),
  });
  return response.data;
}

export async function cancelWaitlistEntry(entryId: string, managementToken: string) {
  const response = await api.post<WaitlistEntry>(
    `/api/v1/waitlist-entries/${entryId}/cancel`,
    {},
    { headers: { Authorization: `Bearer ${managementToken}` } },
  );
  return response.data;
}
