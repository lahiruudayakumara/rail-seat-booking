export type TabType = "booking" | "lookup" | "schedules" | "help" | "account";

export const PATHS = {
  HOME: "/",
  LOOKUP: "/lookup",
  SCHEDULES: "/schedules",
  HELP: "/help",
  ACCOUNT: "/account",
  NOT_FOUND: "*",
} as const;

export type PathType = (typeof PATHS)[keyof typeof PATHS];
