import { useMemo } from "react";
import { motion, AnimatePresence, useReducedMotion } from "motion/react";
import { useEventStore } from "../store";
import type { ResizeState, RetiringState, PastArray } from "../store";

interface CanvasArrayProps {
  length: number;
  capacity: number;
  resize: ResizeState | null;
  retiring: RetiringState | null;
  pastArrays: PastArray[];
}

const colors = {
  filledNormal: {
    borderColor: "#10b981",
    backgroundColor: "rgba(6, 78, 59, 1)",
    color: "#a7f3d0",
  },
  emptyNormal: {
    borderColor: "#52525b",
    backgroundColor: "rgba(39, 39, 42, 1)",
    color: "#71717a",
  },
  old: {
    borderColor: "#f59e0b",
    backgroundColor: "rgba(120, 53, 15, 0.3)",
    color: "#fde68a",
  },
  oldCopied: {
    borderColor: "#52525b",
    backgroundColor: "rgba(39, 39, 42, 0.4)",
    color: "#71717a",
  },
  newEmpty: {
    borderColor: "#52525b",
    backgroundColor: "rgba(39, 39, 42, 0.5)",
    color: "#52525b",
  },
  newArrived: {
    borderColor: "#10b981",
    backgroundColor: "rgba(6, 78, 59, 1)",
    color: "#a7f3d0",
  },
  pastFilled: {
    borderColor: "#52525b",
    backgroundColor: "rgba(39, 39, 42, 0.4)",
    color: "#71717a",
  },
  pastEmpty: {
    borderColor: "rgba(63, 63, 70, 0.5)",
    backgroundColor: "rgba(24, 24, 27, 0.3)",
    color: "#3f3f46",
  },
};

const smoothEase = [0.32, 0.72, 0, 1] as const;
const snap = { duration: 0 };

function Cell({
  index,
  filled,
  variant,
  isFull,
  noMotion,
}: {
  index: number;
  filled: boolean;
  variant: "normal" | "old" | "old-copied" | "new" | "new-arrived" | "past";
  isFull?: boolean;
  noMotion: boolean;
}) {
  const showIndex = variant !== "old-copied" && variant !== "new" && filled;

  let target;
  switch (variant) {
    case "normal":
      target = filled ? colors.filledNormal : colors.emptyNormal;
      break;
    case "old":
      target = colors.old;
      break;
    case "old-copied":
      target = colors.oldCopied;
      break;
    case "new":
      target = colors.newEmpty;
      break;
    case "new-arrived":
      target = colors.newArrived;
      break;
    case "past":
      target = filled ? colors.pastFilled : colors.pastEmpty;
      break;
  }

  if (noMotion) {
    return (
      <div
        className="flex h-12 w-12 items-center justify-center rounded-md border-2 text-sm font-mono font-semibold"
        style={target}
      >
        {showIndex ? index : ""}
      </div>
    );
  }

  return (
    <motion.div
      className="flex h-12 w-12 items-center justify-center rounded-md border-2 text-sm font-mono font-semibold"
      animate={{
        ...target,
        scale:
          variant === "normal" && filled && isFull ? [1, 1.04, 1] : 1,
      }}
      transition={{
        scale:
          variant === "normal" && filled && isFull
            ? { duration: 1.5, ease: "easeInOut", repeat: Infinity }
            : undefined,
        delay: variant === "old-copied" ? index * 0.03 : 0,
      }}
    >
      {showIndex ? index : ""}
    </motion.div>
  );
}

function PastArrayRow({
  arr,
  index,
  noMotion,
  layoutId,
}: {
  arr: PastArray;
  index: number;
  noMotion: boolean;
  layoutId: string;
}) {
  if (noMotion) {
    return (
      <div className="mb-4" style={{ opacity: 0.4 }}>
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
              noMotion
            />
          ))}
        </div>
      </div>
    );
  }

  return (
    <motion.div
      className="mb-4"
      layoutId={layoutId}
      initial={false}
      animate={{ opacity: 0.4 }}
      transition={{
        layout: { duration: 0.3, ease: smoothEase },
        opacity: { duration: 0.3, ease: smoothEase },
      }}
    >
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
            noMotion={false}
          />
        ))}
      </div>
    </motion.div>
  );
}

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
  const copiedSet = useMemo(
    () => (resize ? new Set(resize.copied) : null),
    [resize],
  );
  const commitRetiring = useEventStore((s) => s.commitRetiring);
  const instantSeek = useEventStore((s) => s.instantSeek);
  const reducedMotion = !!useReducedMotion();
  const noMotion = reducedMotion || instantSeek;

  const oldVisible = !!(resize || retiring);
  const oldCap = resize ? resize.oldCap : retiring ? retiring.capacity : 0;
  const oldLength = resize ? length : retiring ? retiring.length : 0;
  const isRetiring = !resize && !!retiring;
  const isFull = !resize && !retiring && length === capacity && capacity > 0;
  const resizeNumber = pastArrays.length + 1;

  return (
    <div className="p-6">
      {pastArrays.map((arr, i) => (
        <PastArrayRow
          key={`past-${i}`}
          arr={arr}
          index={i}
          noMotion={noMotion}
          layoutId={`resize-row-${i + 1}`}
        />
      ))}

      <AnimatePresence mode="popLayout">
        {oldVisible && (
          <motion.div
            key="old-array"
            className="mb-4 overflow-hidden"
            layoutId={noMotion ? undefined : `resize-row-${resizeNumber}`}
            initial={noMotion ? false : { height: 0, opacity: 0 }}
            animate={{
              height: "auto",
              opacity: isRetiring ? 0.4 : 1,
            }}
            exit={{ opacity: 0 }}
            transition={
              noMotion
                ? snap
                : {
                    height: { duration: 0.35, ease: smoothEase },
                    opacity: {
                      duration: isRetiring ? 0.5 : 0.25,
                      ease: smoothEase,
                    },
                    layout: { duration: 0.3, ease: smoothEase },
                  }
            }
            onAnimationComplete={() => {
              if (isRetiring) commitRetiring();
            }}
          >
            <motion.div
              className="mb-2 text-sm font-medium"
              animate={{ color: isRetiring ? "#71717a" : "#fbbf24" }}
              transition={noMotion ? snap : { duration: 0.25, ease: smoothEase }}
            >
              {isRetiring
                ? `resize #${resizeNumber} — capacity ${oldCap}`
                : `old array — capacity ${oldCap}`}
            </motion.div>
            <div className="flex flex-wrap gap-2">
              {Array.from({ length: oldCap }, (_, i) => (
                <Cell
                  key={`old-${i}`}
                  index={i}
                  filled={i < oldLength}
                  variant={oldCellVariant(i, resize, copiedSet)}
                  noMotion={noMotion}
                />
              ))}
            </div>
            <motion.div
              className="mt-3 flex items-center gap-2 text-sm font-medium"
              animate={{ color: isRetiring ? "#52525b" : "#a1a1aa" }}
              transition={noMotion ? snap : { duration: 0.25, ease: smoothEase }}
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
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      <motion.div
        layout={!noMotion}
        transition={noMotion ? snap : { layout: { duration: 0.3, ease: smoothEase } }}
      >
        <div className="mb-2 text-sm font-medium text-emerald-400">
          {resize
            ? `new array — capacity ${resize.newCap}`
            : `current array — capacity ${capacity}`}
        </div>
        <div className="flex flex-wrap gap-2">
          {Array.from({ length: capacity }, (_, i) => {
            const arrived = resize ? copiedSet!.has(i) : false;
            const filled = resize ? arrived : i < length;
            const variant = resize
              ? arrived
                ? "new-arrived"
                : "new"
              : "normal";

            return (
              <Cell
                key={`current-${i}`}
                index={i}
                filled={filled}
                variant={variant}
                isFull={isFull}
                noMotion={noMotion}
              />
            );
          })}
        </div>
      </motion.div>
    </div>
  );
}
