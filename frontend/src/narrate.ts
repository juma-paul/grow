interface Event {
  type: string;
  [key: string]: unknown;
}

export function narrate(event: Event): string | null {
  switch (event.type) {
    case "append_begin": {
      const len = event.length as number;
      const cap = event.capacity as number;
      const val = event.value;
      if (len >= cap) {
        return `Appending ${val} at index ${len}. Array is full (length ${len} = capacity ${cap}) — resize needed.`;
      }
      return `Appending ${val} at index ${len}. Capacity ${cap} has room — no resize.`;
    }

    case "append_end":
      return null;

    case "resize_begin": {
      const oldCap = event.old_cap as number;
      const newCap = event.new_cap as number;
      const newsize = oldCap + 1;
      const term1 = newsize;
      const term2 = newsize >> 3;
      const raw = term1 + term2 + 6;
      const aligned = raw & ~3;
      return `Array is full at capacity ${oldCap}. CPython allocates capacity ${newCap}: (${newsize} + ${term2} + 6) & ~3 = ${aligned}.`;
    }

    case "resize_end": {
      const cost = event.cost as number;
      return `Resize complete. Copied ${cost} element${cost !== 1 ? "s" : ""} to the new backing array.`;
    }

    case "copy_element": {
      const from = event.from as number;
      const val = event.value;
      return `Copying element [${from}] = ${val} to the new array.`;
    }

    case "shrink_begin": {
      const oldCap = event.old_cap as number;
      const newCap = event.new_cap as number;
      return `Length dropped below half of capacity ${oldCap}. Shrinking to ${newCap}.`;
    }

    case "shrink_end": {
      const cost = event.cost as number;
      return `Shrink complete. Copied ${cost} element${cost !== 1 ? "s" : ""} to the smaller array.`;
    }

    case "pop_begin": {
      const len = event.length as number;
      const cap = event.capacity as number;
      const newLen = len - 1;
      if (newLen < cap >> 1) {
        return `Popping element at index ${newLen}. Length ${newLen} < ${cap}/2 — shrink triggered.`;
      }
      return `Popping element at index ${newLen}. No shrink needed.`;
    }

    case "pop_end":
      return null;

    case "insert_begin": {
      const idx = event.index as number;
      const val = event.value;
      const len = event.length as number;
      const shifts = len - idx;
      return `Inserting ${val} at index ${idx}. Must shift ${shifts} element${shifts !== 1 ? "s" : ""} right.`;
    }

    case "insert_end": {
      const cost = event.cost as number;
      return `Insert complete. Cost: ${cost} (shifts + placement).`;
    }

    case "shift_right": {
      const idx = event.index as number;
      return `Shifting element at index ${idx - 1} → ${idx}.`;
    }

    case "extend_begin": {
      const items = event.items as number;
      const len = event.length as number;
      const cap = event.capacity as number;
      const needed = len + items;
      if (needed > cap) {
        return `Extending by ${items} items. Need capacity ${needed}, have ${cap} — one resize up front.`;
      }
      return `Extending by ${items} items. Capacity ${cap} is sufficient — no resize.`;
    }

    case "extend_end": {
      const cost = event.cost as number;
      return `Extend complete. Placed ${cost} element${cost !== 1 ? "s" : ""}.`;
    }

    case "overflow": {
      const needed = event.needed as number;
      const cap = event.capacity as number;
      return `Overflow: need capacity for ${needed} elements but fixed at ${cap}. No-growth strategy cannot resize.`;
    }

    case "limit_exceeded": {
      const reason = event.reason as string;
      return `Execution limit: ${reason}.`;
    }

    default:
      return null;
  }
}
