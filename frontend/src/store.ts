import { create } from "zustand";

interface Event {
  type: string;
  [key: string]: unknown;
}

export type Mode = "presets" | "wrapper" | "auto";
export type Strategy = "cpython" | "doubling" | "1.5x" | "nogrowth";

export interface ResizeState {
  oldCap: number;
  newCap: number;
  copied: number[];
}

export interface PastArray {
  capacity: number;
  length: number;
}

export interface RetiringState {
  capacity: number;
  length: number;
}

interface CostEntry {
  op: number;
  cost: number;
  isResize: boolean;
  amortized: number;
}

export interface LastResize {
  oldCap: number;
  newCap: number;
  needed: number;
  appendNum: number;
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
  totalCost: number;
  resizeCount: number;
  lastResize: LastResize | null;
  instantSeek: boolean;
  isPlaying: boolean;
  speed: number;
  count: number;
  mode: Mode;
  strategy: Strategy;
  addEvents: (events: Event[]) => void;
  stepForward: () => void;
  stepBack: () => void;
  seekTo: (index: number) => void;
  commitRetiring: () => void;
  play: () => void;
  pause: () => void;
  setSpeed: (speed: number) => void;
  setCount: (count: number) => void;
  setMode: (mode: Mode) => void;
  setStrategy: (strategy: Strategy) => void;
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
    totalCost: number;
    resizeCount: number;
    lastResize: LastResize | null;
  },
  event: Event,
) {
  let { length, capacity } = state;
  let resize = state.resize;
  let retiring = state.retiring;
  let pastArrays = state.pastArrays;
  let costs = state.costs;
  let pendingResizeCost = state.pendingResizeCost;
  let totalCost = state.totalCost;
  let resizeCount = state.resizeCount;
  let lastResize = state.lastResize;

  if (event.type === "append_begin") {
    length = (event.length as number) + 1;
  }

  if (event.type === "resize_begin" || event.type === "shrink_begin") {
    if (retiring) {
      pastArrays = [...pastArrays, retiring];
      retiring = null;
    }
    const oldCap = event.old_cap as number;
    const newCap = event.new_cap as number;
    capacity = newCap;
    resize = { oldCap, newCap, copied: [] };
    resizeCount++;
    lastResize = {
      oldCap,
      newCap,
      needed: oldCap + 1,
      appendNum: costs.length + 1,
    };
  }

  if (event.type === "copy_element" && resize) {
    resize = { ...resize, copied: [...resize.copied, event.to as number] };
  }

  if (event.type === "resize_end" || event.type === "shrink_end") {
    if (resize) {
      retiring = { capacity: resize.oldCap, length: resize.copied.length };
    }
    resize = null;
    pendingResizeCost = event.cost as number;
  }

  if (event.type === "append_end") {
    const appendCost = (event.cost as number) + pendingResizeCost;
    totalCost += appendCost;
    const opNum = costs.length + 1;
    costs = [
      ...costs,
      {
        op: opNum,
        cost: appendCost,
        isResize: pendingResizeCost > 0,
        amortized: parseFloat((totalCost / opNum).toFixed(2)),
      },
    ];
    pendingResizeCost = 0;
  }

  return { length, capacity, resize, retiring, pastArrays, costs, pendingResizeCost, totalCost, resizeCount, lastResize };
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
    totalCost: 0,
    resizeCount: 0,
    lastResize: null as LastResize | null,
  };
}

function replayTo(events: Event[], targetIndex: number) {
  let state = freshState();
  for (let i = 0; i <= targetIndex; i++) {
    state = applyEvent(state, events[i]);
  }
  if (state.retiring) {
    state = {
      ...state,
      pastArrays: [...state.pastArrays, state.retiring],
      retiring: null,
    };
  }
  return state;
}

export const useEventStore = create<EventStore>((set, get) => ({
  events: [],
  currentIndex: -1,
  ...freshState(),
  instantSeek: false,
  isPlaying: false,
  speed: 1,
  count: 20,
  mode: "presets",
  strategy: "cpython",

  addEvents: (events) => set({ events }),

  stepForward: () => {
    const { events, currentIndex } = get();
    const next = currentIndex + 1;
    if (next >= events.length) {
      set({ isPlaying: false });
      return;
    }
    const updated = applyEvent(get(), events[next]);
    set({ currentIndex: next, instantSeek: false, ...updated });
  },

  stepBack: () => {
    const { events, currentIndex } = get();
    if (currentIndex < 0) return;
    const target = currentIndex - 1;
    if (target < 0) {
      set({ currentIndex: -1, ...freshState(), instantSeek: true, isPlaying: false });
      return;
    }
    set({ currentIndex: target, instantSeek: true, ...replayTo(events, target) });
  },

  seekTo: (index: number) => {
    const { events } = get();
    if (index < 0) {
      set({ currentIndex: -1, ...freshState(), instantSeek: true, isPlaying: false });
      return;
    }
    const clamped = Math.min(index, events.length - 1);
    set({ currentIndex: clamped, instantSeek: true, ...replayTo(events, clamped) });
  },

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
  setMode: (mode) => set({ mode }),
  setStrategy: (strategy) => set({ strategy }),

  reset: () =>
    set({
      events: [],
      currentIndex: -1,
      ...freshState(),
      isPlaying: false,
    }),
}));
