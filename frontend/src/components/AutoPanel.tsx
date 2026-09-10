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
    <div className="h-full flex flex-col">
      <div className="flex items-center gap-2 px-4 py-1.5 border-b border-[#1e2430] shrink-0">
        <span className="flex items-center gap-1.5 text-[11px] font-medium text-[#8b90a0]">
          <span className="inline-block w-1.5 h-1.5 rounded-full bg-emerald-400" />
          auto.py
        </span>
        <button
          onClick={() => connectCode(source, true)}
          className="ml-auto flex items-center gap-1.5 rounded-md bg-emerald-400 px-3 py-1 text-[12px] font-semibold text-[#0f1117] hover:opacity-90"
        >
          <svg width="8" height="10" viewBox="0 0 10 12" fill="none">
            <path d="M0 0L10 6L0 12V0Z" fill="currentColor" />
          </svg>
          Run
        </button>
      </div>
      <div className="flex-1 min-h-0">
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
