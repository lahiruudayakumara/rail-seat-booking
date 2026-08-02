import { useTranslation } from "react-i18next";
import { TrainFront } from "./icons";

export function Footer() {
  const { t } = useTranslation();

  return (
    <footer className="mt-20 border-t border-stone-200 bg-stone-900 py-12 text-white">
      <div className="mx-auto max-w-6xl px-6 md:px-10">
        <div className="flex flex-col gap-6 md:flex-row md:items-center md:justify-between border-b border-stone-800 pb-8">
          <div className="flex items-center gap-3 text-lg font-bold">
            <span className="logo-mark">
              <TrainFront size={20} />
            </span>
            <span>{t("brand.title")}</span>
          </div>

          <p className="text-xs text-stone-400 max-w-md">
            {t("footer.desc")}
          </p>
        </div>

        <div className="mt-8 flex flex-col gap-4 text-xs text-stone-500 md:flex-row md:items-center md:justify-between">
          <p>{t("footer.copyright")}</p>
          <div className="flex items-center gap-4">
            <span>{t("brand.subwayLine")}</span>
          </div>
        </div>
      </div>
    </footer>
  );
}
