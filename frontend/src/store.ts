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
  push: (event: Event) => void;
  reset: () => void;
}

export const useEventStore = create<EventStore>((set) => ({
  events: [],
  currentIndex: 0,
  length: 0,
  capacity: 0,

  push: (event) =>
    set((state) => {
      const events = [...state.events, event];
      let { length, capacity } = state;

      if (event.type === "append_begin") {
        length = (event.length as number) + 1;
      }
      if (event.type === "resize_begin") {
        capacity = event.new_cap as number;
      }

      return { events, currentIndex: events.length - 1, length, capacity };
    }),

  reset: () => set({ events: [], currentIndex: 0, length: 0, capacity: 0 }),
}));
