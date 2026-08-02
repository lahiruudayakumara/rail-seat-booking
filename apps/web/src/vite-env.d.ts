/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_PAYMENT_PROVIDER?: "sandbox" | "payhere";
}
