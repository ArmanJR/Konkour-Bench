import type { Filters, Outcome, Report, Row } from "./types";

export const initialFilters: Filters = {
  cohort: "text-only",
  dataset: "",
  category: "",
  tier: "",
  search: "",
};
export const tiers = ["minimal", "medium", "max"] as const;
export const tierColors = {
  minimal: "#829db5",
  medium: "#c29b55",
  max: "#357c65",
};
export const number = (value: number | null | undefined, digits = 0) =>
  value == null
    ? "—"
    : value.toLocaleString("en-US", { maximumFractionDigits: digits });
export const percent = (value: number | null) =>
  value == null ? "—" : `${number(value * 100, 1)}%`;
export const money = (value: number | null | undefined, digits = 2) =>
  value == null
    ? "—"
    : `$${value.toLocaleString("en-US", { minimumFractionDigits: digits, maximumFractionDigits: digits })}`;
export const modelName = (model: string) =>
  model
    .split("/")
    .at(-1)!
    .replace(/-(?:\d{4})$/, "")
    .replaceAll("-", " ");
export const datasetName = (id: string) =>
  id.replaceAll("-", " ").replace(/\b\w/g, (c) => c.toUpperCase());
export function matchesOutcome(o: Outcome, f: Filters) {
  return (
    (!f.dataset || o.dataset === f.dataset) &&
    (!f.category || o.category === f.category) &&
    (!f.tier || o.tier === f.tier) &&
    o.model.toLowerCase().includes(f.search.toLowerCase()) &&
    (f.cohort === "full" ||
      (f.cohort === "image-only" ? o.has_image : !o.has_image))
  );
}
export function selectRows(rows: Row[], f: Filters) {
  return rows.filter(
    (r) =>
      r.cohort === f.cohort &&
      r.dataset === f.dataset &&
      r.category === f.category &&
      (!f.tier || r.tier === f.tier) &&
      r.model.toLowerCase().includes(f.search.toLowerCase()),
  );
}
export type SortKey =
  | "model"
  | "accuracy"
  | "cost_usd"
  | "total_tokens"
  | "complete";
export function sortRows(rows: Row[], key: SortKey, ascending: boolean) {
  return [...rows].sort((a, b) => {
    const av = a[key],
      bv = b[key];
    if (av == null && bv != null) return 1;
    if (av != null && bv == null) return -1;
    const value =
      typeof av === "string" && typeof bv === "string"
        ? av.localeCompare(bv)
        : Number(av) - Number(bv);
    return (
      (ascending ? value : -value) ||
      a.model.localeCompare(b.model) ||
      tiers.indexOf(a.tier) - tiers.indexOf(b.tier)
    );
  });
}
export function progress(outcomes: Outcome[]) {
  const complete = outcomes.filter((o) => o.status === "complete").length;
  const excluded = outcomes.filter(
    (o) => o.status === "unsupported" || o.status === "skipped",
  ).length;
  return {
    complete,
    excluded,
    total: outcomes.length,
    eligible: outcomes.length - excluded,
  };
}
export function leaderboardCSV(rows: Row[]) {
  const cell = (value: unknown) => {
    let text = value == null ? "" : String(value);
    if (/^[=+@\-\t\r]/.test(text)) text = `'${text}`;
    return `"${text.replaceAll('"', '""')}"`;
  };
  const fields: (keyof Row)[] = [
    "model",
    "tier",
    "cohort",
    "dataset",
    "category",
    "status",
    "total",
    "complete",
    "correct",
    "accuracy",
    "cost_usd",
    "total_tokens",
    "cost_frontier",
    "token_frontier",
    "pareto_3d",
  ];
  return [fields, ...rows.map((row) => fields.map((field) => row[field]))]
    .map((row) => row.map(cell).join(","))
    .join("\r\n");
}
export async function fetchReport(signal: AbortSignal): Promise<Report> {
  const response = await fetch("/api/report", { signal, cache: "no-store" });
  if (!response.ok)
    throw new Error(
      `Unable to load results (HTTP ${response.status}). Check the dashboard server and try again.`,
    );
  const doc = await response.json();
  if (
    !doc ||
    ![
      "rows",
      "outcomes",
      "models",
      "datasets",
      "categories",
      "questions",
    ].every((key) => Array.isArray(doc[key])) ||
    typeof doc.generated !== "string"
  ) {
    throw new Error("The server returned an invalid benchmark report.");
  }
  return doc as Report;
}
