import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import App from "./App";
import { MathText } from "./Questions";
import { fixture } from "./test-fixtures";

beforeEach(() => {
  window.history.replaceState(null, "", "#overview");
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue({ ok: true, json: async () => fixture() }),
  );
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

it("loads real report metrics, changes cohorts, and excludes incomplete rows on request", async () => {
  render(<App />);
  expect(screen.getByRole("status").textContent).toContain(
    "Loading your benchmark",
  );
  await screen.findByRole("button", { name: "test/alpha, medium details" });
  fireEvent.click(screen.getByLabelText("Show incomplete"));
  const table = screen.getAllByRole("table")[0];
  expect(within(table).queryByText("Incomplete")).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "With images" }));
  await screen.findByRole("button", { name: "test/beta, medium details" });
  expect(screen.getAllByText("100%").length).toBeGreaterThan(0);
  fireEvent.change(screen.getByLabelText("Search models"), {
    target: { value: "no-such-model" },
  });
  expect(screen.getByText("No results match these filters.")).toBeTruthy();
});

it("preserves existing results when refresh fails, then recovers", async () => {
  const fetch = vi.mocked(globalThis.fetch);
  render(<App />);
  await screen.findByRole("button", { name: "test/alpha, medium details" });
  fetch.mockResolvedValueOnce({ ok: false, status: 503 } as Response);
  fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
  expect((await screen.findByRole("alert")).textContent).toContain(
    "Showing the last loaded results",
  );
  expect(
    screen.getByRole("button", { name: "test/alpha, medium details" }),
  ).toBeTruthy();
  fireEvent.click(screen.getByRole("button", { name: "Try again →" }));
  await waitFor(() => expect(screen.queryByRole("alert")).toBeNull());
});

it("shows original Persian math, answer keys and outcomes in the question explorer", async () => {
  window.history.replaceState(null, "", "#questions");
  render(<App />);
  await screen.findByText("Key: option 2");
  expect(screen.getByText("مقدار", { exact: false })).toBeTruthy();
  expect(document.querySelector(".katex")).toBeTruthy();
  expect(within(screen.getByRole("table")).getByText("Correct")).toBeTruthy();
  fireEvent.change(screen.getByLabelText("Answer filter"), {
    target: { value: "wrong" },
  });
  expect(screen.getByText("No answers match this filter.")).toBeTruthy();
  fireEvent.click(screen.getByRole("button", { name: "With images" }));
  expect(screen.getByAltText("Diagram for question 2")).toBeTruthy();
});

it("supports navigation without reloading the report", async () => {
  render(<App />);
  await screen.findByRole("button", { name: "test/alpha, medium details" });
  act(() => {
    window.location.hash = "spending";
    window.dispatchEvent(new HashChangeEvent("hashchange"));
  });
  expect(
    screen.getByRole("heading", { name: "Run spending", level: 1 }),
  ).toBeTruthy();
  expect(screen.getByText("run-1")).toBeTruthy();
  expect(globalThis.fetch).toHaveBeenCalledTimes(1);
});

it("renders input text as text and does not trust KaTeX HTML extensions", () => {
  const { container } = render(
    <MathText
      text={"<script>bad()</script> $\\href{javascript:alert(1)}{x}$"}
    />,
  );
  expect(container.querySelector("script")).toBeNull();
  expect(container.querySelector('a[href^="javascript:"]')).toBeNull();
  expect(container.textContent).toContain("<script>");
});
