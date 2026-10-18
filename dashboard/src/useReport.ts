import { useEffect, useState } from "react";
import { fetchReport } from "./data";
import type { Report } from "./types";

export function useReport(autoRefresh: boolean) {
  const [data, setData] = useState<Report | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [revision, setRevision] = useState(0);
  useEffect(() => {
    const controller = new AbortController();
    let timer: ReturnType<typeof setTimeout>;
    async function load() {
      setLoading(true);
      try {
        const report = await fetchReport(controller.signal);
        if (!controller.signal.aborted) {
          setData(report);
          setError("");
        }
      } catch (e) {
        if (!controller.signal.aborted)
          setError(e instanceof Error ? e.message : "Unable to load results.");
      } finally {
        if (!controller.signal.aborted) {
          setLoading(false);
          if (autoRefresh) timer = setTimeout(load, 30_000);
        }
      }
    }
    void load();
    return () => {
      controller.abort();
      clearTimeout(timer);
    };
  }, [autoRefresh, revision]);
  return { data, error, loading, refresh: () => setRevision((v) => v + 1) };
}
