import { useEventStore } from "../store";
import type { Mode, Strategy } from "../store";

const MODES: { key: Mode; label: string; title: string }[] = [
  { key: "presets", label: "P", title: "Presets" },
  { key: "wrapper", label: "W", title: "Wrapper" },
  { key: "auto", label: "A", title: "Auto" },
];

const STRATEGIES: { key: Strategy; label: string }[] = [
  { key: "cpython", label: "CPython" },
  { key: "doubling", label: "Doubling" },
  { key: "1.5x", label: "1.5×" },
  { key: "nogrowth", label: "No growth" },
];

export default function Rail() {
  const mode = useEventStore((s) => s.mode);
  const setMode = useEventStore((s) => s.setMode);
  const strategy = useEventStore((s) => s.strategy);
  const setStrategy = useEventStore((s) => s.setStrategy);
  const length = useEventStore((s) => s.length);
  const capacity = useEventStore((s) => s.capacity);

  return (
    <aside className="hidden md:flex flex-col gap-6 border-r border-zinc-700 bg-zinc-900/50 p-4 overflow-y-auto">
      <div className="text-lg font-bold tracking-tight text-white">grow</div>

      <div>
        <div className="mb-2 text-[11px] font-medium uppercase tracking-wider text-zinc-500">
          Mode
        </div>
        <div className="flex gap-1">
          {MODES.map((m) => (
            <button
              key={m.key}
              onClick={() => setMode(m.key)}
              title={m.title}
              aria-pressed={mode === m.key}
              className={`flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                mode === m.key
                  ? "bg-emerald-600 text-white"
                  : "text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
              }`}
            >
              {m.label}
            </button>
          ))}
        </div>
      </div>

      <div>
        <div className="mb-2 text-[11px] font-medium uppercase tracking-wider text-zinc-500">
          Strategy
        </div>
        <div className="grid grid-cols-2 gap-1">
          {STRATEGIES.map((s) => (
            <button
              key={s.key}
              onClick={() => setStrategy(s.key)}
              aria-pressed={strategy === s.key}
              className={`rounded-md px-2 py-1.5 text-xs font-medium transition-colors ${
                strategy === s.key
                  ? "bg-zinc-700 text-white ring-1 ring-emerald-500"
                  : "text-zinc-400 hover:bg-zinc-800 hover:text-zinc-200"
              }`}
            >
              {s.label}
            </button>
          ))}
        </div>
      </div>

      <div className="mt-auto">
        <div className="mb-2 text-[11px] font-medium uppercase tracking-wider text-zinc-500">
          Stats
        </div>
        <div className="space-y-1 text-sm">
          <div className="flex justify-between">
            <span className="text-zinc-500">Length</span>
            <span className="font-mono tabular-nums text-zinc-300">
              {length}
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-zinc-500">Capacity</span>
            <span className="font-mono tabular-nums text-zinc-300">
              {capacity}
            </span>
          </div>
        </div>
      </div>
    </aside>
  );
}
