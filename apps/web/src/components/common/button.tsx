import type { ButtonHTMLAttributes, ReactNode } from "react";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary";
  children: ReactNode;
}

export function Button({ variant = "primary", className = "", children, ...props }: ButtonProps) {
  const baseClass = variant === "primary" ? "primary-button" : "secondary-button";
  return (
    <button className={`${baseClass} ${className}`.trim()} {...props}>
      {children}
    </button>
  );
}
