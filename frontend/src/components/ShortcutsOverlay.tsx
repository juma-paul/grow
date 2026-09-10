const SHORTCUTS = [
  { key: "Space", action: "Play / Pause" },
  { key: "→", action: "Step forward" },
  { key: "←", action: "Step back" },
  { key: "?", action: "Toggle this help" },
];

export default function ShortcutsOverlay({
  onClose,
}: {
  onClose: () => void;
}) {
  return (
    <div
      className="absolute inset-0 z-20 flex items-center justify-center bg-black/50"
      onClick={onClose}
    >
      <div
        className="rounded-xl border border-[#262d3d] bg-[#171b24] px-6 py-5 shadow-2xl max-w-xs w-full"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="text-[13px] font-semibold text-[#e2e4ea] mb-4">
          Keyboard shortcuts
        </div>
        <div className="flex flex-col gap-2.5">
          {SHORTCUTS.map((s) => (
            <div key={s.key} className="flex items-center justify-between">
              <span className="text-[12px] text-[#8b90a0]">{s.action}</span>
              <kbd className="rounded bg-[#252b3a] border border-[#262d3d] px-2 py-0.5 font-[var(--font-mono)] text-[11px] text-[#e2e4ea]">
                {s.key}
              </kbd>
            </div>
          ))}
        </div>
        <div className="mt-4 text-center text-[10px] text-[#565b6b]">
          Press <kbd className="rounded bg-[#252b3a] border border-[#262d3d] px-1 py-px font-[var(--font-mono)] text-[10px]">?</kbd> or click outside to close
        </div>
      </div>
    </div>
  );
}
