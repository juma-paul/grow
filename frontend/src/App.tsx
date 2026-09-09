import CanvasArray from "./components/CanvasArray";
import CostGraph from "./components/CostGraph";
import PresetsPanel from "./components/PresetsPanel";
import WrapperPanel from "./components/WrapperPanel";
import AutoPanel from "./components/AutoPanel";
import Rail from "./components/Rail";
import ExplanationPanel from "./components/ExplanationPanel";
import { useEventStore } from "./store";
import { usePlayback } from "./hooks/usePlayback";

const SPEEDS = [0.5, 1, 2, 5];

function App() {
  const length = useEventStore((s) => s.length);
  const capacity = useEventStore((s) => s.capacity);
  const resize = useEventStore((s) => s.resize);
  const retiring = useEventStore((s) => s.retiring);
  const pastArrays = useEventStore((s) => s.pastArrays);
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
  const mode = useEventStore((s) => s.mode);
  const strategy = useEventStore((s) => s.strategy);
  const setMode = useEventStore((s) => s.setMode);
  const setStrategy = useEventStore((s) => s.setStrategy);
  usePlayback();

  const hasEvents = totalEvents > 0;
  const atEnd = currentIndex >= totalEvents - 1;
  const atStart = currentIndex < 0;

  return (
    <div className="grid h-screen grid-cols-1 md:grid-cols-[240px_1fr] bg-zinc-900 text-white">
      <Rail />

      {/* Mobile header — visible on small screens only */}
      <div className="flex md:hidden items-center gap-3 border-b border-zinc-700 px-4 py-2">
        <span className="text-lg font-bold tracking-tight">grow</span>
        <div className="flex gap-1">
          {(["presets", "wrapper", "auto"] as const).map((m) => (
            <button
              key={m}
              onClick={() => setMode(m)}
              aria-pressed={mode === m}
              className={`rounded px-2 py-0.5 text-xs font-medium ${
                mode === m
                  ? "bg-emerald-600 text-white"
                  : "text-zinc-400"
              }`}
            >
              {m[0].toUpperCase()}
            </button>
          ))}
        </div>
        <select
          value={strategy}
          onChange={(e) => setStrategy(e.target.value as typeof strategy)}
          aria-label="Growth strategy"
          className="ml-auto rounded bg-zinc-800 border border-zinc-600 px-1 py-0.5 text-xs"
        >
          <option value="cpython">CPython</option>
          <option value="doubling">Doubling</option>
          <option value="1.5x">1.5×</option>
          <option value="nogrowth">No growth</option>
        </select>
      </div>

      {/* Center stage */}
      <main className="flex flex-col min-h-0 overflow-hidden">
        {/* Tab content */}
        {mode === "presets" && <PresetsPanel />}
        {mode === "wrapper" && <WrapperPanel />}
        {mode === "auto" && <AutoPanel />}

        {/* Transport bar — playback controls only */}
        {hasEvents && (
          <div className="flex items-center gap-3 border-b border-zinc-700 px-4 py-2">
            <div className="flex items-center gap-1">
              <button
                onClick={stepBack}
                disabled={atStart}
                aria-label="Step back"
                className="rounded px-2 py-1 text-sm font-medium hover:bg-zinc-700 disabled:opacity-30 disabled:cursor-not-allowed"
              >
                ⏮
              </button>
              <button
                onClick={isPlaying ? pause : play}
                disabled={atEnd && !isPlaying}
                aria-label={isPlaying ? "Pause" : "Play"}
                className="rounded px-2 py-1 text-sm font-medium hover:bg-zinc-700 disabled:opacity-30 disabled:cursor-not-allowed"
              >
                {isPlaying ? "⏸" : "▶"}
              </button>
              <button
                onClick={stepForward}
                disabled={atEnd}
                aria-label="Step forward"
                className="rounded px-2 py-1 text-sm font-medium hover:bg-zinc-700 disabled:opacity-30 disabled:cursor-not-allowed"
              >
                ⏭
              </button>
            </div>
            <div className="flex items-center gap-1">
              {SPEEDS.map((s) => (
                <button
                  key={s}
                  onClick={() => setSpeed(s)}
                  aria-label={`Speed ${s}x`}
                  aria-pressed={speed === s}
                  className={`rounded px-2 py-1 text-xs font-medium ${
                    speed === s
                      ? "bg-zinc-600 text-white"
                      : "text-zinc-400 hover:bg-zinc-700 hover:text-zinc-200"
                  }`}
                >
                  {s}×
                </button>
              ))}
            </div>
            <span className="ml-auto text-xs tabular-nums text-zinc-500">
              event {currentIndex + 1}/{totalEvents}
            </span>
          </div>
        )}

        {/* Scrubber */}
        {hasEvents && (
          <div className="flex items-center gap-3 border-b border-zinc-800 px-4 py-1.5">
            <input
              type="range"
              min={-1}
              max={totalEvents - 1}
              value={currentIndex}
              onChange={(e) => seekTo(Number(e.target.value))}
              aria-label="Event scrubber"
              className="h-1.5 flex-1 cursor-pointer appearance-none rounded-full bg-zinc-700 accent-emerald-500 [&::-webkit-slider-thumb]:h-3 [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-emerald-400"
            />
            <span className="shrink-0 text-xs tabular-nums text-zinc-500">
              {currentIndex + 1} / {totalEvents}
            </span>
          </div>
        )}

        {/* Viz area */}
        <div className="min-h-0 flex-1 overflow-auto">
          <CanvasArray
            length={length}
            capacity={capacity}
            resize={resize}
            retiring={retiring}
            pastArrays={pastArrays}
          />
        </div>

        <ExplanationPanel />
        <CostGraph />
      </main>
    </div>
  );
}

export default App;
