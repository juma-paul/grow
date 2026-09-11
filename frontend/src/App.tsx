import CanvasArray from "./components/CanvasArray";
import CostGraph from "./components/CostGraph";
import PresetsPanel from "./components/PresetsPanel";
import WrapperPanel from "./components/WrapperPanel";
import AutoPanel from "./components/AutoPanel";
import BottomDock from "./components/BottomDock";
import Rail from "./components/Rail";
import ExplanationPanel from "./components/ExplanationPanel";
import Onboarding from "./components/Onboarding";
import ShortcutsOverlay from "./components/ShortcutsOverlay";
import LandingPage from "./components/LandingPage";
import { useEventStore } from "./store";
import { usePlayback } from "./hooks/usePlayback";
import { useKeyboardShortcuts } from "./hooks/useKeyboardShortcuts";
import { useTheme } from "./hooks/useTheme";

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
      <div className="flex items-center gap-3 border-b border-[var(--border)] px-4 py-2 bg-[var(--surface-1)] shrink-0">
        <div className="flex items-center gap-1">
          <button
            onClick={stepBack}
            disabled={atStart}
            aria-label="Step back"
            className="rounded px-2 py-1 text-sm font-medium hover:bg-[var(--surface-2)] disabled:opacity-30 disabled:cursor-not-allowed"
          >
            ⏮
          </button>
          <button
            onClick={isPlaying ? pause : play}
            disabled={atEnd && !isPlaying}
            aria-label={isPlaying ? "Pause" : "Play"}
            className="rounded px-2 py-1 text-sm font-medium hover:bg-[var(--surface-2)] disabled:opacity-30 disabled:cursor-not-allowed"
          >
            {isPlaying ? "⏸" : "▶"}
          </button>
          <button
            onClick={stepForward}
            disabled={atEnd}
            aria-label="Step forward"
            className="rounded px-2 py-1 text-sm font-medium hover:bg-[var(--surface-2)] disabled:opacity-30 disabled:cursor-not-allowed"
          >
            ⏭
          </button>
        </div>
        <div className="flex items-center gap-0.5 bg-[var(--surface-2)] rounded-md p-0.5">
          {SPEEDS.map((s) => (
            <button
              key={s}
              onClick={() => setSpeed(s)}
              aria-label={`Speed ${s}x`}
              aria-pressed={speed === s}
              className={`rounded px-2 py-1 text-[10px] font-medium font-[var(--font-mono)] ${
                speed === s
                  ? "bg-[var(--surface-3)] text-[var(--text-0)]"
                  : "text-[var(--text-2)] hover:text-[var(--text-1)]"
              }`}
            >
              {s}×
            </button>
          ))}
        </div>
        <span className="ml-auto font-[var(--font-mono)] text-[10px] tabular-nums text-[var(--text-2)]">
          event {currentIndex + 1}/{totalEvents}
        </span>
      </div>

      <div className="flex items-center gap-3 border-b border-[var(--border-light)] px-4 py-1.5 shrink-0">
        <input
          type="range"
          min={-1}
          max={totalEvents - 1}
          value={currentIndex}
          onChange={(e) => seekTo(Number(e.target.value))}
          aria-label="Event scrubber"
          className="h-1.5 flex-1 cursor-pointer appearance-none rounded-full bg-[var(--surface-3)] accent-emerald-500 [&::-webkit-slider-thumb]:h-3 [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-emerald-400"
        />
        <span className="shrink-0 font-[var(--font-mono)] text-[10px] tabular-nums text-[var(--text-2)]">
          {currentIndex + 1} / {totalEvents}
        </span>
      </div>
    </>
  );
}

function Visualizer() {
  const length = useEventStore((s) => s.length);
  const capacity = useEventStore((s) => s.capacity);
  const resize = useEventStore((s) => s.resize);
  const retiring = useEventStore((s) => s.retiring);
  const pastArrays = useEventStore((s) => s.pastArrays);
  const totalEvents = useEventStore((s) => s.events.length);
  const mode = useEventStore((s) => s.mode);
  const strategy = useEventStore((s) => s.strategy);
  const setMode = useEventStore((s) => s.setMode);
  const setStrategy = useEventStore((s) => s.setStrategy);
  const lastResize = useEventStore((s) => s.lastResize);
  const error = useEventStore((s) => s.error);
  const setError = useEventStore((s) => s.setError);
  usePlayback();
  const { showHelp, setShowHelp } = useKeyboardShortcuts();
  const { theme, toggle: toggleTheme } = useTheme();

  const hasEvents = totalEvents > 0;
  const showExplanation = !!lastResize;
  const isCodeMode = mode === "wrapper" || mode === "auto";

  return (
    <div
      className={`grid h-screen grid-cols-1 bg-[var(--surface-0)] text-[var(--text-0)] ${
        showExplanation
          ? "md:grid-cols-[220px_1fr_320px]"
          : "md:grid-cols-[220px_1fr]"
      }`}
    >
      <Rail onShowHelp={() => setShowHelp(true)} theme={theme} onToggleTheme={toggleTheme} />

      {/* Mobile header */}
      <div className="flex md:hidden items-center gap-3 border-b border-[var(--border)] px-4 py-2">
        <span className="text-lg font-bold tracking-tight">grow</span>
        <div className="flex gap-1">
          {(["presets", "wrapper", "auto"] as const).map((m) => (
            <button
              key={m}
              onClick={() => setMode(m)}
              aria-pressed={mode === m}
              className={`rounded px-2 py-0.5 text-xs font-medium ${
                mode === m
                  ? "bg-emerald-400/[.12] text-emerald-400"
                  : "text-[var(--text-2)]"
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
          className="ml-auto rounded bg-[var(--surface-2)] border border-[var(--border)] px-1 py-0.5 text-xs"
        >
          <option value="cpython">CPython</option>
          <option value="doubling">Doubling</option>
          <option value="1.5x">1.5×</option>
          <option value="nogrowth">No growth</option>
        </select>
      </div>

      {/* Center stage */}
      <main className="relative flex flex-col min-h-0 overflow-hidden bg-[var(--surface-0)]">
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
        {mode === "presets" && <Onboarding />}
        {showHelp && <ShortcutsOverlay onClose={() => setShowHelp(false)} />}
        {/* Presets mode: buttons at top, then transport, viz, graph */}
        {mode === "presets" && (
          <>
            <PresetsPanel />
            {hasEvents && <TransportBar />}
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
          </>
        )}

        {/* Wrapper/Auto mode: viz on top, transport, bottom dock with code+graph */}
        {isCodeMode && (
          <>
            <div className="min-h-0 flex-1 overflow-auto">
              <CanvasArray
                length={length}
                capacity={capacity}
                resize={resize}
                retiring={retiring}
                pastArrays={pastArrays}
              />
            </div>
            {hasEvents && <TransportBar />}
            <BottomDock
              codeContent={mode === "wrapper" ? <WrapperPanel theme={theme} /> : <AutoPanel theme={theme} />}
              graphContent={<CostGraph />}
            />
          </>
        )}
      </main>

      {/* Right sidebar */}
      {showExplanation && (
        <div className="hidden md:flex min-h-0">
          <ExplanationPanel />
        </div>
      )}
    </div>
  );
}

function App() {
  const page = useEventStore((s) => s.page);
  if (page === "home") return <LandingPage />;
  return <Visualizer />;
}

export default App;
