import { render, screen } from "@testing-library/react";
import { vi } from "vitest";
import { AppRouter } from "@/routes/router";

vi.stubGlobal(
  "fetch",
  vi.fn(
    async () =>
      new Response(JSON.stringify({ items: [] }), {
        status: 200,
        headers: { "Content-Type": "application/json" },
      }),
  ),
);

test("renders the journey search page", async () => {
  render(<AppRouter />);
  expect(screen.getByRole("heading", { name: /where are you travelling/i })).toBeInTheDocument();
  expect(screen.getByRole("button", { name: /find trains/i })).toBeInTheDocument();
});
