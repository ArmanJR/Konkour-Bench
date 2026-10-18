import { afterEach, describe, expect, it, vi } from "vitest";
import {
  fetchReport,
  initialFilters,
  leaderboardCSV,
  matchesOutcome,
  progress,
  selectRows,
  sortRows,
} from "./data";
import { fixture, outcome, row } from "./test-fixtures";

afterEach(() => vi.unstubAllGlobals());
describe("benchmark comparisons", () => {
  it("uses the exact cohort and category aggregates instead of double-counting rows", () => {
    const rows = [
      row(),
      row({ cohort: "full" }),
      row({ category: "ریاضی" }),
      row({ dataset: "exam" }),
      row({ tier: "max" }),
    ];
    expect(selectRows(rows, { ...initialFilters, tier: "medium" })).toEqual([
      rows[0],
    ]);
    expect(
      selectRows(rows, {
        ...initialFilters,
        category: "ریاضی",
        search: "ALPHA",
      }),
    ).toEqual([rows[2]]);
  });
  it("keeps unknown metrics last in both sort directions and preserves zero", () => {
    const rows = [
      row({ model: "unknown", cost_usd: null }),
      row({ model: "free", cost_usd: 0 }),
      row({ model: "paid", cost_usd: 1 }),
    ];
    expect(sortRows(rows, "cost_usd", true).map((r) => r.model)).toEqual([
      "free",
      "paid",
      "unknown",
    ]);
    expect(sortRows(rows, "cost_usd", false).map((r) => r.model)).toEqual([
      "paid",
      "free",
      "unknown",
    ]);
  });
  it("does not turn unsupported or skipped questions into successful evaluations", () => {
    expect(
      progress([
        outcome(),
        outcome({ status: "not_run" }),
        outcome({ status: "unsupported" }),
        outcome({ status: "skipped" }),
      ]),
    ).toEqual({ complete: 1, eligible: 2, excluded: 2, total: 4 });
    expect(matchesOutcome(outcome({ has_image: true }), initialFilters)).toBe(
      false,
    );
    expect(
      matchesOutcome(outcome({ has_image: true }), {
        ...initialFilters,
        cohort: "image-only",
      }),
    ).toBe(true);
  });
  it("exports filtered values with CSV escaping and protects formula cells", () => {
    const csv = leaderboardCSV([
      row({ model: '=formula,"x"', accuracy: null }),
    ]);
    expect(csv).toContain('"\'=formula,""x"""');
    expect(csv).toContain('"accuracy"');
    expect(csv).not.toContain("null");
  });
});
describe("API errors", () => {
  it("rejects HTTP errors and malformed documents", async () => {
    const fetch = vi
      .fn()
      .mockResolvedValueOnce({ ok: false, status: 500 })
      .mockResolvedValueOnce({ ok: true, json: async () => ({ rows: [] }) })
      .mockResolvedValueOnce({ ok: true, json: async () => fixture() });
    vi.stubGlobal("fetch", fetch);
    const signal = new AbortController().signal;
    await expect(fetchReport(signal)).rejects.toThrow("HTTP 500");
    await expect(fetchReport(signal)).rejects.toThrow(
      "invalid benchmark report",
    );
    await expect(fetchReport(signal)).resolves.toMatchObject({
      models: ["test/alpha", "test/beta"],
    });
  });
});
