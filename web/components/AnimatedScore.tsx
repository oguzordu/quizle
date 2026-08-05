"use client";

import { useEffect, useRef, useState } from "react";
import { useLocale } from "@/lib/i18n";

const BASE_POINTS = 100;

/**
 * Shows a score that just changed as two chips — "+100" then, a beat later,
 * a labeled "+45 ⏱️ Hız Bonusu" chip — before counting up from the previous
 * total to the new one. The second chip spells out that the extra points
 * came from answering fast, rather than relying on a bare "+45" that reads
 * as an unexplained bonus. When nothing was awarded this round
 * (pointsThisRound is 0/undefined), it just renders the flat number.
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
  const [step, setStep] = useState<"base" | "bonus" | "counting" | "done">(
    pointsThisRound ? "base" : "done"
  );
  const [display, setDisplay] = useState(pointsThisRound ? previousTotal : newTotal);
  const raf = useRef<number>(0);
  const { t } = useLocale();

  useEffect(() => {
    // This effect drives a staged animation timeline (chip -> chip -> count
    // up) in response to a fresh reveal arriving as props — it's
    // synchronizing with an external animation clock, not deriving render
    // output, so the setState calls inside are intentional.
    if (!pointsThisRound) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setDisplay(newTotal);
      return;
    }
    setStep("base");
    setDisplay(previousTotal);

    const bonusTimer = setTimeout(() => setStep("bonus"), 500);
    const countTimer = setTimeout(() => {
      setStep("counting");
      const start = performance.now();
      const duration = 500;
      function tick(now: number) {
        const progress = Math.min((now - start) / duration, 1);
        setDisplay(Math.round(previousTotal + (newTotal - previousTotal) * progress));
        if (progress < 1) {
          raf.current = requestAnimationFrame(tick);
        } else {
          setStep("done");
        }
      }
      raf.current = requestAnimationFrame(tick);
    }, 1100);

    return () => {
      clearTimeout(bonusTimer);
      clearTimeout(countTimer);
      cancelAnimationFrame(raf.current);
    };
  }, [previousTotal, newTotal, pointsThisRound]);

  const bonus = pointsThisRound ? pointsThisRound - BASE_POINTS : 0;

  if ((step === "base" || step === "bonus") && pointsThisRound) {
    return (
      <span className="flex items-center gap-1.5">
        <span className="animate-[pop_0.25s_ease-out] rounded-md bg-emerald-100 px-1.5 py-0.5 text-sm font-bold text-emerald-700 dark:bg-emerald-950 dark:text-emerald-400">
          +{BASE_POINTS}
        </span>
        {step === "bonus" && bonus > 0 && (
          <span className="flex animate-[pop_0.25s_ease-out] items-center gap-1 rounded-md bg-amber-100 px-1.5 py-0.5 text-sm font-bold text-amber-700 dark:bg-amber-950 dark:text-amber-400">
            <span>+{bonus}</span>
            <span className="text-xs font-semibold">⏱️ {t("speedBonus")}</span>
          </span>
        )}
      </span>
    );
  }

  return <span className="font-bold tabular-nums text-purple-900">{display}</span>;
}
