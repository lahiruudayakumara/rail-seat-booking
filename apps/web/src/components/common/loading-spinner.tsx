import { RefreshCw } from "./icons";

interface LoadingSpinnerProps {
  label: string;
}

export function LoadingSpinner({ label }: LoadingSpinnerProps) {
  return (
    <div className="flex items-center justify-center gap-2 py-4 text-stone-600 font-medium" role="status" aria-live="polite">
      <RefreshCw size={18} className="animate-spin text-maroon-800" />
      <span>{label}</span>
    </div>
  );
}
