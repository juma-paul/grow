import { useEventStore } from "../store";
import type { Mode, Strategy } from "../store";

const MODES: { key: Mode; letter: string; label: string }[] = [
  { key: "presets", letter: "P", label: "Presets" },
  { key: "wrapper", letter: "W", label: "Wrapper" },
  { key: "auto", letter: "A", label: "Auto" },
];

const STRATEGIES: { key: Strategy; label: string; subtitle: string }[] = [
  { key: "cpython", label: "CPython", subtitle: "~1.125×" },
  { key: "doubling", label: "Doubling", subtitle: "2×" },
  { key: "1.5x", label: "Java", subtitle: "1.5×" },
  { key: "nogrowth", label: "No growth", subtitle: "fixed" },
];

export default function Rail({ onShowHelp }: { onShowHelp?: () => void }) {
  const mode = useEventStore((s) => s.mode);
  const setMode = useEventStore((s) => s.setMode);
  const strategy = useEventStore((s) => s.strategy);
  const setStrategy = useEventStore((s) => s.setStrategy);
  const length = useEventStore((s) => s.length);
  const capacity = useEventStore((s) => s.capacity);
  const resizeCount = useEventStore((s) => s.resizeCount);
  const costs = useEventStore((s) => s.costs);

  const amortized =
    costs.length > 0 ? costs[costs.length - 1].amortized : 0;

  return (
    <aside className="hidden md:flex flex-col bg-[#171b24] border-r border-[#262d3d] overflow-y-auto">
      {/* Header */}
      <div className="flex items-center justify-between px-5 py-4 border-b border-[#1e2430]">
        <div className="flex items-center gap-2">
          <span className="inline-block h-2 w-2 rounded-full bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.25)]" />
          <span className="font-[var(--font-mono)] text-xl font-semibold tracking-tight text-emerald-400">
            grow
          </span>
        </div>
        <span className="text-[11px] text-[#565b6b] cursor-pointer hover:text-[#8b90a0]">
          Home
        </span>
      </div>

      {/* Mode */}
      <div className="px-3 pt-4 pb-2">
        <div className="mb-2 px-2 text-[10px] font-semibold uppercase tracking-[1px] text-[#565b6b]">
          Mode
        </div>
        <div className="flex flex-col gap-0.5">
          {MODES.map((m) => (
            <button
              key={m.key}
              onClick={() => setMode(m.key)}
              aria-pressed={mode === m.key}
              className={`flex items-center gap-2.5 rounded-md px-2.5 py-2 text-[13px] font-medium transition-colors text-left w-full ${
                mode === m.key
                  ? "bg-emerald-400/[.12] text-emerald-400"
                  : "text-[#8b90a0] hover:bg-[#1e2330] hover:text-[#e2e4ea]"
              }`}
            >
              <span
                className={`flex items-center justify-center w-[18px] h-[18px] rounded text-[11px] font-semibold font-[var(--font-mono)] ${
                  mode === m.key
                    ? "bg-emerald-400/25 text-emerald-400"
                    : "bg-[#252b3a] text-[#565b6b]"
                }`}
              >
                {m.letter}
              </span>
              <span>{m.label}</span>
            </button>
          ))}
        </div>
      </div>

      {/* Growth Strategy */}
      <div className="px-3 py-4 border-t border-[#1e2430]">
        <div className="mb-2 px-2 text-[10px] font-semibold uppercase tracking-[1px] text-[#565b6b]">
          Growth Strategy
        </div>
        <div className="grid grid-cols-2 gap-1.5">
          {STRATEGIES.map((s) => (
            <button
              key={s.key}
              onClick={() => setStrategy(s.key)}
              aria-pressed={strategy === s.key}
              className={`flex flex-col items-center rounded-md px-2 py-[7px] text-center transition-colors border ${
                strategy === s.key
                  ? "border-emerald-400 bg-emerald-400/[.12] text-emerald-400"
                  : "border-[#262d3d] bg-[#1e2330] text-[#8b90a0] hover:border-[#565b6b] hover:text-[#e2e4ea]"
              }`}
            >
              <span className="font-[var(--font-mono)] text-[10px] font-medium">
                {s.label}
              </span>
              <span className="text-[10px] text-[#565b6b] mt-0.5">
                {s.subtitle}
              </span>
            </button>
          ))}
        </div>
      </div>

      {/* Stats */}
      <div className="mt-auto flex flex-col gap-2.5 px-5 py-4 border-t border-[#1e2430]">
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-[#565b6b]">Length</span>
          <span className="font-[var(--font-mono)] text-[13px] font-medium tabular-nums text-emerald-400">
            {length}
          </span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-[#565b6b]">Capacity</span>
          <span className="font-[var(--font-mono)] text-[13px] font-medium tabular-nums text-[#e2e4ea]">
            {capacity}
          </span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-[#565b6b]">Resizes</span>
          <span className="font-[var(--font-mono)] text-[13px] font-medium tabular-nums text-amber-500">
            {resizeCount}
          </span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-[#565b6b]">Amortized</span>
          <span className="font-[var(--font-mono)] text-[13px] font-medium tabular-nums text-violet-400">
            {amortized.toFixed(2)}
          </span>
        </div>
      </div>

      {/* Help */}
      {onShowHelp && (
        <div className="px-5 pb-4">
          <button
            onClick={onShowHelp}
            className="flex items-center justify-center w-6 h-6 rounded border border-[#262d3d] bg-[#1e2330] text-[11px] font-[var(--font-mono)] text-[#565b6b] hover:text-[#8b90a0] hover:border-[#565b6b] transition-colors"
            aria-label="Keyboard shortcuts"
          >
            ?
          </button>
        </div>
      )}
    </aside>
  );
}
