import { Outlet, useLocation, useNavigate } from "react-router-dom";
import type { TabType } from "@/types";
import { Footer, NoticeBanner } from "@/components";
import { useBookingFlow } from "@/hooks/use-booking-flow";
import { HeaderSection } from "@/layouts/main/header";

function pathToTab(pathname: string): TabType {
    if (pathname.startsWith("/lookup")) return "lookup";
    if (pathname.startsWith("/schedules")) return "schedules";
    if (pathname.startsWith("/help")) return "help";
    return "booking";
}

function tabToPath(tab: TabType): string {
    switch (tab) {
        case "booking":
            return "/";
        case "lookup":
            return "/lookup";
        case "schedules":
            return "/schedules";
        case "help":
            return "/help";
        default:
            return "/";
    }
}

export function Main() {
    const { notice } = useBookingFlow();
    const location = useLocation();
    const navigate = useNavigate();

    const activeTab = pathToTab(location.pathname);

    const handleTabChange = (tab: TabType) => {
        const targetPath = tabToPath(tab);
        if (location.pathname !== targetPath) {
            navigate(targetPath);
        }
    };

    return (
        <div className="min-h-screen flex flex-col bg-[#f7f6f5] text-stone-900">
            <HeaderSection activeTab={activeTab} setActiveTab={handleTabChange} />
            <main className="flex-1 mx-auto -mt-16 w-full max-w-6xl px-5 pb-16 md:px-10">
                <NoticeBanner message={notice} />
                <Outlet />
            </main>
            <Footer />
        </div>
    );
}

export default Main;
