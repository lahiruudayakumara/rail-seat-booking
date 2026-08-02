import { createBrowserRouter } from "react-router-dom";
import HelpPage from "@/pages/main/help";
import HomePage from "@/pages/main/home";
import LookupPage from "@/pages/main/lookup";
import NotFoundPage from "@/pages/not-found";
import SchedulesPage from "@/pages/main/schedules";
import Main from "@/layouts/main";

export const router = createBrowserRouter([
  {
    path: "/",
    Component: Main,
    children: [
      {
        index: true,
        Component: HomePage,
      },
      {
        path: "lookup",
        Component: LookupPage,
      },
      {
        path: "schedules",
        Component: SchedulesPage,
      },
      {
        path: "help",
        Component: HelpPage,
      },
    ],
  },
  {
    path: "*",
    Component: NotFoundPage,
  },
]);