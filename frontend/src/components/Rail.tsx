import { useEventStore } from "../store";
import type { Mode, Strategy, Backend } from "../store";

const MODES: { key: Mode; letter: string; label: string }[] = [
  { key: "presets", letter: "P", label: "Presets" },
  { key: "wrapper", letter: "W", label: "Wrapper" },
  { key: "auto", letter: "A", label: "Auto" },
];

const STRATEGIES: { key: Strategy; label: string; subtitle: string }[] = [
  { key: "cpython", label: "CPython", subtitle: "~1.125×" },
  { key: "doubling", label: "Doubling", subtitle: "2×" },
  { key: "1.5x", label: "1.5×", subtitle: "grow by half" },
  { key: "nogrowth", label: "No growth", subtitle: "fixed" },
];

export default function Rail({
  onShowHelp,
  theme,
  onToggleTheme,
}: {
  onShowHelp?: () => void;
  theme?: "dark" | "light";
  onToggleTheme?: () => void;
}) {
  const mode = useEventStore((s) => s.mode);
  const setMode = useEventStore((s) => s.setMode);
  const strategy = useEventStore((s) => s.strategy);
  const setStrategy = useEventStore((s) => s.setStrategy);
  const backend = useEventStore((s) => s.backend);
  const setBackend = useEventStore((s) => s.setBackend);
  const length = useEventStore((s) => s.length);
  const capacity = useEventStore((s) => s.capacity);
  const resizeCount = useEventStore((s) => s.resizeCount);
  const costs = useEventStore((s) => s.costs);

  const amortized =
    costs.length > 0 ? costs[costs.length - 1].amortized : 0;

  return (
    <aside className="hidden md:flex flex-col bg-[var(--surface-1)] border-r border-[var(--border)] overflow-y-auto">
      {/* Header */}
      <div className="flex items-center justify-between px-5 py-4 border-b border-[var(--border-light)]">
        <div className="flex items-center gap-2">
          <span className="inline-block h-2 w-2 rounded-full bg-emerald-400 shadow-[0_0_8px_rgba(52,211,153,0.25)]" />
          <span className="font-[var(--font-mono)] text-xl font-semibold tracking-tight text-emerald-400">
            grow
          </span>
        </div>
        <span className="text-[11px] text-[var(--text-2)] cursor-pointer hover:text-[var(--text-1)]">
          Home
        </span>
      </div>

      {/* Mode */}
      <div className="px-3 pt-4 pb-2">
        <div className="mb-2 px-2 text-[10px] font-semibold uppercase tracking-[1px] text-[var(--text-2)]">
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
                  : "text-[var(--text-1)] hover:bg-[var(--surface-2)] hover:text-[var(--text-0)]"
              }`}
            >
              <span
                className={`flex items-center justify-center w-[18px] h-[18px] rounded text-[11px] font-semibold font-[var(--font-mono)] ${
                  mode === m.key
                    ? "bg-emerald-400/25 text-emerald-400"
                    : "bg-[var(--surface-3)] text-[var(--text-2)]"
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
      <div className="px-3 py-4 border-t border-[var(--border-light)]">
        <div className="mb-2 px-2 text-[10px] font-semibold uppercase tracking-[1px] text-[var(--text-2)]">
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
                  : "border-[var(--border)] bg-[var(--surface-2)] text-[var(--text-1)] hover:border-[var(--text-2)] hover:text-[var(--text-0)]"
              }`}
            >
              <span className="font-[var(--font-mono)] text-[10px] font-medium">
                {s.label}
              </span>
              <span className="text-[10px] text-[var(--text-2)] mt-0.5">
                {s.subtitle}
              </span>
            </button>
          ))}
        </div>
      </div>

      {/* Backend */}
      <div className="px-3 py-4 border-t border-[var(--border-light)]">
        <div className="mb-2 px-2 text-[10px] font-semibold uppercase tracking-[1px] text-[var(--text-2)]">
          Backend
        </div>
        <div className="grid grid-cols-2 gap-1.5">
          {(
            [
              { key: "simulator" as Backend, label: "Simulator", subtitle: "Go engine" },
              { key: "cpython" as Backend, label: "CPython", subtitle: "real python3" },
            ] as const
          ).map((b) => (
            <button
              key={b.key}
              onClick={() => setBackend(b.key)}
              aria-pressed={backend === b.key}
              className={`flex flex-col items-center rounded-md px-2 py-[7px] text-center transition-colors border ${
                backend === b.key
                  ? "border-emerald-400 bg-emerald-400/[.12] text-emerald-400"
                  : "border-[var(--border)] bg-[var(--surface-2)] text-[var(--text-1)] hover:border-[var(--text-2)] hover:text-[var(--text-0)]"
              }`}
            >
              <span className="font-[var(--font-mono)] text-[10px] font-medium">
                {b.label}
              </span>
              <span className="text-[10px] text-[var(--text-2)] mt-0.5">
                {b.subtitle}
              </span>
            </button>
          ))}
        </div>
        {backend === "cpython" && (
          <div className="mt-2 px-2 py-1.5 rounded bg-amber-500/10 border border-amber-500/20 text-[10px] text-amber-500 leading-snug">
            Runs Python on this machine
          </div>
        )}
      </div>

      {/* Stats */}
      <div className="mt-auto flex flex-col gap-2.5 px-5 py-4 border-t border-[var(--border-light)]">
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-[var(--text-2)]">Length</span>
          <span className="font-[var(--font-mono)] text-[13px] font-medium tabular-nums text-emerald-400">
            {length}
          </span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-[var(--text-2)]">Capacity</span>
          <span className="font-[var(--font-mono)] text-[13px] font-medium tabular-nums text-[var(--text-0)]">
            {capacity}
          </span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-[var(--text-2)]">Resizes</span>
          <span className="font-[var(--font-mono)] text-[13px] font-medium tabular-nums text-amber-500">
            {resizeCount}
          </span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-[11px] text-[var(--text-2)]">Amortized</span>
          <span className="font-[var(--font-mono)] text-[13px] font-medium tabular-nums text-violet-400">
            {amortized.toFixed(2)}
          </span>
        </div>
      </div>

      {/* Footer actions */}
      <div className="flex items-center gap-2 px-5 pb-4">
        {onToggleTheme && (
          <button
            onClick={onToggleTheme}
            className="flex items-center justify-center w-6 h-6 rounded border border-[var(--border)] bg-[var(--surface-2)] text-[11px] text-[var(--text-2)] hover:text-[var(--text-1)] hover:border-[var(--text-2)] transition-colors"
            aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
          >
            {theme === "dark" ? "☀" : "☾"}
          </button>
        )}
        {onShowHelp && (
          <button
            onClick={onShowHelp}
            className="flex items-center justify-center w-6 h-6 rounded border border-[var(--border)] bg-[var(--surface-2)] text-[11px] font-[var(--font-mono)] text-[var(--text-2)] hover:text-[var(--text-1)] hover:border-[var(--text-2)] transition-colors"
            aria-label="Keyboard shortcuts"
          >
            ?
          </button>
        )}
      </div>
    </aside>
  );
}
