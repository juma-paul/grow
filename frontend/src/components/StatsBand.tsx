import { useEffect, useRef, useState } from "react";

interface Stats {
  total_runs: number;
  runs_today: number;
  elements_allocated: number;
}

function useCountUp(target: number, duration = 800): number {
  const [value, setValue] = useState(0);
  const prev = useRef(0);
  const raf = useRef(0);

  useEffect(() => {
    const start = prev.current;
    const delta = target - start;
    if (delta === 0) return;

    const t0 = performance.now();
    const step = (now: number) => {
      const progress = Math.min((now - t0) / duration, 1);
      const eased = 1 - Math.pow(1 - progress, 3);
      setValue(Math.round(start + delta * eased));
      if (progress < 1) {
        raf.current = requestAnimationFrame(step);
      } else {
        prev.current = target;
      }
    };
    raf.current = requestAnimationFrame(step);
    return () => cancelAnimationFrame(raf.current);
  }, [target, duration]);

  return value;
}

function fmt(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + "M";
  if (n >= 1_000) return (n / 1_000).toFixed(1) + "K";
  return n.toLocaleString();
}

export default function StatsBand() {
  const [stats, setStats] = useState<Stats | null>(null);

  useEffect(() => {
    let active = true;
    const load = () => {
      fetch("/stats")
        .then((r) => (r.ok ? r.json() : null))
        .then((data) => {
          if (active && data) setStats(data);
        })
        .catch(() => {});
    };
    load();
    const id = setInterval(load, 15_000);
    return () => {
      active = false;
      clearInterval(id);
    };
  }, []);

  const runs = useCountUp(stats?.total_runs ?? 0);
  const today = useCountUp(stats?.runs_today ?? 0);
  const elements = useCountUp(stats?.elements_allocated ?? 0);

  if (!stats) return null;

  return (
    <div className="px-5 py-3 border-t border-[var(--border-light)]">
      <div className="mb-2 px-0 text-[10px] font-semibold uppercase tracking-[1px] text-[var(--text-2)]">
        Community
      </div>
      <div className="flex flex-col gap-1.5">
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-[var(--text-2)]">Total runs</span>
          <span className="font-[var(--font-mono)] text-[13px] font-medium tabular-nums text-[var(--text-0)]">
            {fmt(runs)}
          </span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-[var(--text-2)]">Today</span>
          <span className="font-[var(--font-mono)] text-[13px] font-medium tabular-nums text-[var(--text-0)]">
            {fmt(today)}
          </span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-[var(--text-2)]">Elements</span>
          <span className="font-[var(--font-mono)] text-[13px] font-medium tabular-nums text-[var(--text-0)]">
            {fmt(elements)}
          </span>
        </div>
      </div>
    </div>
  );
}
