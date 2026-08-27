import { create } from "zustand";

// Mirrors the Go Event interface - every event has a type field
interface Event {
  type: string;
  [key: string]: unknown;
}

interface EventStore {
  events: Event[];
  currentIndex: number;
  length: number;
  capacity: number;
  isPlaying: boolean;
  speed: number;
  addEvents: (events: Event[]) => void;
  stepForward: () => void;
  play: () => void;
  pause: () => void;
  reset: () => void;
}

function applyEvent(state: { length: number; capacity: number }, event: Event) {
  let { length, capacity } = state;
  if (event.type === "append_begin") {
    length = (event.length as number) + 1;
  }
  if (event.type === "resize_begin" || event.type === "shrink_begin") {
    capacity = event.new_cap as number;
  }
  return { length, capacity };
}

export const useEventStore = create<EventStore>((set, get) => ({
  events: [],
  currentIndex: -1,
  length: 0,
  capacity: 0,
  isPlaying: false,
  speed: 1,

  addEvents: (events) => set({ events }),

  stepForward: () => {
    const { events, currentIndex } = get();
    const next = currentIndex + 1;
    if (next >= events.length) {
      set({ isPlaying: false });
      return;
    }
    const updated = applyEvent(get(), events[next]);
    set({ currentIndex: next, ...updated });
  },

  play: () => set({ isPlaying: true }),
  pause: () => set({ isPlaying: false }),

  reset: () =>
    set({
      events: [],
      currentIndex: -1,
      length: 0,
      capacity: 0,
      isPlaying: false,
    }),
}));
