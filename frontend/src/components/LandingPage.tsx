import { useEffect, useState } from "react";
import { useEventStore } from "../store";
import { useTheme } from "../hooks/useTheme";

interface Stats {
  total_runs: number;
  runs_today: number;
  elements_allocated: number;
  countries?: Record<string, number>;
}


function fmt(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + "M";
  if (n >= 1_000) return (n / 1_000).toFixed(1) + "K";
  return n.toLocaleString();
}

const FEATURES = [
  {
    title: "Four growth strategies",
    desc: "Compare CPython's ~1.125x, doubling, 1.5x, and fixed allocation side by side.",
  },
  {
    title: "Step-by-step playback",
    desc: "Scrub through every append, resize, and copy operation at your own pace.",
  },
  {
    title: "Amortized cost graph",
    desc: "See how the cost-per-operation converges to O(1) as elements accumulate.",
  },
  {
    title: "Write real Python",
    desc: "Paste your own list code and watch the allocator respond to appends, pops, and extends.",
  },
];

export default function LandingPage() {
  const setPage = useEventStore((s) => s.setPage);
  const { theme, toggle: toggleTheme } = useTheme();
  const [stats, setStats] = useState<Stats | null>(null);

  useEffect(() => {
    fetch("/stats")
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => {
        if (data) setStats(data);
      })
      .catch(() => {});
  }, []);

  return (
    <div className="min-h-screen bg-[var(--surface-0)] text-[var(--text-0)]">
      {/* Nav */}
      <nav className="flex items-center justify-between px-6 py-4 max-w-5xl mx-auto">
        <div className="flex items-center gap-2">
          <span className="inline-block h-2.5 w-2.5 rounded-full bg-emerald-400 shadow-[0_0_10px_rgba(52,211,153,0.3)]" />
          <span className="font-[var(--font-mono)] text-2xl font-semibold tracking-tight text-emerald-400">
            grow
          </span>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={toggleTheme}
            className="flex items-center justify-center w-8 h-8 rounded-lg border border-[var(--border)] bg-[var(--surface-1)] text-sm text-[var(--text-2)] hover:text-[var(--text-1)] hover:border-[var(--text-2)] transition-colors"
            aria-label={`Switch to ${theme === "dark" ? "light" : "dark"} mode`}
          >
            {theme === "dark" ? "☀" : "☾"}
          </button>
          <a
            href="https://github.com/juma-paul/grow"
            target="_blank"
            rel="noopener noreferrer"
            className="text-sm text-[var(--text-2)] hover:text-[var(--text-1)] transition-colors"
          >
            GitHub
          </a>
        </div>
      </nav>

      {/* Hero */}
      <section className="px-6 pt-16 pb-20 max-w-5xl mx-auto text-center">
        <h1 className="text-4xl sm:text-5xl font-bold tracking-tight leading-tight mb-6">
          See how Python lists{" "}
          <span className="text-emerald-400">actually grow</span>
        </h1>
        <p className="text-lg sm:text-xl text-[var(--text-1)] max-w-2xl mx-auto mb-10 leading-relaxed">
          An interactive visualizer for CPython's list allocation strategy.
          Watch appends trigger resizes, trace element copies, and understand
          why amortized cost stays O(1).
        </p>
        <button
          onClick={() => setPage("app")}
          className="inline-flex items-center gap-2 rounded-lg bg-emerald-500 hover:bg-emerald-400 text-white font-semibold px-8 py-3.5 text-lg transition-colors shadow-lg shadow-emerald-500/20"
        >
          Launch visualizer
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" className="ml-1">
            <path d="M7 4l6 6-6 6" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </button>
      </section>

      {/* Stats */}
      {stats && (
        <section className="px-6 pb-16 max-w-5xl mx-auto">
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 max-w-2xl mx-auto">
            <div className="flex flex-col items-center gap-1 rounded-xl bg-[var(--surface-1)] border border-[var(--border)] px-4 py-5">
              <span className="font-[var(--font-mono)] text-2xl font-bold tabular-nums text-emerald-400">
                {fmt(stats.total_runs)}
              </span>
              <span className="text-xs text-[var(--text-2)]">total runs</span>
            </div>
            <div className="flex flex-col items-center gap-1 rounded-xl bg-[var(--surface-1)] border border-[var(--border)] px-4 py-5">
              <span className="font-[var(--font-mono)] text-2xl font-bold tabular-nums text-amber-500">
                {fmt(stats.runs_today)}
              </span>
              <span className="text-xs text-[var(--text-2)]">today</span>
            </div>
            <div className="flex flex-col items-center gap-1 rounded-xl bg-[var(--surface-1)] border border-[var(--border)] px-4 py-5">
              <span className="font-[var(--font-mono)] text-2xl font-bold tabular-nums text-violet-400">
                {fmt(stats.elements_allocated)}
              </span>
              <span className="text-xs text-[var(--text-2)]">elements allocated</span>
            </div>
            {stats.countries && Object.keys(stats.countries).length > 0 && (
              <div className="flex flex-col items-center gap-1 rounded-xl bg-[var(--surface-1)] border border-[var(--border)] px-4 py-5">
                <span className="font-[var(--font-mono)] text-2xl font-bold tabular-nums text-sky-400">
                  {Object.keys(stats.countries).length}
                </span>
                <span className="text-xs text-[var(--text-2)]">countries</span>
              </div>
            )}
          </div>
        </section>
      )}

      {/* Features */}
      <section className="px-6 pb-20 max-w-5xl mx-auto">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-5 max-w-3xl mx-auto">
          {FEATURES.map((f) => (
            <div
              key={f.title}
              className="rounded-xl bg-[var(--surface-1)] border border-[var(--border)] p-6"
            >
              <h3 className="font-semibold text-[15px] mb-2">{f.title}</h3>
              <p className="text-sm text-[var(--text-1)] leading-relaxed">
                {f.desc}
              </p>
            </div>
          ))}
        </div>
      </section>

      {/* How it works */}
      <section className="px-6 pb-20 max-w-5xl mx-auto text-center">
        <h2 className="text-2xl font-bold mb-8">How it works</h2>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 max-w-3xl mx-auto">
          <div className="flex flex-col items-center gap-3">
            <div className="flex items-center justify-center w-10 h-10 rounded-full bg-emerald-400/[.12] text-emerald-400 font-[var(--font-mono)] font-bold text-lg">
              1
            </div>
            <p className="text-sm text-[var(--text-1)]">
              Pick a growth strategy or write your own Python code
            </p>
          </div>
          <div className="flex flex-col items-center gap-3">
            <div className="flex items-center justify-center w-10 h-10 rounded-full bg-emerald-400/[.12] text-emerald-400 font-[var(--font-mono)] font-bold text-lg">
              2
            </div>
            <p className="text-sm text-[var(--text-1)]">
              The server simulates every append, resize, and copy
            </p>
          </div>
          <div className="flex flex-col items-center gap-3">
            <div className="flex items-center justify-center w-10 h-10 rounded-full bg-emerald-400/[.12] text-emerald-400 font-[var(--font-mono)] font-bold text-lg">
              3
            </div>
            <p className="text-sm text-[var(--text-1)]">
              Step through the animation and see amortized cost converge
            </p>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-[var(--border-light)] px-6 py-6 max-w-5xl mx-auto">
        <div className="flex items-center justify-between text-xs text-[var(--text-2)]">
          <span>A learning tool by Juma Paul</span>
          <a
            href="https://github.com/juma-paul/grow"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-[var(--text-1)] transition-colors"
          >
            Source on GitHub
          </a>
        </div>
      </footer>
    </div>
  );
}
