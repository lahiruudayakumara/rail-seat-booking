export type TabType = "booking" | "lookup" | "schedules" | "help";

export const PATHS = {
  HOME: "/",
  LOOKUP: "/lookup",
  SCHEDULES: "/schedules",
  HELP: "/help",
  NOT_FOUND: "*",
} as const;

export type PathType = (typeof PATHS)[keyof typeof PATHS];
