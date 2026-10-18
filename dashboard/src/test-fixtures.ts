import type { Outcome, Report, Row } from "./types";

export function row(overrides: Partial<Row> = {}): Row {
  return {
    model: "test/alpha",
    tier: "medium",
    cohort: "text-only",
    dataset: "",
    category: "",
    status: "complete",
    total: 2,
    complete: 2,
    correct: 1,
    incorrect: 1,
    unsupported: 0,
    skipped: 0,
    failed: 0,
    accuracy: 0.5,
    cost_usd: 0.02,
    cost_per_question_usd: 0.01,
    total_tokens: 100,
    tokens_per_question: 50,
    prompt_tokens: 40,
    completion_tokens: 60,
    reasoning_tokens: 20,
    cached_input_tokens: 0,
    sent_efforts: "medium",
    cost_frontier: true,
    token_frontier: true,
    pareto_3d: true,
    ...overrides,
  };
}
export function outcome(overrides: Partial<Outcome> = {}): Outcome {
  return {
    question: "exam:1",
    model: "test/alpha",
    tier: "medium",
    status: "complete",
    dataset: "exam",
    category: "ریاضی",
    has_image: false,
    expected: 2,
    answer: 2,
    correct: true,
    cached: false,
    usage: { cost: 0.01, total_tokens: 50 },
    ...overrides,
  };
}
export function fixture(): Report {
  return {
    generated: "2026-10-18T12:00:00Z",
    scoring_revision: "test-revision",
    protocol: "test-protocol",
    datasets: ["exam"],
    categories: ["ریاضی"],
    models: ["test/alpha", "test/beta"],
    rows: [
      row(),
      row({
        model: "test/beta",
        status: "incomplete",
        accuracy: null,
        cost_usd: null,
        total_tokens: null,
        complete: 1,
      }),
      row({ model: "test/beta", cohort: "image-only", accuracy: 1 }),
    ],
    outcomes: [
      outcome(),
      outcome({
        question: "exam:2",
        model: "test/beta",
        has_image: true,
        answer: 1,
        correct: false,
      }),
      outcome({
        model: "test/beta",
        status: "not_run",
        answer: 0,
        correct: false,
        usage: { cost: null, total_tokens: null },
      }),
    ],
    questions: [
      {
        id: "exam:1",
        number: 1,
        text: "مقدار $x^2$؟",
        options: [1, 2, 3, 4].map((n) => ({ number: n, text: String(n) })),
      },
      {
        id: "exam:2",
        number: 2,
        text: "سؤال تصویری",
        image: "data:image/png;base64,AQI=",
        options: [1, 2, 3, 4].map((n) => ({ number: n, text: String(n) })),
      },
    ],
    run_spending: [
      {
        run_id: "run-1",
        attempts: 2,
        known_cost_usd: 0.02,
        known_retry_cost_usd: 0,
        attempts_with_unknown_cost: 1,
      },
    ],
  };
}
