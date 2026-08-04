"use client";

import { useEffect, useState } from "react";

/** Live countdown bar that fills down from 100% to 0% as `deadline` approaches. */
export function CountdownBar({ deadline }: { deadline: string }) {
  const [fraction, setFraction] = useState(1);

  useEffect(() => {
    const deadlineMs = new Date(deadline).getTime();
    const startMs = Date.now();
    const totalMs = Math.max(deadlineMs - startMs, 1);

    let frame: number;
    function tick() {
      const remaining = deadlineMs - Date.now();
      setFraction(Math.max(remaining / totalMs, 0));
      if (remaining > 0) frame = requestAnimationFrame(tick);
    }
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [deadline]);

  const color = fraction > 0.5 ? "bg-indigo-500" : fraction > 0.2 ? "bg-amber-500" : "bg-red-500";

  return (
    <div className="h-2 w-full overflow-hidden rounded-full bg-zinc-200 dark:bg-zinc-800">
      <div
        className={`h-full rounded-full transition-[width] duration-100 ease-linear ${color}`}
        style={{ width: `${fraction * 100}%` }}
      />
    </div>
  );
}
