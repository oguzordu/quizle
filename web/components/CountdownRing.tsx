"use client";

import { useEffect, useState } from "react";

/** Circular countdown showing whole seconds remaining, filling out as a ring. */
export function CountdownRing({ deadline }: { deadline: string }) {
  const [fraction, setFraction] = useState(1);
  const [secondsLeft, setSecondsLeft] = useState(0);

  useEffect(() => {
    const deadlineMs = new Date(deadline).getTime();
    const startMs = Date.now();
    const totalMs = Math.max(deadlineMs - startMs, 1);

    let frame: number;
    function tick() {
      const remaining = deadlineMs - Date.now();
      setFraction(Math.max(remaining / totalMs, 0));
      setSecondsLeft(Math.max(Math.ceil(remaining / 1000), 0));
      if (remaining > 0) frame = requestAnimationFrame(tick);
    }
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [deadline]);

  const radius = 20;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference * (1 - fraction);
  const color = fraction > 0.5 ? "#fbbf24" : fraction > 0.2 ? "#f97316" : "#f87171";

  return (
    <div className="relative flex h-14 w-14 items-center justify-center">
      <svg width="56" height="56" viewBox="0 0 48 48" className="-rotate-90">
        <circle cx="24" cy="24" r={radius} fill="none" stroke="rgba(255,255,255,0.25)" strokeWidth="4" />
        <circle
          cx="24"
          cy="24"
          r={radius}
          fill="none"
          stroke={color}
          strokeWidth="4"
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          style={{ transition: "stroke-dashoffset 0.2s linear" }}
        />
      </svg>
      <span className="absolute text-sm font-bold text-white">{secondsLeft}</span>
    </div>
  );
}
