import { useState } from "react";
import { modelName, money, number, percent, tierColors, tiers } from "./data";
import type { Row } from "./types";

export function ParetoChart({ rows }: { rows: Row[] }) {
  const [metric, setMetric] = useState<"cost" | "tokens">("cost");
  const [active, setActive] = useState<Row | null>(null);
  const key =
    metric === "cost" ? "cost_per_question_usd" : "tokens_per_question";
  const frontier = metric === "cost" ? "cost_frontier" : "token_frontier";
  const points = rows.filter(
    (r) => r.status === "complete" && r.accuracy != null && r[key] != null,
  );
  const max = Math.max(0, ...points.map((r) => r[key]!)) * 1.12 || 1;
  const x = (v: number) => 58 + (v / max) * 490;
  const y = (v: number) => 254 - v * 210;
  const line = points
    .filter((r) => r[frontier])
    .sort((a, b) => a[key]! - b[key]!);
  return (
    <section className="panel chart-panel">
      <div className="panel-heading">
        <div>
          <h2>Performance & efficiency</h2>
          <p>Higher accuracy. Fewer resources.</p>
        </div>
        <div className="segmented small" aria-label="Chart resource">
          {(["cost", "tokens"] as const).map((m) => (
            <button
              key={m}
              aria-pressed={m === metric}
              onClick={() => {
                setMetric(m);
                setActive(null);
              }}
            >
              {m === "cost" ? "Cost" : "Tokens"}
            </button>
          ))}
        </div>
      </div>
      {points.length ? (
        <svg
          viewBox="0 0 580 305"
          role="img"
          aria-label={`Accuracy versus ${metric} per question. Dashed line marks the Pareto frontier.`}
        >
          {[0, 0.25, 0.5, 0.75, 1].map((v) => (
            <g key={v}>
              <line
                x1="58"
                x2="550"
                y1={y(v)}
                y2={y(v)}
                className="grid-line"
              />
              <text x="46" y={y(v) + 4} textAnchor="end">
                {v * 100}%
              </text>
              <text x={x(v * max)} y="277" textAnchor="middle">
                {metric === "cost" ? money(v * max, 3) : number(v * max)}
              </text>
            </g>
          ))}
          <text x="58" y="20" className="axis-label">
            Accuracy
          </text>
          <text x="305" y="302" textAnchor="middle">
            {metric === "cost"
              ? "USD per question"
              : "Native tokens per question"}
          </text>
          {line.length > 1 && (
            <polyline
              points={line
                .map((r) => `${x(r[key]!)},${y(r.accuracy!)}`)
                .join(" ")}
              className="frontier-line"
            />
          )}
          {points.map((r) => (
            <circle
              key={r.model + r.tier}
              cx={x(r[key]!)}
              cy={y(r.accuracy!)}
              r={r[frontier] ? 6 : 4.5}
              fill={tierColors[r.tier]}
              className={
                r[frontier] ? "plot-point frontier-point" : "plot-point"
              }
              tabIndex={0}
              aria-label={`${r.model}, ${r.tier}, accuracy ${percent(r.accuracy)}, ${metric} ${r[key]}`}
              onMouseEnter={() => setActive(r)}
              onFocus={() => setActive(r)}
            >
              <title>{`${r.model} · ${r.tier}: ${percent(r.accuracy)}, ${metric === "cost" ? money(r[key], 4) : number(r[key])}`}</title>
            </circle>
          ))}
        </svg>
      ) : (
        <div className="chart-empty">
          <span className="empty-symbol">↗</span>
          <strong>Waiting for complete comparisons</strong>
          <p>
            Points appear when a model finishes this cohort and its accounting
            is available.
          </p>
        </div>
      )}
      <div className="chart-legend">
        {tiers.map((t) => (
          <span key={t}>
            <i style={{ background: tierColors[t] }} />
            {t}
          </span>
        ))}
        <span className="frontier-legend">┄ Pareto frontier</span>
      </div>
      <p className="chart-caption" aria-live="polite">
        {active && points.includes(active)
          ? `${modelName(active.model)} · ${active.tier} · ${percent(active.accuracy)} · ${metric === "cost" ? money(active[key], 4) : number(active[key])} / question`
          : "Focus or hover over a point to inspect a result."}
      </p>
    </section>
  );
}

export function EffortChart({ rows }: { rows: Row[] }) {
  const models = [
    ...new Set(rows.filter((r) => r.accuracy != null).map((r) => r.model)),
  ]
    .sort(
      (a, b) =>
        Math.max(
          ...rows.filter((r) => r.model === b).map((r) => r.accuracy ?? -1),
        ) -
        Math.max(
          ...rows.filter((r) => r.model === a).map((r) => r.accuracy ?? -1),
        ),
    )
    .slice(0, 6);
  return (
    <section className="panel chart-panel">
      <div className="panel-heading">
        <div>
          <h2>Does more thinking help?</h2>
          <p>Accuracy by requested reasoning effort</p>
        </div>
        <span className="muted-label">TOP 6 MODELS</span>
      </div>
      {models.length ? (
        <div className="effort-chart">
          {models.map((model) => (
            <div className="effort-model" key={model}>
              <span title={model}>{modelName(model)}</span>
              <div>
                {tiers.map((tier) => {
                  const r = rows.find(
                    (row) => row.model === model && row.tier === tier,
                  );
                  return (
                    <div className="effort-bar-row" key={tier}>
                      <span className="sr-only">{tier}</span>
                      <div className="effort-track">
                        <div
                          style={{
                            width: `${(r?.accuracy ?? 0) * 100}%`,
                            background: tierColors[tier],
                          }}
                        />
                      </div>
                      <small>{percent(r?.accuracy ?? null)}</small>
                    </div>
                  );
                })}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="chart-empty">
          <span className="empty-symbol">≋</span>
          <strong>No complete scores yet</strong>
          <p>
            Completed model results will show how accuracy changes with
            reasoning effort.
          </p>
        </div>
      )}
      <div className="chart-legend">
        {tiers.map((t) => (
          <span key={t}>
            <i style={{ background: tierColors[t] }} />
            {t}
          </span>
        ))}
      </div>
    </section>
  );
}
