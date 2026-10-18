import { Fragment, useState } from "react";
import {
  leaderboardCSV,
  modelName,
  money,
  number,
  percent,
  sortRows,
} from "./data";
import type { SortKey } from "./data";
import type { Row } from "./types";

export function Status({ status }: { status: string }) {
  const label: Record<string, string> = {
    complete: "Complete",
    incomplete: "Incomplete",
    unsupported: "Unsupported",
    ineligible: "Image skips",
    not_run: "Not evaluated",
    infra_error: "API error",
    skipped: "Image skipped",
  };
  return (
    <span className={`status status-${status}`}>{label[status] ?? status}</span>
  );
}

export function Leaderboard({
  rows,
  limit,
  onViewAll,
}: {
  rows: Row[];
  limit?: number;
  onViewAll?: () => void;
}) {
  const [sort, setSort] = useState<SortKey>("accuracy");
  const [ascending, setAscending] = useState(false);
  const [includeIncomplete, setIncludeIncomplete] = useState(true);
  const [expanded, setExpanded] = useState("");
  const filtered = rows.filter(
    (r) => includeIncomplete || r.status === "complete",
  );
  const sorted = sortRows(filtered, sort, ascending);
  const ranked = sortRows(
    rows.filter((r) => r.accuracy != null),
    "accuracy",
    false,
  );
  const visible = limit ? sorted.slice(0, limit) : sorted;
  function changeSort(key: SortKey) {
    if (sort === key) setAscending(!ascending);
    else {
      setSort(key);
      setAscending(key !== "accuracy" && key !== "complete");
    }
  }
  function exportCSV() {
    const url = URL.createObjectURL(
      new Blob(["\uFEFF", leaderboardCSV(sorted)], {
        type: "text/csv;charset=utf-8",
      }),
    );
    const link = document.createElement("a");
    link.href = url;
    link.download = "konkour-leaderboard.csv";
    link.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  }
  const header = (key: SortKey, label: string) => (
    <th
      aria-sort={
        sort === key ? (ascending ? "ascending" : "descending") : "none"
      }
    >
      <button onClick={() => changeSort(key)}>
        {label}
        <span aria-hidden="true">
          {sort === key ? (ascending ? " ↑" : " ↓") : " ↕"}
        </span>
      </button>
    </th>
  );
  return (
    <section className="panel leaderboard">
      <div className="panel-heading">
        <div>
          <h2>
            Model leaderboard <span className="count">{rows.length}</span>
          </h2>
          <p>One result per model and reasoning effort</p>
        </div>
        <div className="table-actions">
          <label className="check-label">
            <input
              type="checkbox"
              checked={includeIncomplete}
              onChange={(e) => setIncludeIncomplete(e.target.checked)}
            />
            Show incomplete
          </label>
          <button
            className="button"
            onClick={exportCSV}
            disabled={!sorted.length}
          >
            ↓ Export CSV
          </button>
        </div>
      </div>
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th>#</th>
              {header("model", "Model")}
              <th>Effort</th>
              {header("accuracy", "Accuracy")}
              {header("cost_usd", "Cost")}
              {header("total_tokens", "Tokens")}
              {header("complete", "Completed")}
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {visible.map((r) => {
              const id = r.model + r.tier;
              return (
                <Fragment key={id}>
                  <tr className={expanded === id ? "row-selected" : ""}>
                    <td className="rank">
                      {r.accuracy == null
                        ? "—"
                        : ranked.findIndex(
                            (row) => row.accuracy === r.accuracy,
                          ) + 1}
                    </td>
                    <td>
                      <button
                        className="model-button"
                        aria-label={`${r.model}, ${r.tier} details`}
                        aria-expanded={expanded === id}
                        onClick={() => setExpanded(expanded === id ? "" : id)}
                      >
                        <span
                          className={`provider-icon provider-${r.model.split("/")[0]}`}
                        >
                          {r.model.split("/")[0].slice(0, 2).toUpperCase()}
                        </span>
                        <span>
                          <strong>{modelName(r.model)}</strong>
                          <small>{r.model.split("/")[0]}</small>
                        </span>
                        <span className="expand-icon">
                          {expanded === id ? "−" : "+"}
                        </span>
                      </button>
                    </td>
                    <td>
                      <span className={`effort effort-${r.tier}`}>
                        {r.tier}
                      </span>
                    </td>
                    <td>
                      <div className="accuracy-cell">
                        <strong>{percent(r.accuracy)}</strong>
                        {r.pareto_3d && (
                          <span
                            title="On the accuracy / cost / tokens Pareto frontier"
                            className="pareto-mark"
                          >
                            ◆
                          </span>
                        )}
                        <div className="mini-track">
                          <span
                            style={{ width: `${(r.accuracy ?? 0) * 100}%` }}
                          />
                        </div>
                      </div>
                    </td>
                    <td className="numeric">{money(r.cost_usd, 3)}</td>
                    <td className="numeric">{number(r.total_tokens)}</td>
                    <td>
                      <span className="numeric">
                        {r.complete}
                        <span className="secondary"> / {r.total}</span>
                      </span>
                    </td>
                    <td>
                      <Status status={r.status} />
                    </td>
                  </tr>
                  {expanded === id && (
                    <tr className="detail-row">
                      <td colSpan={8}>
                        <div className="result-details">
                          <div>
                            <small>Correct / evaluated</small>
                            <strong>
                              {r.correct} / {r.complete}
                            </strong>
                          </div>
                          <div>
                            <small>Cost / question</small>
                            <strong>{money(r.cost_per_question_usd, 5)}</strong>
                          </div>
                          <div>
                            <small>Prompt tokens</small>
                            <strong>{number(r.prompt_tokens)}</strong>
                          </div>
                          <div>
                            <small>Completion tokens</small>
                            <strong>{number(r.completion_tokens)}</strong>
                          </div>
                          <div>
                            <small>Reasoning tokens</small>
                            <strong>{number(r.reasoning_tokens)}</strong>
                          </div>
                          <div>
                            <small>Sent effort</small>
                            <strong>{r.sent_efforts || "—"}</strong>
                          </div>
                        </div>
                        <p className="detail-note">
                          {r.model} · Reasoning tokens are included in
                          completion tokens.{" "}
                          {r.status !== "complete" &&
                            "This result is not eligible for ranking."}
                        </p>
                      </td>
                    </tr>
                  )}
                </Fragment>
              );
            })}
            {!visible.length && (
              <tr>
                <td colSpan={8} className="empty-cell">
                  No results match these filters.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
      <div className="panel-footer">
        <span>
          ◆ Pareto optimal across accuracy, cost, and tokens. Incomplete results
          are unranked.
        </span>
        {onViewAll && (
          <button className="text-button" onClick={onViewAll}>
            View all results →
          </button>
        )}
      </div>
    </section>
  );
}
