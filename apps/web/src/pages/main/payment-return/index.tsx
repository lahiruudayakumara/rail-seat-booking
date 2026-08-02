import { Helmet } from "react-helmet-pro";
import { PaymentReturnView } from "@/sections/main/payment/payment-return-view";

export default function PaymentReturnPage({ cancelled = false }: { cancelled?: boolean }) {
  return <><Helmet><title>Payment Status | Lanka Rail Reserve</title></Helmet><PaymentReturnView cancelled={cancelled} /></>;
}

export function PaymentCancelPage() {
  return <PaymentReturnPage cancelled />;
}
