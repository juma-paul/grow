import { useEffect } from "react";
import { useEventStore } from "../store";

export function usePlayback() {
    const isPlaying = useEventStore((state) => state.isPlaying)
    const speed = useEventStore((state) => state.speed)
    const stepForward = useEventStore((state) => state.stepForward)

    useEffect(() => {
        if (!isPlaying) return

        const interval = setInterval(() => {
            stepForward()
        }, 100 / speed)

        return () => clearInterval(interval)
    }, [isPlaying, speed, stepForward])
}