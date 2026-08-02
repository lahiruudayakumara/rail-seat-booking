import { useTranslation } from "react-i18next";

export function LanguageSelector() {
  const { i18n } = useTranslation();

  const currentLanguage = i18n.language?.slice(0, 2) || "en";

  const changeLanguage = (lng: string) => {
    i18n.changeLanguage(lng);
  };

  return (
    <div className="flex items-center gap-1 rounded-lg border border-white/20 bg-black/20 p-1 text-xs font-bold text-white">
      <button
        type="button"
        onClick={() => changeLanguage("en")}
        className={`cursor-pointer rounded px-2.5 py-1 transition-colors ${
          currentLanguage === "en"
            ? "bg-white text-stone-900"
            : "text-white/80 hover:text-white"
        }`}
      >
        English
      </button>
      <button
        type="button"
        onClick={() => changeLanguage("si")}
        className={`cursor-pointer rounded px-2.5 py-1 transition-colors ${
          currentLanguage === "si"
            ? "bg-white text-stone-900"
            : "text-white/80 hover:text-white"
        }`}
      >
        සිංහල
      </button>
      <button
        type="button"
        onClick={() => changeLanguage("ta")}
        className={`cursor-pointer rounded px-2.5 py-1 transition-colors ${
          currentLanguage === "ta"
            ? "bg-white text-stone-900"
            : "text-white/80 hover:text-white"
        }`}
      >
        தமிழ்
      </button>
    </div>
  );
}
