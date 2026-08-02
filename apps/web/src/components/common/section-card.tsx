import type { ReactNode } from "react";

interface SectionCardProps {
  title: string;
  icon: ReactNode;
  children: ReactNode;
}

export function SectionCard({ title, icon, children }: SectionCardProps) {
  return (
    <section className="panel mt-8 p-5 md:p-8">
      <div className="mb-6 flex items-center gap-3 text-[var(--green)]">
        {icon}
        <h2 className="font-serif text-3xl text-[var(--ink)]">{title}</h2>
      </div>
      {children}
    </section>
  );
}
