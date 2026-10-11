import { create } from "zustand";

interface Event {
  type: string;
  [key: string]: unknown;
}

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

export interface ElementInfo {
  id: number;
  label: string;
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
  elements: (ElementInfo | null)[];
  nextElementId: number;
  pendingLabel: string;
  pendingInsertIndex: number;
  instantSeek: boolean;
  isPlaying: boolean;
  speed: number;
  count: number;
  strategy: Strategy;
  error: string | null;
  setError: (error: string | null) => void;
  addEvents: (events: Event[]) => void;
  stepForward: () => void;
  stepBack: () => void;
  seekTo: (index: number) => void;
  commitRetiring: () => void;
  play: () => void;
  pause: () => void;
  setSpeed: (speed: number) => void;
  setCount: (count: number) => void;
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
    elements: (ElementInfo | null)[];
    nextElementId: number;
    pendingLabel: string;
    pendingInsertIndex: number;
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
  let elements = state.elements;
  let nextElementId = state.nextElementId;
  let pendingLabel = state.pendingLabel;
  let pendingInsertIndex = state.pendingInsertIndex;

  if (event.type === "init") {
    length = event.length as number;
    capacity = event.capacity as number;
    elements = [];
    for (let i = 0; i < length; i++) {
      elements.push({ id: i, label: String(i) });
    }
    nextElementId = length;
  }

  if (event.type === "append_begin") {
    pendingLabel = String(event.value ?? elements.length);
  }

  // pop_begin: element removal IS the pop
  if (event.type === "pop_begin") {
    const popIdx = event.index as number;
    length = (event.length as number) - 1;
    if (popIdx === length) {
      elements = elements.slice(0, -1);
    } else {
      elements = [...elements];
      elements[popIdx] = null;
    }
  }

  if (event.type === "shift_left") {
    const idx = event.index as number;
    elements = [...elements];
    elements[idx] = elements[idx + 1];
    elements[idx + 1] = null;
  }

  if (event.type === "insert_begin") {
    pendingInsertIndex = event.index as number;
    pendingLabel = String(event.value ?? "");
    elements = [...elements, null];
  }

  if (event.type === "shift_right") {
    const idx = event.index as number;
    elements = [...elements];
    elements[idx] = elements[idx - 1];
    elements[idx - 1] = null;
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
    elements = [...elements, { id: nextElementId, label: pendingLabel }];
    nextElementId++;
    length++;
  }

  if (event.type === "insert_end") {
    elements = [...elements];
    elements[pendingInsertIndex] = { id: nextElementId, label: pendingLabel };
    nextElementId++;
    pendingInsertIndex = -1;
    length++;
  }

  if (event.type === "extend_end") {
    const count = event.cost as number;
    const newElements = [...elements];
    for (let i = 0; i < count; i++) {
      newElements.push({ id: nextElementId + i, label: String(elements.length + i) });
    }
    elements = newElements;
    nextElementId += count;
    length += count;
  }

  if (event.type === "pop_end") {
    while (elements.length > 0 && elements[elements.length - 1] === null) {
      elements = elements.slice(0, -1);
    }
  }

  if (event.type === "append_end" || event.type === "pop_end" || event.type === "insert_end" || event.type === "extend_end") {
    const opCost = (event.cost as number) + pendingResizeCost;
    totalCost += opCost;
    const opNum = costs.length + 1;
    costs = [
      ...costs,
      {
        op: opNum,
        cost: opCost,
        isResize: pendingResizeCost > 0,
        amortized: parseFloat((totalCost / opNum).toFixed(2)),
      },
    ];
    pendingResizeCost = 0;
  }

  if (event.type === "append_batch") {
    const count = event.count as number;
    const toLen = event.to_len as number;
    const newElements = [...elements];
    for (let i = 0; i < count; i++) {
      newElements.push({ id: nextElementId + i, label: String(elements.length + i) });
    }
    elements = newElements;
    nextElementId += count;
    length = toLen;
    for (let i = 0; i < count; i++) {
      totalCost += 1;
      const opNum = costs.length + 1;
      costs = [
        ...costs,
        {
          op: opNum,
          cost: 1,
          isResize: false,
          amortized: parseFloat((totalCost / opNum).toFixed(2)),
        },
      ];
    }
  }

  return {
    length, capacity, resize, retiring, pastArrays, costs,
    pendingResizeCost, totalCost, resizeCount, lastResize,
    elements, nextElementId, pendingLabel, pendingInsertIndex,
  };
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
    elements: [] as (ElementInfo | null)[],
    nextElementId: 0,
    pendingLabel: "",
    pendingInsertIndex: -1,
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
  error: null,
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
  setError: (error) => set({ error }),
  setStrategy: (strategy) => set({ strategy }),

  reset: () =>
    set({
      events: [],
      currentIndex: -1,
      ...freshState(),
      isPlaying: false,
      error: null,
    }),
}));
