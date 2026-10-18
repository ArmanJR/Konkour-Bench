import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { fixture } from "./test-fixtures";
import { useReport } from "./useReport";

afterEach(() => {
  cleanup();
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

it("refreshes every 30 seconds and stops polling after unmount", async () => {
  vi.useFakeTimers();
  const fetch = vi
    .fn()
    .mockResolvedValue({ ok: true, json: async () => fixture() });
  vi.stubGlobal("fetch", fetch);
  const { result, unmount } = renderHook(() => useReport(true));
  await act(async () => {});
  expect(result.current.data?.models).toHaveLength(2);
  expect(fetch).toHaveBeenCalledTimes(1);
  await act(async () => {
    await vi.advanceTimersByTimeAsync(30_000);
  });
  expect(fetch).toHaveBeenCalledTimes(2);
  unmount();
  await vi.advanceTimersByTimeAsync(60_000);
  expect(fetch).toHaveBeenCalledTimes(2);
  expect(fetch.mock.calls[1][1].signal.aborted).toBe(true);
});

it("does not replace a newer report with an aborted response", async () => {
  let resolveFirst!: (value: unknown) => void;
  const old = new Promise((resolve) => {
    resolveFirst = resolve;
  });
  const newer = { ...fixture(), scoring_revision: "newest" };
  vi.stubGlobal(
    "fetch",
    vi
      .fn()
      .mockReturnValueOnce(old)
      .mockResolvedValueOnce({ ok: true, json: async () => newer }),
  );
  const { result } = renderHook(() => useReport(false));
  await act(async () => {
    result.current.refresh();
  });
  expect(result.current.data?.scoring_revision).toBe("newest");
  await act(async () => {
    resolveFirst({ ok: true, json: async () => fixture() });
  });
  expect(result.current.data?.scoring_revision).toBe("newest");
});
