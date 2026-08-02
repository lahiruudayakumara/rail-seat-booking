import { Helmet } from "react-helmet-pro";
import { AccountView } from "@/sections/main/account/account-view";

export default function AccountPage() {
  return (
    <>
      <Helmet>
        <title>Passenger Account | Lanka Rail Reserve</title>
      </Helmet>
      <AccountView />
    </>
  );
}
