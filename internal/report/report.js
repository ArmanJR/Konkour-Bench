"use strict";

// Pure helpers are shared with the Node regression tests.
function matchesScope(o, s) {
  return (
    (!s.dataset || o.dataset === s.dataset) &&
    (!s.category || o.category === s.category) &&
    (s.cohort === "full" ||
      (s.cohort === "image-only" ? o.has_image : !o.has_image))
  );
}
function difficultyBins(questions) {
  const bins = [0, 0, 0, 0];
  for (const q of questions) {
    if (q.panel > 0) bins[Math.min(3, Math.floor((q.correct / q.panel) * 4))]++;
  }
  return bins;
}
function compareOutcomes(a, b) {
  const other = new Map(
    b.filter((o) => o.status === "complete").map((o) => [o.question, o]),
  );
  const result = { both: [], a: [], b: [], neither: [], agreement: 0, n: 0 };
  for (const o of a) {
    const p = other.get(o.question);
    if (o.status !== "complete" || !p) continue;
    result.n++;
    if (
      !o.invalid &&
      !p.invalid &&
      o.answer >= 1 &&
      o.answer <= 4 &&
      o.answer === p.answer
    )
      result.agreement++;
    result[
      o.correct ? (p.correct ? "both" : "a") : p.correct ? "b" : "neither"
    ].push(o.question);
  }
  return result;
}
function rankSpan(r) {
  return r.low === r.high ? String(r.low) : `${r.low}–${r.high}`;
}
// Name the configurations at a separation group's lowest or highest score.
function separationExtreme(group, panel, pick) {
  if (!group.correct.length) return null;
  const value = pick(...group.correct);
  return {
    accuracy: value / group.questions,
    configurations: panel.filter((_, i) => group.correct[i] === value),
  };
}
function sortedRows(rows, key, ascending) {
  return [...rows].sort((a, b) => {
    if (a[key] == null) return b[key] == null ? 0 : 1;
    if (b[key] == null) return -1;
    const diff =
      typeof a[key] === "string"
        ? a[key].localeCompare(b[key])
        : Number(a[key]) - Number(b[key]);
    return (
      (ascending ? diff : -diff) ||
      a.model.localeCompare(b.model) ||
      a.tier.localeCompare(b.tier)
    );
  });
}

function initReport(doc, data) {
  data = { ...data, rows: data.rows || [], outcomes: data.outcomes || [] };
  const $ = (id) => doc.getElementById(id);
  const number = (v, digits = 1) =>
    v == null
      ? "—"
      : Number(v).toLocaleString("en-US", { maximumFractionDigits: digits });
  const pct = (v) => (v == null ? "—" : `${number(v * 100, 2)}%`);
  const usd = (v) => (v == null ? "—" : `$${number(v, v < 0.01 ? 5 : 4)}`);
  const signed = (v) =>
    v == null ? "—" : `${v > 0 ? "+" : ""}${number(v * 100, 2)} pp`;
  // Compact one-decimal forms keep the wider diagnostic tables readable.
  const pct1 = (v) => (v == null ? "—" : `${number(v * 100, 1)}%`);
  const points = (v) =>
    v == null
      ? "—"
      : `${v > 0 ? "+" : v < 0 ? "−" : ""}${number(Math.abs(v) * 100, 1)} pp`;
  const short = modelDisplayName;
  const key = (r) => `${r.model}\u0000${r.tier}`;
  const label = (r) => `${short(r.model)} / ${r.tier}`;
  const qNumber = (id) => id.slice(id.lastIndexOf(":") + 1);
  const bank = new Map((data.question_bank || []).map((q) => [q.id, q]));
  let view,
    scope,
    scopeRows = [],
    scopedOutcomes = [],
    selectedQuestion = "",
    sortKey = "accuracy",
    ascending = false;

  const charts = createReportCharts(
    doc,
    data,
    () => scope,
    () =>
      [$("model-search").value.trim(), $("tier").value]
        .filter(Boolean)
        .join(" · "),
  );

  function el(tag, text, parent, className) {
    const node = doc.createElement(tag);
    if (text != null) node.textContent = text;
    if (className) node.className = className;
    if (parent) parent.append(node);
    return node;
  }
  function clear(id) {
    const node = $(id);
    node.replaceChildren();
    return node;
  }
  function empty(parent, text = "No results for this selection.") {
    return el("p", text, parent, "empty");
  }
  function table(id, columns, rows, sortable = false) {
    const wrap = clear(id);
    if (!rows.length) {
      empty(wrap);
      return;
    }
    wrap.tabIndex = 0;
    wrap.setAttribute(
      "aria-label",
      `${id.replaceAll("-", " ")} table; scroll horizontally if needed`,
    );
    const t = el("table", null, wrap),
      tr = el("tr", null, el("thead", null, t));
    for (const c of columns) {
      const th = el("th", null, tr);
      th.scope = "col";
      if (sortable && c.key) {
        th.setAttribute(
          "aria-sort",
          sortKey === c.key ? (ascending ? "ascending" : "descending") : "none",
        );
        const b = el(
          "button",
          c.label + (sortKey === c.key ? (ascending ? " ↑" : " ↓") : ""),
          th,
        );
        b.type = "button";
        b.onclick = () => {
          ascending = sortKey === c.key ? !ascending : c.key === "model";
          sortKey = c.key;
          renderPerformance();
        };
      } else th.textContent = c.label;
    }
    const body = el("tbody", null, t);
    for (const row of rows) {
      const tr = el("tr", null, body);
      for (const c of columns) {
        const td = el("td", null, tr, c.className),
          content = c.format(row);
        if (content?.nodeType) td.append(content);
        else td.textContent = content ?? "—";
      }
    }
  }
  function modelIdentity(model, detail = "") {
    const node = el("span", null, null, "model-identity"),
      source = modelLogo(model);
    node.title = model;
    if (source) {
      const img = el("img", null, node, "model-logo");
      img.src = source;
      img.alt = "";
      img.width = 24;
      img.height = 24;
      img.setAttribute("aria-hidden", "true");
    }
    const copy = el("span", null, node, "model-copy");
    el("span", short(model), copy, "model-name");
    if (detail) el("span", detail, copy, "sub");
    return node;
  }
  function modelCell(r) {
    return modelIdentity(
      r.model,
      `${r.model.split("/")[0]} · ${r.tier} → ${r.sent_efforts || "—"}`,
    );
  }
  function qLink(id, parent) {
    const b = el("button", `Q${qNumber(id)}`, parent, "question-link");
    b.type = "button";
    b.title = id;
    b.onclick = () => {
      selectedQuestion = id;
      $("question-search").value = "";
      $("question-filter").value = "all";
      renderQuestions();
      $("questions").scrollIntoView({ behavior: "instant" });
      $("question-detail").tabIndex = -1;
      $("question-detail").focus({ preventScroll: true });
    };
    return b;
  }
  function filteredRows() {
    const search = $("model-search").value.toLowerCase().trim(),
      tier = $("tier").value;
    return scopeRows.filter(
      (r) =>
        (!search || modelMatchesQuery(r.model, search)) &&
        (!tier || r.tier === tier),
    );
  }
  function renderOverview() {
    const box = clear("insights");
    for (const insight of view.insights) {
      const card = el("article", null, box, "insight");
      el("div", insight.label, card, "insight-label");
      el("div", insight.value, card, "insight-value");
      el("p", insight.detail, card);
    }
    if (!view.insights.length) empty(box);
    charts.coverage(view.coverage);
  }
  function renderPerformance() {
    const rows = filteredRows(),
      ranked = rows.filter((r) => r.accuracy != null),
      visible = new Set(rows.map(key)),
      rankIntervals = new Map(view.ranks.map((r) => [key(r), r]));
    charts.accuracy(sortedRows(ranked, "accuracy", false));
    charts.ranks(
      view.ranks.filter((r) => visible.has(key(r))),
      view.panel.length,
      view.rank_resamples,
    );
    const allRanked = scopeRows.filter((r) => r.accuracy != null);
    table(
      "leaderboard",
      [
        {
          label: "Rank",
          format: (r) =>
            1 + allRanked.filter((v) => v.accuracy > r.accuracy).length,
        },
        {
          key: "model",
          label: "Model / requested → sent",
          format: modelCell,
          className: "model",
        },
        { key: "accuracy", label: "Accuracy", format: (r) => pct(r.accuracy) },
        { label: "Correct", format: (r) => `${r.correct} / ${r.total}` },
        {
          label: "95% interval",
          format: (r) =>
            r.accuracy_interval
              ? `${pct(r.accuracy_interval.low)}–${pct(r.accuracy_interval.high)}`
              : "—",
        },
        {
          label: "Rank, 90% resampled",
          format: (r) => {
            const interval = rankIntervals.get(key(r));
            return interval ? rankSpan(interval) : "—";
          },
        },
        {
          key: "cost_usd",
          label: "API cost / exam",
          format: (r) => usd(r.cost_usd),
        },
        {
          key: "tokens_per_question",
          label: "Tokens / Q",
          format: (r) => number(r.tokens_per_question, 0),
        },
        {
          label: "Frontier",
          format: (r) =>
            [r.cost_frontier && "cost", r.token_frontier && "tokens"]
              .filter(Boolean)
              .join(" · ") || "—",
        },
      ],
      sortedRows(ranked, sortKey, ascending),
      true,
    );
    const estimates = view.estimates.filter((e) => visible.has(key(e)));
    charts.estimates(estimates);
    table(
      "estimate-table",
      [
        {
          label: "Model / effort",
          format: (e) => modelIdentity(e.model, e.tier),
          className: "model",
        },
        {
          label: "Correct / done / total",
          format: (e) => `${e.correct} / ${e.complete} / ${e.total}`,
        },
        {
          label: "Bounds",
          format: (e) =>
            `${pct1(e.correct / e.total)}–${pct1((e.correct + e.unresolved) / e.total)}`,
        },
        {
          label: "Panel accuracy: unresolved vs done",
          format: (e) =>
            `${pct1(e.unresolved_panel_accuracy)} vs ${pct1(e.completed_panel_accuracy)}`,
        },
        {
          label: "Same model, unresolved Qs",
          format: (e) =>
            e.same_model.length
              ? e.same_model
                  .map((c) => `${c.label} ${c.n}/${e.unresolved}`)
                  .join(" · ")
              : "—",
        },
        {
          label: "Estimate (correct ± SD)",
          format: (e) =>
            `${pct1(e.estimate)} (${number(e.expected_correct, 1)} ± ${number(e.expected_correct_sd, 1)})`,
        },
        {
          label: "Held-out: predicted vs actual",
          format: (e) =>
            e.check_predicted == null
              ? "—"
              : `${pct1(e.check_predicted)} vs ${pct1(e.unresolved_panel_accuracy)}`,
        },
      ],
      estimates,
    );
    const unranked = rows.filter((r) => r.accuracy == null);
    $("unranked-summary").textContent =
      `${unranked.length} unranked configurations · coverage and possible completion bounds`;
    table(
      "unranked-table",
      [
        {
          label: "Model / requested → sent",
          format: modelCell,
          className: "model",
        },
        { label: "Status", format: (r) => r.status },
        { label: "Completed", format: (r) => `${r.complete} / ${r.total}` },
        { label: "Known correct", format: (r) => r.correct },
        { label: "Unresolved", format: (r) => r.failed },
        {
          label: "Unsupported / skipped",
          format: (r) => `${r.unsupported} / ${r.skipped}`,
        },
        {
          label: "Completion bounds",
          format: (r) =>
            r.completion_bounds
              ? `${pct(r.completion_bounds.low)}–${pct(r.completion_bounds.high)}`
              : "—",
        },
      ],
      unranked,
    );
    charts.scatter(ranked, "cost");
    charts.scatter(ranked, "tokens");
    const distribution = (r, name, format) => {
      const d = r[name];
      return d?.n ? `${format(d.median)} / ${format(d.p90)} (n=${d.n})` : "—";
    };
    table(
      "resource-table",
      [
        { label: "Model / effort", format: modelCell, className: "model" },
        { label: "Completed", format: (r) => `${r.complete}/${r.total}` },
        {
          label: "USD median / p90",
          format: (r) => distribution(r, "cost_distribution", usd),
        },
        {
          label: "Tokens median / p90",
          format: (r) =>
            distribution(r, "token_distribution", (v) => number(v, 0)),
        },
        {
          label: "Seconds median / p90",
          format: (r) =>
            distribution(r, "latency_seconds", (v) => number(v, 1)),
        },
        { label: "Invalid answers", format: (r) => r.invalid },
      ],
      rows.filter((r) => r.complete > 0),
    );
    const length = view.length;
    charts.answerLength(
      length.rows.filter((r) => visible.has(key(r))),
      length,
    );
    const summary = [];
    if (length.ratio != null)
      summary.push(
        `Holding question and configuration fixed, wrong answers used ${number(length.ratio, 2)}× the output tokens of correct answers (${number(length.answers, 0)} answers; ${length.contrast_questions} questions have both correct and wrong answers).`,
      );
    if (length.longest)
      summary.push(
        `Each complete configuration's longest 10% of answers were ${pct(length.longest_correct / length.longest)} correct (${length.longest_correct} of ${length.longest}), versus ${pct(length.answers_correct / length.answers)} across all ${number(length.answers, 0)} answers with known output tokens.`,
      );
    clear("length-summary").textContent = summary.length
      ? summary.join(" ")
      : "No complete configurations with known output tokens in this scope.";
  }
  function renderEffort() {
    const [from, to] = $("effort-pair").value.split("|");
    const rows = view.efforts
      .filter((e) => e.from === from && e.to === to && e.paired > 0)
      .sort((a, b) => b.delta - a.delta || a.model.localeCompare(b.model));
    charts.effort(rows, from, to);
    table(
      "effort-table",
      [
        {
          label: "Model",
          format: (r) => modelIdentity(r.model),
          className: "model",
        },
        {
          label: "Paired / selected",
          format: (r) =>
            `${r.paired} / ${r.total}${r.paired < r.total ? " · partial*" : ""}`,
        },
        { label: "Gains", format: (r) => `+${r.gained}` },
        { label: "Regressions", format: (r) => `−${r.lost}` },
        { label: "Net accuracy", format: (r) => signed(r.delta) },
        {
          label: "Cost multiplier",
          format: (r) =>
            r.cost_ratio == null ? "—" : `${number(r.cost_ratio, 2)}×`,
        },
        {
          label: "Token multiplier",
          format: (r) =>
            r.token_ratio == null ? "—" : `${number(r.token_ratio, 2)}×`,
        },
      ],
      rows,
    );
  }
  function renderSubjects() {
    const categoryRows = data.rows.filter(
      (r) =>
        r.cohort === scope.cohort && r.dataset === scope.dataset && r.category,
    );
    const categories = [...new Set(categoryRows.map((r) => r.category))];
    const overall = data.rows.filter(
      (r) =>
        r.cohort === scope.cohort && r.dataset === scope.dataset && !r.category,
    );
    charts.subjects(overall, categories, categoryRows);
    const mix = clear("subject-mix");
    for (const category of categories) {
      const n = categoryRows.find((r) => r.category === category)?.total || 0;
      const card = el("div", null, mix, "subject-card");
      el("span", category, card).dir = "auto";
      el("b", `${n} questions`, card);
    }
    table(
      "heatmap",
      [
        { label: "Model / effort", format: modelCell, className: "model" },
        ...categories.map((category) => ({
          label: category,
          format: (r) => {
            const item = categoryRows.find(
                (v) => key(v) === key(r) && v.category === category,
              ),
              cell = el("div", null, null, "heat-cell");
            if (!item) {
              cell.textContent = "—";
              return cell;
            }
            if (item.accuracy != null) {
              cell.textContent = pct(item.accuracy);
              el("span", `${item.correct} / ${item.total}`, cell);
              cell.style.background = `rgba(74, 141, 107, ${0.07 + item.accuracy * 0.23})`;
              cell.style.padding = "6px 12px";
              cell.style.borderRadius = "4px";
            } else {
              cell.textContent = item.status;
              el("span", `${item.complete} / ${item.total} completed`, cell);
            }
            if (category === scope.category)
              cell.classList.add("selected-column");
            return cell;
          },
        })),
      ],
      overall,
    );
  }
  function renderSeparation() {
    charts.separation(view.separation, view.panel);
    const extreme = (g, pick) => {
      const e = separationExtreme(g, view.panel, pick);
      if (!e) return "—";
      const names =
        e.configurations.length > 3
          ? `${e.configurations.length} configurations`
          : e.configurations.map(label).join(", ");
      return `${pct(e.accuracy)} · ${names}`;
    };
    table(
      "separation-table",
      [
        {
          label: "Subject / section",
          format: (g) => {
            const node = el(
              "span",
              g.level === "subject"
                ? g.category
                : `↳ ${g.section || "Untagged questions"}`,
            );
            node.dir = "auto";
            return node;
          },
        },
        { label: "Questions", format: (g) => g.questions },
        {
          label: "All right / none right",
          format: (g) => `${g.ceiling} / ${g.floor}`,
        },
        { label: "Mean accuracy", format: (g) => pct(g.mean) },
        { label: "Lowest", format: (g) => extreme(g, Math.min) },
        { label: "Highest", format: (g) => extreme(g, Math.max) },
        {
          label: "Share of variance",
          format: (g) => pct(g.variance_share),
        },
      ],
      view.separation,
    );
  }
  function renderDifficulty() {
    charts.difficulty(difficultyBins(view.questions), view.coverage.ranked);
  }
  function renderQuestions() {
    const query = $("question-search").value.trim().toLowerCase(),
      filter = $("question-filter").value;
    const questions = view.questions.filter((q) => {
      const content = bank.get(q.id),
        match =
          !query ||
          [q.id, q.category, content?.text].some((t) =>
            t?.toLowerCase().includes(query),
          );
      return (
        match &&
        (filter === "all" ||
          (filter === "review" && q.review) ||
          (filter === "hard" && q.panel > 0 && q.correct * 2 < q.panel) ||
          (filter === "image" && q.has_image))
      );
    });
    $("panel-note").textContent =
      `${view.coverage.ranked} complete configurations in the difficulty panel. All ${view.coverage.configurations} selected configurations remain visible in the answer details. Questions are ordered from lowest to highest panel accuracy.`;
    $("question-count").textContent =
      `${questions.length} of ${view.questions.length} questions shown`;
    if (!questions.some((q) => q.id === selectedQuestion))
      selectedQuestion = questions[0]?.id || "";
    const list = clear("question-list");
    for (const q of questions) {
      const button = el("button", null, list, "question-button");
      button.type = "button";
      button.setAttribute("aria-pressed", String(q.id === selectedQuestion));
      button.title = q.id;
      el(
        "strong",
        `Q${qNumber(q.id)} · ${q.panel ? `${q.correct}/${q.panel} correct` : "No complete panel"}`,
        button,
      );
      el(
        "span",
        `${q.category}${q.section ? ` · ${q.section}` : ""}${q.has_image ? " · image" : ""}${q.review ? " · review" : ""}`,
        button,
        "sub",
      ).dir = "auto";
      const bar = el("div", null, button, "mini-bar");
      el("span", null, bar).style.width =
        `${q.panel ? (q.correct / q.panel) * 100 : 0}%`;
      button.onclick = () => {
        selectedQuestion = q.id;
        renderQuestions();
      };
    }
    if (!questions.length) {
      empty(list, "No questions match these filters.");
      empty(
        clear("question-detail"),
        "Choose another filter to inspect a question.",
      );
      return;
    }
    renderQuestion(view.questions.find((q) => q.id === selectedQuestion));
  }
  function renderQuestion(q) {
    const box = clear("question-detail"),
      content = bank.get(q.id);
    el("h3", `Question ${qNumber(q.id)}`, box);
    el(
      "p",
      `${q.id} · ${q.category}${q.section ? ` · ${q.section}` : ""}`,
      box,
      "note",
    ).dir = "auto";
    if (q.review)
      el(
        "p",
        "Review candidate: most complete configurations chose the same answer other than the supplied key.",
        box,
        "callout",
      );
    if (content) {
      el("div", content.text, box, "question-text").lang = "fa";
      const addImage = (src, parent, alt) => {
        if (!/^data:image\/(png|jpeg|gif);base64,/.test(src || "")) return;
        const img = el("img", null, parent);
        img.src = src;
        img.alt = alt;
      };
      addImage(content.image, box, `Question ${content.number} figure`);
      for (const option of content.options) {
        const row = el(
          "div",
          null,
          box,
          `option${option.number === q.expected ? " key" : ""}`,
        );
        el(
          "span",
          `${option.number}${option.number === q.expected ? " · KEY" : ""}`,
          row,
          "option-label",
        );
        const body = el("div", null, row, "option-content");
        el("div", option.text, body, "option-text").lang = "fa";
        addImage(option.image, body, `Option ${option.number} figure`);
      }
      if (typeof renderMathInElement === "function")
        renderMathInElement(box, {
          delimiters: [
            { left: "$$", right: "$$", display: true },
            { left: "\\[", right: "\\]", display: true },
            { left: "\\(", right: "\\)", display: false },
            { left: "$", right: "$", display: false },
          ],
          throwOnError: false,
          trust: false,
        });
    } else
      el(
        "p",
        "Question text is unavailable in this report snapshot.",
        box,
        "note",
      );
    el("div", null, box).id = "answer-chart";
    charts.answers(q);
    el(
      "p",
      `${q.correct} / ${q.panel} panel configurations correct. Key: option ${q.expected}.`,
      box,
      "note",
    );
    const wrap = el("div", null, box, "table-wrap");
    wrap.id = "question-answers";
    table(
      "question-answers",
      [
        {
          label: "Model / effort",
          format: (o) => modelIdentity(o.model, o.tier),
          className: "model",
        },
        {
          label: "Answer",
          format: (o) =>
            o.status === "complete"
              ? o.invalid
                ? "Invalid"
                : o.answer || "Invalid"
              : "—",
        },
        {
          label: "Result",
          format: (o) =>
            o.status === "complete"
              ? o.correct
                ? "Correct"
                : "Wrong"
              : o.status,
        },
        { label: "USD", format: (o) => usd(o.usage.cost) },
        { label: "Tokens", format: (o) => number(o.usage.total_tokens, 0) },
        {
          label: "Audit detail",
          format: (o) => {
            const d = el("details");
            el("summary", "Inspect", d).title =
              `${o.model} / ${o.tier} audit detail`;
            el(
              "p",
              `Backend: ${o.backend || "—"} · Provider: ${o.provider || "—"} · Sent effort: ${o.sent_effort || "—"}`,
              d,
              "note",
            );
            el(
              "p",
              `Selected run: ${o.run_id || "—"} · Original inference: ${o.origin_run || "—"} · ${o.cached ? "Reused cache" : "Recorded attempt"}`,
              d,
              "note",
            );
            if (o.invalid || o.note)
              el(
                "p",
                [o.invalid, o.note].filter(Boolean).join(" · "),
                d,
                "note",
              );
            el("code", o.request_hash || "No request dispatched", d);
            return d;
          },
        },
      ],
      scopedOutcomes.filter((o) => o.question === q.id),
    );
  }
  function setupComparison() {
    const ranked = scopeRows.filter((r) => r.accuracy != null);
    for (const [id, index] of [
      ["compare-a", 0],
      ["compare-b", 1],
    ]) {
      const previous = $(id).value,
        select = clear(id);
      for (const r of ranked) {
        const option = el("option", label(r), select);
        option.value = key(r);
      }
      const fallback = ranked[Math.min(index, ranked.length - 1)];
      select.value = ranked.some((r) => key(r) === previous)
        ? previous
        : fallback
          ? key(fallback)
          : "";
      select.disabled = !ranked.length;
    }
    renderComparison();
  }
  function renderComparison() {
    const box = clear("pair-results"),
      a = $("compare-a").value,
      b = $("compare-b").value;
    if (!a || !b) {
      empty(box, "No complete configurations available for paired comparison.");
      return;
    }
    const pair = compareOutcomes(
        scopedOutcomes.filter((o) => key(o) === a),
        scopedOutcomes.filter((o) => key(o) === b),
      ),
      chart = el("div", null, box, "chart-card");
    chart.id = "pair-chart";
    charts.pair(
      pair,
      scopeRows.find((r) => key(r) === a),
      scopeRows.find((r) => key(r) === b),
    );
    el(
      "p",
      `Net advantage of B: ${signed(pair.n ? (pair.b.length - pair.a.length) / pair.n : null)} · Same valid answer: ${pair.agreement}/${pair.n} · Either correct (oracle): ${pct(pair.n ? (pair.n - pair.neither.length) / pair.n : null)}`,
      box,
      "callout",
    );
    for (const [title, ids] of [
      ["Both wrong", pair.neither],
      ["Only A correct", pair.a],
      ["Only B correct", pair.b],
    ]) {
      if (!ids.length) continue;
      const block = el("div", null, box);
      block.style.marginTop = "18px";
      el("h3", title, block);
      const links = el("div", null, block, "question-links");
      ids.forEach((id) => qLink(id, links));
    }
  }
  function renderHealth() {
    const box = clear("health"),
      c = view.coverage,
      failures = el("div", null, box, "health-card");
    el("h3", "What prevented a valid answer?", failures);
    if (!c.reasons.length)
      el("p", "All selected outcomes are valid completions.", failures, "note");
    else {
      const list = el("ul", null, failures);
      for (const reason of c.reasons)
        el("li", `${number(reason.n, 0)} × ${reason.label}`, list);
    }
    const accounting = el("div", null, box, "health-card");
    el("h3", "Accounting quality", accounting);
    const completed = scopedOutcomes.filter((o) => o.status === "complete"),
      paid = completed.filter((o) => o.backend !== "claude");
    const facts = [
      `${c.invalid} invalid completions are scored wrong.`,
      `${paid.filter((o) => o.usage.cost == null).length} / ${paid.length} completed paid-API answers have unknown charges.`,
      `${completed.filter((o) => o.backend === "claude").length} subscription completions have no reported actual dollar charge.`,
      `${completed.filter((o) => o.usage.total_tokens == null).length} completed answers have unknown total tokens.`,
      `${completed.filter((o) => o.duration_seconds == null).length} completed answers have unavailable or invalid original timing.`,
      `${c.token_anomalies} completed answers report reasoning tokens greater than completion tokens. Values are preserved as received; do not infer a reasoning share from these records.`,
    ];
    const list = el("ul", null, accounting);
    facts.forEach((f) => el("li", f, list));
    table(
      "providers",
      [
        {
          label: "Model",
          format: (e) => modelIdentity(e.model),
          className: "model",
        },
        { label: "Provider", format: (e) => e.provider },
        {
          label: "Tiers served",
          format: (e) => e.tiers.map((t) => `${t.label} ${t.n}`).join(" · "),
        },
        { label: "Answers", format: (e) => e.answers },
        { label: "Invalid", format: (e) => e.invalid },
        { label: "Accuracy", format: (e) => pct1(e.correct / e.answers) },
        { label: "Expected", format: (e) => pct1(e.expected) },
        { label: "vs expected", format: (e) => points(e.excess) },
        {
          label: "vs other providers (95% CI)",
          format: (e) =>
            e.gap == null
              ? "—"
              : `${points(e.gap)} ${e.gap_interval ? `(${points(e.gap_interval.low)} to ${points(e.gap_interval.high)})` : "· no CI: a side has < 10 answers"}`,
        },
      ],
      view.providers,
    );
  }
  function renderProvenance() {
    table(
      "spending",
      [
        { label: "Run", format: (r) => r.run_id },
        { label: "New attempts", format: (r) => r.attempts },
        {
          label: "Subscription attempts",
          format: (r) => r.subscription_attempts || 0,
        },
        { label: "Known charges", format: (r) => usd(r.known_cost_usd) },
        {
          label: "Included retry cost",
          format: (r) => usd(r.known_retry_cost_usd),
        },
        {
          label: "Unknown paid charges",
          format: (r) =>
            data.backend === "claude" ? "N/A" : r.attempts_with_unknown_cost,
        },
      ],
      data.run_spending || [],
    );
    table(
      "sources",
      [
        {
          label: "Run / started (UTC)",
          format: (r) => {
            const node = el("span", r.id);
            el(
              "span",
              r.created.replace("T", " ").replace(/\.\d+Z$/, " UTC"),
              node,
              "sub",
            );
            return node;
          },
        },
        {
          label: "Backend / CLI version",
          format: (r) =>
            `${r.backend}${r.cli_version ? ` / ${r.cli_version}` : ""}`,
        },
        { label: "Selected jobs", format: (r) => r.selected_jobs },
        { label: "Origin answers", format: (r) => r.origin_answers },
        {
          label: "Timeout / budget",
          format: (r) =>
            `${number(r.timeout_seconds / 60)} min / ${usd(r.budget_usd)}`,
        },
      ],
      data.source_runs || [],
    );
    const meta = clear("dataset-metadata");
    for (const [id, content] of Object.entries(data.dataset_metadata || {})) {
      el("h3", id, meta);
      el("pre", JSON.stringify(content, null, 2), meta);
    }
    if (data.warnings?.length) {
      const list = el("ul", null, meta);
      for (const warning of data.warnings) el("li", warning, list);
    }
  }
  function render() {
    scope = {
      cohort: $("cohort").value,
      dataset: $("dataset").value,
      category: $("category").value,
    };
    scopeRows = data.rows.filter(
      (r) =>
        r.cohort === scope.cohort &&
        r.dataset === scope.dataset &&
        r.category === scope.category,
    );
    scopedOutcomes = data.outcomes.filter((o) => matchesScope(o, scope));
    view = (data.analysis || []).find(
      (v) =>
        v.cohort === scope.cohort &&
        v.dataset === scope.dataset &&
        v.category === scope.category,
    ) || {
      coverage: {
        questions: 0,
        images: 0,
        models: 0,
        configurations: 0,
        ranked: 0,
        slots: 0,
        complete: 0,
        statuses: {},
        reasons: [],
        invalid: 0,
        token_anomalies: 0,
      },
      insights: [],
      efforts: [],
      questions: [],
      panel: [],
      ranks: [],
      rank_resamples: 0,
      separation: [],
      length: { rows: [], answers: 0, answers_correct: 0, longest: 0 },
      estimates: [],
      providers: [],
    };
    $("scope-summary").textContent =
      `${view.coverage.questions} questions · ${view.coverage.ranked}/${view.coverage.configurations} configurations ranked · ${scope.cohort}${scope.category ? ` · ${scope.category}` : ""}`;
    renderOverview();
    renderDifficulty();
    renderPerformance();
    renderEffort();
    renderSubjects();
    renderSeparation();
    renderQuestions();
    setupComparison();
    renderHealth();
  }
  for (const id of ["cohort", "dataset", "category"])
    $(id).addEventListener("change", render);
  $("model-search").addEventListener("input", renderPerformance);
  $("tier").addEventListener("change", renderPerformance);
  $("effort-pair").addEventListener("change", renderEffort);
  $("question-search").addEventListener("input", renderQuestions);
  $("question-filter").addEventListener("change", renderQuestions);
  for (const id of ["compare-a", "compare-b"])
    $(id).addEventListener("change", renderComparison);
  $("reset").onclick = () => {
    $("cohort").value = "full";
    for (const id of [
      "dataset",
      "category",
      "tier",
      "model-search",
      "question-search",
    ])
      $(id).value = "";
    $("question-filter").value = "all";
    $("effort-pair").value = "medium|max";
    selectedQuestion = "";
    sortKey = "accuracy";
    ascending = false;
    render();
  };
  $("print").onclick = () => doc.defaultView.print();
  renderProvenance();
  render();
  return { render };
}
if (typeof module !== "undefined" && module.exports)
  module.exports = {
    matchesScope,
    compareOutcomes,
    sortedRows,
    difficultyBins,
    rankSpan,
    separationExtreme,
    initReport,
  };
if (typeof document !== "undefined" && document.getElementById("data")) {
  const start = () =>
    initReport(
      document,
      JSON.parse(document.getElementById("data").textContent),
    );
  // Font metrics must be ready before wrapping chart captions and laying out labels.
  if (document.fonts)
    document.fonts.load("500 16px Inter").then(start, (err) => {
      console.error(
        "Report typography could not load; using the system fallback.",
        err,
      );
      start();
    });
  else start();
}
