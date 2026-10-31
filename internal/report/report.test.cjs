const test = require("node:test");
const assert = require("node:assert/strict");
const { readFileSync } = require("node:fs");
const { join } = require("node:path");
const {
  modelBrand,
  modelLogo,
  modelDisplayName,
  modelMatchesQuery,
} = require("./logos.js");
const {
  pngDimensions,
  frontierLabelPositions,
  chartPNG,
  countAxisMaximum,
  heatmapColor,
} = require("./charts.js");
const {
  matchesScope,
  compareOutcomes,
  sortedRows,
  difficultyBins,
  rankSpan,
  separationExtreme,
} = require("./report.js");

test("readable model names preserve versions and support canonical and display-name searches", () => {
  assert.equal(
    modelDisplayName("anthropic/claude-opus-5.5"),
    "Claude Opus 5.5",
  );
  assert.equal(modelDisplayName("openai/gpt-6.1-sol"), "GPT-6.1 Sol");
  assert.equal(
    modelDisplayName("deepseek/deepseek-v4-pro-0813"),
    "DeepSeek V4 Pro 0813",
  );
  assert.equal(modelDisplayName("qwen/qwen3.8-max-0902"), "Qwen 3.8 Max 0902");
  assert.equal(
    modelDisplayName("unknown/raw-model-v8.1:beta"),
    "raw-model-v8.1:beta",
  );
  assert.equal(modelDisplayName("unqualified-model"), "unqualified-model");
  for (const query of [
    "claude opus",
    "  OPUS 5.5 ",
    "anthropic",
    "claude-opus-5.5",
    "",
  ]) {
    assert.ok(modelMatchesQuery("anthropic/claude-opus-5.5", query), query);
  }
  assert.equal(modelMatchesQuery("anthropic/claude-opus-5.5", "sonnet"), false);
});

test("model logos cover benchmark families and leave unknown models unbranded", () => {
  for (const [model, brand] of [
    ["openai/gpt-6.1-sol", "openai"],
    ["anthropic/claude-opus-5.5", "claude-color"],
    ["google/gemini-3.8-flash", "gemini-color"],
    ["deepseek/deepseek-v4-pro-0813", "deepseek-color"],
    ["qwen/qwen3.8-27b", "qwen-color"],
    ["x-ai/grok-4.7", "grok"],
    ["xiaomi/mimo-v2.6-pro", "xiaomimimo"],
    ["z-ai/glm-5.3", "zai"],
  ]) {
    assert.equal(modelBrand(model), brand);
    const svg = readFileSync(join(__dirname, "logos", `${brand}.svg`), "utf8");
    assert.match(svg, /^<svg\b/);
    assert.match(svg, /viewBox="0 0 24 24"/);
    assert.doesNotMatch(
      svg,
      /<script\b|<foreignObject\b|\son\w+=|(?:href|src)=["'](?!#|data:)/i,
    );
    const source = `data:image/svg+xml;base64,${Buffer.from(svg).toString("base64")}`;
    assert.equal(modelLogo(model, { [brand]: source }), source);
    assert.equal(modelLogo(model, {}), "");
  }
  for (const model of [
    null,
    "",
    "unknown/gemini-3",
    "x-ai/not-grok",
    "google/other",
    "toString",
    "__proto__",
  ]) {
    assert.equal(modelBrand(model), null);
    assert.equal(modelLogo(model, { openai: "unused" }), "");
  }
});

test("PNG exports use double resolution while bounding tall and oversized canvases", () => {
  assert.deepEqual(pngDimensions(1200, 760), { width: 2400, height: 1520 });
  for (const [w, h] of [
    [1200, 50000],
    [50000, 1200],
    [20000, 20000],
  ]) {
    const size = pngDimensions(w, h);
    assert.ok(size.width <= 16384 && size.height <= 16384);
    assert.ok(size.width * size.height <= 32000000);
    assert.ok(Math.abs(size.width / size.height - w / h) / (w / h) < 0.002);
  }
  for (const value of [0, -1, NaN, Infinity])
    assert.throws(() => pngDimensions(value, 100), /invalid dimensions/);
});

test("direct frontier labels stay near their points without collisions or clipping", () => {
  const bounds = { left: 96, right: 1158, top: 190, bottom: 618 };
  const targets = [
    { x: 854, y: 297, width: 192, height: 42 },
    { x: 591, y: 310, width: 166, height: 42 },
    { x: 538, y: 361, width: 166, height: 42 },
    { x: 487, y: 399, width: 194, height: 42 },
    { x: 335, y: 425, width: 158, height: 42 },
    { x: 185, y: 476, width: 158, height: 42 },
  ];
  const original = structuredClone(targets);
  const points = [...targets, { x: 570, y: 361 }, { x: 650, y: 336 }];
  const segments = targets.slice(1).map((p, i) => [targets[i], p]);
  const positions = frontierLabelPositions(targets, bounds, points, segments);
  const overlaps = (a, b) =>
    a.x < b.x + b.width &&
    a.x + a.width > b.x &&
    a.y < b.y + b.height &&
    a.y + a.height > b.y;
  assert.equal(positions.length, targets.length);
  for (const [i, box] of positions.entries()) {
    assert.ok(box.x >= bounds.left && box.x + box.width <= bounds.right);
    assert.ok(box.y >= bounds.top && box.y + box.height <= bounds.bottom);
    assert.ok(positions.every((other, j) => i === j || !overlaps(box, other)));
    assert.ok(
      points.every(
        (p) =>
          !overlaps(box, { x: p.x - 10, y: p.y - 10, width: 20, height: 20 }),
      ),
    );
    const point = targets[i];
    assert.ok(
      Math.hypot(
        Math.max(box.x - point.x, 0, point.x - box.x - box.width),
        Math.max(box.y - point.y, 0, point.y - box.y - box.height),
      ) <= 70,
      "a direct label must remain next to its own marker",
    );
    for (const [a, b] of segments) {
      for (let t = 0; t <= 1; t += 0.01) {
        const x = a.x + (b.x - a.x) * t,
          y = a.y + (b.y - a.y) * t;
        assert.ok(
          !(
            x > box.x &&
            x < box.x + box.width &&
            y > box.y &&
            y < box.y + box.height
          ),
          "labels must not obscure the frontier",
        );
      }
    }
  }
  assert.deepEqual(
    targets,
    original,
    "label placement must not move data points",
  );
  assert.deepEqual(
    frontierLabelPositions(targets, bounds, points, segments),
    positions,
    "placement must be deterministic",
  );
});

test("direct frontier labels handle empty panels, ties, and plot edges", () => {
  const bounds = { left: 0, right: 1000, top: 0, bottom: 500 };
  assert.deepEqual(frontierLabelPositions([], bounds, []), []);
  for (const points of [
    [
      { x: 0, y: 0 },
      { x: 1000, y: 500 },
    ],
    [
      { x: 500, y: 250 },
      { x: 500, y: 250 },
    ],
  ]) {
    const positions = frontierLabelPositions(
      points.map((p) => ({ ...p, width: 180, height: 42 })),
      bounds,
      points,
    );
    for (const box of positions) {
      assert.ok(
        box.x >= 0 &&
          box.y >= 0 &&
          box.x + box.width <= 1000 &&
          box.y + box.height <= 500,
      );
      assert.ok(
        points.every(
          (p) =>
            p.x < box.x ||
            p.x > box.x + box.width ||
            p.y < box.y ||
            p.y > box.y + box.height,
        ),
      );
    }
    const [a, b] = positions;
    assert.ok(
      a.x + a.width <= b.x ||
        b.x + b.width <= a.x ||
        a.y + a.height <= b.y ||
        b.y + b.height <= a.y,
    );
  }
  assert.throws(
    () =>
      frontierLabelPositions(
        [{ x: 0, y: 0, width: 1100, height: 42 }],
        bounds,
        [],
      ),
    /label is larger than the plot/,
  );
});

test("count axes have integer ticks and heatmap text meets 4.5:1 contrast", () => {
  for (const n of [0, 1, 6, 23, 80, 1001]) {
    const max = countAxisMaximum(n);
    assert.ok(max >= n && max >= 4);
    assert.equal(max % 4, 0);
  }
  for (let i = 0; i <= 100; i++) {
    const color = heatmapColor(i / 100);
    const rgb = color.fill.match(/\d+/g).map(Number);
    const linear = rgb.map((n) =>
      n / 255 <= 0.04045 ? n / 255 / 12.92 : ((n / 255 + 0.055) / 1.055) ** 2.4,
    );
    const light = linear[0] * 0.2126 + linear[1] * 0.7152 + linear[2] * 0.0722;
    const contrast =
      color.text === "#fff" ? 1.05 / (light + 0.05) : (light + 0.05) / 0.05;
    assert.ok(contrast >= 4.5, `accuracy=${i}%, contrast=${contrast}`);
  }
});

test("PNG encoding preserves the source SVG and releases temporary URLs on success and failure", async () => {
  for (const mode of ["success", "image-error", "no-canvas", "encode-error"]) {
    const revoked = [],
      cloneAttributes = {},
      draws = [];
    const result = new Blob(["encoded"], { type: "image/png" });
    let fontsReady = false;
    const copy = {
      setAttribute: (k, v) => {
        cloneAttributes[k] = v;
      },
      removeAttribute: () => {},
      querySelectorAll: () => [],
    };
    const svg = {
      viewBox: { baseVal: { width: 1200, height: 760 } },
      cloneNode: () => copy,
    };
    const canvas = {
      getContext: () =>
        mode === "no-canvas"
          ? null
          : { fillRect: () => {}, drawImage: (...args) => draws.push(args) },
      toBlob: (callback) => callback(mode === "encode-error" ? null : result),
    };
    const doc = {
      fonts: {
        ready: Promise.resolve().then(() => {
          fontsReady = true;
        }),
      },
      createElement: () => canvas,
      defaultView: {
        Blob,
        setTimeout,
        clearTimeout,
        URL: {
          createObjectURL: () => "blob:chart",
          revokeObjectURL: (url) => revoked.push(url),
        },
        XMLSerializer: class {
          serializeToString(node) {
            assert.ok(
              fontsReady,
              "export must wait for the chart's font metrics",
            );
            assert.equal(node, copy);
            return "<svg/>";
          }
        },
        Image: class {
          set src(_value) {
            queueMicrotask(() =>
              mode === "image-error" ? this.onerror() : this.onload(),
            );
          }
        },
      },
    };
    if (mode === "success") {
      assert.equal(await chartPNG(svg, doc), result);
      assert.deepEqual([canvas.width, canvas.height], [2400, 1520]);
      assert.deepEqual(draws[0].slice(1), [0, 0, 2400, 1520]);
    } else
      await assert.rejects(
        chartPNG(svg, doc),
        /render this chart|Canvas 2D|PNG encoding failed/,
      );
    assert.deepEqual(revoked, ["blob:chart"]);
    assert.equal(cloneAttributes.xmlns, "http://www.w3.org/2000/svg");
    assert.equal(cloneAttributes.width, 1200);
    assert.equal(svg.width, undefined);
  }
});

test("difficulty histogram includes zero and perfect scores but not unknown panels", () => {
  assert.deepEqual(
    difficultyBins([
      { correct: 0, panel: 0 },
      { correct: 0, panel: 4 },
      { correct: 1, panel: 4 },
      { correct: 2, panel: 4 },
      { correct: 3, panel: 4 },
      { correct: 4, panel: 4 },
    ]),
    [1, 1, 1, 2],
  );
});

test("embedded math fonts preserve every complete CSS font-face rule", () => {
  const css = readFileSync(join(__dirname, "math.generated.css"), "utf8");
  const faces = [...css.matchAll(/@font-face\{[^{}]*\}/g)];
  assert.ok(faces.length >= 20);
  assert.equal(faces.length, [...css.matchAll(/@font-face/g)].length);
  for (const [face] of faces)
    assert.match(
      face,
      /src:url\(data:font\/woff2;base64,[A-Za-z0-9+/=]+\) format\('woff2'\)/,
    );
  assert.ok(!css.includes("url(fonts/"));
});

test("cohort, dataset and category filters preserve their intersection", () => {
  const o = { dataset: "exam", category: "math", has_image: true };
  assert.equal(
    matchesScope(o, { cohort: "full", dataset: "", category: "" }),
    true,
  );
  assert.equal(
    matchesScope(o, { cohort: "text-only", dataset: "", category: "" }),
    false,
  );
  assert.equal(
    matchesScope(o, {
      cohort: "image-only",
      dataset: "exam",
      category: "math",
    }),
    true,
  );
  assert.equal(
    matchesScope(o, {
      cohort: "image-only",
      dataset: "other",
      category: "math",
    }),
    false,
  );
});
test("paired comparison uses matching completed questions and counts invalid as wrong", () => {
  const o = (question, correct, answer, status = "complete", invalid = "") => ({
    question,
    correct,
    answer,
    status,
    invalid,
  });
  const a = [
    o("1", true, 1),
    o("2", true, 2),
    o("3", false, 4),
    o("4", false, 0, "complete", "invalid"),
    o("5", true, 1),
  ];
  const b = [
    o("4", false, 0, "complete", "invalid"),
    o("3", true, 3),
    o("2", false, 3),
    o("1", true, 1),
    o("5", false, 0, "infra_error"),
  ];
  assert.deepEqual(compareOutcomes(a, b), {
    n: 4,
    both: ["1"],
    a: ["2"],
    b: ["3"],
    neither: ["4"],
    agreement: 1,
  });
  assert.equal(compareOutcomes([], b).n, 0);
});
test("sorting keeps unknown accounting last in either direction and preserves zero", () => {
  const rows = [
    { model: "a", tier: "max", cost: null },
    { model: "b", tier: "max", cost: 0 },
    { model: "c", tier: "max", cost: 1 },
  ];
  assert.deepEqual(
    sortedRows(rows, "cost", true).map((r) => r.model),
    ["b", "c", "a"],
  );
  assert.deepEqual(
    sortedRows(rows, "cost", false).map((r) => r.model),
    ["c", "b", "a"],
  );
  assert.equal(rows[0].model, "a");
});
test("rank intervals collapse equal bounds and name every tied extreme", () => {
  assert.equal(rankSpan({ low: 1, high: 9 }), "1–9");
  assert.equal(rankSpan({ low: 3, high: 3 }), "3");
  const panel = [
    { model: "a/x", tier: "max" },
    { model: "b/y", tier: "minimal" },
    { model: "c/z", tier: "medium" },
  ];
  const group = { questions: 8, correct: [8, 0, 8] };
  assert.deepEqual(separationExtreme(group, panel, Math.max), {
    accuracy: 1,
    configurations: [panel[0], panel[2]],
  });
  assert.deepEqual(separationExtreme(group, panel, Math.min), {
    accuracy: 0,
    configurations: [panel[1]],
  });
  assert.equal(
    separationExtreme({ questions: 1, correct: [] }, [], Math.max),
    null,
  );
});
