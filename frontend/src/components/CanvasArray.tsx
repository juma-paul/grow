import { useEffect } from "react";
import { useEventStore } from "../store";

interface ResizeState {
  oldCap: number;
  newCap: number;
  copied: number[];
}

interface RetiringState {
  capacity: number;
  length: number;
}

interface PastArray {
  capacity: number;
  length: number;
}

interface CanvasArrayProps {
  length: number;
  capacity: number;
  resize: ResizeState | null;
  retiring: RetiringState | null;
  pastArrays: PastArray[];
}

function Cell({
  index,
  filled,
  variant,
  isFull,
}: {
  index: number;
  filled: boolean;
  variant:
    | "normal"
    | "old"
    | "old-copied"
    | "new"
    | "new-arrived"
    | "past";
  isFull?: boolean;
}) {
  const styles: Record<string, string> = {
    normal: filled
      ? "border-emerald-500 bg-emerald-900 text-emerald-200"
      : "border-zinc-600 bg-zinc-800 text-zinc-500",
    old: "border-amber-500 bg-amber-900/30 text-amber-200",
    "old-copied":
      "border-zinc-600 bg-zinc-800/40 text-zinc-500 transition-[border-color,background-color,color] duration-500 ease-out",
    new: "border-zinc-600 bg-zinc-800/50 text-zinc-600",
    "new-arrived":
      "border-emerald-500 bg-emerald-900 text-emerald-200",
    past: filled
      ? "border-zinc-600 bg-zinc-800/40 text-zinc-500"
      : "border-zinc-700/50 bg-zinc-900/30 text-zinc-700",
  };

  // Anticipation: subtle amber pulse on filled cells when array is full
  const pulseStyle =
    variant === "normal" && filled && isFull
      ? { animation: "pulse-warn 1.5s ease-in-out infinite" }
      : undefined;

  // Secondary action: bounce on arrival in new array
  const arriveStyle =
    variant === "new-arrived"
      ? { animation: "cell-arrive 350ms cubic-bezier(0, 0, 0.2, 1) both" }
      : undefined;

  return (
    <div
      className={`flex h-12 w-12 items-center justify-center rounded-md border-2 text-sm font-mono font-semibold ${styles[variant]}`}
      style={pulseStyle || arriveStyle}
    >
      {variant === "old-copied" || variant === "new" ? "" : filled ? index : ""}
    </div>
  );
}

function PastArrayRow({ arr, index }: { arr: PastArray; index: number }) {
  return (
    <div className="mb-4 opacity-40">
      <div className="mb-1 text-xs text-zinc-500">
        resize #{index + 1} — capacity {arr.capacity}
      </div>
      <div className="flex flex-wrap gap-2">
        {Array.from({ length: arr.capacity }, (_, i) => (
          <Cell
            key={`past-${index}-${i}`}
            index={i}
            filled={i < arr.length}
            variant="past"
          />
        ))}
      </div>
    </div>
  );
}

// Single function to determine cell variant in old array section.
// Same DOM nodes stay mounted across resize → retiring.
// When retiring, all cells get "old-copied" which has a CSS transition
// so they smoothly fade from amber to gray.
function oldCellVariant(
  index: number,
  resize: ResizeState | null,
  copiedSet: Set<number> | null,
): "old" | "old-copied" {
  if (resize && copiedSet) {
    return copiedSet.has(index) ? "old-copied" : "old";
  }
  return "old-copied";
}

export default function CanvasArray({
  length,
  capacity,
  resize,
  retiring,
  pastArrays,
}: CanvasArrayProps) {
  const copiedSet = resize ? new Set(resize.copied) : null;
  const commitRetiring = useEventStore((s) => s.commitRetiring);

  // After retiring fades to match past array visuals, commit it.
  // The 800ms gives time for: color fade (500ms) + settle (300ms)
  useEffect(() => {
    if (!retiring) return;
    const timer = setTimeout(() => commitRetiring(), 800);
    return () => clearTimeout(timer);
  }, [retiring, commitRetiring]);

  const oldVisible = !!(resize || retiring);
  const oldCap = resize ? resize.oldCap : retiring ? retiring.capacity : 0;
  const oldLength = resize ? length : retiring ? retiring.length : 0;
  const isRetiring = !resize && !!retiring;
  const isFull = !resize && !retiring && length === capacity && capacity > 0;
  const resizeNumber = pastArrays.length + 1;

  return (
    <div className="p-6">
      {/* Past arrays — always visible, grayed out */}
      {pastArrays.map((arr, i) => (
        <PastArrayRow key={`past-${i}`} arr={arr} index={i} />
      ))}

      {/* Old array section — ONE set of DOM nodes across resize → retiring.
          Uses clipPath (GPU-composited, no content distortion) instead of scaleY.
          During retiring: opacity fades to 0.4 to match past array appearance,
          then commitRetiring swaps it seamlessly. */}
      <div
        className="mb-4 transition-[clip-path,opacity] ease-out"
        style={{
          clipPath: oldVisible
            ? "inset(0 0 0 0)"
            : "inset(0 0 100% 0)",
          opacity: isRetiring ? 0.4 : oldVisible ? 1 : 0,
          transitionDuration: oldVisible ? "500ms" : "400ms",
          transitionTimingFunction: oldVisible
            ? "cubic-bezier(0, 0, 0.58, 1)"
            : "cubic-bezier(0.42, 0, 1, 1)",
        }}
      >
        {oldVisible && (
          <>
            <div
              className="mb-2 text-sm font-medium transition-colors duration-500 ease-out"
              style={{ color: isRetiring ? "#71717a" : "#fbbf24" }}
            >
              {isRetiring
                ? `resize #${resizeNumber} — capacity ${oldCap}`
                : `old array — capacity ${oldCap}`}
            </div>
            <div className="flex flex-wrap gap-2">
              {Array.from({ length: oldCap }, (_, i) => (
                <Cell
                  key={`old-${i}`}
                  index={i}
                  filled={i < oldLength}
                  variant={oldCellVariant(i, resize, copiedSet)}
                />
              ))}
            </div>
            <div
              className="mt-3 flex items-center gap-2 text-sm font-medium transition-colors duration-500 ease-out"
              style={{ color: isRetiring ? "#52525b" : "#a1a1aa" }}
            >
              {resize ? (
                <>
                  <span>↓</span>
                  <span>
                    copying {resize.copied.length}/
                    {Math.min(length, resize.oldCap)} elements
                  </span>
                  <span>↓</span>
                </>
              ) : (
                <span>done</span>
              )}
            </div>
          </>
        )}
      </div>

      {/* Current / new array — always in the same DOM position */}
      <div>
        <div className="mb-2 text-sm font-medium text-emerald-400">
          {resize
            ? `new array — capacity ${resize.newCap}`
            : `current array — capacity ${capacity}`}
        </div>
        <div className="flex flex-wrap gap-2">
          {Array.from({ length: capacity }, (_, i) => {
            if (resize) {
              const arrived = copiedSet!.has(i);
              return (
                <Cell
                  key={`new-${i}`}
                  index={i}
                  filled={arrived}
                  variant={arrived ? "new-arrived" : "new"}
                />
              );
            }
            return (
              <Cell
                key={i}
                index={i}
                filled={i < length}
                variant="normal"
                isFull={isFull}
              />
            );
          })}
        </div>
      </div>
    </div>
  );
}
