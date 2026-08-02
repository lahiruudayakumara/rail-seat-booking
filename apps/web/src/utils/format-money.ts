import type { FareQuote, Money } from "@/types";

export function formatMoney(value: Money | FareQuote): string {
  return new Intl.NumberFormat("en-LK", {
    style: "currency",
    currency: value.currency,
  }).format(value.amountMinor / 10 ** value.currencyScale);
}

export default formatMoney;
