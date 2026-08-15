import { Editor } from "@monaco-editor/react";

function App() {
  return (
    <div className="flex h-screen flex-col bg-zinc-900 text-white">
      <h1 className="px-4 py-2 text-lg font-semibold">grow</h1>
      <Editor
        height="100%"
        defaultLanguage="python"
        defaultValue={
          "# Try some python list operations\nlst = []\nfor i in range(100):\n    lst.append(i)\n"
        }
        theme="vs-dark"
      />
    </div>
  );
}

export default App;
