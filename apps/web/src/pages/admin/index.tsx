import { Helmet } from "react-helmet-pro";
import { AdminDashboardView } from "@/sections/admin/dashboard-view";

export function AdminPage() {
  return (
    <>
      <Helmet>
        <title>Operations Dashboard | Lanka Rail Reserve</title>
        <meta name="robots" content="noindex,nofollow" />
      </Helmet>
      <AdminDashboardView />
    </>
  );
}

export default AdminPage;
