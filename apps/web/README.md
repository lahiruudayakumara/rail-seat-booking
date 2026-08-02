# Booking Web Application

React and TypeScript passenger interface for segment-based train booking. It uses Vite, TanStack Query, React Hook Form, Zod and Tailwind CSS.

From the repository root:

```bash
pnpm install
pnpm dev
```

Set `VITE_API_BASE_URL` to the Go API origin; it defaults to `http://localhost:8080`. The recommended full-stack workflow remains `docker compose up --build` from the repository root.

The flow supports journey search, train selection, segment-aware seat availability, a coach seat map, fare quote, passenger validation, booking confirmation and cancellation. A booking `409` refreshes availability, clears the lost seat and preserves passenger-entered data.
