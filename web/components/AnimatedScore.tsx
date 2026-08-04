"use client";

import { useEffect, useRef, useState } from "react";

const BASE_POINTS = 100;

/**
 * Shows a score that just changed as "100 + 45" for a beat, then counts up
 * from the previous total to the new one. When nothing was awarded this
 * round (pointsThisRound is 0/undefined), it just renders the flat number.
 */
export function AnimatedScore({
  previousTotal,
  newTotal,
  pointsThisRound,
}: {
  previousTotal: number;
  newTotal: number;
  pointsThisRound?: number;
}) {
  const [phase, setPhase] = useState<"breakdown" | "counting" | "done">(
    pointsThisRound ? "breakdown" : "done"
  );
  const [display, setDisplay] = useState(pointsThisRound ? previousTotal : newTotal);
  const raf = useRef<number>(0);

  useEffect(() => {
    // This effect drives a requestAnimationFrame-based animation timeline in
    // response to new props arriving (a fresh reveal) — it's synchronizing
    // with an external animation clock, not just deriving render output, so
    // the setState calls inside are intentional.
    if (!pointsThisRound) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setDisplay(newTotal);
      return;
    }
    setPhase("breakdown");
    setDisplay(previousTotal);

    const breakdownTimer = setTimeout(() => {
      setPhase("counting");
      const start = performance.now();
      const duration = 500;
      function tick(now: number) {
        const progress = Math.min((now - start) / duration, 1);
        setDisplay(Math.round(previousTotal + (newTotal - previousTotal) * progress));
        if (progress < 1) {
          raf.current = requestAnimationFrame(tick);
        } else {
          setPhase("done");
        }
      }
      raf.current = requestAnimationFrame(tick);
    }, 900);

    return () => {
      clearTimeout(breakdownTimer);
      cancelAnimationFrame(raf.current);
    };
  }, [previousTotal, newTotal, pointsThisRound]);

  if (phase === "breakdown" && pointsThisRound) {
    const bonus = pointsThisRound - BASE_POINTS;
    return (
      <span className="font-bold tabular-nums">
        {display}{" "}
        <span className="text-sm font-medium text-emerald-600 dark:text-emerald-400">
          ({BASE_POINTS}
          {bonus > 0 ? ` + ${bonus}` : ""})
        </span>
      </span>
    );
  }

  return <span className="font-bold tabular-nums">{display}</span>;
}
