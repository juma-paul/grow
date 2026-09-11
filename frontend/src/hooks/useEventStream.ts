import { useCallback, useEffect, useRef } from "react";
import { useEventStore } from "../store";

export function useEventStream() {
  const addEvents = useEventStore((state) => state.addEvents);
  const play = useEventStore((state) => state.play);
  const reset = useEventStore((state) => state.reset);
  const setError = useEventStore((state) => state.setError);
  const count = useEventStore((state) => state.count);
  const strategy = useEventStore((state) => state.strategy);
  const wsRef = useRef<WebSocket | null>(null);

  const connectPreset = useCallback(
    (overrideCount?: number) => {
      wsRef.current?.close();
      reset();
      const n = overrideCount ?? count;
      const buffer: { type: string; [key: string]: unknown }[] = [];
      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      const ws = new WebSocket(
        `${proto}//${window.location.host}/execute?count=${n}&strategy=${strategy}`,
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
    [addEvents, play, reset, setError, count, strategy],
  );

  const backend = useEventStore((state) => state.backend);

  const connectCode = useCallback(
    (source: string, rewrite: boolean) => {
      wsRef.current?.close();
      reset();
      const buffer: { type: string; [key: string]: unknown }[] = [];
      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      let endpoint: string;
      if (backend === "cpython") {
        endpoint = "observe";
      } else {
        endpoint = rewrite ? "auto" : "auto?rewrite=false";
      }
      const ws = new WebSocket(
        `${proto}//${window.location.host}/${endpoint}`,
      );
      wsRef.current = ws;

      ws.onopen = () => {
        ws.send(source);
      };

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
    [addEvents, play, reset, setError, backend],
  );

  useEffect(() => {
    return () => {
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, []);

  return { connect: connectPreset, connectCode };
}
