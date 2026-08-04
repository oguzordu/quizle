const COLORS = ["#6366f1", "#d946ef", "#f59e0b", "#10b981", "#ef4444", "#3b82f6"];

// Generated once at module load (not during render) so the "random"-looking
// values stay stable across re-renders without upsetting React's purity
// rules — every mount of <Confetti /> reuses this same precomputed burst,
// which is indistinguishable from fresh randomness for a one-off celebration.
const PIECES = Array.from({ length: 40 }, (_, i) => ({
  left: Math.random() * 100,
  delay: Math.random() * 0.6,
  duration: 1.8 + Math.random() * 1.2,
  color: COLORS[i % COLORS.length],
  size: 6 + Math.random() * 6,
}));

/** A short burst of falling confetti — pure CSS, no canvas or library. */
export function Confetti() {
  return (
    <div aria-hidden className="pointer-events-none absolute inset-x-0 top-0 h-0 overflow-visible">
      {PIECES.map((p, i) => (
        <span
          key={i}
          style={{
            position: "absolute",
            left: `${p.left}%`,
            top: 0,
            width: p.size,
            height: p.size * 0.4,
            backgroundColor: p.color,
            animation: `confettiFall ${p.duration}s ease-in ${p.delay}s forwards`,
            borderRadius: 2,
          }}
        />
      ))}
    </div>
  );
}
