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

The default configuration evaluates the 80-question humanities 1405 booklet 1. Its source PDF, images, and answer key are in `datasets/humanities-1405/`. Six questions include images; 74 are text-only. The shared 200-question answer key includes 120 keys reserved for later booklets.

The booklet contains 20 mathematics, 30 Persian language/literature, 15 social sciences, and 15 psychology questions. Source spot-checking corrected question 15, option 2 from `2/8` to the PDF's `3/8`. The supplied answer key is unchanged; it has not been independently verified against an official key.

Run a small benchmark against the original sample and generate its report:

```sh
./bin/konkour-bench run --config benchmark.sample.json \
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

Repeating a run reuses completed generations, including malformed answers and refusals. The run summary reports new attempts, cache hits, and new spending. Generation tokens remain uncapped. Each request now defaults to a one-hour emergency deadline (`--timeout 1h`), with a warning after ten minutes (`--warn-after 10m`). Set either duration to `0` to disable it. The default `--budget-usd 50` stops new dispatch once recorded new spending reaches $50 or a potentially billed attempt has unknown cost. It is not a hard billing cap: concurrent requests can overshoot it, and provider cancellation support varies. Set a separate OpenRouter key credit limit for account protection; use `--budget-usd 0` only to explicitly disable the local spending threshold. `--concurrency` defaults to four, with at most one active question per model.

## React dashboard

Build and start the local dashboard (use Node.js 24 or newer for the frontend toolchain):

```sh
make app
```

Open http://127.0.0.1:8080. The React dashboard reads the SQLite results and refreshes every 30 seconds, including while a benchmark is running. It includes sortable model leaderboards with CSV export, cost/token Pareto charts, reasoning effort comparisons, subject accuracy, original questions with LaTeX and images, and a run spending ledger. Incomplete results stay unranked; unknown accounting stays unknown. Viewing the dashboard does not call OpenRouter or require an API key.

After the first build, start it directly:

```sh
./bin/konkour-bench serve
./bin/konkour-bench serve --run latest --addr 127.0.0.1:8081
```

The default uses current questions and compatible recorded history. `--run` selects a saved snapshot, as with `report`. The server binds to localhost by default and has no authentication; keep it local unless you provide access control. `--assets` selects the built frontend directory (default `dashboard/dist`). Offline HTML/CSV/JSON/SVG reports remain available with `report`.

For frontend development, keep `konkour-bench serve` running on port 8080 and run `npm --prefix dashboard run dev` in another terminal. Vite proxies `/api` to the Go server. Run `make dashboard-check` for TypeScript, lint, and component/data tests. `npm --prefix dashboard run format` formats the frontend source. Generated bundles and installed packages are ignored by Git; `package-lock.json` pins the frontend dependencies.

## Add models or questions

Edit `benchmark.json` to add an OpenRouter model ID. `models` shows currently advertised image and reasoning capabilities; `run --model provider/model` benchmarks just that model. Reports combine compatible results from different runs.

The configuration contains dataset entries with stable IDs and paths to question and answer files. Paths are relative to the configuration file; image paths are relative to the question file. Add another entry for each booklet so question numbers can repeat across booklets. CLI question selectors use `dataset-id:number`.

Keep the supplied JSON structure:

- Question records have `number`, `category`, `question`, four `options`, and an optional `image` path.
- Options have a unique `number` from 1 to 4, optional `text`, and an optional `image` path. An image-only option can have `"text": null`.
- Answer records have `number` and `correct_option` from 1 to 4.

Preserve the original Persian, English, or Arabic text and inline LaTeX. JSON backslashes must be escaped. PNG, JPEG, and GIF images are validated and sent unchanged. Missing or corrupt images stop validation; they are never silently removed. Extra answer keys are allowed with a warning so partial booklets work. Every included question needs a key. Missing categories are grouped as `Uncategorized`.

`benchmark.sample.json` selects the original sample. The sample has 15 questions, including two with images, and 24 answer keys. The nine extra keys are intentionally ignored. The sample is a fixture for exercising the benchmark, not a representative evaluation dataset.

After adding questions, rerun the benchmark; unchanged requests remain cached. After correcting keys or categories, regenerate the report without inference:

```sh
./bin/konkour-bench validate
./bin/konkour-bench report --out reports/rescored
```

Without `--run`, reports use the current dataset/key snapshot and compatible recorded requests. Newly added questions remain uncompleted until evaluated. `report --run RUN_ID` reproduces a saved run’s original dataset, answer keys, and question/model selection. Every report records a scoring revision and protocol fingerprint.

## Evaluation rules

Every request uses the same concise Persian instruction, original question/options, and labeled images. Categories, exam names, filenames, and answer keys are excluded. The structured response must be exactly an object such as `{"answer": 3}` with one integer from 1 to 4. No prose extraction, JSON repair, tools, or model fallback is used.

New inference requests use streaming; response-header IDs and raw SSE data are saved as they arrive. The final assembled response must pass the same strict answer validation. Heartbeat comments are not evidence of model reasoning progress. A dropped stream, malformed event, or deadline leaves the outcome unknown, even if partial text looks like a valid answer.

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

Run snapshots, request bodies, responses, attempts, and accounting are stored in `data/bench.db`. Request hashes include the API destination, model/routing, prompt, schema, sent effort, transport, and actual image bytes. Existing completed non-streaming results are reused through an explicit legacy-cache lookup; uncertain legacy requests still require an explicit retry. Run snapshots record the deadline and spending threshold. Internal grading metadata and credentials do not affect the request hash.

```sh
./bin/konkour-bench inspect --hash REQUEST_HASH
```

`run` retries explicit rate-limit and transient server errors up to three times per request variant, respecting `Retry-After`. Authentication/account errors stop the run. Ctrl-C preserves completed work.

A connection loss after dispatch or a process crash can leave an outcome unknown. Such requests are not automatically resent, because they might already have generated a billed answer. Review the stored attempts first, then explicitly opt in if needed:

```sh
./bin/konkour-bench run --retry-unknown --model provider/model
```

This can incur duplicate charges. New runners hold a kernel file lock for the database. A second runner is refused, and process death releases the lock automatically. `status` and reports recover stale pending records as unknown; legacy PID records that cannot be verified remain unverified. Elapsed time is time since dispatch, not proof of ongoing generation or billing. A request still owned by a live process cannot be reclaimed. There is no automatic resampling of completed answers. Use a separate `--db` for an intentionally independent experiment, and keep its report separate.

Check local ownership and saved OpenRouter IDs without making API calls:

```sh
./bin/konkour-bench status --run latest
./bin/konkour-bench inspect --attempt 123
```

`inspect --attempt` returns recorded stream events. `reconcile` retrieves accounting for completed **and interrupted** attempts, including older retries. It never recovers an answer from token/cost metadata or marks an interrupted request correct. For a historical request whose ID was missing, copy the generation ID from OpenRouter Activity:

```sh
./bin/konkour-bench reconcile --attempt 123 --generation-id gen-EXAMPLE
```

Manual reconciliation checks the generation ID, model, and creation time against the local attempt. Charges stay attached to that attempt, so reconciling an old retry cannot change a newer answer. Unknown costs remain unknown until the metadata reports a terminal outcome. Reconciliation uses the API destination recorded with the original run.

### Supervised runs on macOS

Prepare a launchd configuration after building the binary. This command does **not** start inference, and the plist contains the env-file path rather than its contents:

```sh
./bin/konkour-bench supervisor --env-file .env --budget-usd 50 --timeout 1h
```

When ready to resume, load the prepared job from a terminal:

```sh
launchctl bootstrap "gui/$(id -u)" "$PWD/data/supervisor/konkour-bench.plist"
```

launchd owns the worker independently of this chat or terminal session. It runs once and does not restart after completion, a credit failure, or a crash. Logs are in `data/supervisor/runner.log` and `runner-summary.log`. The CLI writes the final or partial report to `--out` on a graceful exit. After a forced kill, use `status` and `report` to recover bookkeeping and rebuild it.

To stop dispatch and cancel the local connections, then unload the job:

```sh
launchctl kill SIGINT "gui/$(id -u)/ai.konkour-bench.runner"
launchctl bootout "gui/$(id -u)/ai.konkour-bench.runner"
```

Stopping the client does not guarantee provider billing stops. Review interrupted attempts before adding `--retry-unknown`. To change the service configuration, use a new `--service-file` path; existing files are not overwritten. The optional `--env-file` supports a literal `OPENROUTER_API_KEY=value` assignment, with optional quotes or `export`, and does not execute shell expressions.

Useful flags include `--config`, `--db`, `--effort minimal,medium,max`, `--question`, and `--log-json`. Commands write summaries to stdout and operational logs to stderr. Use `COMMAND --help` for the full flag list. Keep the database and `.env` private; both are ignored by Git.

## Development

```sh
make build
make check
```

Checks include Go formatting, `go vet`, race-enabled tests, builds, and pinned Oxlint for the report JavaScript. Node/npm is needed only for the JavaScript development check; the built CLI and generated reports need no Node runtime. Tests use local mock HTTP servers and never call paid APIs. No browser automation is required for the automated checks.

API references: [structured outputs](https://openrouter.ai/docs/guides/features/structured-outputs), [models](https://openrouter.ai/docs/guides/overview/models), [image inputs](https://openrouter.ai/docs/guides/overview/multimodal/image-understanding), [reasoning](https://openrouter.ai/docs/guides/best-practices/reasoning-tokens), [Exacto](https://openrouter.ai/docs/guides/routing/model-variants/exacto), and [usage accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting).
