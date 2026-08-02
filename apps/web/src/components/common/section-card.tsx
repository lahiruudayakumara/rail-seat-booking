import type { ReactNode } from "react";

interface SectionCardProps {
  title: string;
  icon: ReactNode;
  children: ReactNode;
}

export function SectionCard({ title, icon, children }: SectionCardProps) {
  return (
    <section className="panel p-5 md:p-8">
      <div className="mb-6 flex items-center gap-3 text-[#6b1724]">
        {icon}
        <h2 className="font-heading text-2xl font-bold text-stone-900">{title}</h2>
      </div>
      {children}
    </section>
  );
}
