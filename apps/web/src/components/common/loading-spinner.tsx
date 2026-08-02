import { RefreshCw } from "./icons";

interface LoadingSpinnerProps {
  label: string;
}

export function LoadingSpinner({ label }: LoadingSpinnerProps) {
  return (
    <div className="loading" role="status" aria-live="polite">
      <RefreshCw size={18} className="animate-spin text-maroon-800" />
      <span>{label}</span>
    </div>
  );
}
