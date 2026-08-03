import { forwardRef, type ButtonHTMLAttributes, type ReactNode } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary";
  children: ReactNode;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(function Button({ variant = "primary", className = "", children, ...props }, ref) {
  const baseClass = variant === "primary" ? "primary-button" : "secondary-button";
  return (
    <button ref={ref} className={`${baseClass} ${className}`.trim()} {...props}>
      {children}
    </button>
  );
});
