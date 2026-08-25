export function tween(
  from: number,
  to: number,
  duration: number,
  onUpdate: (value: number) => void,
  onComplete?: () => void,
) {
  const start = performance.now();

  function tick(now: number) {
    const elapsed = now - start;
    const t = Math.min(elapsed / duration, 1);

    // Ease-out: deceleratestoward the end (feels natural)
    const eased = 1 - (1 - t) * (1 - t);

    const value = from + (to - from) * eased;
    onUpdate(value);

    if (t < 1) {
      requestAnimationFrame(tick);
    } else {
      onComplete?.();
    }
  }

  requestAnimationFrame(tick);
}
