import CanvasArray from "./components/CanvasArray";
import CodeDrawer from "./components/CodeDrawer";
import { useEventStream } from "./hooks/useEventStream";
import { useEventStore } from "./store";
import { usePlayback } from "./hooks/usePlayback";

const PRESETS = [5, 10, 15, 20];

function App() {
  const length = useEventStore((state) => state.length);
  const capacity = useEventStore((state) => state.capacity);
  const resize = useEventStore((state) => state.resize);
  const retiring = useEventStore((state) => state.retiring);
  const pastArrays = useEventStore((state) => state.pastArrays);
  const currentIndex = useEventStore((state) => state.currentIndex);
  const totalEvents = useEventStore((state) => state.events.length);
  const count = useEventStore((state) => state.count);
  const setCount = useEventStore((state) => state.setCount);
  const { connect } = useEventStream();
  usePlayback();

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
        <span className="text-sm text-zinc-400">
          length: {length} / capacity: {capacity}
          {totalEvents > 0 && ` — event ${currentIndex + 1}/${totalEvents}`}
        </span>
      </div>

      <div className="flex-1 overflow-auto pb-10">
        <CanvasArray
          length={length}
          capacity={capacity}
          resize={resize}
          retiring={retiring}
          pastArrays={pastArrays}
        />
      </div>

      <CodeDrawer />
    </div>
  );
}

export default App;
