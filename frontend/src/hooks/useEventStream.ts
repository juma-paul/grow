import { useCallback } from "react";
import { useEventStore } from "../store";

export function useEventStream() {
  const addEvents = useEventStore((state) => state.addEvents);
  const play = useEventStore((state) => state.play);
  const reset = useEventStore((state) => state.reset);
  const count = useEventStore((state) => state.count);

  const connect = useCallback(() => {
    reset();
    const buffer: { type: string; [key: string]: unknown }[] = [];
    const ws = new WebSocket(`ws://localhost:8080/execute?count=${count}`);

    ws.onmessage = (msg) => {
      buffer.push(JSON.parse(msg.data));
    };

    ws.onclose = () => {
      addEvents(buffer);
      play();
    };

    ws.onerror = (err) => {
      console.error("WebSocket error:", err);
    };

    return ws;
  }, [addEvents, play, reset, count]);

  return { connect };
}
