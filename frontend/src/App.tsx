import CanvasArray from "./components/CanvasArray";
import CostGraph from "./components/CostGraph";
import CodeDrawer from "./components/CodeDrawer";
import { useEventStream } from "./hooks/useEventStream";
import { useEventStore } from "./store";
import { usePlayback } from "./hooks/usePlayback";

const PRESETS = [5, 10, 15, 20];
const SPEEDS = [0.5, 1, 2, 5];

function App() {
  const length = useEventStore((state) => state.length);
  const capacity = useEventStore((state) => state.capacity);
  const resize = useEventStore((state) => state.resize);
  const retiring = useEventStore((state) => state.retiring);
  const pastArrays = useEventStore((state) => state.pastArrays);
  const currentIndex = useEventStore((state) => state.currentIndex);
  const totalEvents = useEventStore((state) => state.events.length);
  const isPlaying = useEventStore((state) => state.isPlaying);
  const count = useEventStore((state) => state.count);
  const setCount = useEventStore((state) => state.setCount);
  const play = useEventStore((state) => state.play);
  const pause = useEventStore((state) => state.pause);
  const stepForward = useEventStore((state) => state.stepForward);
  const stepBack = useEventStore((state) => state.stepBack);
  const speed = useEventStore((state) => state.speed);
  const setSpeed = useEventStore((state) => state.setSpeed);
  const { connect } = useEventStream();
  usePlayback();

  const hasEvents = totalEvents > 0;
  const atEnd = currentIndex >= totalEvents - 1;
  const atStart = currentIndex < 0;

  return (
    <div className="relative flex h-screen flex-col bg-zinc-900 text-white">
      <div className="flex flex-wrap items-center gap-4 px-6 py-3 border-b border-zinc-700">
        <h1 className="text-xl font-bold">grow</h1>
        <select
          value={count}
          onChange={(e) => setCount(Number(e.target.value))}
          className="rounded bg-zinc-800 border border-zinc-600 px-2 py-1 text-sm"
        >
          {PRESETS.map((n) => (
            <option key={n} value={n}>
              Append {n}
            </option>
          ))}
        </select>
        <button
          onClick={() => connect()}
          className="rounded bg-emerald-600 px-4 py-1 text-sm font-medium hover:bg-emerald-500"
        >
          Run
        </button>
        {hasEvents && (
          <div className="flex items-center gap-1">
            <button
              onClick={stepBack}
              disabled={atStart}
              className="rounded px-2 py-1 text-sm font-medium hover:bg-zinc-700 disabled:opacity-30 disabled:cursor-not-allowed"
              title="Step back"
            >
              ⏮
            </button>
            <button
              onClick={isPlaying ? pause : play}
              disabled={atEnd && !isPlaying}
              className="rounded px-2 py-1 text-sm font-medium hover:bg-zinc-700 disabled:opacity-30 disabled:cursor-not-allowed"
              title={isPlaying ? "Pause" : "Play"}
            >
              {isPlaying ? "⏸" : "▶"}
            </button>
            <button
              onClick={stepForward}
              disabled={atEnd}
              className="rounded px-2 py-1 text-sm font-medium hover:bg-zinc-700 disabled:opacity-30 disabled:cursor-not-allowed"
              title="Step forward"
            >
              ⏭
            </button>
          </div>
        )}
        {hasEvents && (
          <div className="flex items-center gap-1">
            {SPEEDS.map((s) => (
              <button
                key={s}
                onClick={() => setSpeed(s)}
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
        )}
        <span className="text-sm text-zinc-400">
          length: {length} / capacity: {capacity}
          {hasEvents && ` — event ${currentIndex + 1}/${totalEvents}`}
        </span>
      </div>

      <div className="min-h-0 flex-1 overflow-auto">
        <CanvasArray
          length={length}
          capacity={capacity}
          resize={resize}
          retiring={retiring}
          pastArrays={pastArrays}
        />
      </div>

      <CostGraph />

      <CodeDrawer />
    </div>
  );
}

export default App;
