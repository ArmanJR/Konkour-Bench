"use strict";

// Display typography only; canonical IDs continue to identify every result.
const modelDisplayNames = new Map([
  ["anthropic/claude-opus-5.5", "Claude Opus 5.5"],
  ["anthropic/claude-sonnet-5.5", "Claude Sonnet 5.5"],
  ["google/gemini-3.8-flash", "Gemini 3.8 Flash"],
  ["google/gemini-4-argon", "Gemini 4 Argon"],
  ["openai/gpt-6.1-sol", "GPT-6.1 Sol"],
  ["openai/gpt-6-luna", "GPT-6 Luna"],
  ["deepseek/deepseek-v4.1-flash-1017", "DeepSeek V4.1 Flash 1017"],
  ["deepseek/deepseek-v4-pro-0813", "DeepSeek V4 Pro 0813"],
  ["qwen/qwen3.8-27b", "Qwen 3.8 27B"],
  ["qwen/qwen3.8-max-0902", "Qwen 3.8 Max 0902"],
  ["x-ai/grok-4.7", "Grok 4.7"],
  ["xiaomi/mimo-v2.6-pro", "MiMo V2.6 Pro"],
  ["xiaomi/mimo-v2.6-flash", "MiMo V2.6 Flash"],
  ["z-ai/glm-5.3", "GLM 5.3"],
  ["z-ai/glm-5.3-flash", "GLM 5.3 Flash"],
]);

function modelDisplayName(model) {
  return (
    modelDisplayNames.get(model) || model.split("/").slice(1).join("/") || model
  );
}

function modelMatchesQuery(model, query) {
  return `${model} ${modelDisplayName(model)}`
    .toLowerCase()
    .includes(query.toLowerCase().trim());
}

// Match model families, not arbitrary substrings or the inference backend.
function modelBrand(model) {
  if (typeof model !== "string") return null;
  const families = [
    ["openai/", "openai"],
    ["anthropic/claude-", "claude-color"],
    ["google/gemini-", "gemini-color"],
    ["deepseek/", "deepseek-color"],
    ["qwen/", "qwen-color"],
    ["x-ai/grok-", "grok"],
    ["xiaomi/mimo-", "xiaomimimo"],
    ["z-ai/glm-", "zai"],
  ];
  return families.find(([prefix]) => model.startsWith(prefix))?.[1] || null;
}

function modelLogo(
  model,
  logos = typeof reportLogos === "undefined" ? {} : reportLogos,
) {
  const brand = modelBrand(model);
  return brand ? logos[brand] || "" : "";
}

if (typeof module !== "undefined" && module.exports)
  module.exports = {
    modelBrand,
    modelLogo,
    modelDisplayName,
    modelMatchesQuery,
  };
