import { useEventStore } from "../store";
import { snippets, CPYTHON_VERSION } from "../cpython-snippets";
import { narrate } from "../narrate";

const highlights: Record<string, { title: string; lines: [number, number] }> = {
  "list_resize:growth-formula": { title: "Growth formula", lines: [27, 33] },
  "list_resize:shrink-check": { title: "Shrink check", lines: [14, 18] },
  "list_resize:memcpy": { title: "Element copy", lines: [44, 50] },
  "list_resize:overflow": { title: "Capacity overflow", lines: [14, 18] },
  "list_insert:ins1": { title: "Insert & shift", lines: [12, 24] },
  "list_append:PyList_Append": { title: "Append", lines: [7, 14] },
  "list_pop:PyList_Pop": { title: "Pop & shrink", lines: [14, 18] },
  "list_extend:PyList_Extend": { title: "Extend with length hint", lines: [27, 33] },
};

export default function ExplanationPanel() {
  const focusedRef = useEventStore((s) => s.focusedRef);
  const currentIndex = useEventStore((s) => s.currentIndex);
  const events = useEventStore((s) => s.events);

  if (!focusedRef) return null;

  const snippet = snippets[focusedRef];
  const highlight = highlights[focusedRef];

  if (!snippet || !highlight) return null;

  const currentEvent = currentIndex >= 0 ? events[currentIndex] : null;
  const sentence = currentEvent ? narrate(currentEvent) : null;

  const lines = snippet.split("\n");

  return (
    <div className="border-t border-zinc-700 bg-zinc-850 max-h-72 overflow-auto">
      <div className="flex items-baseline gap-2 px-4 py-2 border-b border-zinc-800">
        <span className="text-sm font-semibold text-emerald-400">{highlight.title}</span>
        <span className="text-xs text-zinc-500">CPython {CPYTHON_VERSION}</span>
      </div>
      <div className="px-4 py-2">
        {sentence && (
          <p className="text-xs text-zinc-300 leading-relaxed mb-3">{sentence}</p>
        )}
        <pre className="text-xs leading-5 overflow-x-auto">
          {lines.map((line, i) => {
            const lineNum = i + 1;
            const highlighted =
              lineNum >= highlight.lines[0] && lineNum <= highlight.lines[1];
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
