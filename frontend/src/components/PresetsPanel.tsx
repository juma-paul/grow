import { useEventStore } from "../store";
import { useEventStream } from "../hooks/useEventStream";

const SCENARIOS = [5, 10, 15, 20, 50, 100];

export default function PresetsPanel() {
  const count = useEventStore((s) => s.count);
  const setCount = useEventStore((s) => s.setCount);
  const { connect } = useEventStream();

  return (
    <div className="flex flex-wrap items-center gap-2 border-b border-[var(--border)] px-4 py-3">
      {SCENARIOS.map((n) => (
        <button
          key={n}
          onClick={() => {
            setCount(n);
            connect(n);
          }}
          aria-pressed={count === n}
          className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
            count === n
              ? "bg-emerald-600 text-white"
              : "bg-[var(--surface-2)] text-[var(--text-1)] hover:bg-[var(--surface-3)]"
          }`}
        >
          Append {n}
        </button>
      ))}
    </div>
  );
}
