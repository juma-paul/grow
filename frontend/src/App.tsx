import CanvasArray from "./components/CanvasArray";
import CostGraph from "./components/CostGraph";
import PresetsPanel from "./components/PresetsPanel";
import Rail from "./components/Rail";
import { useEventStore } from "./store";
import { usePlayback } from "./hooks/usePlayback";

const SPEEDS = [0.5, 1, 2, 5];

function TransportBar() {
  const currentIndex = useEventStore((s) => s.currentIndex);
  const totalEvents = useEventStore((s) => s.events.length);
  const isPlaying = useEventStore((s) => s.isPlaying);
  const play = useEventStore((s) => s.play);
  const pause = useEventStore((s) => s.pause);
  const stepForward = useEventStore((s) => s.stepForward);
  const stepBack = useEventStore((s) => s.stepBack);
  const seekTo = useEventStore((s) => s.seekTo);
  const speed = useEventStore((s) => s.speed);
  const setSpeed = useEventStore((s) => s.setSpeed);

  const atEnd = currentIndex >= totalEvents - 1;
  const atStart = currentIndex < 0;

  return (
    <>
      <div className="flex items-center gap-3 border-b border-(--border) px-4 py-2 bg-(--surface-1) shrink-0">
        <div className="flex items-center gap-1">
          <button
            onClick={stepBack}
            disabled={atStart}
            aria-label="Step back"
            className="rounded px-2 py-1 text-sm font-medium hover:bg-(--surface-2) disabled:opacity-30 disabled:cursor-not-allowed"
          >
            ⏮
          </button>
          <button
            onClick={isPlaying ? pause : play}
            disabled={atEnd && !isPlaying}
            aria-label={isPlaying ? "Pause" : "Play"}
            className="rounded px-2 py-1 text-sm font-medium hover:bg-(--surface-2) disabled:opacity-30 disabled:cursor-not-allowed"
          >
            {isPlaying ? "⏸" : "▶"}
          </button>
          <button
            onClick={stepForward}
            disabled={atEnd}
            aria-label="Step forward"
            className="rounded px-2 py-1 text-sm font-medium hover:bg-(--surface-2) disabled:opacity-30 disabled:cursor-not-allowed"
          >
            ⏭
          </button>
        </div>
        <div className="flex items-center gap-0.5 bg-(--surface-2) rounded-md p-0.5">
          {SPEEDS.map((s) => (
            <button
              key={s}
              onClick={() => setSpeed(s)}
              aria-label={`Speed ${s}x`}
              aria-pressed={speed === s}
              className={`rounded px-2 py-1 text-[10px] font-medium font-mono ${
                speed === s
                  ? "bg-(--surface-3) text-(--text-0)"
                  : "text-(--text-2) hover:text-(--text-1)"
              }`}
            >
              {s}×
            </button>
          ))}
        </div>
        <span className="ml-auto font-mono text-[10px] tabular-nums text-(--text-2)">
          event {currentIndex + 1}/{totalEvents}
        </span>
      </div>

      <div className="flex items-center gap-3 border-b border-(--border-light) px-4 py-1.5 shrink-0">
        <input
          type="range"
          min={-1}
          max={totalEvents - 1}
          value={currentIndex}
          onChange={(e) => seekTo(Number(e.target.value))}
          aria-label="Event scrubber"
          className="h-1.5 flex-1 cursor-pointer appearance-none rounded-full bg-(--surface-3) accent-emerald-500 [&::-webkit-slider-thumb]:h-3 [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-emerald-400"
        />
        <span className="shrink-0 font-mono text-[10px] tabular-nums text-(--text-2)">
          {currentIndex + 1} / {totalEvents}
        </span>
      </div>
    </>
  );
}

function App() {
  const length = useEventStore((s) => s.length);
  const capacity = useEventStore((s) => s.capacity);
  const resize = useEventStore((s) => s.resize);
  const retiring = useEventStore((s) => s.retiring);
  const pastArrays = useEventStore((s) => s.pastArrays);
  const elements = useEventStore((s) => s.elements);
  const totalEvents = useEventStore((s) => s.events.length);
  const strategy = useEventStore((s) => s.strategy);
  const setStrategy = useEventStore((s) => s.setStrategy);
  const error = useEventStore((s) => s.error);
  const setError = useEventStore((s) => s.setError);
  usePlayback();

  const hasEvents = totalEvents > 0;

  return (
    <div className="grid h-screen grid-cols-1 md:grid-cols-[220px_1fr] bg-(--surface-0) text-(--text-0)">
      <Rail />

      {/* Mobile header */}
      <div className="flex md:hidden items-center gap-3 border-b border-(--border) px-4 py-2">
        <span className="text-lg font-bold tracking-tight">grow</span>
        <select
          value={strategy}
          onChange={(e) => setStrategy(e.target.value as typeof strategy)}
          aria-label="Growth strategy"
          className="ml-auto rounded bg-(--surface-2) border border-(--border) px-1 py-0.5 text-xs"
        >
          <option value="cpython">CPython</option>
          <option value="doubling">Doubling</option>
          <option value="1.5x">1.5×</option>
          <option value="nogrowth">No growth</option>
        </select>
      </div>

      {/* Center stage */}
      <main className="relative flex flex-col min-h-0 overflow-hidden bg-(--surface-0)">
        {error && (
          <div className="flex items-center gap-3 bg-red-500/10 border-b border-red-500/20 px-4 py-2.5 text-sm text-red-400 shrink-0">
            <span className="font-medium">{error}</span>
            <button
              onClick={() => setError(null)}
              className="ml-auto text-red-400/60 hover:text-red-400 text-xs"
            >
              dismiss
            </button>
          </div>
        )}
        <PresetsPanel />
        {hasEvents && <TransportBar />}
        <div className="min-h-0 flex-1 overflow-auto">
          <CanvasArray
            length={length}
            capacity={capacity}
            resize={resize}
            retiring={retiring}
            pastArrays={pastArrays}
            elements={elements}
          />
        </div>
        <CostGraph />
      </main>
    </div>
  );
}

export default App;
