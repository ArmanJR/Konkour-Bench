import { lazy, Suspense, useEffect, useState } from "react";
import { EffortChart, ParetoChart } from "./Charts";
import { Leaderboard } from "./Leaderboard";
import {
  datasetName,
  initialFilters,
  matchesOutcome,
  modelName,
  money,
  number,
  percent,
  progress,
  selectRows,
  sortRows,
  tiers,
} from "./data";
import { useReport } from "./useReport";
import type { Cohort, Filters, Report, Row } from "./types";

const Questions = lazy(() =>
  import("./Questions").then((module) => ({ default: module.Questions })),
);

const pages = [
  {
    id: "overview",
    label: "Overview",
    icon: "◫",
    title: "Benchmark overview",
    description: "A clearer view of how language models perform on Konkour.",
  },
  {
    id: "leaderboard",
    label: "Leaderboard",
    icon: "≡",
    title: "Model leaderboard",
    description:
      "Compare accuracy, inference cost, and token use on the same questions.",
  },
  {
    id: "questions",
    label: "Question explorer",
    icon: "▤",
    title: "Question explorer",
    description:
      "Inspect the original questions, answer keys, and model responses.",
  },
  {
    id: "spending",
    label: "Run spending",
    icon: "◷",
    title: "Run spending",
    description:
      "Recorded charges for new requests across your saved benchmark runs.",
  },
];

function FiltersBar({
  data,
  filters,
  update,
}: {
  data: Report;
  filters: Filters;
  update: (patch: Partial<Filters>) => void;
}) {
  return (
    <div className="filters">
      <div className="filter-row">
        <div className="segmented" aria-label="Question cohort">
          {(
            [
              ["text-only", "Text only"],
              ["full", "Full dataset"],
              ["image-only", "With images"],
            ] as const
          ).map(([value, label]) => (
            <button
              key={value}
              aria-pressed={filters.cohort === value}
              onClick={() => update({ cohort: value })}
            >
              {label}
            </button>
          ))}
        </div>
        <span className="filter-divider" />
        <label>
          <span>Dataset</span>
          <select
            aria-label="Dataset"
            value={filters.dataset}
            onChange={(e) => update({ dataset: e.target.value })}
          >
            <option value="">All datasets</option>
            {data.datasets.map((d) => (
              <option key={d} value={d}>
                {datasetName(d)}
              </option>
            ))}
          </select>
        </label>
        <label>
          <span>Subject</span>
          <select
            aria-label="Subject"
            value={filters.category}
            onChange={(e) => update({ category: e.target.value })}
          >
            <option value="">All subjects</option>
            {data.categories.map((c) => (
              <option dir="auto" key={c}>
                {c}
              </option>
            ))}
          </select>
        </label>
        <label>
          <span>Reasoning</span>
          <select
            aria-label="Reasoning effort"
            value={filters.tier}
            onChange={(e) => update({ tier: e.target.value })}
          >
            <option value="">All efforts</option>
            {tiers.map((t) => (
              <option key={t}>{t}</option>
            ))}
          </select>
        </label>
        <label className="search">
          <span aria-hidden="true">⌕</span>
          <input
            aria-label="Search models"
            placeholder="Search models…"
            value={filters.search}
            onChange={(e) => update({ search: e.target.value })}
          />
        </label>
      </div>
    </div>
  );
}

function Stats({ data, filters }: { data: Report; filters: Filters }) {
  const outcomes = data.outcomes.filter((o) => matchesOutcome(o, filters));
  const p = progress(outcomes);
  const spend = data.run_spending ?? [];
  const unknown = spend.reduce((n, r) => n + r.attempts_with_unknown_cost, 0);
  const spendingScope = data.run_id ? "Selected run" : "All saved runs";
  const stats = [
    {
      label: "Questions",
      value: number(new Set(outcomes.map((o) => o.question)).size),
      detail:
        filters.cohort === "text-only"
          ? "Text-only comparison cohort"
          : filters.cohort === "image-only"
            ? "Questions with visual content"
            : "Every question in the selection",
      icon: "▤",
    },
    {
      label: "Models",
      value: number(new Set(outcomes.map((o) => o.model)).size),
      detail: filters.tier
        ? `${filters.tier} reasoning effort`
        : "Across three reasoning efforts",
      icon: "◎",
    },
    {
      label: "Evaluations completed",
      value: number(p.complete),
      detail: `${number(p.total)} planned · ${number(p.excluded)} unsupported or skipped`,
      icon: "✓",
    },
    {
      label: "Recorded spend",
      value: money(spend.reduce((n, r) => n + r.known_cost_usd, 0)),
      detail: unknown
        ? `${spendingScope} · ${unknown} charges unknown`
        : `${spendingScope} · new requests only`,
      icon: "$",
    },
  ];
  return (
    <div className="stats">
      {stats.map((stat, i) => (
        <section className={`stat-card stat-${i}`} key={stat.label}>
          <div>
            <span>{stat.label}</span>
            <i aria-hidden="true">{stat.icon}</i>
          </div>
          <strong>{stat.value}</strong>
          <p>{stat.detail}</p>
          {i === 2 && (
            <div
              className="stat-progress"
              aria-label={`${p.complete} of ${p.eligible} eligible evaluations complete`}
            >
              <span
                style={{
                  width: `${p.eligible ? (p.complete / p.eligible) * 100 : 0}%`,
                }}
              />
            </div>
          )}
        </section>
      ))}
    </div>
  );
}

function Subjects({
  data,
  filters,
  rows,
}: {
  data: Report;
  filters: Filters;
  rows: Row[];
}) {
  const models = sortRows(rows, "accuracy", false).slice(0, 12);
  return (
    <section className="panel">
      <div className="panel-heading">
        <div>
          <h2>Across the subjects</h2>
          <p>
            Category accuracy for the first 12 results in the current selection
          </p>
        </div>
        <span className="muted-label">ACCURACY</span>
      </div>
      <div className="table-scroll">
        <table className="heatmap">
          <thead>
            <tr>
              <th>Model / effort</th>
              {data.categories.map((c) => (
                <th dir="auto" key={c}>
                  {c}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {models.map((r) => (
              <tr key={r.model + r.tier}>
                <td>
                  <span className="model-name">{modelName(r.model)}</span>{" "}
                  <span className={`effort effort-${r.tier}`}>{r.tier}</span>
                </td>
                {data.categories.map((category) => {
                  const row = data.rows.find(
                    (v) =>
                      v.model === r.model &&
                      v.tier === r.tier &&
                      v.cohort === filters.cohort &&
                      v.dataset === filters.dataset &&
                      v.category === category,
                  );
                  return (
                    <td key={category}>
                      <span
                        className="heat-cell"
                        title={row?.status ?? "No questions"}
                        style={
                          row?.accuracy != null
                            ? {
                                background: `rgba(53,124,101,${0.06 + row.accuracy * 0.2})`,
                                color: "#235440",
                              }
                            : {}
                        }
                      >
                        {percent(row?.accuracy ?? null)}
                      </span>
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      <div className="panel-footer">
        Each cell uses its subject’s full denominator. A dash means the subject
        result is incomplete or unavailable.
      </div>
    </section>
  );
}

function Spending({ data }: { data: Report }) {
  const runs = [...(data.run_spending ?? [])].reverse();
  return (
    <>
      <section className="panel">
        <div className="panel-heading">
          <div>
            <h2>
              Spending ledger <span className="count">{runs.length}</span>
            </h2>
            <p>
              {data.run_id ? "Selected run" : "All saved runs"} · cache hits add
              no new charges
            </p>
          </div>
        </div>
        <div className="table-scroll">
          <table>
            <thead>
              <tr>
                <th>Run</th>
                <th>New attempts</th>
                <th>Recorded cost</th>
                <th>Included retry cost</th>
                <th>Unknown charges</th>
              </tr>
            </thead>
            <tbody>
              {runs.map((r) => (
                <tr key={r.run_id}>
                  <td>
                    <code>{r.run_id}</code>
                  </td>
                  <td>{number(r.attempts)}</td>
                  <td>{money(r.known_cost_usd, 4)}</td>
                  <td>{money(r.known_retry_cost_usd, 4)}</td>
                  <td>
                    {r.attempts_with_unknown_cost ? (
                      <span className="status status-incomplete">
                        {r.attempts_with_unknown_cost} unknown
                      </span>
                    ) : (
                      "None"
                    )}
                  </td>
                </tr>
              ))}
              {!runs.length && (
                <tr>
                  <td colSpan={5} className="empty-cell">
                    No requests have been recorded yet.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        <div className="panel-footer">
          Recorded cost is a lower bound when charges are unknown. Leaderboard
          costs describe retained answers; this ledger includes retries.
        </div>
      </section>
      <section className="panel provenance">
        <h2>Result provenance</h2>
        <dl>
          <dt>Scoring revision</dt>
          <dd>{data.scoring_revision}</dd>
          <dt>Evaluation protocol</dt>
          <dd>{data.protocol}</dd>
          <dt>Results</dt>
          <dd>
            {data.run_id || "Current dataset and compatible recorded history"}
          </dd>
        </dl>
      </section>
    </>
  );
}

function readPage() {
  return pages.find((p) => p.id === window.location.hash.slice(1)) ?? pages[0];
}

export default function App() {
  const [page, setPage] = useState(readPage);
  const [filters, setFilters] = useState<Filters>(initialFilters);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const { data, error, loading, refresh } = useReport(autoRefresh);
  useEffect(() => {
    const update = () => {
      if (pages.some((p) => p.id === window.location.hash.slice(1)))
        setPage(readPage());
    };
    window.addEventListener("hashchange", update);
    return () => window.removeEventListener("hashchange", update);
  }, []);
  const update = (patch: Partial<Filters>) =>
    setFilters((f) => ({ ...f, ...patch }));
  const rows = data ? selectRows(data.rows, filters) : [];
  const cohortLabels: Record<Cohort, string> = {
    "text-only": "Text only",
    full: "Full dataset",
    "image-only": "With images",
  };
  return (
    <div className="app-shell">
      <a className="skip-link" href="#main-content">
        Skip to content
      </a>
      <aside className="sidebar">
        <a className="brand" href="#overview">
          <span className="brand-mark">
            K<span>∷</span>
          </span>
          <span>
            Konkour <strong>Bench</strong>
            <small>LANGUAGE MODEL EVALUATION</small>
          </span>
        </a>
        <div className="nav-label">WORKSPACE</div>
        <nav aria-label="Main navigation">
          {pages.map((p) => (
            <a
              key={p.id}
              href={`#${p.id}`}
              aria-current={page.id === p.id ? "page" : undefined}
            >
              <span aria-hidden="true">{p.icon}</span>
              {p.label}
              {p.id === "leaderboard" && data && (
                <small>{data.models.length}</small>
              )}
            </a>
          ))}
        </nav>
        <div className="sidebar-note">
          <div className="exam-icon">ک</div>
          <strong>The Konkour benchmark</strong>
          <p>
            Original exam questions.
            <br />A common test for every model.
          </p>
          <span>Persian · English · Arabic</span>
        </div>
        <div className="sidebar-bottom">
          <span className="local-dot" />
          Local workspace<span className="version">v0.1</span>
        </div>
      </aside>
      <div className="workspace">
        <header className="topbar">
          <div>
            <span className="breadcrumb">Workspace</span>
            <span className="breadcrumb-slash">/</span>
            <strong>{page.label}</strong>
          </div>
          <div className="topbar-right">
            <span className="routing-tag">Exacto routing</span>
            <span className="avatar" aria-label="Konkour Bench">
              KB
            </span>
          </div>
        </header>
        <main id="main-content" tabIndex={-1}>
          <div className="page-heading">
            <div>
              <div className="eyebrow">KONKOUR BENCH</div>
              <h1>{page.title}</h1>
              <p>{page.description}</p>
            </div>
            <div className="refresh-controls">
              <label className="auto-refresh">
                <input
                  type="checkbox"
                  checked={autoRefresh}
                  onChange={(e) => setAutoRefresh(e.target.checked)}
                />
                Auto-refresh
              </label>
              <button className="button" disabled={loading} onClick={refresh}>
                <span aria-hidden="true">↻</span>
                {loading ? "Refreshing…" : "Refresh"}
              </button>
            </div>
          </div>
          {error && (
            <div className="error-banner" role="alert">
              <strong>
                {data
                  ? "Refresh failed. Showing the last loaded results."
                  : "Results could not be loaded."}
              </strong>
              <p>{error}</p>
              <button className="text-button" onClick={refresh}>
                Try again →
              </button>
            </div>
          )}
          {!data && !error && (
            <div className="loading-state" role="status">
              <div className="loading-bars">
                <i />
                <i />
                <i />
              </div>
              <h2>Loading your benchmark</h2>
              <p>Reading saved results and scoring the current dataset…</p>
            </div>
          )}
          {data && (
            <>
              <div className="dataset-summary">
                <span className="dataset-label">
                  ▤{" "}
                  {data.datasets.length === 1
                    ? datasetName(data.datasets[0])
                    : `${data.datasets.length} datasets`}
                </span>
                <span>
                  Updated{" "}
                  {new Date(data.generated).toLocaleTimeString([], {
                    hour: "2-digit",
                    minute: "2-digit",
                    second: "2-digit",
                  })}
                  {autoRefresh ? " · every 30 seconds" : ""}
                </span>
              </div>
              {page.id !== "spending" && (
                <FiltersBar data={data} filters={filters} update={update} />
              )}
              <Stats
                data={data}
                filters={page.id === "spending" ? initialFilters : filters}
              />
              {page.id === "overview" && (
                <>
                  <div className="charts-grid">
                    <ParetoChart rows={rows} />
                    <EffortChart
                      rows={selectRows(data.rows, { ...filters, tier: "" })}
                    />
                  </div>
                  <Leaderboard
                    rows={rows}
                    limit={8}
                    onViewAll={() => {
                      window.location.hash = "leaderboard";
                    }}
                  />
                  <Subjects data={data} filters={filters} rows={rows} />
                </>
              )}
              {page.id === "leaderboard" && (
                <>
                  <Leaderboard rows={rows} />
                  <div className="charts-grid">
                    <ParetoChart rows={rows} />
                    <EffortChart
                      rows={selectRows(data.rows, { ...filters, tier: "" })}
                    />
                  </div>
                </>
              )}
              {page.id === "questions" && (
                <Suspense
                  fallback={
                    <div className="panel empty-state" role="status">
                      Loading question explorer…
                    </div>
                  }
                >
                  <Questions
                    questions={data.questions}
                    outcomes={data.outcomes.filter((o) =>
                      matchesOutcome(o, filters),
                    )}
                  />
                </Suspense>
              )}
              {page.id === "spending" && <Spending data={data} />}
              <footer className="page-footer">
                <span>
                  Konkour Bench <span className="footer-dot">·</span>{" "}
                  {page.id === "spending"
                    ? "Recorded spending"
                    : cohortLabels[filters.cohort]}
                </span>
                <span>
                  Equal question weights · One answer per request · No
                  generation token cap
                </span>
              </footer>
            </>
          )}
        </main>
      </div>
    </div>
  );
}
