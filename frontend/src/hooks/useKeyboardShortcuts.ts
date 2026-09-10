import { useEffect, useState } from "react";
import { useEventStore } from "../store";

export function useKeyboardShortcuts() {
  const [showHelp, setShowHelp] = useState(false);

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      if ((e.target as HTMLElement)?.closest?.(".monaco-editor")) return;

      const {
        isPlaying,
        play,
        pause,
        stepForward,
        stepBack,
        events,
      } = useEventStore.getState();

      switch (e.key) {
        case " ":
          e.preventDefault();
          if (events.length === 0) return;
          isPlaying ? pause() : play();
          break;
        case "ArrowRight":
          e.preventDefault();
          stepForward();
          break;
        case "ArrowLeft":
          e.preventDefault();
          stepBack();
          break;
        case "?":
          setShowHelp((v) => !v);
          break;
      }
    };

    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  return { showHelp, setShowHelp };
}
