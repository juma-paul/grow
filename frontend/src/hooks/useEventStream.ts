import { useCallback, useEffect, useRef } from "react";
import { useEventStore } from "../store";

export interface Scenario {
  op: "append" | "insert" | "extend";
  count: number;
  pos?: "start" | "middle" | "end";
  seed?: number;
}

export function useEventStream() {
  const addEvents = useEventStore((state) => state.addEvents);
  const play = useEventStore((state) => state.play);
  const reset = useEventStore((state) => state.reset);
  const setError = useEventStore((state) => state.setError);
  const strategy = useEventStore((state) => state.strategy);
  const wsRef = useRef<WebSocket | null>(null);

  const run = useCallback(
    (scenario: Scenario) => {
      wsRef.current?.close();
      reset();

      const params = new URLSearchParams({
        op: scenario.op,
        count: String(scenario.count),
        strategy,
      });
      if (scenario.pos) params.set("pos", scenario.pos);
      if (scenario.seed) params.set("seed", String(scenario.seed));

      const buffer: { type: string; [key: string]: unknown }[] = [];
      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      const ws = new WebSocket(
        `${proto}//${window.location.host}/execute?${params.toString()}`,
      );
      wsRef.current = ws;

      ws.onmessage = (msg) => {
        buffer.push(JSON.parse(msg.data));
      };

      ws.onclose = (e) => {
        if (wsRef.current === ws) {
          const reason = e.reason;
          if (reason && reason !== "done") {
            setError(reason);
          }
          addEvents(buffer);
          if (buffer.length > 0 && (!reason || reason === "done")) {
            play();
          }
        }
      };

      ws.onerror = (err) => {
        console.error("WebSocket error:", err);
      };
    },
    [addEvents, play, reset, setError, strategy],
  );

  useEffect(() => {
    return () => {
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, []);

  return { run };
}
