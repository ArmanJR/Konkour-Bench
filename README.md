# Konkour Bench

A Go benchmark for four-option Iranian university entrance exam questions, using OpenRouter or the Claude Code CLI with a Claude subscription. Each question is evaluated independently at `minimal`, `medium`, and `max` reasoning effort. Results are cached locally and scored into offline HTML, CSV, JSON, and SVG reports.

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

The booklet contains 20 mathematics, 30 Persian language/literature, 15 social sciences, and 15 psychology questions. Literature questions carry section tags in booklet order: literary history and stylistics (21–28), rhetorical devices (29–36), prosody and rhyme (37–44), one integrated question (45), and semantic correspondence (46–50). Source spot-checking corrected question 15, option 2 from `2/8` to the PDF's `3/8`. The supplied answer key is unchanged; it has not been independently verified against an official key.

Run a small benchmark against the original sample and generate its report:

```sh
./bin/konkour-bench run --config benchmark.sample.json \
  --model openai/gpt-6-luna \
  --question sample:2,sample:9
./bin/konkour-bench report --run latest --out reports/smoke
```

Open `reports/smoke/index.html` in a browser. The HTML is self-contained, including charts, Persian questions, images, equations, and fonts. It includes scoped findings, coverage, sortable rankings with accuracy intervals and resampled rank intervals, difficulty-based estimates for configurations with unresolved answers, paired effort comparisons, subject scores and score separation by subject and section, cost/token frontiers, resource distributions, output length of correct versus wrong answers, question difficulty, shared errors, serving-provider comparisons, and run provenance. Adjacent CSV, JSON, and SVG files support further analysis. Full transcripts remain in the database.

Run all configured models and questions:

```sh
./bin/konkour-bench run
./bin/konkour-bench report
```

Repeating a run reuses completed generations, including malformed answers and refusals. The run summary reports new attempts, cache hits, and new spending. Generation tokens remain uncapped. Each request defaults to a 20-minute emergency deadline (`--timeout 20m`), with a warning after ten minutes (`--warn-after 10m`). Set either duration to `0` to disable it. Changing the timeout requires restarting the worker; an existing worker retains the timeout it started with. `--concurrency` defaults to four, with at most one active question per model.

The default `--budget-usd 50` stops new dispatch once recorded new spending reaches $50. If a timed-out request or another response is missing its cost, the worker temporarily pauses new dispatch and automatically retrieves generation accounting, allowing up to five minutes for delayed metadata. Active requests can finish during this wait. Once all pending charges are known and the budget permits, the same run continues automatically; interrupted questions remain unresolved and are not retried. Missing generation IDs or accounting that remains unavailable stop the run with an actionable error. This is not a hard billing cap: concurrent requests can overshoot it, and provider cancellation support varies. Set a separate OpenRouter key credit limit for account protection. `--budget-usd 0` disables the local spending guard and its automatic accounting wait.

## Claude subscription backend

Install Claude Code 2.1.296 or newer and sign in with `claude auth login`. The backend uses your existing subscription login, including its usage limits; it does not need `OPENROUTER_API_KEY`. API-key and third-party provider authentication are not accepted. `--claude-bin /absolute/path/to/claude` selects another installation.

```sh
./bin/konkour-bench models --backend claude
./bin/konkour-bench run --backend claude \
  --model anthropic/claude-sonnet-5.5 --effort medium \
  --question humanities-1405-booklet-1:1
./bin/konkour-bench report --run latest --out reports/claude
./bin/konkour-bench serve --run latest
```

After changing answer keys or categories, `report --backend claude` rescores the current dataset against compatible cached CLI results without invoking Claude. `serve --backend claude` shows the same view. These commands use the most recent saved Claude CLI version; `--run RUN_ID` instead reproduces a saved run's original snapshot.

For the complete Claude benchmark, use `run --backend claude`. It selects the configured `anthropic/claude-opus-5.5` and `anthropic/claude-sonnet-5.5` models, all questions, and all three tiers. Explicitly selecting another model fails before inference. OpenRouter remains the default backend for existing commands; `benchmark.json` is unchanged.

To benchmark all configured models together, routing Claude to the subscription CLI and everything else to OpenRouter, use:

```sh
./bin/konkour-bench run --backend mixed --env-file .env --budget-usd 30
./bin/konkour-bench report --backend mixed --out reports/mixed
./bin/konkour-bench serve --backend mixed
```

Mixed runs retain the individual backend cache keys, so completed OpenRouter and CLI answers are reused. Claude's older OpenRouter answers remain a different protocol and are not reused as CLI answers. All tiers are included, with up to four requests overall and at most one Claude CLI request at a time. The dollar budget applies only to OpenRouter. If one backend hits an account or quota error, it stops its own dispatch while the other can continue. Claude never falls back to paid OpenRouter inference. Mixed reports show both backends together; Claude has accuracy/token scores but no cost frontier membership. Subscription attempts are counted separately from unknown OpenRouter charges in the spending ledger. `supervisor --backend mixed` prepares a persistent worker with the same routing; supply `--env-file` for OpenRouter authentication.

Each question starts a fresh `claude -p` process in an empty temporary directory. The question and inline image blocks arrive through stdin; the concise system instruction and JSON schema are explicit arguments. Safe/restricted mode, disabled skills/hooks/MCP, and an empty tool list prevent the model from reading local answer keys or using search. Claude's `StructuredOutput` tool remains available solely for the answer schema. There is no conversation resume or session persistence. The CLI may use multiple internal turns to produce that structured result; this is a distinct evaluation protocol from a direct OpenRouter generation.

The tiers request `low`, `medium`, and `max`, respectively. Explicit effort rejections can use the existing same-band fallbacks. The benchmark sets no token cap; the CLI's own model limits still apply. Claude Code can clamp effort under organization policy, and its stream does not always expose the applied effort, so recorded tiers describe the requested setting. A different generated model is rejected rather than scored under the requested model. `:exacto` is specific to OpenRouter and does not apply to Claude CLI sessions. See the official [headless guide](https://code.claude.com/docs/en/headless), [CLI reference](https://code.claude.com/docs/en/cli-reference), and [effort configuration](https://code.claude.com/docs/en/model-config#adjust-effort-level).

Claude concurrency defaults to **one**; `--concurrency` can raise it, with one active question per model. The same 20-minute emergency timeout applies and terminates the local process group. Authentication, subscription-limit, and unexpected CLI failures stop dispatch. Rerunning reuses completed answers. An interrupted or uncertain request still requires `--retry-unknown` to repeat it. The benchmark does not automatically retry subscription-limit failures or switch to API billing. If your Claude account has paid **extra usage** enabled, disable it in Claude to stay within the subscription; detecting an overage event and stopping locally is not a hard billing cap.

Claude's `total_cost_usd` is an API-price estimate, **not your subscription charge**. It is stored as `estimated_api_cost_usd` in normalized usage and the run summary. Actual cost stays unknown, so these results participate in accuracy/token rankings but not cost Pareto frontiers. `--budget-usd` defaults to zero for Claude and rejects a nonzero value because it cannot cap subscription consumption. Tokens include uncached input, cache creation, cache reads, and output; unavailable reasoning counts remain unknown. The backend, CLI version, invocation arguments, prompt, schema, effort, and image bytes separate its cache from OpenRouter and from other CLI versions.

Use `inspect --hash HASH` for the exact stored invocation and result, and `inspect --attempt ID` for raw CLI events, session/request IDs, and stderr. `reconcile` only handles OpenRouter attempts. For a macOS background worker, `supervisor --backend claude` prepares a launchd job with the resolved CLI executable and needs no `--env-file`; starting the worker remains a separate action. Future backends implement `internal/backbone.Client` and supply deterministic request encoding and capability metadata.

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

- Question records have `number`, `category`, `question`, four `options`, and optional `image` and `section` fields. A section subdivides a category for reporting only; it is never sent to a model.
- Options have a unique `number` from 1 to 4, optional `text`, and an optional `image` path. An image-only option can have `"text": null`.
- Answer records have `number` and `correct_option` from 1 to 4.

Preserve the original Persian, English, or Arabic text and inline LaTeX. JSON backslashes must be escaped. PNG, JPEG, and GIF images are validated and sent unchanged. Missing or corrupt images stop validation; they are never silently removed. Extra answer keys are allowed with a warning so partial booklets work. Every included question needs a key. Missing categories are grouped as `Uncategorized`.

`benchmark.sample.json` selects the original sample. The sample has 15 questions, including two with images, and 24 answer keys. The nine extra keys are intentionally ignored. The sample is a fixture for exercising the benchmark, not a representative evaluation dataset.

After adding questions, rerun the benchmark; unchanged requests remain cached. After correcting keys, categories, or sections, regenerate the report without inference:

```sh
./bin/konkour-bench validate
./bin/konkour-bench report --out reports/rescored
```

Without `--run`, reports use the current dataset/key snapshot and compatible recorded requests. Newly added questions remain uncompleted until evaluated. `report --run RUN_ID` reproduces a saved run’s original dataset, answer keys, and question/model selection. Every report records a scoring revision and protocol fingerprint.

### Analyze a completed benchmark

To combine the mixed benchmark with its resume runs and compatible cached answers:

```sh
./bin/konkour-bench report --backend mixed --out reports/humanities-1405-mixed
```

The output folder is a report label, not a saved run ID. Use an actual timestamped ID with `--run` to select a single run snapshot. Report generation uses local records and performs no inference or accounting API calls.

The checked-in [humanities 1405 mixed report](released-reports/humanities-1405-mixed/index.html) includes 13 models, with its exports and [eight reasoning observations](released-reports/humanities-1405-mixed/reasoning-observations.md) alongside it. The default configuration excludes MiMo-v2.6-Pro and MiMo-v2.6-Flash because neither completed an evaluation in this benchmark; historical records remain in the database.

The report starts with the full exam. Cohort, dataset, and subject controls recompute the analysis scope; the model search and effort display filter only hide performance/resource rows, without changing frontier membership. Incomplete configurations remain unranked, with best/worst possible completion bounds where applicable. These bounds are distinct from the 95% Wilson intervals shown for complete scores. Wilson intervals describe an independent-question sampling model, not run-to-run variability or generalization across exams.

Effort comparisons show gains and regressions on questions completed at both settings, and label partial pairs. Question difficulty uses a fixed panel of configurations complete in the selected scope. Review flags identify a strict majority choosing the same non-key option; they do not establish a key error or change grading. The shared-error comparison's “either correct” result is an oracle upper bound, not an ensemble score. Subscription estimates never enter paid cost rankings. Resource quantiles use known measurements on completed answers, with linear interpolation at `(n-1)*p`; original inference timing survives cache reuse.

Five diagnostics use the same complete-configuration panel. Rank intervals come from 2,000 paired question resamples with a fixed seed; ties share the better rank. Unresolved-answer estimates apply only when every missing answer is an unresolved request: a Rasch model fits item difficulty to the panel and ability to the configuration's completed answers. A held-out check refits it for each complete configuration on the same questions, and the same model's other tiers are shown on those questions; the estimate is descriptive and never ranks. Score separation attributes between-configuration variance in total correct to subjects and sections, as cov(group, total) / var(total). Answer length compares output tokens of correct and wrong answers, including a ratio from log tokens with question and configuration fixed effects. Provider comparisons use panel accuracy on the same questions excluding the same model, with intervals clustered by question and shown only with at least 10 answers on each side.

Exports include `leaderboard.csv`, `outcomes.csv`, `questions.csv`, `effort.csv`, `ranks.csv`, `estimates.csv`, `separation.csv`, `length.csv`, and `providers.csv`, plus `results.json` with all scoped analyses. The per-analysis CSVs include cohort/dataset/category columns: select a scope before aggregating, because the same observations appear in multiple scopes. The spending ledger covers new attempts in matching-protocol runs; the provenance table also identifies earlier inference origins reused from cache.

Every diagram has a **Download PNG** button. Images include the current selection, title, units, legends, and interpretation notes, and export at double resolution (bounded for unusually large charts). Export works offline. Pareto diagrams place model names, requested effort, and accuracy beside frontier points without connector lines. Gray markers show the remaining configurations; marker shapes distinguish efforts. Hover or focus a point for exact resource use. Display filters hide points without recomputing frontier membership.

Model-family logos appear alongside names in tables and charts, including exported PNGs. Artwork is embedded, so shared HTML remains offline; unknown model families keep their text labels. Asset sources and update instructions are in [internal/report/logos/README.md](internal/report/logos/README.md).

Charts use an embedded Inter typeface with readable model labels, consistent weights, and letter spacing. The same font travels with PNG exports. Exact model IDs remain in hover details and data exports, and model search accepts both IDs and display names. Font sources and update instructions are in [internal/report/fonts/README.md](internal/report/fonts/README.md).

Report code and templates live in `internal/report/`. The committed math assets are generated from the dashboard's existing pinned KaTeX dependency; after updating that dependency, run `npm --prefix dashboard ci --no-audit --no-fund` and `make report-assets`. Do not edit `math.generated.*` directly. `make check` includes the report's Go and dependency-free JavaScript regression tests.

## Evaluation rules

Every request uses the same concise Persian instruction, original question/options, and labeled images. Categories, exam names, filenames, and answer keys are excluded. The structured response must be exactly an object such as `{"answer": 3}` with one integer from 1 to 4. No prose extraction, JSON repair, research tools, or benchmark-selected model fallback is used.

New inference requests use streaming; response-header IDs and raw SSE data are saved as they arrive. The final assembled response must pass the same strict answer validation. Heartbeat comments are not evidence of model reasoning progress. A dropped stream, malformed event, or deadline leaves the outcome unknown, even if partial text looks like a valid answer.

All OpenRouter requests carry `:exacto` and require support for the supplied parameters. Reasoning is enabled; temperature, top-p, and seed are left at provider defaults. The client never sends `max_tokens`, `max_completion_tokens`, or `reasoning.max_tokens`. Provider limits and internal effort translation can still apply.

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
./bin/konkour-bench supervisor --env-file .env --budget-usd 50 --timeout 20m
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
