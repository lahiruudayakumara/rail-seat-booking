import { useTranslation } from "react-i18next";
import { FAQ_ITEMS } from "@/data";

const HelpView = () => {
  const { t } = useTranslation();

  return (
    <section className="panel p-6 md:p-10 max-w-3xl mx-auto my-10">
      <div className="border-b border-stone-200 pb-5 mb-6">
        <p className="section-kicker">{t("nav.help")}</p>
        <h2 className="font-heading text-3xl font-extrabold text-stone-900">
          {t("help.title")}
        </h2>
      </div>

      <div className="space-y-6">
        {FAQ_ITEMS.map((item) => (
          <div key={item.id} className="border-b border-stone-200 pb-5 last:border-b-0">
            <h3 className="text-lg font-bold text-maroon-900 mb-2">
              {t(item.questionKey)}
            </h3>
            <p className="text-sm text-stone-600 leading-relaxed">
              {t(item.answerKey)}
            </p>
          </div>
        ))}
      </div>
    </section>
  );
}

export default HelpView;
