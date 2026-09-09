import { useState } from "react";
import { Editor } from "@monaco-editor/react";
import { useEventStream } from "../hooks/useEventStream";

const DEFAULT_CODE = `# Write normal Python — list ops are auto-rewritten
x = []
for i in range(20):
    x.append(i)
`;

export default function AutoPanel() {
  const [source, setSource] = useState(DEFAULT_CODE);
  const { connectCode } = useEventStream();

  return (
    <div className="flex flex-col border-b border-zinc-800">
      <div className="flex items-center gap-2 px-4 py-2">
        <span className="text-xs font-medium uppercase tracking-wider text-zinc-500">
          Auto
        </span>
        <button
          onClick={() => connectCode(source, true)}
          className="ml-auto rounded bg-emerald-600 px-3 py-1 text-sm font-medium hover:bg-emerald-500"
        >
          Run
        </button>
      </div>
      <div className="h-48">
        <Editor
          height="100%"
          defaultLanguage="python"
          value={source}
          onChange={(v) => setSource(v ?? "")}
          theme="vs-dark"
          options={{
            minimap: { enabled: false },
            fontSize: 13,
            lineNumbers: "on",
            scrollBeyondLastLine: false,
            padding: { top: 8 },
          }}
        />
      </div>
    </div>
  );
}
