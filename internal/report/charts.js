"use strict";

function countAxisMaximum(max) {
  return Math.max(4, Math.ceil(max / 4) * 4);
}

function heatmapColor(value) {
  const rgb = [239 - value * 231, 246 - value * 129, 241 - value * 137].map(
    Math.round,
  );
  const linear = rgb.map((n) => {
    const s = n / 255;
    return s <= 0.04045 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
  });
  const luminance =
    linear[0] * 0.2126 + linear[1] * 0.7152 + linear[2] * 0.0722;
  return {
    fill: `rgb(${rgb.join(",")})`,
    text: luminance > 0.179 ? "#000" : "#fff",
  };
}

// Bound canvas allocations without cropping tall figures or changing aspect ratio.
function pngDimensions(width, height) {
  if (!(width > 0 && height > 0 && Number.isFinite(width + height)))
    throw new Error("The chart has invalid dimensions.");
  const scale = Math.min(
    2,
    16384 / width,
    16384 / height,
    Math.sqrt(32000000 / (width * height)),
  );
  return {
    width: Math.max(1, Math.floor(width * scale)),
    height: Math.max(1, Math.floor(height * scale)),
  };
}

// Place compact labels next to their own data points. Candidate positions are
// scored against markers, the frontier, and other labels; data never moves.
function frontierLabelPositions(labels, bounds, points, segments = []) {
  const overlap = (a, b, pad = 6) =>
    Math.max(
      0,
      Math.min(a.x + a.width + pad, b.x + b.width) - Math.max(a.x - pad, b.x),
    ) *
    Math.max(
      0,
      Math.min(a.y + a.height + pad, b.y + b.height) - Math.max(a.y - pad, b.y),
    );
  const crosses = (box, [a, b]) => {
    let enter = 0,
      leave = 1;
    for (const [axis, extent] of [
      ["x", "width"],
      ["y", "height"],
    ]) {
      const delta = b[axis] - a[axis];
      const low = box[axis] - 4,
        high = box[axis] + box[extent] + 4;
      if (delta === 0) {
        if (a[axis] < low || a[axis] > high) return false;
      } else {
        const t1 = (low - a[axis]) / delta,
          t2 = (high - a[axis]) / delta;
        enter = Math.max(enter, Math.min(t1, t2));
        leave = Math.min(leave, Math.max(t1, t2));
        if (enter > leave) return false;
      }
    }
    return true;
  };
  const candidates = labels.map(({ x, y, width, height }) => {
    if (
      width > bounds.right - bounds.left ||
      height > bounds.bottom - bounds.top
    )
      throw new Error("The frontier label is larger than the plot.");
    return [16, 32, 48].flatMap((gap) =>
      [
        [x + gap, y - height - gap],
        [x - width - gap, y - height - gap],
        [x + gap, y + gap],
        [x - width - gap, y + gap],
        [x - width / 2, y - height - gap],
        [x - width / 2, y + gap],
        [x + gap, y - height / 2],
        [x - width - gap, y - height / 2],
      ].map(([left, top], preference) => {
        const box = {
          x: Math.max(bounds.left, Math.min(bounds.right - width, left)),
          y: Math.max(bounds.top, Math.min(bounds.bottom - height, top)),
          width,
          height,
        };
        const distance = Math.hypot(
          Math.max(box.x - x, 0, x - box.x - width),
          Math.max(box.y - y, 0, y - box.y - height),
        );
        const collisions = points.filter((p) =>
          overlap(
            box,
            {
              x: p.x - 10,
              y: p.y - 10,
              width: 20,
              height: 20,
            },
            3,
          ),
        ).length;
        return {
          ...box,
          penalty:
            collisions * 100000 +
            segments.filter((line) => crosses(box, line)).length * 10000 +
            distance +
            preference * 2,
        };
      }),
    );
  });
  const placed = [];
  // A few deterministic relaxation passes resolve close frontier neighbours.
  for (let pass = 0; pass < 4; pass++) {
    candidates.forEach((options, i) => {
      let best,
        bestScore = Infinity;
      for (const option of options) {
        const score =
          option.penalty +
          placed.reduce(
            (sum, other, j) =>
              sum + (i === j ? 0 : overlap(option, other) * 1000),
            0,
          );
        if (score < bestScore) {
          best = option;
          bestScore = score;
        }
      }
      placed[i] = best;
    });
  }
  return placed.map(({ x, y, width, height }) => ({ x, y, width, height }));
}

async function chartPNG(svg, doc = svg.ownerDocument) {
  if (doc.fonts) await doc.fonts.ready;
  const win = doc.defaultView,
    box = svg.viewBox.baseVal;
  const size = pngDimensions(box.width, box.height);
  const copy = svg.cloneNode(true);
  copy.setAttribute("xmlns", "http://www.w3.org/2000/svg");
  copy.setAttribute("width", box.width);
  copy.setAttribute("height", box.height);
  copy.removeAttribute("class");
  copy
    .querySelectorAll("[tabindex]")
    .forEach((node) => node.removeAttribute("tabindex"));
  const source = new win.XMLSerializer().serializeToString(copy);
  const url = win.URL.createObjectURL(
    new win.Blob([source], { type: "image/svg+xml;charset=utf-8" }),
  );
  try {
    const img = new win.Image();
    await new Promise((resolve, reject) => {
      const timeout = win.setTimeout(
        () => reject(new Error("Timed out rendering the chart image.")),
        15000,
      );
      img.onload = () => {
        win.clearTimeout(timeout);
        resolve();
      };
      img.onerror = () => {
        win.clearTimeout(timeout);
        reject(new Error("The browser could not render this chart."));
      };
      img.src = url;
    });
    const canvas = doc.createElement("canvas");
    canvas.width = size.width;
    canvas.height = size.height;
    const ctx = canvas.getContext("2d");
    if (!ctx) throw new Error("PNG export requires Canvas 2D support.");
    ctx.fillStyle = "#ffffff";
    ctx.fillRect(0, 0, size.width, size.height);
    ctx.drawImage(img, 0, 0, size.width, size.height);
    return await new Promise((resolve, reject) =>
      canvas.toBlob(
        (blob) =>
          blob
            ? resolve(blob)
            : reject(
                new Error("PNG encoding failed. Try a smaller selection."),
              ),
        "image/png",
      ),
    );
  } finally {
    win.URL.revokeObjectURL(url);
  }
}

function createReportCharts(doc, data, getScope, getFilters) {
  const ink = "#182f38",
    muted = "#52656d",
    grid = "#e2e9e5";
  const colors = { minimal: "#4e6e9b", medium: "#087568", max: "#a95427" };
  const format = (v, digits = 2) =>
    Number(v).toLocaleString("en-US", { maximumFractionDigits: digits });
  const percent = (v) => `${format(v * 100)}%`;
  const money = (v) =>
    `$${format(v, v > 0 && v < 0.01 ? Math.max(5, 2 - Math.floor(Math.log10(v))) : 4)}`;
  const measure = doc.createElement("canvas").getContext("2d");

  function shape(tag, attrs, parent, value) {
    const e = doc.createElementNS("http://www.w3.org/2000/svg", tag);
    if (tag === "text")
      attrs = {
        "font-family": "Inter, Arial, sans-serif",
        "font-kerning": "normal",
        "font-size": 14,
        fill: muted,
        ...attrs,
      };
    for (const [key, v] of Object.entries(attrs)) e.setAttribute(key, v);
    if (value != null) e.textContent = value;
    parent.append(e);
    return e;
  }
  function text(s, x, y, value, attrs = {}) {
    return shape("text", { x, y, ...attrs }, s, value);
  }
  function modelText(s, x, y, model, attrs = {}, suffix = "") {
    const label = text(s, x, y, modelDisplayName(model) + suffix, {
      "font-size": 15,
      "font-weight": 550,
      "letter-spacing": 0.15,
      fill: ink,
      "data-model-name": model,
      ...attrs,
    });
    shape("title", {}, label, model);
    return label;
  }
  function logo(s, model, x, y, size = 26) {
    const source = modelLogo(model);
    if (source)
      shape(
        "image",
        {
          x,
          y,
          width: size,
          height: size,
          href: source,
          "aria-hidden": "true",
          "data-model-logo": model,
        },
        s,
      );
  }
  function wrap(value, width, size = 14) {
    if (measure) measure.font = `${size}px Inter, Arial, sans-serif`;
    const fits = (s) =>
      (measure ? measure.measureText(s).width : s.length * size * 0.6) <= width;
    const lines = [];
    let line = "";
    for (const word of String(value).split(/\s+/)) {
      if (line && !fits(`${line} ${word}`)) {
        lines.push(line);
        line = "";
      }
      for (const char of (line ? " " : "") + word) {
        if (line && !fits(line + char)) {
          lines.push(line);
          line = "";
        }
        line += char;
      }
    }
    if (line) lines.push(line);
    return lines;
  }
  function lines(s, x, y, value, width, attrs = {}, step = 18) {
    const all = wrap(value, width, attrs["font-size"] || 14);
    all.forEach((line, i) => text(s, x, y + i * step, line, attrs));
    return all.length;
  }
  function scopeText(override = {}) {
    const scope = { ...getScope(), ...override };
    return [
      scope.dataset || (data.datasets || []).join(" + ") || "All datasets",
      {
        full: "Full exam",
        "text-only": "Text-only",
        "image-only": "Image-only",
      }[scope.cohort] || scope.cohort,
      scope.category || "All subjects",
    ].join("  ·  ");
  }
  function chart(id, title, height, options = {}) {
    const {
      width = 1200,
      subtitle = "",
      footer = "",
      display = false,
      scope = {},
      suffix = "",
    } = options;
    const container = doc.getElementById(id);
    container.replaceChildren();
    const toolbar = doc.createElement("div");
    toolbar.className = "chart-toolbar";
    container.append(toolbar);
    const button = doc.createElement("button");
    button.type = "button";
    button.className = "secondary chart-download";
    button.textContent = "Download PNG";
    button.setAttribute("aria-label", `Download PNG: ${title}`);
    toolbar.append(button);
    const status = doc.createElement("span");
    status.className = "note";
    status.setAttribute("role", "status");
    toolbar.append(status);
    const scroll = doc.createElement("div");
    scroll.className = "chart-scroll";
    scroll.tabIndex = 0;
    scroll.setAttribute(
      "aria-label",
      `${title}; scroll horizontally on small screens`,
    );
    container.append(scroll);
    scroll.style.setProperty("--chart-min-width", `${Math.min(width, 1000)}px`);
    const s = shape(
      "svg",
      {
        viewBox: `0 0 ${width} ${height}`,
        width,
        height,
        class: "report-chart",
        role: "img",
        "aria-label": title,
        "font-family": "Inter, Arial, sans-serif",
        "font-size": 14,
        fill: ink,
      },
      scroll,
    );
    shape("title", {}, s, title);
    const fontCSS = doc.getElementById("chart-fonts")?.textContent;
    if (fontCSS) shape("style", { "data-chart-font": "true" }, s, fontCSS);
    shape("rect", { x: 0, y: 0, width, height, fill: "#fff" }, s);
    text(s, 32, 30, "KONKOUR BENCH / RESULTS", {
      "font-size": 11,
      "letter-spacing": 1.5,
      fill: "#087568",
    });
    text(s, 32, 63, title, {
      "font-size": 26,
      "font-weight": 650,
      "letter-spacing": -0.35,
      fill: ink,
    });
    const selection = { ...getScope(), ...scope },
      filters = display ? getFilters() : "";
    let context = scopeText(selection);
    if (display) {
      if (filters) context += `  ·  Display: ${filters}`;
    }
    lines(s, 32, 88, context, width - 64, { "font-size": 13 }, 17);
    text(s, 32, 127, subtitle, { "font-size": 13 });
    shape(
      "line",
      {
        x1: 32,
        x2: width - 32,
        y1: height - 64,
        y2: height - 64,
        stroke: grid,
      },
      s,
    );
    lines(s, 32, height - 42, footer, width - 64, { "font-size": 12 }, 16);
    const dimensions = pngDimensions(width, height);
    button.title = `Save ${dimensions.width} × ${dimensions.height} pixels, with title, labels, and current selection`;
    button.onclick = async () => {
      button.disabled = true;
      status.textContent = "Preparing PNG…";
      try {
        const blob = await chartPNG(s, doc),
          win = doc.defaultView,
          url = win.URL.createObjectURL(blob);
        const a = doc.createElement("a");
        a.href = url;
        a.download =
          [
            "konkour",
            id,
            selection.cohort,
            selection.dataset ||
              (data.datasets || []).join("+") ||
              "all-datasets",
            selection.category,
            filters,
            suffix,
          ]
            .filter(Boolean)
            .join("-")
            .replace(/[^\p{L}\p{N}._-]+/gu, "-") + ".png";
        doc.body.append(a);
        a.click();
        a.remove();
        win.setTimeout(() => win.URL.revokeObjectURL(url), 1000);
        status.textContent = `Saved ${dimensions.width} × ${dimensions.height} PNG`;
      } catch (err) {
        console.error("Chart PNG export failed", id, err);
        status.textContent = `Export failed: ${err.message}`;
      } finally {
        button.disabled = false;
      }
    };
    return s;
  }
  function empty(id, message) {
    const container = doc.getElementById(id);
    container.replaceChildren();
    const p = doc.createElement("p");
    p.className = "empty";
    p.textContent = message;
    container.append(p);
  }
  function marker(s, x, y, tier, size = 6, active = true) {
    const attrs = {
      fill: active ? colors[tier] || colors.medium : "#fff",
      stroke: colors[tier] || colors.medium,
      "stroke-width": active ? 2 : 1.6,
    };
    if (tier === "medium")
      return shape(
        "rect",
        {
          x: x - size,
          y: y - size,
          width: size * 2,
          height: size * 2,
          rx: 1.5,
          ...attrs,
        },
        s,
      );
    if (tier === "max")
      return shape(
        "path",
        {
          d: `M${x} ${y - size - 1}L${x + size + 1} ${y}L${x} ${y + size + 1}L${x - size - 1} ${y}Z`,
          ...attrs,
        },
        s,
      );
    return shape("circle", { cx: x, cy: y, r: size, ...attrs }, s);
  }
  function legend(s, x, y) {
    ["minimal", "medium", "max"].forEach((tier, i) => {
      marker(s, x + i * 128, y - 4, tier, 5);
      text(s, x + 13 + i * 128, y, tier, { "font-size": 13 });
    });
  }

  function accuracy(rows) {
    if (!rows.length)
      return empty(
        "accuracy-chart",
        "No complete configurations match these display filters.",
      );
    const height = 265 + rows.length * 40;
    const s = chart(
      "accuracy-chart",
      "Accuracy by model and reasoning effort",
      height,
      {
        display: true,
        subtitle:
          "Point: observed accuracy  ·  Whisker: 95% Wilson interval  ·  Equal question weights",
        footer:
          "Intervals assume independent question sampling; they do not measure run-to-run variation. Ties retain equal rank.",
      },
    );
    const x = (v) => 410 + v * 620,
      bottom = 180 + (rows.length - 1) * 40 + 20;
    text(s, 32, 158, "MODEL / REQUESTED EFFORT", {
      "font-size": 11,
      "font-weight": 700,
    });
    text(s, 1075, 158, "SCORE", {
      "font-size": 11,
      "text-anchor": "middle",
      "font-weight": 700,
    });
    text(s, 1170, 158, "CORRECT", {
      "font-size": 11,
      "text-anchor": "end",
      "font-weight": 700,
    });
    for (let i = 0; i <= 4; i++) {
      const v = i / 4;
      shape(
        "line",
        {
          x1: x(v),
          x2: x(v),
          y1: 165,
          y2: bottom,
          stroke: i === 1 ? "#b6c7bf" : grid,
          "stroke-dasharray": i === 1 ? "4 4" : "none",
        },
        s,
      );
      text(s, x(v), 154, percent(v), {
        "font-size": 12,
        "text-anchor": "middle",
      });
    }
    rows.forEach((r, i) => {
      const y = 184 + i * 40;
      logo(s, r.model, 32, y - 16);
      modelText(s, 70, y - 2, r.model);
      text(s, 70, y + 14, `${r.model.split("/")[0]} · ${r.tier}`, {
        "font-size": 11,
      });
      if (r.accuracy_interval) {
        const lo = x(r.accuracy_interval.low),
          hi = x(r.accuracy_interval.high),
          color = colors[r.tier];
        shape(
          "line",
          {
            x1: lo,
            x2: hi,
            y1: y,
            y2: y,
            stroke: color,
            "stroke-width": 2.5,
            opacity: 0.7,
          },
          s,
        );
        for (const px of [lo, hi])
          shape(
            "line",
            {
              x1: px,
              x2: px,
              y1: y - 4,
              y2: y + 4,
              stroke: color,
              "stroke-width": 1.5,
            },
            s,
          );
      }
      marker(s, x(r.accuracy), y, r.tier, 5);
      text(s, 1075, y + 5, percent(r.accuracy), {
        "font-size": 15,
        "font-weight": 700,
        "text-anchor": "middle",
        fill: ink,
      });
      text(s, 1170, y + 5, `${r.correct}/${r.total}`, {
        "font-size": 13,
        "text-anchor": "end",
      });
    });
    text(
      s,
      410,
      bottom + 28,
      "Dashed guide: 25% uniform-guessing expectation",
      { "font-size": 12 },
    );
  }

  function scatter(rows, metric) {
    const metricKey =
        metric === "cost" ? "cost_per_question_usd" : "tokens_per_question",
      frontierKey = metric === "cost" ? "cost_frontier" : "token_frontier";
    const points = rows.filter((r) => r[metricKey] > 0);
    if (!points.length)
      return empty(
        `${metric}-chart`,
        "No complete configurations with positive known accounting.",
      );
    const frontier = points
      .filter((r) => r[frontierKey])
      .sort(
        (a, b) =>
          b.accuracy - a.accuracy ||
          b[metricKey] - a[metricKey] ||
          a.model.localeCompare(b.model),
      );
    const height = Math.max(820, 400 + frontier.length * 70),
      left = 96,
      right = 1148,
      top = 220,
      bottom = height - 190;
    const s = chart(
      `${metric}-chart`,
      metric === "cost" ? "Accuracy and API cost" : "Accuracy and token use",
      height,
      {
        display: true,
        subtitle: `${points.length} complete configurations  ·  Colored points: Pareto frontier  ·  Gray points: other configurations`,
        footer:
          metric === "cost"
            ? "Actual API charges only; subscription usage and unknown charges are excluded. Frontier membership uses all configurations in this scope."
            : "Native tokenizers differ between models. Reasoning is included in completion tokens. Frontier membership uses all configurations in this scope.",
      },
    );
    const low = Math.floor(
        Math.log10(Math.min(...points.map((r) => r[metricKey]))),
      ),
      high = Math.max(
        low + 1,
        Math.ceil(Math.log10(Math.max(...points.map((r) => r[metricKey])))),
      );
    const yMin = Math.max(
      0,
      Math.floor(Math.min(...points.map((r) => r.accuracy)) * 10 - 0.5) / 10,
    );
    const x = (v) =>
        left + ((Math.log10(v) - low) / (high - low)) * (right - left),
      y = (v) => bottom - ((v - yMin) / (1 - yMin)) * (bottom - top);
    text(s, left, 179, "ACCURACY ↑", {
      "font-size": 11,
      "font-weight": 650,
      "letter-spacing": 0.8,
    });
    const yStep = 1 - yMin > 0.5 ? 0.2 : 0.1;
    const yTicks = [yMin, 1];
    for (
      let v = Math.ceil((yMin + 1e-8) / yStep) * yStep;
      v < 1 - 1e-8;
      v += yStep
    )
      yTicks.push(v);
    for (const v of yTicks) {
      shape(
        "line",
        { x1: left, x2: right, y1: y(v), y2: y(v), stroke: "#eaf0ed" },
        s,
      );
      text(s, left - 18, y(v) + 5, percent(v), {
        "text-anchor": "end",
        "font-size": 13,
      });
    }
    for (let i = low; i <= high; i++) {
      shape(
        "line",
        {
          x1: x(10 ** i),
          x2: x(10 ** i),
          y1: bottom,
          y2: bottom + 6,
          stroke: "#b9c8c1",
        },
        s,
      );
      text(
        s,
        x(10 ** i),
        bottom + 28,
        metric === "cost" ? money(10 ** i) : format(10 ** i, 0),
        { "font-size": 13, "text-anchor": "middle" },
      );
    }
    text(
      s,
      (left + right) / 2,
      bottom + 59,
      metric === "cost"
        ? "USD per question · logarithmic scale"
        : "Native tokens per question · logarithmic scale",
      { "text-anchor": "middle", "font-size": 14 },
    );
    legend(s, left + 5, bottom + 102);
    text(
      s,
      right,
      bottom + 102,
      metric === "cost"
        ? "Lower cost and higher accuracy are better"
        : "Fewer tokens and higher accuracy are better",
      { "text-anchor": "end", "font-size": 12 },
    );

    if (frontier.length > 1)
      shape(
        "polyline",
        {
          points: [...frontier]
            .reverse()
            .map((r) => `${x(r[metricKey])},${y(r.accuracy)}`)
            .join(" "),
          fill: "none",
          stroke: "#147d6c",
          "stroke-width": 2.75,
          "stroke-linejoin": "round",
          "data-frontier-line": "true",
        },
        s,
      );
    // Draw frontier markers last so coincident comparison points cannot hide them.
    for (const r of [...points.filter((r) => !r[frontierKey]), ...frontier]) {
      const title = `${modelDisplayName(r.model)} / ${r.tier}: ${percent(r.accuracy)} · ${metric === "cost" ? money(r[metricKey]) : format(r[metricKey], 0)} per question${r[frontierKey] ? " · Pareto frontier" : ""} (${r.model})`;
      const dot = marker(
        s,
        x(r[metricKey]),
        y(r.accuracy),
        r.tier,
        r[frontierKey] ? 8 : 5,
      );
      dot.setAttribute("stroke", r[frontierKey] ? "#ffffff" : "#b4c2bc");
      dot.setAttribute("stroke-width", r[frontierKey] ? "2.5" : "1.5");
      if (!r[frontierKey]) dot.setAttribute("fill", "#f1f5f3");
      dot.setAttribute(
        "data-pareto-point",
        r[frontierKey] ? "frontier" : "comparison",
      );
      dot.setAttribute("tabindex", "0");
      dot.setAttribute("aria-label", title);
      shape("title", {}, dot, title);
    }
    const labels = frontier.map((r) => {
      const name = modelDisplayName(r.model),
        detail = `${r.tier} · ${percent(r.accuracy)}`;
      if (measure) measure.font = "600 16px Inter, Arial, sans-serif";
      const nameWidth = measure
        ? measure.measureText(name).width + name.length * 0.1
        : name.length * 10;
      if (measure) measure.font = "400 12px Inter, Arial, sans-serif";
      const detailWidth = measure
        ? measure.measureText(detail).width
        : detail.length * 8;
      return {
        x: x(r[metricKey]),
        y: y(r.accuracy),
        width: Math.ceil(Math.max(nameWidth, detailWidth) + 32),
        height: 42,
        detail,
      };
    });
    const positions = frontierLabelPositions(
      labels,
      { left, right: right + 10, top: 190, bottom: bottom - 12 },
      points.map((r) => ({ x: x(r[metricKey]), y: y(r.accuracy) })),
      labels.slice(1).map((p, i) => [labels[i], p]),
    );
    frontier.forEach((r, i) => {
      const { x: lx, y: ly } = positions[i];
      const label = shape(
        "g",
        {
          "data-frontier-label": "true",
          "aria-label": `${modelDisplayName(r.model)}, ${labels[i].detail}`,
        },
        s,
      );
      logo(label, r.model, lx, ly + 1, 22);
      const halo = {
        stroke: "#fff",
        "stroke-width": 5,
        "stroke-linejoin": "round",
        "paint-order": "stroke fill",
      };
      modelText(label, lx + 30, ly + 18, r.model, {
        "font-size": 16,
        "font-weight": 600,
        "letter-spacing": 0.1,
        ...halo,
      });
      text(label, lx + 30, ly + 37, labels[i].detail, {
        "font-size": 12,
        ...halo,
      });
    });
    if (!frontier.length)
      text(s, right, 179, "No frontier points match the display filters", {
        "text-anchor": "end",
        "font-size": 12,
      });
  }

  function effort(rows, from, to) {
    if (!rows.length)
      return empty(
        "effort-chart",
        "No completed question pairs for these tiers in this scope.",
      );
    const height = 285 + rows.length * 48,
      bottom = height - 114;
    const s = chart(
      "effort-chart",
      `Reasoning effort: ${from} → ${to}`,
      height,
      {
        subtitle:
          "Net accuracy change on the same questions · Gains minus regressions, in percentage points",
        footer:
          "* Partial pair: only jointly completed questions are compared; missing questions can bias the result. Requested tiers are not equal budgets across models.",
        suffix: `${from}-${to}`,
      },
    );
    const extent =
      Math.ceil(Math.max(0.05, ...rows.map((r) => Math.abs(r.delta))) * 20) /
      20;
    const x = (v) => 750 + (v / extent) * 230;
    text(s, 32, 159, "MODEL / PAIRED QUESTIONS", {
      "font-size": 11,
      "font-weight": 700,
    });
    text(s, 1138, 159, "NET CHANGE", {
      "font-size": 11,
      "font-weight": 700,
      "text-anchor": "end",
    });
    for (const v of [-extent, -extent / 2, 0, extent / 2, extent]) {
      shape(
        "line",
        {
          x1: x(v),
          x2: x(v),
          y1: 175,
          y2: bottom,
          stroke: v === 0 ? "#7f9690" : grid,
          "stroke-width": v === 0 ? 1.5 : 1,
        },
        s,
      );
      text(
        s,
        x(v),
        bottom + 25,
        `${v > 0 ? "+" : ""}${format(v * 100, 1)} pp`,
        { "text-anchor": "middle", "font-size": 12 },
      );
    }
    rows.forEach((r, i) => {
      const y = 196 + i * 48,
        color = r.delta >= 0 ? colors.medium : "#b3433b";
      logo(s, r.model, 32, y - 16);
      modelText(s, 70, y - 2, r.model);
      text(
        s,
        70,
        y + 15,
        `${r.paired}/${r.total} paired${r.paired < r.total ? " *" : ""} · ${r.gained} gained · ${r.lost} lost`,
        { "font-size": 12 },
      );
      shape(
        "line",
        {
          x1: x(0),
          x2: x(r.delta),
          y1: y,
          y2: y,
          stroke: color,
          "stroke-width": 5,
          "stroke-linecap": "round",
        },
        s,
      );
      shape(
        "circle",
        {
          cx: x(r.delta),
          cy: y,
          r: 6,
          fill: color,
          stroke: "#fff",
          "stroke-width": 1.5,
        },
        s,
      );
      text(
        s,
        1138,
        y + 5,
        `${r.delta > 0 ? "+" : ""}${format(r.delta * 100)} pp`,
        {
          "text-anchor": "end",
          "font-size": 16,
          "font-weight": 700,
          fill: color,
        },
      );
    });
  }

  // All bar charts use an explicit zero baseline and label every count.
  function bars(
    s,
    items,
    {
      x = 300,
      top = 190,
      width = 690,
      step = 55,
      max = Math.max(1, ...items.map((d) => d.value)),
      denominator,
      labelX = 32,
    } = {},
  ) {
    const bottom = top + (items.length - 1) * step + 19;
    max = countAxisMaximum(max);
    for (let i = 0; i <= 4; i++) {
      const value = (max * i) / 4,
        px = x + (width * i) / 4;
      shape(
        "line",
        { x1: px, x2: px, y1: top - 28, y2: bottom, stroke: grid },
        s,
      );
      text(s, px, bottom + 24, format(value, 1), {
        "font-size": 12,
        "text-anchor": "middle",
      });
    }
    items.forEach((item, i) => {
      const y = top + i * step;
      text(s, labelX, y + 5, item.label, { "font-size": 14, fill: ink });
      shape(
        "rect",
        {
          x,
          y: y - 13,
          width: (width * item.value) / max,
          height: 27,
          rx: 3,
          fill: item.color || colors.medium,
        },
        s,
      );
      text(
        s,
        x + width + 28,
        y + 5,
        `${format(item.value, 0)}${denominator ? ` / ${format(denominator, 0)}` : ""}`,
        { "font-size": 15, "font-weight": 700, fill: ink },
      );
    });
  }

  function difficulty(bins, panel) {
    if (!bins.some(Boolean))
      return empty(
        "difficulty-chart",
        "No complete configuration panel is available.",
      );
    const total = bins.reduce((a, b) => a + b, 0);
    const s = chart(
      "difficulty-chart",
      "Question difficulty across the complete panel",
      515,
      {
        subtitle: `${total} questions · ${panel} complete configurations · Grouped by the fraction of panel configurations answering correctly`,
        footer:
          "Each question appears once. Configurations from the same model are correlated; these are descriptive votes, not independent observations.",
      },
    );
    bars(
      s,
      bins.map((value, i) => ({
        value,
        label: [
          "Below 25% correct",
          "25% to below 50%",
          "50% to below 75%",
          "75% to 100%",
        ][i],
        color: ["#a15a3b", "#c08b5d", "#6c9e8c", "#087568"][i],
      })),
      {
        top: 200,
        step: 57,
        max: Math.max(4, Math.ceil(Math.max(...bins) / 4) * 4),
      },
    );
    text(s, 645, 428, "Number of questions", {
      "text-anchor": "middle",
      "font-size": 13,
    });
  }

  function coverage(c) {
    const names = {
      complete: "Completed",
      unsupported: "Unsupported",
      skipped: "Image skips",
      infra_error: "Unresolved",
      not_run: "Not run",
      pending: "Pending",
    };
    const palette = {
      complete: "#087568",
      unsupported: "#a0afa6",
      skipped: "#b58b35",
      infra_error: "#b3433b",
      not_run: "#b9c4ca",
      pending: "#4e6e9b",
    };
    const statuses = Object.entries(c.statuses).filter(([, n]) => n > 0);
    const height = 330 + Math.ceil(statuses.length / 3) * 62;
    const s = chart("coverage", "Coverage of configured answer slots", height, {
      subtitle: `${format(c.complete, 0)} / ${format(c.slots, 0)} completed · ${c.questions} questions · ${c.images} with images · ${c.models} models`,
      footer:
        "One slot is one question × model × requested effort. Completed includes invalid answers, which count as wrong. Other statuses remain unranked.",
    });
    let x = 32;
    statuses.forEach(([status, n]) => {
      const width = (n / (c.slots || 1)) * 1136;
      shape(
        "rect",
        { x, y: 173, width, height: 44, fill: palette[status] || "#7f8d91" },
        s,
      );
      x += width;
    });
    statuses.forEach(([status, n], i) => {
      const x = 32 + (i % 3) * 380,
        y = 258 + Math.floor(i / 3) * 62;
      shape(
        "rect",
        {
          x,
          y: y - 12,
          width: 12,
          height: 12,
          rx: 2,
          fill: palette[status] || "#7f8d91",
        },
        s,
      );
      text(s, x + 22, y, names[status] || status, {
        "font-size": 14,
        fill: ink,
      });
      text(
        s,
        x + 22,
        y + 23,
        `${format(n, 0)} · ${percent(n / (c.slots || 1))}`,
        { "font-size": 17, "font-weight": 700, fill: ink },
      );
    });
  }

  function subjects(overall, categories, categoryRows) {
    if (!categories.length || !overall.length)
      return empty("subjects-chart", "No subject results for this selection.");
    const rows = [...overall].sort(
      (a, b) =>
        (b.accuracy ?? -1) - (a.accuracy ?? -1) ||
        a.model.localeCompare(b.model) ||
        a.tier.localeCompare(b.tier),
    );
    const width = Math.max(1200, 420 + categories.length * 175),
      height = 326 + rows.length * 41;
    const s = chart("subjects-chart", "Accuracy by subject", height, {
      width,
      scope: { category: "" },
      subtitle:
        "Each cell uses its subject denominator · Gray cells show completed / selected; they have no accuracy score",
      footer:
        "Cell color encodes accuracy from 0% (light) to 100% (dark). Rows sort by overall accuracy; incomplete configurations follow. Requested effort appears beside each model.",
    });
    const cellWidth = (width - 420) / categories.length;
    text(s, 32, 177, "MODEL / EFFORT", { "font-size": 11, "font-weight": 700 });
    categories.forEach((category, i) => {
      const x = 398 + i * cellWidth + cellWidth / 2;
      text(s, x, 170, category, {
        "text-anchor": "middle",
        "font-size": 16,
        fill: ink,
      });
      text(
        s,
        x,
        191,
        `${categoryRows.find((r) => r.category === category)?.total || 0} questions`,
        { "text-anchor": "middle", "font-size": 12 },
      );
    });
    const lookup = new Map(
      categoryRows.map((r) => [`${r.model}\0${r.tier}\0${r.category}`, r]),
    );
    rows.forEach((r, i) => {
      const y = 209 + i * 41;
      logo(s, r.model, 32, y + 2, 24);
      modelText(s, 68, y + 16, r.model, { "font-size": 14 });
      text(s, 375, y + 16, r.tier, {
        "text-anchor": "end",
        "font-size": 12,
        fill: colors[r.tier],
      });
      categories.forEach((category, j) => {
        const item = lookup.get(`${r.model}\0${r.tier}\0${category}`),
          value = item?.accuracy,
          x = 398 + j * cellWidth;
        const color =
          value == null ? { fill: "#f1f3f2", text: ink } : heatmapColor(value);
        shape(
          "rect",
          {
            x: x + 3,
            y,
            width: cellWidth - 6,
            height: 34,
            rx: 3,
            fill: color.fill,
          },
          s,
        );
        text(
          s,
          x + cellWidth / 2,
          y + 22,
          item
            ? value == null
              ? `${item.complete}/${item.total} done`
              : `${percent(value)} · ${item.correct}/${item.total}`
            : "—",
          {
            "text-anchor": "middle",
            "font-size": 12,
            fill: color.text,
          },
        );
      });
    });
    const legendY = height - 90;
    text(s, 32, legendY + 4, "ACCURACY", { "font-size": 11 });
    [0, 0.25, 0.5, 0.75, 1].forEach((v, i) => {
      const x = 140 + i * 118;
      shape(
        "rect",
        {
          x,
          y: legendY - 10,
          width: 30,
          height: 16,
          rx: 2,
          fill: heatmapColor(v).fill,
        },
        s,
      );
      text(s, x + 38, legendY + 3, percent(v), { "font-size": 12 });
    });
  }

  function answers(q) {
    if (!q.panel)
      return empty(
        "answer-chart",
        "No complete configuration panel is available for answer frequencies.",
      );
    const s = chart(
      "answer-chart",
      `Question ${q.id.split(":").at(-1)} · answer distribution`,
      495,
      {
        width: 720,
        subtitle: `${q.correct}/${q.panel} panel configurations correct · Supplied key: option ${q.expected}`,
        footer:
          "Counts use the fixed complete panel. Model/effort votes are correlated. The supplied key is not independently verified.",
        suffix: q.id,
      },
    );
    bars(
      s,
      [1, 2, 3, 4, 0].map((i) => ({
        value: q.answers[i],
        label: i ? `Option ${i}${i === q.expected ? " · KEY" : ""}` : "Invalid",
        color:
          i === q.expected ? colors.medium : i === 0 ? "#a15a3b" : "#8faaa0",
      })),
      { x: 170, top: 183, width: 400, step: 43, max: q.panel },
    );
    text(s, 370, 418, "Number of panel configurations", {
      "text-anchor": "middle",
      "font-size": 12,
    });
  }

  function pair(result, a, b) {
    const s = chart("pair-chart", "Shared and complementary answers", 600, {
      subtitle: `${result.n} paired questions · Both configurations completed the selected scope`,
      footer:
        "Either correct is an oracle upper bound that knows the answer key, not the accuracy of a deployable ensemble. Invalid completions count as wrong.",
      suffix: `${a.model}-${a.tier}-vs-${b.model}-${b.tier}`,
    });
    text(s, 32, 168, "A", { "font-size": 15, "font-weight": 600, fill: ink });
    text(s, 32, 194, "B", { "font-size": 15, "font-weight": 600, fill: ink });
    logo(s, a.model, 56, 152, 20);
    logo(s, b.model, 56, 178, 20);
    modelText(s, 86, 168, a.model, {}, ` / ${a.tier}`);
    modelText(s, 86, 194, b.model, {}, ` / ${b.tier}`);
    bars(
      s,
      [
        {
          label: "Both correct",
          value: result.both.length,
          color: colors.medium,
        },
        { label: "Only A correct", value: result.a.length, color: "#4e6e9b" },
        { label: "Only B correct", value: result.b.length, color: "#a17a39" },
        { label: "Both wrong", value: result.neither.length, color: "#a15a3b" },
      ],
      { top: 249, step: 58, max: Math.max(1, result.n), denominator: result.n },
    );
    text(s, 645, 495, "Number of paired questions", {
      "text-anchor": "middle",
      "font-size": 13,
    });
  }

  function configurationLabel(s, r, y) {
    logo(s, r.model, 32, y - 16);
    modelText(s, 70, y - 2, r.model);
    text(s, 70, y + 14, `${r.model.split("/")[0]} · ${r.tier}`, {
      "font-size": 11,
    });
  }
  function columnHeading(s, x, y, value, anchor = "start") {
    text(s, x, y, value, {
      "font-size": 11,
      "font-weight": 700,
      "text-anchor": anchor,
    });
  }

  function ranks(rows, total, resamples) {
    if (!rows.length)
      return empty(
        "rank-chart",
        "Rank intervals need at least two complete configurations in this scope.",
      );
    const height = 265 + rows.length * 40,
      bottom = 200 + (rows.length - 1) * 40;
    const s = chart(
      "rank-chart",
      "Rank uncertainty under question resampling",
      height,
      {
        display: true,
        subtitle: `Bar: 5th–95th percentile rank across ${format(resamples, 0)} paired question resamples  ·  Marker: observed rank`,
        footer:
          "Ties share the better rank. Resampling treats questions as exchangeable; it does not measure run-to-run variation. Ranks always use every complete configuration in the scope, including any hidden by display filters.",
      },
    );
    const x = (rank) => 410 + ((rank - 1) / Math.max(1, total - 1)) * 590;
    columnHeading(s, 32, 158, "MODEL / REQUESTED EFFORT");
    columnHeading(s, 1075, 158, "90% RANKS", "middle");
    columnHeading(s, 1170, 158, "FIRST", "end");
    const ticks = new Set([1, total]);
    for (let t = 5; t < total; t += 5) ticks.add(t);
    for (const t of ticks) {
      shape(
        "line",
        { x1: x(t), x2: x(t), y1: 165, y2: bottom, stroke: grid },
        s,
      );
      text(s, x(t), 154, String(t), {
        "font-size": 12,
        "text-anchor": "middle",
      });
    }
    rows.forEach((r, i) => {
      const y = 184 + i * 40,
        color = colors[r.tier] || colors.medium,
        span = r.low === r.high ? String(r.low) : `${r.low}–${r.high}`;
      configurationLabel(s, r, y);
      const bar = shape(
        "line",
        {
          x1: x(r.low),
          x2: x(r.high),
          y1: y,
          y2: y,
          stroke: color,
          "stroke-width": 8,
          "stroke-linecap": "round",
          opacity: 0.3,
        },
        s,
      );
      const title = `${modelDisplayName(r.model)} / ${r.tier}: observed rank ${r.rank}, 90% interval ${span}, median ${r.median}, first in ${percent(r.first)} of resamples`;
      shape("title", {}, bar, title);
      shape("title", {}, marker(s, x(r.rank), y, r.tier, 5), title);
      text(s, 1075, y + 5, span, {
        "font-size": 15,
        "font-weight": 700,
        "text-anchor": "middle",
        fill: ink,
      });
      text(s, 1170, y + 5, percent(r.first), {
        "font-size": 13,
        "text-anchor": "end",
      });
    });
    text(
      s,
      410,
      bottom + 28,
      "Rank 1 is the most accurate. FIRST: share of resamples ranked first, ties included.",
      { "font-size": 12 },
    );
  }

  function estimates(rows) {
    if (!rows.length)
      return empty(
        "estimate-chart",
        "No unranked configuration in this scope is missing only unresolved answers.",
      );
    const height = 300 + rows.length * 48,
      bottom = 204 + (rows.length - 1) * 48;
    const s = chart(
      "estimate-chart",
      "Unresolved answers: bounds and a difficulty-based estimate",
      height,
      {
        display: true,
        subtitle:
          "Band: every unresolved answer wrong → every one correct  ·  Open marker: rate on completed answers  ·  Filled marker: Rasch estimate",
        footer:
          "The estimate assumes unresolved questions behave like completed questions of the same panel difficulty. It is not a score and never enters rankings. When the held-out check predicts more than the panel actually scored, treat it as optimistic.",
      },
    );
    const values = rows.flatMap((r) => [
      r.correct / r.total,
      r.correct / r.complete,
      r.estimate,
    ]);
    const lo = Math.max(0, Math.floor(Math.min(...values) * 10) / 10),
      x = (v) => 410 + ((v - lo) / (1 - lo) || 0) * 580,
      step = 1 - lo > 0.3 ? 0.1 : 0.05;
    columnHeading(s, 32, 158, "MODEL / REQUESTED EFFORT");
    columnHeading(s, 1050, 158, "ESTIMATE", "middle");
    columnHeading(s, 1170, 158, "UNRESOLVED", "end");
    for (let v = lo; v <= 1 + 1e-9; v += step) {
      shape(
        "line",
        { x1: x(v), x2: x(v), y1: 165, y2: bottom, stroke: grid },
        s,
      );
      text(s, x(v), 154, percent(v), {
        "font-size": 12,
        "text-anchor": "middle",
      });
    }
    rows.forEach((r, i) => {
      const y = 184 + i * 48,
        low = r.correct / r.total,
        high = (r.correct + r.unresolved) / r.total,
        observed = r.correct / r.complete;
      configurationLabel(s, r, y);
      const band = shape(
        "rect",
        {
          x: x(low),
          y: y - 8,
          width: Math.max(2, x(high) - x(low)),
          height: 16,
          rx: 3,
          fill: "#dbe6e0",
        },
        s,
      );
      shape(
        "title",
        {},
        band,
        `Bounds ${percent(low)}–${percent(high)} (${r.unresolved} unresolved of ${r.total})`,
      );
      shape(
        "title",
        {},
        marker(s, x(observed), y, r.tier, 6, false),
        `Completed answers: ${r.correct}/${r.complete} correct (${percent(observed)})`,
      );
      shape(
        "title",
        {},
        marker(s, x(r.estimate), y, r.tier, 6),
        `Rasch estimate ${percent(r.estimate)} · about ${format(r.expected_correct, 1)} ± ${format(r.expected_correct_sd, 1)} of ${r.total} correct`,
      );
      text(s, 1050, y + 5, percent(r.estimate), {
        "font-size": 15,
        "font-weight": 700,
        "text-anchor": "middle",
        fill: ink,
      });
      text(s, 1170, y + 5, `${r.unresolved}/${r.total}`, {
        "font-size": 13,
        "text-anchor": "end",
      });
    });
    const legendY = bottom + 34;
    shape(
      "rect",
      {
        x: 410,
        y: legendY - 11,
        width: 26,
        height: 14,
        rx: 3,
        fill: "#dbe6e0",
      },
      s,
    );
    text(s, 444, legendY, "Completion bounds", { "font-size": 13 });
    // Shape and color encode requested effort, as in the other charts.
    shape(
      "circle",
      {
        cx: 616,
        cy: legendY - 4,
        r: 5,
        fill: "#fff",
        stroke: ink,
        "stroke-width": 1.6,
      },
      s,
    );
    text(s, 629, legendY, "Completed-answer rate", { "font-size": 13 });
    shape("circle", { cx: 826, cy: legendY - 4, r: 5, fill: ink }, s);
    text(s, 839, legendY, "Rasch estimate", { "font-size": 13 });
    text(s, 1170, legendY, "Marker shape: requested effort", {
      "font-size": 12,
      "text-anchor": "end",
    });
  }

  function separation(groups, panel) {
    if (!groups.length)
      return empty(
        "separation-chart",
        "Variance shares need at least two complete configurations in this scope.",
      );
    const rowHeight = 46,
      top = 200,
      height = top + groups.length * rowHeight + 150,
      stripLeft = 540,
      stripRight = 900,
      sx = (v) => stripLeft + v * (stripRight - stripLeft),
      shareLeft = 960,
      shareWidth = 140,
      bottom = top + groups.length * rowHeight;
    const s = chart(
      "separation-chart",
      "Where complete configurations separate",
      height,
      {
        subtitle:
          "Dots: each complete configuration's accuracy on the group  ·  Bar: share of between-configuration variance in total correct",
        footer:
          "Share = cov(group score, total score) / var(total score) across complete configurations. Subject shares sum to 100%, and a subject's section shares sum to its share; a group that runs against the total is negative. Same-model configurations are correlated.",
      },
    );
    columnHeading(s, 32, 162, "SUBJECT / SECTION");
    columnHeading(s, 420, 162, "QUESTIONS", "middle");
    columnHeading(s, 495, 162, "ALL RIGHT", "middle");
    columnHeading(s, stripLeft, 162, "ACCURACY BY CONFIGURATION");
    columnHeading(s, shareLeft, 162, "SHARE OF VARIANCE");
    for (const v of [0, 0.25, 0.5, 0.75, 1]) {
      shape(
        "line",
        { x1: sx(v), x2: sx(v), y1: top - 12, y2: bottom, stroke: grid },
        s,
      );
      text(s, sx(v), top - 18, percent(v), {
        "font-size": 11,
        "text-anchor": "middle",
      });
    }
    groups.forEach((g, i) => {
      const y = top + i * rowHeight + 22,
        subject = g.level === "subject";
      if (subject && i > 0)
        shape(
          "line",
          {
            x1: 32,
            x2: 1168,
            y1: y - 26,
            y2: y - 26,
            stroke: "#cdd8d3",
          },
          s,
        );
      text(
        s,
        subject ? 32 : 58,
        y + 5,
        subject ? g.category : g.section || "Untagged questions",
        {
          "font-size": subject ? 15 : 14,
          "font-weight": subject ? 650 : 450,
          fill: ink,
        },
      );
      text(s, 420, y + 5, String(g.questions), {
        "font-size": 14,
        "text-anchor": "middle",
        fill: ink,
      });
      text(s, 495, y + 5, String(g.ceiling), {
        "font-size": 14,
        "text-anchor": "middle",
      });
      shape(
        "line",
        {
          x1: sx(g.min),
          x2: sx(g.max),
          y1: y,
          y2: y,
          stroke: "#b9c8c1",
          "stroke-width": 2,
        },
        s,
      );
      g.correct.forEach((n, j) => {
        const cfg = panel[j],
          v = n / g.questions,
          // A fixed vertical offset keeps coincident configurations visible.
          dot = shape(
            "circle",
            {
              cx: sx(v),
              cy: y + (((j * 7) % 5) - 2) * 3.5,
              r: 4.5,
              fill: colors[cfg.tier] || colors.medium,
              "fill-opacity": 0.7,
              stroke: "#fff",
              "stroke-width": 1,
            },
            s,
          );
        shape(
          "title",
          {},
          dot,
          `${modelDisplayName(cfg.model)} / ${cfg.tier}: ${n}/${g.questions} (${percent(v)})`,
        );
      });
      shape(
        "line",
        {
          x1: sx(g.mean),
          x2: sx(g.mean),
          y1: y - 13,
          y2: y + 13,
          stroke: ink,
          "stroke-width": 2,
        },
        s,
      );
      if (g.variance_share == null) {
        text(s, 1170, y + 5, "—", { "text-anchor": "end" });
        return;
      }
      shape(
        "rect",
        {
          x: shareLeft,
          y: y - 9,
          width: Math.max(0, Math.min(1, g.variance_share)) * shareWidth,
          height: 18,
          rx: 3,
          fill: subject ? colors.medium : "#6c9e8c",
        },
        s,
      );
      text(s, 1170, y + 5, percent(g.variance_share), {
        "font-size": 14,
        "font-weight": 700,
        "text-anchor": "end",
        fill: ink,
      });
    });
    const legendY = bottom + 30;
    ["minimal", "medium", "max"].forEach((tier, i) => {
      shape(
        "circle",
        {
          cx: stripLeft + i * 110,
          cy: legendY - 4,
          r: 4.5,
          fill: colors[tier],
          "fill-opacity": 0.7,
        },
        s,
      );
      text(s, stripLeft + 10 + i * 110, legendY, tier, { "font-size": 12 });
    });
    shape(
      "line",
      {
        x1: stripLeft + 340,
        x2: stripLeft + 340,
        y1: legendY - 12,
        y2: legendY + 4,
        stroke: ink,
        "stroke-width": 2,
      },
      s,
    );
    text(s, stripLeft + 350, legendY, "mean", { "font-size": 12 });
    text(
      s,
      32,
      legendY,
      "ALL RIGHT: questions every configuration answered correctly",
      {
        "font-size": 12,
      },
    );
  }

  function answerLength(rows, summary) {
    const points = rows.filter(
      (r) => r.correct_median_tokens > 0 || r.wrong_median_tokens > 0,
    );
    if (!points.length)
      return empty(
        "length-chart",
        "No complete configurations with known output tokens match these display filters.",
      );
    const sorted = [...points].sort(
      (a, b) =>
        (b.ratio ?? -1) - (a.ratio ?? -1) ||
        a.model.localeCompare(b.model) ||
        a.tier.localeCompare(b.tier),
    );
    const height = 300 + sorted.length * 40,
      bottom = 200 + (sorted.length - 1) * 40;
    const s = chart(
      "length-chart",
      "Output tokens of correct and wrong answers",
      height,
      {
        display: true,
        subtitle:
          "Median output tokens per answer within each complete configuration  ·  Logarithmic scale",
        footer: `${summary.ratio == null ? "" : `Holding question and configuration fixed, wrong answers used ${format(summary.ratio)}× the output tokens of correct answers (${format(summary.answers, 0)} answers). `}Medians also reflect question difficulty: hard questions are both longer and more often wrong. Invalid answers count as wrong.`,
      },
    );
    const values = sorted
      .flatMap((r) => [r.correct_median_tokens, r.wrong_median_tokens])
      .filter((v) => v > 0);
    const low = Math.floor(Math.log10(Math.min(...values))),
      high = Math.max(low + 1, Math.ceil(Math.log10(Math.max(...values)))),
      x = (v) => 410 + ((Math.log10(v) - low) / (high - low)) * 570;
    columnHeading(s, 32, 158, "MODEL / REQUESTED EFFORT");
    columnHeading(s, 1085, 158, "WRONG ÷ CORRECT", "middle");
    columnHeading(s, 1170, 158, "N", "end");
    for (let i = low; i <= high; i++) {
      shape(
        "line",
        { x1: x(10 ** i), x2: x(10 ** i), y1: 165, y2: bottom, stroke: grid },
        s,
      );
      text(s, x(10 ** i), 154, format(10 ** i, 0), {
        "font-size": 12,
        "text-anchor": "middle",
      });
    }
    const right = "#087568",
      wrong = "#a15a3b";
    sorted.forEach((r, i) => {
      const y = 184 + i * 40;
      configurationLabel(s, r, y);
      if (r.correct_median_tokens > 0 && r.wrong_median_tokens > 0)
        shape(
          "line",
          {
            x1: x(r.correct_median_tokens),
            x2: x(r.wrong_median_tokens),
            y1: y,
            y2: y,
            stroke: "#c9d4cf",
            "stroke-width": 3,
          },
          s,
        );
      for (const [value, color, name, n] of [
        [r.correct_median_tokens, right, "correct", r.correct],
        [r.wrong_median_tokens, wrong, "wrong", r.wrong],
      ]) {
        if (!(value > 0)) continue;
        const dot = shape(
          "circle",
          {
            cx: x(value),
            cy: y,
            r: 6,
            fill: color,
            stroke: "#fff",
            "stroke-width": 1.5,
          },
          s,
        );
        shape(
          "title",
          {},
          dot,
          `${modelDisplayName(r.model)} / ${r.tier}: median ${format(value, 0)} output tokens on ${n} ${name} answers`,
        );
      }
      text(s, 1085, y + 5, r.ratio == null ? "—" : `${format(r.ratio, 1)}×`, {
        "font-size": 15,
        "font-weight": 700,
        "text-anchor": "middle",
        fill: ink,
      });
      text(s, 1170, y + 5, `${r.correct} / ${r.wrong}`, {
        "font-size": 13,
        "text-anchor": "end",
      });
    });
    const legendY = bottom + 34;
    for (const [i, color, label] of [
      [0, right, "Correct answers (median)"],
      [1, wrong, "Wrong answers (median)"],
    ]) {
      shape(
        "circle",
        { cx: 416 + i * 220, cy: legendY - 4, r: 6, fill: color },
        s,
      );
      text(s, 428 + i * 220, legendY, label, { "font-size": 13 });
    }
    text(s, 1170, legendY, "N: correct / wrong answers", {
      "font-size": 12,
      "text-anchor": "end",
    });
  }

  return {
    accuracy,
    scatter,
    effort,
    difficulty,
    coverage,
    subjects,
    answers,
    pair,
    ranks,
    estimates,
    separation,
    answerLength,
  };
}

if (typeof module !== "undefined" && module.exports)
  module.exports = {
    countAxisMaximum,
    heatmapColor,
    pngDimensions,
    frontierLabelPositions,
    chartPNG,
    createReportCharts,
  };
