import { Check } from "@/components";
import { useAppSelector } from "@/store";

const steps = ["Journey", "Train", "Seat", "Passenger"];

export function BookingProgress() {
  const searched = useAppSelector((state) => state.search.searched);
  const { runId, selectedSeat, booking, group } = useAppSelector((state) => state.booking);
  const current = booking || group ? 4 : selectedSeat ? 3 : runId ? 2 : searched ? 1 : 0;

  return (
    <nav className="panel px-4 py-4 md:px-7" aria-label="Booking progress">
      <ol className="grid grid-cols-4 gap-1">
        {steps.map((step, index) => {
          const complete = index < current || Boolean(booking || group);
          const active = index === current && !booking && !group;
          return (
            <li key={step} className="relative flex flex-col items-center text-center">
              {index > 0 && <span className={`absolute right-1/2 top-4 h-0.5 w-full ${index <= current ? "bg-[#851e2e]" : "bg-stone-200"}`} aria-hidden="true" />}
              <span className={`relative z-10 flex h-8 w-8 items-center justify-center rounded-full border-2 text-xs font-extrabold ${complete ? "border-[#851e2e] bg-[#851e2e] text-white" : active ? "border-[#851e2e] bg-white text-[#851e2e]" : "border-stone-200 bg-white text-stone-400"}`}>
                {complete ? <Check size={14} /> : index + 1}
              </span>
              <span className={`mt-2 text-[10px] font-bold uppercase tracking-wide sm:text-xs ${active || complete ? "text-stone-800" : "text-stone-400"}`}>{step}</span>
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
