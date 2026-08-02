import { useTranslation } from "react-i18next";
import { STATIONS_SCHEDULE_DATA } from "@/data";

const SchedulesView = () => {
  const { t } = useTranslation();

  return (
    <section className="panel p-6 md:p-10 max-w-4xl mx-auto my-10">
      <div className="border-b border-stone-200 pb-5 mb-6">
        <p className="section-kicker">{t("nav.schedules")}</p>
        <h2 className="font-heading text-3xl font-extrabold text-stone-900">
          {t("schedules.title")}
        </h2>
        <p className="mt-2 text-sm text-stone-600">
          {t("schedules.sub")}
        </p>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-stone-300 bg-stone-100 text-stone-700 uppercase font-mono text-xs">
              <th className="py-3 px-4">{t("schedules.positionHeader")}</th>
              <th className="py-3 px-4">{t("schedules.codeHeader")}</th>
              <th className="py-3 px-4">{t("schedules.stationHeader")}</th>
              <th className="py-3 px-4">{t("schedules.distanceHeader")}</th>
            </tr>
          </thead>
          <tbody>
            {STATIONS_SCHEDULE_DATA.map((st) => (
              <tr key={st.code} className="border-b border-stone-200 hover:bg-stone-50">
                <td className="py-3 px-4 font-mono font-bold text-maroon-900">#{st.position + 1}</td>
                <td className="py-3 px-4 font-mono font-bold text-stone-700">{st.code}</td>
                <td className="py-3 px-4 font-bold text-stone-900">
                  {t(`stations.${st.name}`, st.name)}
                </td>
                <td className="py-3 px-4 font-mono text-stone-600">{st.dist}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}

export default SchedulesView;