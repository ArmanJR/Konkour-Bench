export type Tier = "minimal" | "medium" | "max";
export type Cohort = "text-only" | "full" | "image-only";
export interface Row {
  model: string;
  tier: Tier;
  cohort: Cohort;
  dataset: string;
  category: string;
  status: string;
  total: number;
  complete: number;
  correct: number;
  incorrect: number;
  unsupported: number;
  skipped: number;
  failed: number;
  accuracy: number | null;
  cost_usd: number | null;
  cost_per_question_usd: number | null;
  total_tokens: number | null;
  tokens_per_question: number | null;
  prompt_tokens: number | null;
  completion_tokens: number | null;
  reasoning_tokens: number | null;
  cached_input_tokens: number | null;
  sent_efforts: string;
  cost_frontier: boolean;
  token_frontier: boolean;
  pareto_3d: boolean;
}
export interface Outcome {
  question: string;
  model: string;
  tier: Tier;
  status: string;
  dataset: string;
  category: string;
  has_image: boolean;
  expected: number;
  answer: number;
  correct: boolean;
  cached: boolean;
  sent_effort?: string;
  note?: string;
  invalid?: string;
  provider?: string;
  request_hash?: string;
  run_id?: string;
  usage: {
    cost: number | null;
    total_tokens: number | null;
    completion_tokens_details?: { reasoning_tokens: number | null };
  };
}
export interface Question {
  id: string;
  number: number;
  text: string;
  image?: string;
  options: { number: number; text: string; image?: string }[];
}
export interface Spending {
  run_id: string;
  attempts: number;
  known_cost_usd: number;
  known_retry_cost_usd: number;
  attempts_with_unknown_cost: number;
}
export interface Report {
  backend?: "openrouter" | "claude" | "mixed";
  cost_basis?: string;
  generated: string;
  scoring_revision: string;
  protocol: string;
  run_id?: string;
  rows: Row[];
  outcomes: Outcome[];
  datasets: string[];
  categories: string[];
  models: string[];
  questions: Question[];
  run_spending: Spending[] | null;
}
export interface Filters {
  cohort: Cohort;
  dataset: string;
  category: string;
  tier: string;
  search: string;
}
