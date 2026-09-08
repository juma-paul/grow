import { create } from "zustand";

// Mirrors the Go Event interface - every event has a type field
interface Event {
  type: string;
  [key: string]: unknown;
}

interface ResizeState {
  oldCap: number;
  newCap: number;
  copied: number[];
}

interface PastArray {
  capacity: number;
  length: number;
}

// Brief holding state: old array fades amber → gray before collapsing
interface RetiringState {
  capacity: number;
  length: number;
}

// One entry per append operation — cost = 1 for normal, 1+N for resize
interface CostEntry {
  op: number;
  cost: number;
  isResize: boolean;
  amortized: number;
}

interface EventStore {
  events: Event[];
  currentIndex: number;
  length: number;
  capacity: number;
  resize: ResizeState | null;
  retiring: RetiringState | null;
  pastArrays: PastArray[];
  costs: CostEntry[];
  pendingResizeCost: number;
  isPlaying: boolean;
  speed: number;
  count: number;
  addEvents: (events: Event[]) => void;
  stepForward: () => void;
  stepBack: () => void;
  commitRetiring: () => void;
  play: () => void;
  pause: () => void;
  setSpeed: (speed: number) => void;
  setCount: (count: number) => void;
  reset: () => void;
}

function applyEvent(
  state: {
    length: number;
    capacity: number;
    resize: ResizeState | null;
    retiring: RetiringState | null;
    pastArrays: PastArray[];
    costs: CostEntry[];
    pendingResizeCost: number;
  },
  event: Event,
) {
  let { length, capacity } = state;
  let resize = state.resize;
  let retiring = state.retiring;
  let pastArrays = state.pastArrays;
  let costs = state.costs;
  let pendingResizeCost = state.pendingResizeCost;

  if (event.type === "append_begin") {
    length = (event.length as number) + 1;
  }

  if (event.type === "resize_begin" || event.type === "shrink_begin") {
    // If a previous resize is still retiring, commit it now
    if (retiring) {
      pastArrays = [...pastArrays, retiring];
      retiring = null;
    }
    capacity = event.new_cap as number;
    resize = {
      oldCap: event.old_cap as number,
      newCap: event.new_cap as number,
      copied: [],
    };
  }

  if (event.type === "copy_element" && resize) {
    resize = { ...resize, copied: [...resize.copied, event.to as number] };
  }

  if (event.type === "resize_end" || event.type === "shrink_end") {
    // Phase 1: move to retiring (amber → gray fade happens here)
    // Phase 2: commitRetiring() collapses and adds to pastArrays
    if (resize) {
      retiring = { capacity: resize.oldCap, length: resize.copied.length };
    }
    resize = null;
    pendingResizeCost = event.cost as number;
  }

  if (event.type === "append_end") {
    const totalCost = (event.cost as number) + pendingResizeCost;
    const prevTotal = costs.reduce((sum, c) => sum + c.cost, 0);
    const opNum = costs.length + 1;
    costs = [
      ...costs,
      {
        op: opNum,
        cost: totalCost,
        isResize: pendingResizeCost > 0,
        amortized: parseFloat(((prevTotal + totalCost) / opNum).toFixed(2)),
      },
    ];
    pendingResizeCost = 0;
  }

  return { length, capacity, resize, retiring, pastArrays, costs, pendingResizeCost };
}

function freshState() {
  return {
    length: 0,
    capacity: 0,
    resize: null as ResizeState | null,
    retiring: null as RetiringState | null,
    pastArrays: [] as PastArray[],
    costs: [] as CostEntry[],
    pendingResizeCost: 0,
  };
}

function replayTo(events: Event[], targetIndex: number) {
  let state = freshState();
  for (let i = 0; i <= targetIndex; i++) {
    state = applyEvent(state, events[i]);
  }
  return state;
}

export const useEventStore = create<EventStore>((set, get) => ({
  events: [],
  currentIndex: -1,
  ...freshState(),
  isPlaying: false,
  speed: 1,
  count: 20,

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

  stepBack: () => {
    const { events, currentIndex } = get();
    if (currentIndex < 0) return;
    const target = currentIndex - 1;
    if (target < 0) {
      set({ currentIndex: -1, ...freshState(), isPlaying: false });
      return;
    }
    set({ currentIndex: target, ...replayTo(events, target) });
  },

  // Called after the amber→gray fade completes (~800ms after resize_end)
  commitRetiring: () => {
    const { retiring, pastArrays } = get();
    if (!retiring) return;
    set({
      retiring: null,
      pastArrays: [...pastArrays, retiring],
    });
  },

  play: () => set({ isPlaying: true }),
  pause: () => set({ isPlaying: false }),
  setSpeed: (speed) => set({ speed }),
  setCount: (count) => set({ count }),

  reset: () =>
    set({
      events: [],
      currentIndex: -1,
      ...freshState(),
      isPlaying: false,
    }),
}));
