import { useCallback } from "react";
import { useEventStore } from "../store";

export function useEventStream() {
  const push = useEventStore((state) => state.push);
  const reset = useEventStore((state) => state.reset);

  const connect = useCallback(() => {
    reset();
    const ws = new WebSocket("ws://localhost:8080/execute");

    ws.onmessage = (msg) => {
      const event = JSON.parse(msg.data);
      push(event);
    };

    ws.onerror = (err) => {
      console.error("WebSocket error", err);
    };

    ws.onclose = () => {
      console.log("WebSocket closed");
    };

    return ws;
  }, [push, reset]);

  return { connect };
}
