# Konkour Bench

A Go benchmark for four-option Iranian university entrance exam questions, using OpenRouter. Each question is evaluated independently at `minimal`, `medium`, and `max` reasoning effort. Results are cached locally and scored into offline HTML, CSV, JSON, and SVG reports.

## Quick start

Requires Go 1.27 or newer. SQLite is embedded through a pure-Go driver; no database server is needed.

```sh
go build -o bin/konkour-bench ./cmd/konkour-bench
./bin/konkour-bench validate
./bin/konkour-bench models
```

Set `OPENROUTER_API_KEY` in the environment. If you keep it in a trusted local `.env` file, load it into your shell first:

```sh
set -a
. ./.env
set +a
```

The application does not automatically read `.env`. Keys are never included in stored request bodies, logs, or reports.

Run a small benchmark and generate its report:

```sh
./bin/konkour-bench run \
  --model openai/gpt-6-luna \
  --question sample:2,sample:9
./bin/konkour-bench report --run latest --out reports/smoke
```

Open `reports/smoke/index.html` in a browser. It works offline, with filters, sortable leaderboards, accuracy bars, category heatmaps, effort comparisons, and cost/token Pareto charts. The adjacent CSV, JSON, and SVG files can be shared or analyzed separately. Full transcripts remain in the database.

Run all configured models and questions:

```sh
./bin/konkour-bench run
./bin/konkour-bench report
```

Repeating a run reuses completed generations, including malformed answers and refusals. The run summary reports new attempts, cache hits, and new spending. There is no generation token cap or default generation deadline. Use `--timeout 1h` if you want a wall-clock limit. `--concurrency` defaults to four, with at most one active question per model.

## Add models or questions

Edit `benchmark.json` to add an OpenRouter model ID. `models` shows currently advertised image and reasoning capabilities; `run --model provider/model` benchmarks just that model. Reports combine compatible results from different runs.

The configuration contains dataset entries with stable IDs and paths to question and answer files. Paths are relative to the configuration file; image paths are relative to the question file. Add another entry for each booklet so question numbers can repeat across booklets. CLI question selectors use `dataset-id:number`.

Keep the supplied JSON structure:

- Question records have `number`, `category`, `question`, four `options`, and an optional `image` path.
- Options have a unique `number` from 1 to 4, optional `text`, and an optional `image` path. An image-only option can have `"text": null`.
- Answer records have `number` and `correct_option` from 1 to 4.

Preserve the original Persian, English, or Arabic text and inline LaTeX. JSON backslashes must be escaped. PNG, JPEG, and GIF images are validated and sent unchanged. Missing or corrupt images stop validation; they are never silently removed. Extra answer keys are allowed with a warning so partial booklets work. Every included question needs a key. Missing categories are grouped as `Uncategorized`.

The sample has 15 questions, including two with images, and 24 answer keys. The nine extra keys are intentionally ignored. The sample is a fixture for exercising the benchmark, not a representative evaluation dataset.

After adding questions, rerun the benchmark; unchanged requests remain cached. After correcting keys or categories, regenerate the report without inference:

```sh
./bin/konkour-bench validate
./bin/konkour-bench report --out reports/rescored
```

Without `--run`, reports use the current dataset/key snapshot and compatible recorded requests. Newly added questions remain uncompleted until evaluated. `report --run RUN_ID` reproduces a saved run’s original dataset, answer keys, and question/model selection. Every report records a scoring revision and protocol fingerprint.

## Evaluation rules

Every request uses the same concise Persian instruction, original question/options, and labeled images. Categories, exam names, filenames, and answer keys are excluded. The structured response must be exactly an object such as `{"answer": 3}` with one integer from 1 to 4. No prose extraction, JSON repair, tools, or model fallback is used.

All model requests carry `:exacto` and require support for the supplied parameters. Reasoning is enabled; temperature, top-p, and seed are left at provider defaults. The client never sends `max_tokens`, `max_completion_tokens`, or `reasoning.max_tokens`. Provider limits and internal effort translation can still apply.

Effort compatibility is restricted to these bands:

| Requested tier | Allowed values, in order |
| --- | --- |
| `minimal` | `minimal`, `low` |
| `medium` | `medium` |
| `max` | `max`, `xhigh`, `high` |

Catalog capabilities select a supported value. When the catalog accepts all gateway values, the requested value is tried first; only an explicit effort rejection permits an alternate value in its band. Every sent value is recorded. A model without selectable reasoning, or without a value in the requested band, appears as unsupported. OpenRouter can translate values internally, so these tiers are request settings rather than measured internal thinking budgets.

Leaderboards use three fixed cohorts:

- **Text-only:** comparable across text-only and multimodal models.
- **Full:** all selected questions; models that skip image questions are ineligible.
- **Image-only:** only questions containing an image in the question or an option.

Every question has equal weight. Incorrect, invalid, refused, and truncated answers count as wrong. Unresolved API/network failures make a row incomplete and unranked. Filtering by category or dataset recomputes the same rules on that subset. A complete subset can rank even when another subset remains incomplete.

Costs use actual returned account charges. Token counts use the model’s native tokenizer; reasoning tokens are already part of completion tokens. Missing values remain unknown. Cached generations retain their original inference costs, while the separate spending ledger tracks new attempts and retry overhead. If accounting is missing and a generation ID exists, retrieve it without repeating inference:

```sh
./bin/konkour-bench reconcile
./bin/konkour-bench report
```

Pareto comparisons maximize accuracy and minimize cost or total tokens within the same cohort, dataset, and category. Three-objective membership considers all three together. Ties stay on the frontier; incomplete results or missing relevant accounting are excluded. Model/effort display filters hide rows but do not change frontier membership.

## Inspect and resume

Run snapshots, request bodies, responses, attempts, and accounting are stored in `data/bench.db`. Request hashes include the API destination, model/routing, prompt, schema, sent effort, and actual image bytes. Internal grading metadata and credentials do not affect the request hash.

```sh
./bin/konkour-bench inspect --hash REQUEST_HASH
```

`run` retries explicit rate-limit and transient server errors up to three times per request variant, respecting `Retry-After`. Authentication/account errors stop the run. Ctrl-C preserves completed work.

A connection loss after dispatch or a process crash can leave an outcome unknown. Such requests are not automatically resent, because they might already have generated a billed answer. Review the stored attempts first, then explicitly opt in if needed:

```sh
./bin/konkour-bench run --retry-unknown --model provider/model
```

This can incur duplicate charges. A request still owned by a live process cannot be reclaimed. There is no automatic resampling of completed answers. Use a separate `--db` for an intentionally independent experiment, and keep its report separate.

Useful flags include `--config`, `--db`, `--effort minimal,medium,max`, `--question`, and `--log-json`. Commands write summaries to stdout and operational logs to stderr. Use `COMMAND --help` for the full flag list. Keep the database and `.env` private; both are ignored by Git.

## Development

```sh
make build
make check
```

Checks include Go formatting, `go vet`, race-enabled tests, builds, and pinned Oxlint for the report JavaScript. Node/npm is needed only for the JavaScript development check; the built CLI and generated reports need no Node runtime. Tests use local mock HTTP servers and never call paid APIs. No browser automation is required for the automated checks.

API references: [structured outputs](https://openrouter.ai/docs/guides/features/structured-outputs), [models](https://openrouter.ai/docs/guides/overview/models), [image inputs](https://openrouter.ai/docs/guides/overview/multimodal/image-understanding), [reasoning](https://openrouter.ai/docs/guides/best-practices/reasoning-tokens), [Exacto](https://openrouter.ai/docs/guides/routing/model-variants/exacto), and [usage accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting).
