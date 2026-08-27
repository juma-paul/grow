import { Editor } from "@monaco-editor/react";
import CanvasArray from "./components/CanvasArray";
import { useEventStream } from "./hooks/useEventStream";
import { useEventStore } from "./store";
import { usePlayback } from "./hooks/usePlayback";

function App() {
  const length = useEventStore((state) => state.length);
  const capacity = useEventStore((state) => state.capacity);
  const currentIndex = useEventStore((state) => state.currentIndex);
  const totalEvents = useEventStore((state) => state.events.length);
  const { connect } = useEventStream();
  usePlayback();

  return (
    <div className="flex h-screen flex-col bg-zinc-900 text-white">
      <div className="flex items-center gap-4 px-4 py-2 border-b border-zinc-700">
        <h1 className="text-lg font-semibold">grow</h1>
        <button
          onClick={() => connect()}
          className="rounded bg-emerald-600 px-3 py-1 text-sm hover:bg-emerald-500"
        >
          Run
        </button>
        <span className="text-sm text-zinc-400">
          length: {length} / capacity: {capacity} — event {currentIndex + 1}/
          {totalEvents}
        </span>
      </div>
      <div className="flex flex-1 overflow-hidden">
        <div className="w-1/2 overflow-auto border-r border-zinc-700">
          <CanvasArray length={length} capacity={capacity} />
        </div>
        <div className="w-1/2">
          <Editor
            height="100%"
            defaultLanguage="python"
            defaultValue={
              "# Try some Python list operations\nlst = []\nfor i in range(100):\n    lst.append(i)\n"
            }
            theme="vs-dark"
          />
        </div>
      </div>
    </div>
  );
}

export default App;
