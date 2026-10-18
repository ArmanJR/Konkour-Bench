import { useState } from "react";
import katex from "katex";
import { modelName, money, number } from "./data";
import { Status } from "./Leaderboard";
import type { Outcome, Question } from "./types";

export function MathText({ text }: { text: string }) {
  return (
    <>
      {text.split(/(\$[^$]+\$)/g).map((part, i) =>
        part.startsWith("$") && part.endsWith("$") ? (
          <span
            key={i}
            dir="ltr"
            className="math"
            dangerouslySetInnerHTML={{
              __html: katex.renderToString(part.slice(1, -1), {
                throwOnError: false,
                trust: false,
                strict: "ignore",
                maxSize: 20,
                maxExpand: 1000,
              }),
            }}
          />
        ) : (
          part
        ),
      )}
    </>
  );
}

export function Questions({
  questions,
  outcomes,
}: {
  questions: Question[];
  outcomes: Outcome[];
}) {
  const [selected, setSelected] = useState("");
  const [search, setSearch] = useState("");
  const [resultFilter, setResultFilter] = useState("all");
  const included = new Set(outcomes.map((o) => o.question));
  const list = questions
    .filter(
      (q) =>
        included.has(q.id) &&
        (String(q.number).includes(search) ||
          q.text.toLowerCase().includes(search.toLowerCase())),
    )
    .sort(
      (a, b) =>
        a.id.split(":")[0].localeCompare(b.id.split(":")[0]) ||
        a.number - b.number,
    );
  const question = list.find((q) => q.id === selected) ?? list[0];
  const results = outcomes.filter((o) => o.question === question?.id);
  const visible = results.filter(
    (o) =>
      resultFilter === "all" ||
      (resultFilter === "correct"
        ? o.status === "complete" && o.correct
        : resultFilter === "wrong"
          ? o.status === "complete" && !o.correct
          : o.status !== "complete"),
  );
  return (
    <section className="question-layout">
      <div className="panel question-list">
        <div className="panel-heading">
          <h2>
            Questions <span className="count">{list.length}</span>
          </h2>
        </div>
        <label className="search question-search">
          <span aria-hidden="true">⌕</span>
          <input
            placeholder="Find a question…"
            aria-label="Find a question"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </label>
        <div className="question-choices">
          {list.map((q) => {
            const o = outcomes.find((r) => r.question === q.id);
            return (
              <button
                key={q.id}
                aria-pressed={question?.id === q.id}
                onClick={() => setSelected(q.id)}
              >
                <span>Q{String(q.number).padStart(2, "0")}</span>
                <div>
                  <strong dir="auto">{o?.category}</strong>
                  <small>{q.id.split(":")[0]}</small>
                </div>
                {o?.has_image && (
                  <span
                    title="Includes an image"
                    aria-label="Includes an image"
                  >
                    ▧
                  </span>
                )}
              </button>
            );
          })}
          {!list.length && <p className="empty-cell">No matching questions.</p>}
        </div>
      </div>
      {question ? (
        <div className="question-content">
          <section className="panel question-paper">
            <div className="panel-heading">
              <div>
                <span className="eyebrow">QUESTION {question.number}</span>
                <h2 dir="auto">{results[0]?.category}</h2>
              </div>
              <span className="status status-complete">
                Key: option {results[0]?.expected}
              </span>
            </div>
            <div className="question-original" dir="auto">
              <p>
                <MathText text={question.text} />
              </p>
              {question.image && (
                <img
                  src={question.image}
                  alt={`Diagram for question ${question.number}`}
                />
              )}
              <ol className="options">
                {question.options.map((option) => (
                  <li
                    key={option.number}
                    className={
                      option.number === results[0]?.expected
                        ? "correct-option"
                        : ""
                    }
                  >
                    <span className="option-number">{option.number}</span>
                    <div>
                      <MathText text={option.text} />
                      {option.image && (
                        <img
                          src={option.image}
                          alt={`Option ${option.number}`}
                        />
                      )}
                    </div>
                    {option.number === results[0]?.expected && (
                      <span className="answer-check" aria-label="Answer key">
                        ✓
                      </span>
                    )}
                  </li>
                ))}
              </ol>
            </div>
          </section>
          <section className="panel">
            <div className="panel-heading">
              <div>
                <h2>Model answers</h2>
                <p>
                  {results.filter((o) => o.status === "complete").length}{" "}
                  evaluated ·{" "}
                  {
                    results.filter((o) => o.status === "complete" && o.correct)
                      .length
                  }{" "}
                  correct
                </p>
              </div>
              <label className="inline-select">
                <span className="sr-only">Answer filter</span>
                <select
                  value={resultFilter}
                  onChange={(e) => setResultFilter(e.target.value)}
                >
                  <option value="all">All outcomes</option>
                  <option value="correct">Correct</option>
                  <option value="wrong">Incorrect</option>
                  <option value="pending">Not completed</option>
                </select>
              </label>
            </div>
            <div className="table-scroll">
              <table>
                <thead>
                  <tr>
                    <th>Model</th>
                    <th>Effort</th>
                    <th>Answer</th>
                    <th>Result</th>
                    <th>Cost</th>
                    <th>Tokens</th>
                  </tr>
                </thead>
                <tbody>
                  {visible.map((o) => (
                    <tr key={o.model + o.tier}>
                      <td>
                        <strong className="model-name">
                          {modelName(o.model)}
                        </strong>
                        <small className="cell-subtitle">
                          {o.provider || o.model.split("/")[0]}
                        </small>
                      </td>
                      <td>
                        <span
                          className={`effort effort-${o.tier}`}
                          title={`Sent: ${o.sent_effort || "none"}`}
                        >
                          {o.tier}
                        </span>
                      </td>
                      <td>{o.answer || "—"}</td>
                      <td>
                        {o.status === "complete" ? (
                          <span
                            className={`status ${o.correct ? "status-complete" : "status-wrong"}`}
                          >
                            {o.invalid
                              ? "Invalid"
                              : o.correct
                                ? "Correct"
                                : "Incorrect"}
                          </span>
                        ) : (
                          <Status status={o.status} />
                        )}
                      </td>
                      <td>{money(o.usage.cost, 5)}</td>
                      <td>{number(o.usage.total_tokens)}</td>
                    </tr>
                  ))}
                  {!visible.length && (
                    <tr>
                      <td colSpan={6} className="empty-cell">
                        No answers match this filter.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </section>
        </div>
      ) : (
        <div className="panel empty-state">
          <h2>No questions in this selection</h2>
          <p>Try another cohort, subject, or search.</p>
        </div>
      )}
    </section>
  );
}
