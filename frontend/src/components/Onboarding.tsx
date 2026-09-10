import { useState } from "react";
import { useEventStore } from "../store";
import { useEventStream } from "../hooks/useEventStream";

const STORAGE_KEY = "grow-onboarded";

export default function Onboarding() {
  const totalEvents = useEventStore((s) => s.events.length);
  const setCount = useEventStore((s) => s.setCount);
  const { connect } = useEventStream();

  const [dismissed, setDismissed] = useState(() => {
    try {
      return localStorage.getItem(STORAGE_KEY) === "1";
    } catch {
      return false;
    }
  });

  if (dismissed || totalEvents > 0) return null;

  const handleClick = () => {
    setDismissed(true);
    try {
      localStorage.setItem(STORAGE_KEY, "1");
    } catch {}
    setCount(100);
    connect(100);
  };

  return (
    <div className="absolute inset-0 z-10 flex items-center justify-center">
      <div className="flex flex-col items-center gap-4 rounded-xl border border-[var(--border)] bg-[var(--surface-1)] px-8 py-6 shadow-2xl max-w-xs text-center">
        <div className="flex items-center gap-2">
          <span className="inline-block h-2.5 w-2.5 rounded-full bg-emerald-400 shadow-[0_0_6px_rgba(52,211,153,.5)]" />
          <span className="font-[var(--font-mono)] text-base font-semibold text-emerald-400">
            grow
          </span>
        </div>
        <p className="text-[13px] leading-relaxed text-[var(--text-1)]">
          See how Python's <code className="font-[var(--font-mono)] text-amber-500">list</code> grows
          its backing array — resizes, copies, and amortized cost, visualized step by step.
        </p>
        <button
          onClick={handleClick}
          className="rounded-lg bg-emerald-400 px-5 py-2 text-[13px] font-semibold text-[#0f1117] hover:bg-emerald-300 transition-colors"
        >
          Watch 100 appends
        </button>
        <span className="text-[10px] text-[var(--text-2)]">
          or pick a scenario from the top bar
        </span>
      </div>
    </div>
  );
}
