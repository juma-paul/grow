import { useEffect } from "react";
import { useEventStore } from "../store";

// Compute copy delay using ease-in-out curve:
// first and last copies are slower (anticipation + follow-through),
// middle copies are faster (momentum).
function copyDelay(index: number, total: number, baseDelay: number): number {
  if (total <= 1) return baseDelay;
  // Normalize position to [0, 1]
  const t = index / (total - 1);
  // Ease-in-out quadratic: slow → fast → slow
  const curve = t < 0.5 ? 2 * t * t : -1 + (4 - 2 * t) * t;
  // Map curve to delay: edges are 1.4x base, middle is 0.6x base
  const factor = 0.6 + 0.8 * (1 - Math.abs(curve - 0.5) * 2);
  return baseDelay * factor;
}

export function usePlayback() {
  const isPlaying = useEventStore((state) => state.isPlaying);
  const speed = useEventStore((state) => state.speed);
  const events = useEventStore((state) => state.events);
  const currentIndex = useEventStore((state) => state.currentIndex);
  const stepForward = useEventStore((state) => state.stepForward);

  useEffect(() => {
    if (!isPlaying) return;

    const nextIndex = currentIndex + 1;
    const nextEvent = events[nextIndex];
    let delay: number;

    switch (nextEvent?.type) {
      case "resize_begin":
      case "shrink_begin":
        delay = 1000 / speed;
        break;
      case "copy_element": {
        // Count total copies in this resize sequence to compute position
        let totalCopies = 0;
        let copyPosition = 0;
        for (let i = nextIndex; i < events.length; i++) {
          if (events[i].type === "copy_element") totalCopies++;
          else if (
            events[i].type === "resize_end" ||
            events[i].type === "shrink_end"
          )
            break;
        }
        for (let i = nextIndex - 1; i >= 0; i--) {
          if (events[i].type === "copy_element") copyPosition++;
          else break;
        }
        delay = copyDelay(copyPosition, copyPosition + totalCopies, 450) / speed;
        break;
      }
      case "resize_end":
      case "shrink_end":
        delay = 600 / speed;
        break;
      default:
        delay = 400 / speed;
    }

    const timeout = setTimeout(() => {
      stepForward();
    }, delay);

    return () => clearTimeout(timeout);
  }, [isPlaying, speed, currentIndex, events, stepForward]);
}
