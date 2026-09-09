import { useEventStore } from "../store";
import { snippets, CPYTHON_VERSION } from "../cpython-snippets";

const narrations: Record<string, { title: string; lines: [number, number]; note: string }> = {
  "list_resize:growth-formula": {
    title: "Growth formula",
    lines: [27, 33],
    note: "new_allocated = (newsize + newsize/8 + 6) & ~3 — over-allocates by ~12.5% plus padding to a multiple of 4. This gives amortized O(1) appends.",
  },
  "list_resize:shrink-check": {
    title: "Shrink check",
    lines: [14, 18],
    note: "If newsize drops below half the allocated size, realloc shrinks the backing array. This reclaims memory from heavy pop sequences.",
  },
  "list_resize:memcpy": {
    title: "Element copy",
    lines: [44, 50],
    note: "Realloc copies existing elements to the new buffer. This is the O(n) cost you pay on a resize — the reason appends aren't always O(1).",
  },
  "list_resize:overflow": {
    title: "Capacity overflow",
    lines: [14, 18],
    note: "The NoGrowth strategy can't accommodate more elements. In CPython this would be an allocation failure; here the simulator emits an overflow event.",
  },
  "list_insert:ins1": {
    title: "Insert & shift",
    lines: [12, 24],
    note: "ins1 calls list_resize then shifts elements right from the end to make room. The shift loop is O(n) — inserting at index 0 moves everything.",
  },
  "list_append:PyList_Append": {
    title: "Append",
    lines: [7, 14],
    note: "PyList_Append checks capacity and calls list_resize only when the backing array is full. When there's room, it's a single pointer store — O(1).",
  },
  "list_pop:PyList_Pop": {
    title: "Pop & shrink",
    lines: [14, 18],
    note: "Pop decrements the size. If the new size drops below half the allocated capacity, list_resize shrinks the array to reclaim memory.",
  },
  "list_extend:PyList_Extend": {
    title: "Extend with length hint",
    lines: [27, 33],
    note: "Extend pre-sizes the array for all incoming items in one list_resize call — the length-hint optimization. One resize instead of N appends.",
  },
};

export default function ExplanationPanel() {
  const focusedRef = useEventStore((s) => s.focusedRef);

  if (!focusedRef) return null;

  const snippet = snippets[focusedRef];
  const narration = narrations[focusedRef];

  if (!snippet || !narration) return null;

  const lines = snippet.split("\n");

  return (
    <div className="border-t border-zinc-700 bg-zinc-850 max-h-72 overflow-auto">
      <div className="flex items-baseline gap-2 px-4 py-2 border-b border-zinc-800">
        <span className="text-sm font-semibold text-emerald-400">{narration.title}</span>
        <span className="text-xs text-zinc-500">CPython {CPYTHON_VERSION}</span>
      </div>
      <div className="px-4 py-2">
        <p className="text-xs text-zinc-400 leading-relaxed mb-3">{narration.note}</p>
        <pre className="text-xs leading-5 overflow-x-auto">
          {lines.map((line, i) => {
            const lineNum = i + 1;
            const highlighted =
              lineNum >= narration.lines[0] && lineNum <= narration.lines[1];
            return (
              <div
                key={i}
                className={
                  highlighted
                    ? "bg-emerald-900/40 text-emerald-200 -mx-4 px-4"
                    : "text-zinc-500"
                }
              >
                <span className="inline-block w-8 text-right mr-3 select-none text-zinc-600">
                  {lineNum}
                </span>
                {line}
              </div>
            );
          })}
        </pre>
      </div>
    </div>
  );
}
