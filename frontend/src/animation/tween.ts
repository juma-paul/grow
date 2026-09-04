// Easing functions: map t ∈ [0,1] → progress ∈ [0,1]
// Pattern: ease-in = power curve, ease-out = flip, ease-in-out = stitch both halves
export const easings = {
  linear: (t: number) => t,

  // Quadratic (n=2) — subtle, general purpose
  easeInQuad: (t: number) => t * t,
  easeOutQuad: (t: number) => t * (2 - t),
  easeInOutQuad: (t: number) =>
    t < 0.5 ? 2 * t * t : -1 + (4 - 2 * t) * t,

  // Cubic (n=3) — more pronounced
  easeInCubic: (t: number) => t * t * t,
  easeOutCubic: (t: number) => 1 + --t * t * t,
  easeInOutCubic: (t: number) =>
    t < 0.5 ? 4 * t * t * t : 1 + (t - 1) * (2 * t - 2) * (2 * t - 2),

  // Quart (n=4) — snappy
  easeInQuart: (t: number) => t * t * t * t,
  easeOutQuart: (t: number) => 1 - --t * t * t * t,
} as const;

export type EasingName = keyof typeof easings;

export function tween(
  from: number,
  to: number,
  duration: number,
  onUpdate: (value: number) => void,
  onComplete?: () => void,
  easing: EasingName = "easeOutQuad",
) {
  const start = performance.now();
  const easeFn = easings[easing];

  function tick(now: number) {
    const t = Math.min((now - start) / duration, 1);
    const value = from + (to - from) * easeFn(t);
    onUpdate(value);

    if (t < 1) {
      requestAnimationFrame(tick);
    } else {
      onComplete?.();
    }
  }

  requestAnimationFrame(tick);
}
