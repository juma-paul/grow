import { useState, useRef, useCallback, useEffect, type ReactNode } from "react";

type Tab = "code" | "graph";

interface BottomDockProps {
  codeContent: ReactNode;
  graphContent: ReactNode;
}

const MIN_HEIGHT = 36;
const DEFAULT_HEIGHT = 260;
const STORAGE_KEY = "grow-dock";

function loadState(): { height: number; tab: Tab; collapsed: boolean } {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return JSON.parse(raw);
  } catch {}
  return { height: DEFAULT_HEIGHT, tab: "code", collapsed: false };
}

function saveState(height: number, tab: Tab, collapsed: boolean) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ height, tab, collapsed }));
  } catch {}
}

export default function BottomDock({ codeContent, graphContent }: BottomDockProps) {
  const saved = useRef(loadState());
  const [height, setHeight] = useState(saved.current.height);
  const [activeTab, setActiveTab] = useState<Tab>(saved.current.tab);
  const [collapsed, setCollapsed] = useState(saved.current.collapsed);
  const dragging = useRef(false);
  const startY = useRef(0);
  const startH = useRef(0);

  useEffect(() => {
    saveState(height, activeTab, collapsed);
  }, [height, activeTab, collapsed]);

  const onMouseDown = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      dragging.current = true;
      startY.current = e.clientY;
      startH.current = collapsed ? DEFAULT_HEIGHT : height;

      const onMouseMove = (ev: MouseEvent) => {
        if (!dragging.current) return;
        const delta = startY.current - ev.clientY;
        const maxH = window.innerHeight * 0.6;
        const newH = Math.max(80, Math.min(maxH, startH.current + delta));
        setHeight(newH);
        if (collapsed) setCollapsed(false);
      };

      const onMouseUp = () => {
        dragging.current = false;
        document.removeEventListener("mousemove", onMouseMove);
        document.removeEventListener("mouseup", onMouseUp);
        document.body.style.cursor = "";
        document.body.style.userSelect = "";
      };

      document.body.style.cursor = "row-resize";
      document.body.style.userSelect = "none";
      document.addEventListener("mousemove", onMouseMove);
      document.addEventListener("mouseup", onMouseUp);
    },
    [height, collapsed],
  );

  const toggleCollapse = useCallback(() => {
    setCollapsed((c) => !c);
  }, []);

  const switchTab = useCallback(
    (tab: Tab) => {
      if (activeTab === tab && !collapsed) {
        setCollapsed(true);
      } else {
        setActiveTab(tab);
        if (collapsed) setCollapsed(false);
      }
    },
    [activeTab, collapsed],
  );

  const displayH = collapsed ? MIN_HEIGHT : height;

  return (
    <div
      className="flex flex-col border-t border-[var(--border)] bg-[var(--surface-1)] shrink-0 select-none"
      style={{ height: displayH }}
    >
      {/* Drag handle */}
      <div
        onMouseDown={onMouseDown}
        onDoubleClick={toggleCollapse}
        className="flex items-center justify-center h-[4px] cursor-row-resize group shrink-0"
      >
        <div className="w-10 h-[3px] rounded-full bg-[var(--border)] group-hover:bg-[var(--text-2)] transition-colors" />
      </div>

      {/* Tab bar */}
      <div
        className="flex items-center gap-0.5 px-3 border-b border-[var(--border-light)] shrink-0"
        style={{ height: MIN_HEIGHT - 4 }}
      >
        <button
          onClick={() => switchTab("code")}
          className={`px-3 py-1 rounded text-[11px] font-medium transition-colors ${
            activeTab === "code" && !collapsed
              ? "bg-emerald-400/[.12] text-emerald-400"
              : "text-[var(--text-2)] hover:text-[var(--text-1)]"
          }`}
        >
          Code
        </button>
        <button
          onClick={() => switchTab("graph")}
          className={`px-3 py-1 rounded text-[11px] font-medium transition-colors ${
            activeTab === "graph" && !collapsed
              ? "bg-emerald-400/[.12] text-emerald-400"
              : "text-[var(--text-2)] hover:text-[var(--text-1)]"
          }`}
        >
          Graph
        </button>
        <button
          onClick={toggleCollapse}
          aria-label={collapsed ? "Expand dock" : "Collapse dock"}
          className="ml-auto text-[var(--text-2)] hover:text-[var(--text-1)] text-[10px] transition-colors px-1.5 py-0.5 rounded hover:bg-[var(--surface-2)]"
        >
          {collapsed ? "▲" : "▼"}
        </button>
      </div>

      {/* Content */}
      {!collapsed && (
        <div className="flex-1 min-h-0 overflow-hidden">
          <div className={activeTab === "code" ? "h-full" : "hidden"}>
            {codeContent}
          </div>
          <div className={activeTab === "graph" ? "h-full overflow-auto" : "hidden"}>
            {graphContent}
          </div>
        </div>
      )}
    </div>
  );
}
