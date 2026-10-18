"use strict";
const data = JSON.parse(document.getElementById("data").textContent);
const ids = ["cohort", "dataset", "category", "model", "tier"];
const controls = Object.fromEntries(ids.map(id => [id, document.getElementById(id)]));
const ns = "http://www.w3.org/2000/svg";
const palette = ["#236b9e", "#278566", "#9353a6", "#b35e21", "#a43c59", "#657320"];
let sortKey = "accuracy", ascending = false;
const number = (x, digits = 2) => x == null ? "—" : Number(x).toLocaleString(undefined, {maximumFractionDigits: digits});
const percent = x => x == null ? "—" : `${number(x * 100)}%`;
function element(tag, text, parent) {
  const e = document.createElement(tag);
  if (text != null) e.textContent = text;
  if (parent) parent.append(e);
  return e;
}
function shape(tag, attrs, parent, text) {
  const e = document.createElementNS(ns, tag);
  for (const [k, v] of Object.entries(attrs)) e.setAttribute(k, v);
  if (text != null) e.textContent = text;
  parent.append(e);
  return e;
}
function svg(target, height = 360) {
  const container = document.getElementById(target);
  container.replaceChildren();
  return shape("svg", {viewBox: `0 0 760 ${height}`, role: "img", "aria-label": target}, container);
}
function matches(r, ignoreCategory = false, ignoreTier = false) {
  return r.cohort === controls.cohort.value && r.dataset === controls.dataset.value &&
    (ignoreCategory || r.category === controls.category.value) &&
    (!controls.model.value || r.model === controls.model.value) &&
    (ignoreTier || !controls.tier.value || r.tier === controls.tier.value);
}
function chartRows() { return data.rows.filter(r => matches(r) && r.accuracy != null); }
function scatter(rows, metric) {
  const s = svg(`${metric}-chart`);
  const key = metric === "cost" ? "cost_per_question_usd" : "tokens_per_question";
  const points = rows.filter(r => r[key] != null);
  const max = Math.max(...points.map(r => r[key]), 0) || 1;
  const x = v => 70 + v / max * 640, y = v => 300 - v * 250;
  for (let i = 0; i <= 4; i++) {
    const a = i / 4;
    shape("line", {x1: 70, x2: 710, y1: y(a), y2: y(a), stroke: "#e0e5ea"}, s);
    shape("text", {x: 60, y: y(a) + 5, "text-anchor": "end", "font-size": 12}, s, percent(a));
    shape("text", {x: x(a * max), y: 323, "text-anchor": "middle", "font-size": 12}, s, number(a * max, metric === "cost" ? 5 : 0));
  }
  shape("text", {x: 390, y: 350, "text-anchor": "middle", "font-size": 13}, s, metric === "cost" ? "USD per question" : "Total native tokens per question");
  const frontierKey = metric === "cost" ? "cost_frontier" : "token_frontier";
  const frontier = points.filter(r => r[frontierKey]).sort((a, b) => a[key] - b[key]);
  if (frontier.length > 1) shape("polyline", {points: frontier.map(r => `${x(r[key])},${y(r.accuracy)}`).join(" "), fill: "none", stroke: "#ba5a0a", "stroke-dasharray": "4 3"}, s);
  for (const r of points) {
    const c = shape("circle", {cx: x(r[key]), cy: y(r.accuracy), r: 6, fill: r[frontierKey] ? "#ba5a0a" : "#236b9e", opacity: 0.8}, s);
    shape("title", {}, c, `${r.model} / ${r.tier}: ${percent(r.accuracy)}, ${number(r[key], 6)}`);
  }
  if (!points.length) shape("text", {x: 170, y: 180}, s, "No complete results with known accounting");
}
function bars(rows) {
  const height = Math.max(160, rows.length * 27 + 50), s = svg("accuracy-chart", height);
  rows.forEach((r, i) => {
    const y = 18 + i * 27;
    shape("text", {x: 350, y: y + 15, "text-anchor": "end", "font-size": 12}, s, `${r.model} / ${r.tier}`);
    shape("rect", {x: 360, y, width: r.accuracy * 310, height: 20, fill: "#236b9e"}, s);
    shape("text", {x: 680, y: y + 15, "font-size": 12}, s, percent(r.accuracy));
  });
  if (!rows.length) shape("text", {x: 170, y: 80}, s, "No complete results for this selection");
}
function efforts() {
  const rows = data.rows.filter(r => matches(r, false, true) && r.accuracy != null);
  const models = [...new Set(rows.map(r => r.model))], tiers = ["minimal", "medium", "max"];
  const s = svg("effort-chart", 360 + models.length * 18);
  for (let i = 0; i <= 4; i++) shape("text", {x: 55, y: 305 - i * 62.5, "font-size": 12}, s, `${i * 25}%`);
  tiers.forEach((tier, i) => shape("text", {x: 140 + i * 240, y: 330, "text-anchor": "middle"}, s, tier));
  models.forEach((model, i) => {
    const points = tiers.map((tier, j) => {
      const row = rows.find(r => r.model === model && r.tier === tier);
      return row ? {x: 140 + j * 240, y: 300 - row.accuracy * 250, row} : null;
    });
    const color = palette[i % palette.length];
    for (let j = 1; j < points.length; j++) {
      const a = points[j - 1], b = points[j];
      if (a && b) shape("line", {x1: a.x, y1: a.y, x2: b.x, y2: b.y, stroke: color}, s);
    }
    for (const p of points.filter(Boolean)) {
      const c = shape("circle", {cx: p.x, cy: p.y, r: 5, fill: color}, s);
      shape("title", {}, c, `${model}: ${percent(p.row.accuracy)}`);
    }
    shape("text", {x: 80, y: 360 + i * 18, fill: color, "font-size": 12}, s, model);
  });
}
const columns = [
  ["model", "Model", r => r.model], ["tier", "Tier → sent", r => `${r.tier} → ${r.sent_efforts || "—"}`],
  ["status", "Status", r => r.status], ["accuracy", "Accuracy", r => percent(r.accuracy)],
  ["correct", "Correct / total", r => `${r.correct} / ${r.total}`], ["complete", "Completed", r => r.complete],
  ["incorrect", "Incorrect", r => r.incorrect], ["failed", "Unresolved", r => r.failed],
  ["cost_usd", "Cost USD", r => number(r.cost_usd, 6)], ["cost_per_question_usd", "USD / question", r => number(r.cost_per_question_usd, 6)],
  ["total_tokens", "Total tokens", r => number(r.total_tokens, 0)], ["prompt_tokens", "Prompt", r => number(r.prompt_tokens, 0)],
  ["completion_tokens", "Completion", r => number(r.completion_tokens, 0)], ["reasoning_tokens", "Reasoning", r => number(r.reasoning_tokens, 0)],
  ["cached_input_tokens", "Cached input", r => number(r.cached_input_tokens, 0)], ["pareto_3d", "3D Pareto", r => r.pareto_3d ? "Yes" : "—"]
];
const head = document.querySelector("#leaderboard thead tr");
for (const [key, label] of columns) {
  const b = element("button", label, element("th", null, head));
  b.onclick = () => { ascending = sortKey === key ? !ascending : false; sortKey = key; render(); };
}
function render() {
  const rows = data.rows.filter(r => matches(r)).sort((a, b) => {
    if (a[sortKey] == null) return b[sortKey] == null ? 0 : 1;
    if (b[sortKey] == null) return -1;
    const result = typeof a[sortKey] === "string" ? a[sortKey].localeCompare(b[sortKey]) : Number(a[sortKey]) - Number(b[sortKey]);
    return ascending ? result : -result;
  });
  const tbody = document.querySelector("#leaderboard tbody"); tbody.replaceChildren();
  for (const r of rows) {
    const tr = element("tr", null, tbody);
    for (const [, , format] of columns) element("td", format(r), tr);
  }
  document.getElementById("summary").textContent = `${rows.length} model/effort rows; ${rows.filter(r => r.status === "complete").length} complete and ranked.`;
  const ranked = chartRows(); bars(ranked); scatter(ranked, "cost"); scatter(ranked, "tokens"); efforts();
  const heat = document.getElementById("heatmap"); heat.replaceChildren();
  const table = element("table", null, heat), header = element("tr", null, table);
  element("th", "Model / effort", header);
  for (const category of data.categories) element("th", category, header).dir = "auto";
  for (const r of rows) {
    const tr = element("tr", null, table); element("td", `${r.model} / ${r.tier}`, tr);
    for (const category of data.categories) {
      const item = data.rows.find(v => v.model === r.model && v.tier === r.tier && v.cohort === r.cohort && v.dataset === r.dataset && v.category === category);
      const cell = element("td", item ? percent(item.accuracy) : "—", tr);
      if (item?.accuracy != null) cell.style.backgroundColor = `rgba(39,133,102,${item.accuracy * 0.35})`;
    }
  }
  const outcomes = document.querySelector("#outcomes tbody"); outcomes.replaceChildren();
  for (const o of data.outcomes) {
    if ((controls.dataset.value && o.dataset !== controls.dataset.value) || (controls.category.value && o.category !== controls.category.value) ||
        (controls.model.value && o.model !== controls.model.value) || (controls.tier.value && o.tier !== controls.tier.value) ||
        (controls.cohort.value === "text-only" && o.has_image) || (controls.cohort.value === "image-only" && !o.has_image)) continue;
    const tr = element("tr", null, outcomes);
    const values = [o.question, o.category, o.model, `${o.tier} → ${o.sent_effort || "—"}`, o.status,
      `${o.answer || "—"} / ${o.expected}`, o.provider || "—", [o.invalid, o.note, o.request_hash].filter(Boolean).join(" · ")];
    for (const value of values) element("td", value, tr).dir = "auto";
  }
}
for (const control of Object.values(controls)) control.addEventListener("change", render);
render();
