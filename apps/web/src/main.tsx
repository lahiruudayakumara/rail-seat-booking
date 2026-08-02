import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { AppRouter } from "@/routes/router";
import "./i18n";
import "./styles.css";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AppRouter />
  </StrictMode>,
);
