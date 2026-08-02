import { useTranslation } from "react-i18next";
import type { TabType } from "@/types";
import { LanguageSelector } from "../../components";
import { TrainFront } from "../../components/common/icons";

interface HeaderSectionProps {
  activeTab: TabType;
  setActiveTab: (tab: TabType) => void;
}

export function HeaderSection({ activeTab, setActiveTab }: HeaderSectionProps) {
  const { t } = useTranslation();

  return (
    <header className="hero px-6 pb-24 pt-8 md:px-10">
      <nav className="mx-auto flex max-w-6xl flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div
          className="flex items-center gap-3 text-lg font-bold tracking-tight text-white cursor-pointer"
          onClick={() => setActiveTab("booking")}
        >
          <span className="logo-mark">
            <TrainFront size={20} />
          </span>
          <span>{t("brand.title")}</span>
        </div>

        {/* Navigation Bar Links */}
        <div className="flex flex-wrap items-center gap-2 sm:gap-4">
          <button
            type="button"
            onClick={() => setActiveTab("booking")}
            className={`nav-tab ${activeTab === "booking" ? "active" : ""}`}
          >
            {t("nav.booking")}
          </button>

          <button
            type="button"
            onClick={() => setActiveTab("lookup")}
            className={`nav-tab ${activeTab === "lookup" ? "active" : ""}`}
          >
            {t("nav.lookup")}
          </button>

          <button
            type="button"
            onClick={() => setActiveTab("schedules")}
            className={`nav-tab ${activeTab === "schedules" ? "active" : ""}`}
          >
            {t("nav.schedules")}
          </button>

          <button
            type="button"
            onClick={() => setActiveTab("help")}
            className={`nav-tab ${activeTab === "help" ? "active" : ""}`}
          >
            {t("nav.help")}
          </button>

          <LanguageSelector />
        </div>
      </nav>

      <div className="mx-auto mt-14 max-w-6xl text-white">
        <p className="font-mono text-xs font-bold tracking-widest text-amber-200 uppercase">
          {t("brand.subwayLine")}
        </p>

        <h1 className="mt-3 max-w-3xl font-heading text-4xl font-extrabold text-white md:text-6xl">
          {t("brand.heroHeadline")}
        </h1>

        <p className="mt-3 max-w-xl text-base text-white/80">
          {t("brand.heroSub")}
        </p>
      </div>
    </header>
  );
}
