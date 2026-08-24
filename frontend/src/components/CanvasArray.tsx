interface CanvasArrayProps {
  length: number;
  capacity: number;
}

export default function CanvasArray({ length, capacity }: CanvasArrayProps) {
  return (
    <div className="flex flex-wrap gap-1 p-4">
      {Array.from({ length: capacity }, (_, i) => (
        <div
          key={i}
          className={`flex h-7 w-7 items-center justify-center rounded border text-xs font-mono
              ${
                i < length
                  ? "border-emerald-500 bg-emerald-900 text-emerald-200"
                  : "border-zinc-600 bg-zinc-800 text-zinc-500"
              }`}
        >
          {i < length ? i : ""}
        </div>
      ))}
    </div>
  );
}
