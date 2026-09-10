import { useState, type ReactNode } from "react";
import { useEventStore } from "../store";

function Tip({ children, label }: { children: ReactNode; label: string }) {
  const [show, setShow] = useState(false);
  return (
    <span
      className="relative cursor-help border-b border-dashed border-[#565b6b]"
      onMouseEnter={() => setShow(true)}
      onMouseLeave={() => setShow(false)}
    >
      {children}
      {show && (
        <span className="absolute left-1/2 -translate-x-1/2 bottom-full mb-1.5 whitespace-nowrap rounded bg-[#252b3a] border border-[#262d3d] px-2 py-1 text-[10px] font-normal text-[#e2e4ea] shadow-lg z-10">
          {label}
        </span>
      )}
    </span>
  );
}

function cpythonCapacity(needed: number): number {
  return (needed + (needed >> 3) + 6) & ~3;
}

function doublingCapacity(needed: number): number {
  let cap = 4;
  while (cap < needed) cap *= 2;
  return cap;
}

function javaCapacity(needed: number): number {
  let cap = 4;
  while (cap < needed) cap = Math.floor(cap * 1.5);
  return cap;
}

export default function ExplanationPanel() {
  const lastResize = useEventStore((s) => s.lastResize);
  const strategy = useEventStore((s) => s.strategy);

  if (!lastResize) return null;

  const { oldCap, newCap, needed } = lastResize;
  const shift = needed >> 3;
  const raw = needed + shift + 6;
  const aligned = raw & ~3;

  return (
    <aside className="flex flex-col bg-[#171b24] border-l border-[#262d3d] w-full min-h-0 overflow-y-auto">
      {/* Header */}
      <div className="px-5 py-4 border-b border-[#1e2430]">
        <div className="flex items-center gap-2">
          <span className="text-[13px] font-semibold text-[#e2e4ea]">
            Resize event
          </span>
          <span className="rounded font-[var(--font-mono)] text-[10px] font-medium px-[7px] py-[2px] bg-amber-500/10 text-amber-500">
            resize_begin
          </span>
        </div>
      </div>

      {/* Body */}
      <div className="px-5 py-5 flex-1">
        {/* What's happening */}
        <div className="mb-6">
          <div className="text-[10px] font-semibold uppercase tracking-[0.8px] text-[#565b6b] mb-2.5">
            What's happening
          </div>
          <div className="text-[13px] leading-[1.6] text-[#8b90a0]">
            Array is full at capacity{" "}
            <strong className="text-[#e2e4ea] font-semibold">{oldCap}</strong>.
            {" "}Append #{needed} triggers a resize. The{" "}
            <span className="font-[var(--font-mono)] text-[12px] font-medium text-amber-500">
              {strategy === "cpython"
                ? "CPython"
                : strategy === "doubling"
                  ? "Doubling"
                  : strategy === "1.5x"
                    ? "1.5×"
                    : "Fixed"}
            </span>{" "}
            strategy computes new capacity:
          </div>

          {/* Formula breakdown */}
          <div className="mt-3 rounded-[10px] bg-[#1e2330] border border-[#262d3d] p-4">
            <div className="flex justify-between items-baseline py-1">
              <span className="font-[var(--font-mono)] text-[12px] text-[#8b90a0]">
                needed
              </span>
              <span className="font-[var(--font-mono)] text-[13px] font-semibold tabular-nums text-[#e2e4ea]">
                {needed}
              </span>
            </div>
            <div className="flex justify-between items-baseline py-1">
              <span className="font-[var(--font-mono)] text-[12px] text-[#8b90a0]">
                needed{" "}
                <Tip label="divide by 8">&gt;&gt; 3</Tip>
              </span>
              <span className="font-[var(--font-mono)] text-[13px] font-semibold tabular-nums text-[#e2e4ea]">
                {shift}
              </span>
            </div>
            <div className="flex justify-between items-baseline py-1">
              <span className="font-[var(--font-mono)] text-[12px] text-[#8b90a0]">
                <Tip label="padding constant">+ 6</Tip>, then{" "}
                <Tip label="round down to multiple of 4">&amp; ~3</Tip>
              </span>
              <span className="font-[var(--font-mono)] text-[13px] font-semibold tabular-nums text-[#e2e4ea]">
                {raw} → {aligned}
              </span>
            </div>
            <div className="h-px bg-[#262d3d] my-1.5" />
            <div className="flex justify-between items-baseline py-1">
              <span className="font-[var(--font-mono)] text-[12px] text-[#8b90a0]">
                new capacity
              </span>
              <span className="font-[var(--font-mono)] text-[16px] font-semibold tabular-nums text-amber-500">
                {newCap}
              </span>
            </div>
          </div>
        </div>

        {/* Why this formula */}
        <div className="mb-6">
          <div className="text-[10px] font-semibold uppercase tracking-[0.8px] text-[#565b6b] mb-2.5">
            Why this formula
          </div>
          <div className="text-[13px] leading-[1.6] text-[#8b90a0]">
            Grows by{" "}
            <strong className="text-[#e2e4ea] font-semibold">~12.5%</strong>{" "}
            plus padding aligned to 4. More copies than doubling, but wastes
            less memory. Amortized cost stays O(1) because each element "pays
            forward" for the next resize.
          </div>
        </div>

        {/* Compare */}
        <div>
          <div className="text-[10px] font-semibold uppercase tracking-[0.8px] text-[#565b6b] mb-2.5">
            Compare
          </div>
          {oldCap > 0 ? (
            <div className="font-[var(--font-mono)] text-[11px] leading-[1.8] text-[#8b90a0]">
              <div>
                <span className={strategy === "cpython" ? "text-emerald-400" : "text-[#565b6b]"}>
                  CPython
                </span>
                {"  "}{oldCap} → {cpythonCapacity(needed)}{" "}
                <span className="text-[#565b6b]">
                  (+{Math.round(((cpythonCapacity(needed) - oldCap) / oldCap) * 100)}% at this size)
                </span>
              </div>
              <div>
                <span className={strategy === "doubling" ? "text-emerald-400" : "text-[#565b6b]"}>
                  Doubling
                </span>
                {"  "}{oldCap} → {doublingCapacity(needed)}{" "}
                <span className="text-[#565b6b]">
                  (+{Math.round(((doublingCapacity(needed) - oldCap) / oldCap) * 100)}%)
                </span>
              </div>
              <div>
                <span className={strategy === "1.5x" ? "text-emerald-400" : "text-[#565b6b]"}>
                  Java 1.5×
                </span>
                {" "}{oldCap} → {javaCapacity(needed)}{" "}
                <span className="text-[#565b6b]">
                  (+{Math.round(((javaCapacity(needed) - oldCap) / oldCap) * 100)}%)
                </span>
              </div>
              <div>
                <span className={strategy === "nogrowth" ? "text-emerald-400" : "text-[#565b6b]"}>
                  Fixed
                </span>
                {"  "}<span className="text-red-500">overflow</span>
              </div>
            </div>
          ) : (
            <p className="text-[#565b6b] text-[11px]">
              Initial allocation — no prior capacity to compare.
            </p>
          )}
        </div>
      </div>
    </aside>
  );
}
