import { useState } from "react";
import { useEventStore } from "../store";
import { useEventStream, type Scenario } from "../hooks/useEventStream";

type OpType = "append" | "insert" | "extend";

const APPEND_COUNTS = [5, 10, 20, 50, 100];

interface Preset {
  label: string;
  scenario: Scenario;
}

const OP_PRESETS: Record<Exclude<OpType, "append">, Preset[]> = {
  insert: [
    { label: "At start", scenario: { op: "insert", seed: 8, count: 5, pos: "start" } },
    { label: "At middle", scenario: { op: "insert", seed: 8, count: 5, pos: "middle" } },
    { label: "At end", scenario: { op: "insert", seed: 8, count: 5, pos: "end" } },
  ],
  extend: [
    { label: "+10 items", scenario: { op: "extend", seed: 5, count: 10 } },
    { label: "+50 items", scenario: { op: "extend", seed: 5, count: 50 } },
  ],
};

const OP_HINTS: Record<OpType, string> = {
  append: "Add elements one at a time — watch capacity grow when full",
  insert: "Position determines cost — start is O(n), end is O(1)",
  extend: "Bulk add — one resize to fit all new items at once",
};

export default function PresetsPanel() {
  const reset = useEventStore((s) => s.reset);
  const { run } = useEventStream();
  const [op, setOp] = useState<OpType>("append");
  const [activePreset, setActivePreset] = useState<string | null>(null);

  return (
    <div className="flex flex-col gap-2 border-b border-(--border) px-4 py-3">
      <div className="flex items-center gap-1">
        {(Object.keys(OP_HINTS) as OpType[]).map((o) => (
          <button
            key={o}
            onClick={() => { setOp(o); setActivePreset(null); reset(); }}
            aria-pressed={op === o}
            className={`rounded-md px-2.5 py-1 text-xs font-medium transition-colors ${
              op === o
                ? "bg-emerald-400/15 text-emerald-400"
                : "text-(--text-2) hover:text-(--text-1)"
            }`}
          >
            {o[0].toUpperCase() + o.slice(1)}
          </button>
        ))}
      </div>
      <p className="text-[11px] text-(--text-2) leading-tight">{OP_HINTS[op]}</p>
      <div className="flex flex-wrap items-center gap-2">
        {op === "append" ? (
          APPEND_COUNTS.map((n) => (
            <button
              key={n}
              onClick={() => {
                setActivePreset(`append-${n}`);
                run({ op: "append", count: n });
              }}
              aria-pressed={activePreset === `append-${n}`}
              className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                activePreset === `append-${n}`
                  ? "bg-emerald-600 text-white"
                  : "bg-(--surface-2) text-(--text-1) hover:bg-(--surface-3)"
              }`}
            >
              {n}
            </button>
          ))
        ) : (
          OP_PRESETS[op].map((preset) => (
            <button
              key={preset.label}
              onClick={() => {
                setActivePreset(preset.label);
                run(preset.scenario);
              }}
              aria-pressed={activePreset === preset.label}
              className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                activePreset === preset.label
                  ? "bg-emerald-600 text-white"
                  : "bg-(--surface-2) text-(--text-1) hover:bg-(--surface-3)"
              }`}
            >
              {preset.label}
            </button>
          ))
        )}
      </div>
    </div>
  );
}
