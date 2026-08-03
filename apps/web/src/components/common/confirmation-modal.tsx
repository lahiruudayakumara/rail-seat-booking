import { useEffect, useId, useRef } from "react";
import { createPortal } from "react-dom";
import { Button } from "./button";

interface ConfirmationModalProps {
  open: boolean;
  title: string;
  description: string;
  confirmLabel: string;
  cancelLabel?: string;
  isPending?: boolean;
  error?: string;
  tone?: "default" | "danger";
  onConfirm: () => void;
  onClose: () => void;
}

export function ConfirmationModal({
  open,
  title,
  description,
  confirmLabel,
  cancelLabel = "Cancel",
  isPending = false,
  error,
  tone = "default",
  onConfirm,
  onClose,
}: ConfirmationModalProps) {
  const titleId = useId();
  const descriptionId = useId();
  const cancelButtonRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (!open) return;
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    cancelButtonRef.current?.focus();

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape" && !isPending) onClose();
    };
    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [isPending, onClose, open]);

  if (!open) return null;

  return createPortal(
    <div
      className="fixed inset-0 z-[100] flex items-end justify-center bg-stone-950/55 p-4 backdrop-blur-[2px] sm:items-center"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget && !isPending) onClose();
      }}
    >
      <div
        role="alertdialog"
        aria-modal="true"
        aria-labelledby={titleId}
        aria-describedby={descriptionId}
        className="w-full max-w-md rounded-2xl border border-stone-200 bg-white p-5 shadow-2xl sm:p-6"
      >
        <div className={`flex h-10 w-10 items-center justify-center rounded-full text-lg font-black ${tone === "danger" ? "bg-red-100 text-red-700" : "bg-[#6b1724]/10 text-[#6b1724]"}`} aria-hidden="true">
          !
        </div>
        <h2 id={titleId} className="mt-4 font-heading text-xl font-extrabold text-stone-900">{title}</h2>
        <p id={descriptionId} className="mt-2 text-sm leading-6 text-stone-600">{description}</p>
        {error && <p className="mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm font-semibold text-red-700" role="alert">{error}</p>}
        <div className="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <Button ref={cancelButtonRef} variant="secondary" className="w-full sm:w-auto" onClick={onClose} disabled={isPending}>{cancelLabel}</Button>
          <button
            type="button"
            className={`h-[46px] w-full rounded-[10px] px-5 text-sm font-bold text-white transition disabled:cursor-not-allowed disabled:opacity-60 sm:w-auto ${tone === "danger" ? "bg-red-700 hover:bg-red-800" : "bg-[#6b1724] hover:bg-[#540d17]"}`}
            onClick={onConfirm}
            disabled={isPending}
          >
            {isPending ? "Please wait…" : confirmLabel}
          </button>
        </div>
      </div>
    </div>,
    document.body,
  );
}
