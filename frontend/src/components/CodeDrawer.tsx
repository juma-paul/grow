import { useState, useCallback, useEffect, useRef } from "react";
import { Editor } from "@monaco-editor/react";

export default function CodeDrawer() {
  const [isOpen, setIsOpen] = useState(false);
  const [height, setHeight] = useState(256);
  const isDragging = useRef(false);
  const startY = useRef(0);
  const startHeight = useRef(0);

  const onMouseDown = useCallback(
    (e: React.MouseEvent) => {
      isDragging.current = true;
      startY.current = e.clientY;
      startHeight.current = height;
      document.body.style.cursor = "row-resize";
      document.body.style.userSelect = "none";
    },
    [height],
  );

  const onMouseMove = useCallback((e: MouseEvent) => {
    if (!isDragging.current) return;
    // Dragging up increases height (startY is lower on screen)
    const delta = startY.current - e.clientY;
    const newHeight = Math.min(600, Math.max(120, startHeight.current + delta));
    setHeight(newHeight);
  }, []);

  const onMouseUp = useCallback(() => {
    isDragging.current = false;
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
  }, []);

  useEffect(() => {
    window.addEventListener("mousemove", onMouseMove);
    window.addEventListener("mouseup", onMouseUp);
    return () => {
      window.removeEventListener("mousemove", onMouseMove);
      window.removeEventListener("mouseup", onMouseUp);
    };
  }, [onMouseMove, onMouseUp]);

  return (
    <div className="flex-shrink-0">
      {/* Toggle + drag bar */}
      <div className="flex items-center border-t border-zinc-700 bg-zinc-800">
        <button
          onClick={() => setIsOpen(!isOpen)}
          className="flex items-center gap-2 px-4 py-1.5 text-sm text-zinc-400 hover:text-white transition-colors cursor-pointer"
        >
          <span
            className={`transition-transform duration-200 text-xs ${isOpen ? "rotate-180" : ""}`}
          >
            ▲
          </span>
          <span>Code</span>
        </button>

        {isOpen && (
          <div
            onMouseDown={onMouseDown}
            className="flex-1 h-full cursor-row-resize flex items-center justify-center"
          >
            <div className="w-12 h-1 rounded-full bg-zinc-600 hover:bg-emerald-500 transition-colors" />
          </div>
        )}
      </div>

      {/* Editor panel */}
      <div
        className="overflow-hidden transition-all duration-300 ease-out"
        style={{ height: isOpen ? `${height}px` : "0px" }}
      >
        {isOpen && (
          <Editor
            height="100%"
            defaultLanguage="python"
            defaultValue={
              "# Try some Python list operations\nlst = []\nfor i in range(20):\n    lst.append(i)\n"
            }
            theme="vs-dark"
          />
        )}
      </div>
    </div>
  );
}
