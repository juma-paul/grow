import { useEventStore } from "../store";
import type { Strategy } from "../store";

const STRATEGIES: { key: Strategy; label: string; subtitle: string }[] = [
  { key: "cpython", label: "CPython", subtitle: "~1.125×" },
  { key: "doubling", label: "Doubling", subtitle: "2×" },
  { key: "1.5x", label: "1.5×", subtitle: "grow by half" },
  { key: "nogrowth", label: "No growth", subtitle: "fixed" },
];

export default function Rail() {
  const strategy = useEventStore((s) => s.strategy);
  const setStrategy = useEventStore((s) => s.setStrategy);
  const length = useEventStore((s) => s.length);
  const capacity = useEventStore((s) => s.capacity);
  const resizeCount = useEventStore((s) => s.resizeCount);
  const costs = useEventStore((s) => s.costs);

  const amortized =
    costs.length > 0 ? costs[costs.length - 1].amortized : 0;

  return (
    <aside className="hidden md:flex flex-col bg-(--surface-1) border-r border-(--border) overflow-y-auto">
      {/* Header */}
      <div className="flex items-center gap-2 px-5 py-4 border-b border-(--border-light)">
        <span className="inline-block h-2 w-2 rounded-full bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.25)]" />
        <span className="font-mono text-xl font-semibold tracking-tight text-emerald-400">
          grow
        </span>
      </div>

      {/* Growth Strategy */}
      <div className="px-3 py-4">
        <div className="mb-2 px-2 text-[10px] font-semibold uppercase tracking-[1px] text-(--text-2)">
          Growth Strategy
        </div>
        <div className="grid grid-cols-2 gap-1.5">
          {STRATEGIES.map((s) => (
            <button
              key={s.key}
              onClick={() => setStrategy(s.key)}
              aria-pressed={strategy === s.key}
              className={`flex flex-col items-center rounded-md px-2 py-1.75 text-center transition-colors border ${
                strategy === s.key
                  ? "border-emerald-400 bg-emerald-400/12 text-emerald-400"
                  : "border-(--border) bg-(--surface-2) text-(--text-1) hover:border-(--text-2) hover:text-(--text-0)"
              }`}
            >
              <span className="font-mono text-[10px] font-medium">
                {s.label}
              </span>
              <span className="text-[10px] text-(--text-2) mt-0.5">
                {s.subtitle}
              </span>
            </button>
          ))}
        </div>
      </div>

      {/* Stats */}
      <div className="mt-auto flex flex-col gap-2.5 px-5 py-4 border-t border-(--border-light)">
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-(--text-2)">Length</span>
          <span className="font-mono text-[13px] font-medium tabular-nums text-emerald-400">
            {length}
          </span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-(--text-2)">Capacity</span>
          <span className="font-mono text-[13px] font-medium tabular-nums text-(--text-0)">
            {capacity}
          </span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-(--text-2)">Resizes</span>
          <span className="font-mono text-[13px] font-medium tabular-nums text-amber-500">
            {resizeCount}
          </span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-(--text-2)">Amortized</span>
          <span className="font-mono text-[13px] font-medium tabular-nums text-violet-400">
            {amortized.toFixed(2)}
          </span>
        </div>
      </div>
    </aside>
  );
}
