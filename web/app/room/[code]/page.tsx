"use client";

import { Suspense, useEffect, useMemo, useRef, useState } from "react";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { useGameConnection } from "@/lib/useGameConnection";
import { config } from "@/lib/config";
import { Player } from "@/lib/gameTypes";
import { Avatar } from "@/components/Avatar";
import { CountdownRing } from "@/components/CountdownRing";
import { AnimatedScore } from "@/components/AnimatedScore";
import { FlagImage } from "@/components/FlagImage";
import { Confetti } from "@/components/Confetti";
import { finalTitleFor, randomEveryoneWrongQuip, randomLobbyQuip } from "@/lib/quips";
import { useLocale } from "@/lib/i18n";
import { playCorrect, playTick, playWrong } from "@/lib/sound";

function nameFor(id: string | null, roster: Player[], selfId: string | null, you: string) {
  if (!id) return "?";
  if (id === selfId) return you;
  return roster.find((p) => p.id === id)?.name ?? id.slice(0, 6);
}

function avatarFor(id: string, roster: Player[]) {
  return roster.find((p) => p.id === id)?.avatar;
}

function CoinBadge({ score }: { score: number }) {
  return (
    <span className="flex items-center gap-1.5 rounded-full bg-white/15 px-3 py-1.5 text-sm font-bold text-white backdrop-blur">
      {score} <span className="text-base">🪙</span>
    </span>
  );
}

function RoomScreen() {
  const params = useParams<{ code: string }>();
  const searchParams = useSearchParams();
  const router = useRouter();
  const name = searchParams.get("name") ?? "Oyuncu";
  const avatar = searchParams.get("avatar") ?? "";
  const code = params.code.toUpperCase();
  const { state, submitAnswer, rematch } = useGameConnection(code, name, avatar);
  const { t } = useLocale();
  const [starting, setStarting] = useState(false);
  const [rematching, setRematching] = useState(false);

  const lobbyQuip = useMemo(() => randomLobbyQuip(), []);
  const revealQuip = useMemo(() => {
    if (state.phase !== "reveal") return null;
    const nobodyScored = Object.keys(state.pointsAwarded).length === 0;
    return nobodyScored ? randomEveryoneWrongQuip() : null;
  }, [state.phase, state.pointsAwarded]);

  const lastAnnouncedCorrect = useRef<boolean | null>(null);
  useEffect(() => {
    if (state.myAnswerCorrect === lastAnnouncedCorrect.current) return;
    lastAnnouncedCorrect.current = state.myAnswerCorrect;
    if (state.myAnswerCorrect === true) playCorrect();
    if (state.myAnswerCorrect === false) playWrong();
  }, [state.myAnswerCorrect]);

  async function startGame() {
    setStarting(true);
    try {
      await fetch(`${config.apiBase}/rooms/${code}/start`, { method: "POST" });
    } finally {
      setStarting(false);
    }
  }

  async function handleRematch() {
    setRematching(true);
    try {
      await rematch();
    } finally {
      setRematching(false);
    }
  }

  const sortedScores = Object.entries(state.scores).sort(([, a], [, b]) => b - a);
  const myScore = state.selfId ? (state.scores[state.selfId] ?? 0) : 0;

  return (
    <div className="min-h-screen bg-gradient-to-b from-violet-700 via-purple-800 to-purple-950 px-4 py-6 sm:py-10">
      <div className="mx-auto flex max-w-lg flex-col gap-5">
        <header className="flex items-center justify-between">
          <button
            onClick={() => router.push("/")}
            aria-label={t("backToHome")}
            className="flex h-9 w-9 items-center justify-center rounded-full border border-white/30 text-white transition hover:bg-white/10"
          >
            ✕
          </button>
          <span className="flex items-center gap-1.5 text-sm font-medium text-purple-100">
            <span
              className={`h-1.5 w-1.5 rounded-full ${
                state.connected ? "bg-emerald-400" : "animate-pulse bg-amber-400"
              }`}
            />
            {t("roomCode")} <b className="tracking-widest text-white">{code}</b>
          </span>
          <CoinBadge score={myScore} />
        </header>

        {state.phase === "connecting" && (
          <p className="py-10 text-center text-purple-200">{t("connectingToRoom")}</p>
        )}

        {state.phase === "lobby" && (
          <section
            style={{ animation: "fadeSlideIn 0.35s ease-out" }}
            className="flex flex-col items-center gap-4 rounded-3xl bg-white/95 p-6 shadow-2xl backdrop-blur sm:p-8"
          >
            <p className="text-sm text-purple-600">{t("shareCode")}</p>
            <p className="bg-gradient-to-r from-violet-600 to-purple-600 bg-clip-text text-4xl font-extrabold tracking-[0.3em] text-transparent">
              {code}
            </p>
            <ul className="flex flex-wrap justify-center gap-4 py-2">
              {state.roster.map((p) => (
                <li key={p.id} className="flex flex-col items-center gap-1.5">
                  <Avatar avatar={p.avatar} size={44} />
                  <span className="text-xs font-medium text-purple-700">
                    {p.id === state.selfId ? `${p.name} (${t("you")})` : p.name}
                  </span>
                </li>
              ))}
            </ul>
            <p className="text-center text-xs italic text-purple-500">{lobbyQuip}</p>
            <button
              className="w-full rounded-full bg-gradient-to-r from-violet-600 to-purple-600 px-4 py-3.5 font-bold text-white shadow-lg shadow-purple-600/30 transition hover:brightness-110 active:scale-[0.98] disabled:opacity-50"
              onClick={startGame}
              disabled={starting}
            >
              {t("startGame")}
            </button>
          </section>
        )}

        {state.phase === "question" && state.question && (
          <section
            key={state.question.question_id}
            style={{ animation: "fadeSlideIn 0.35s ease-out" }}
            className="flex flex-col gap-4"
          >
            <div className="flex items-center justify-between">
              <span className="rounded-full bg-white/15 px-3 py-1.5 text-xs font-bold text-white backdrop-blur">
                {t("question")} {state.question.index}/{state.question.total}
              </span>
              <CountdownRing deadline={state.question.deadline} />
            </div>

            {state.question.image && (
              <div className="flex justify-center rounded-2xl bg-white/10 p-6 backdrop-blur">
                <FlagImage image={state.question.image} size={88} />
              </div>
            )}

            <p className="text-center text-xl font-bold text-white">{state.question.text}</p>

            <div className="flex flex-col gap-3">
              {state.question.choices.map((choice, i) => {
                const picked = state.myAnswerChoice === i;
                const answered = state.myAnswerChoice !== null;
                const isRevealedCorrect = state.correctChoice === i;
                let style = "bg-white text-purple-900 hover:bg-purple-50 active:scale-[0.98]";
                if (picked && state.myAnswerCorrect === true) {
                  style = "bg-emerald-500 text-white shadow-lg shadow-emerald-500/30 animate-[pop_0.3s_ease-out]";
                } else if (picked && state.myAnswerCorrect === false) {
                  style = "bg-red-500 text-white shadow-lg shadow-red-500/30 animate-[shake_0.4s_ease-in-out]";
                } else if (picked) {
                  style = "bg-white text-purple-900 shadow-lg";
                } else if (isRevealedCorrect) {
                  // We got it wrong (or the timer ran out) — show the right
                  // answer instead of leaving the player guessing.
                  style = "bg-emerald-500 text-white shadow-lg shadow-emerald-500/30 animate-[pop_0.3s_ease-out]";
                } else if (answered) {
                  style = "bg-white/60 text-purple-700";
                }
                return (
                  <button
                    key={i}
                    onClick={() => {
                      playTick();
                      submitAnswer(i);
                    }}
                    disabled={answered}
                    className={`w-full rounded-full px-6 py-4 text-center text-base font-bold uppercase tracking-wide transition disabled:cursor-default ${style}`}
                  >
                    {choice}
                  </button>
                );
              })}
            </div>

            {state.myAnswerCorrect !== null && (
              <p className="text-center text-sm font-medium text-purple-200">
                {state.myAnswerCorrect ? t("correctFeedback") : t("wrongFeedback")}{" "}
                {t("waitingForResult")}
              </p>
            )}
          </section>
        )}

        {state.phase === "reveal" && (
          <section
            style={{ animation: "fadeSlideIn 0.35s ease-out" }}
            className="flex flex-col gap-4"
          >
            <div className="rounded-3xl bg-white/95 p-5 text-center shadow-2xl backdrop-blur">
              <p className="text-sm text-purple-600">
                {t("correctAnswer")}:{" "}
                <b className="text-emerald-600">
                  {state.question?.choices[state.correctChoice ?? -1]}
                </b>
              </p>
              {revealQuip && <p className="mt-1 text-xs italic text-purple-500">{revealQuip}</p>}
            </div>
            <ScoreTable
              sortedScores={sortedScores}
              pointsAwarded={state.pointsAwarded}
              roster={state.roster}
              selfId={state.selfId}
              you={t("you")}
            />
          </section>
        )}

        {state.phase === "finished" && (
          <section
            style={{ animation: "fadeSlideIn 0.35s ease-out" }}
            className="relative flex flex-col gap-4"
          >
            <Confetti />
            <p className="text-center text-3xl font-extrabold text-white">🏆 {t("gameOver")}</p>
            <ul className="flex flex-col gap-2">
              {sortedScores.map(([id, score], i) => {
                const title = finalTitleFor(
                  id,
                  i + 1,
                  sortedScores.length,
                  score,
                  state.hasFastestPlayer && state.fastestPlayerId === id
                );
                const isFirst = i === 0;
                return (
                  <li
                    key={id}
                    className={`flex items-center justify-between rounded-2xl px-4 py-3 shadow-lg ${
                      isFirst ? "bg-gradient-to-r from-amber-300 to-amber-400" : "bg-white/95"
                    }`}
                  >
                    <span className="flex items-center gap-2.5">
                      <Avatar avatar={avatarFor(id, state.roster)} size={32} />
                      <span className="flex flex-col">
                        <span className={`font-bold ${isFirst ? "text-amber-950" : "text-purple-900"}`}>
                          #{i + 1} {nameFor(id, state.roster, state.selfId, t("you"))}
                        </span>
                        <span className={`text-xs ${isFirst ? "text-amber-900" : "text-purple-600"}`}>
                          {title.emoji} {title.title}
                        </span>
                      </span>
                    </span>
                    <span className={`text-lg font-extrabold ${isFirst ? "text-amber-950" : "text-purple-900"}`}>
                      {score}
                    </span>
                  </li>
                );
              })}
            </ul>

            <div className="mt-2 flex flex-col gap-2 sm:flex-row">
              <button
                className="w-full rounded-full bg-gradient-to-r from-violet-600 to-purple-600 px-4 py-3 font-bold text-white shadow-lg shadow-purple-600/30 transition hover:brightness-110 active:scale-[0.98] disabled:opacity-50"
                onClick={handleRematch}
                disabled={rematching}
              >
                {t("playAgain")}
              </button>
              <button
                className="w-full rounded-full border-2 border-white/40 px-4 py-3 font-bold text-white transition hover:bg-white/10 active:scale-[0.98]"
                onClick={() => router.push("/")}
              >
                {t("backToHome")}
              </button>
            </div>
          </section>
        )}
      </div>
    </div>
  );
}

function ScoreTable({
  sortedScores,
  pointsAwarded,
  roster,
  selfId,
  you,
}: {
  sortedScores: [string, number][];
  pointsAwarded: Record<string, number>;
  roster: Player[];
  selfId: string | null;
  you: string;
}) {
  return (
    <ul className="flex flex-col gap-2">
      {sortedScores.map(([id, score], i) => (
        <li
          key={id}
          className="flex items-center justify-between rounded-2xl bg-white/95 px-4 py-3 shadow-lg backdrop-blur"
        >
          <span className="flex items-center gap-2.5">
            <Avatar avatar={avatarFor(id, roster)} size={30} />
            <span className="font-bold text-purple-900">
              #{i + 1} {nameFor(id, roster, selfId, you)}
            </span>
          </span>
          <AnimatedScore
            previousTotal={score - (pointsAwarded[id] ?? 0)}
            newTotal={score}
            pointsThisRound={pointsAwarded[id]}
          />
        </li>
      ))}
    </ul>
  );
}

export default function RoomPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-screen items-center justify-center bg-gradient-to-b from-violet-700 via-purple-800 to-purple-950">
          <p className="text-center text-purple-200">Yükleniyor...</p>
        </div>
      }
    >
      <RoomScreen />
    </Suspense>
  );
}
