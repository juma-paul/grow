import { useCallback, useEffect, useRef } from "react";
import { useEventStore } from "../store";

export function useEventStream() {
  const addEvents = useEventStore((state) => state.addEvents);
  const play = useEventStore((state) => state.play);
  const reset = useEventStore((state) => state.reset);
  const count = useEventStore((state) => state.count);
  const strategy = useEventStore((state) => state.strategy);
  const wsRef = useRef<WebSocket | null>(null);

  const connect = useCallback(() => {
    wsRef.current?.close();
    reset();
    const buffer: { type: string; [key: string]: unknown }[] = [];
    const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
    const ws = new WebSocket(
      `${proto}//${window.location.host}/execute?count=${count}&strategy=${strategy}`,
    );
    wsRef.current = ws;

    ws.onmessage = (msg) => {
      buffer.push(JSON.parse(msg.data));
    };

    ws.onclose = () => {
      if (wsRef.current === ws) {
        addEvents(buffer);
        play();
      }
    };

    ws.onerror = (err) => {
      console.error("WebSocket error:", err);
    };
  }, [addEvents, play, reset, count, strategy]);

  useEffect(() => {
    return () => {
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, []);

  return { connect };
}
