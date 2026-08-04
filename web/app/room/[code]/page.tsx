"use client";

import { Suspense, useEffect, useMemo, useRef, useState } from "react";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { useGameConnection } from "@/lib/useGameConnection";
import { config } from "@/lib/config";
import { Player } from "@/lib/gameTypes";
import { Avatar } from "@/components/Avatar";
import { CountdownBar } from "@/components/CountdownBar";
import { AnimatedScore } from "@/components/AnimatedScore";
import { FlagImage } from "@/components/FlagImage";
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

  return (
    <div className="relative min-h-screen overflow-hidden bg-gradient-to-b from-indigo-50 via-white to-white px-4 py-6 dark:from-zinc-950 dark:via-black dark:to-black sm:py-10">
      <div
        aria-hidden
        className="pointer-events-none absolute -top-24 -right-24 h-72 w-72 rounded-full bg-fuchsia-300/20 blur-3xl dark:bg-fuchsia-800/10"
      />
      <div
        aria-hidden
        className="pointer-events-none absolute -bottom-24 -left-24 h-72 w-72 rounded-full bg-indigo-300/20 blur-3xl dark:bg-indigo-800/10"
      />

      <div className="relative mx-auto flex max-w-lg flex-col gap-5">
        <header className="flex items-center justify-between rounded-xl bg-white/70 px-4 py-2.5 shadow-sm backdrop-blur dark:bg-zinc-900/70">
          <span className="text-sm text-zinc-500">
            {t("roomCode")}{" "}
            <b className="tracking-widest text-zinc-900 dark:text-zinc-50">{code}</b>
          </span>
          <span
            className={`flex items-center gap-1.5 text-xs font-medium ${
              state.connected ? "text-emerald-600" : "text-amber-600"
            }`}
          >
            <span
              className={`h-1.5 w-1.5 rounded-full ${
                state.connected ? "bg-emerald-500" : "animate-pulse bg-amber-500"
              }`}
            />
            {state.connected ? t("connected") : t("connecting")}
          </span>
        </header>

        {state.phase === "connecting" && (
          <p className="py-10 text-center text-zinc-500">{t("connectingToRoom")}</p>
        )}

        {state.phase === "lobby" && (
          <section className="flex flex-col items-center gap-4 rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-xl shadow-indigo-950/5 dark:border-zinc-800 dark:bg-zinc-900 sm:p-8">
            <p className="text-sm text-zinc-500">{t("shareCode")}</p>
            <p className="bg-gradient-to-br from-indigo-600 to-fuchsia-600 bg-clip-text text-4xl font-extrabold tracking-[0.3em] text-transparent">
              {code}
            </p>
            <ul className="flex flex-wrap justify-center gap-4 py-2">
              {state.roster.map((p) => (
                <li key={p.id} className="flex flex-col items-center gap-1.5">
                  <Avatar avatar={p.avatar} size={44} />
                  <span className="text-xs font-medium text-zinc-600 dark:text-zinc-400">
                    {p.id === state.selfId ? `${p.name} (${t("you")})` : p.name}
                  </span>
                </li>
              ))}
            </ul>
            <p className="text-center text-xs italic text-zinc-400">{lobbyQuip}</p>
            <button
              className="w-full rounded-xl bg-gradient-to-br from-indigo-600 to-fuchsia-600 px-4 py-3.5 font-semibold text-white shadow-lg shadow-indigo-600/25 transition hover:brightness-110 active:scale-[0.98] disabled:opacity-50"
              onClick={startGame}
              disabled={starting}
            >
              {t("startGame")}
            </button>
          </section>
        )}

        {state.phase === "question" && state.question && (
          <section className="flex flex-col gap-4 rounded-2xl border border-zinc-200/80 bg-white p-5 shadow-xl shadow-indigo-950/5 dark:border-zinc-800 dark:bg-zinc-900 sm:p-6">
            <div className="flex items-center justify-between">
              <span className="rounded-full bg-indigo-100 px-2.5 py-1 text-xs font-semibold text-indigo-700 dark:bg-indigo-950 dark:text-indigo-300">
                {t("question")} {state.question.index}/{state.question.total}
              </span>
            </div>
            <CountdownBar deadline={state.question.deadline} />
            <FlagImage image={state.question.image} />
            <p className="text-center text-lg font-semibold">{state.question.text}</p>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
              {state.question.choices.map((choice, i) => {
                const picked = state.myAnswerChoice === i;
                const answered = state.myAnswerChoice !== null;
                let style =
                  "border-zinc-200 hover:border-indigo-300 hover:bg-indigo-50/50 active:scale-[0.98] dark:border-zinc-700 dark:hover:border-indigo-800 dark:hover:bg-indigo-950/30";
                if (picked && state.myAnswerCorrect === true) {
                  style = "border-emerald-600 bg-emerald-600 text-white shadow-md shadow-emerald-600/25";
                } else if (picked && state.myAnswerCorrect === false) {
                  style = "border-red-600 bg-red-600 text-white shadow-md shadow-red-600/25";
                } else if (picked) {
                  style = "border-indigo-600 bg-indigo-600 text-white shadow-md shadow-indigo-600/25";
                } else if (answered) {
                  style = "border-zinc-200 opacity-50 dark:border-zinc-700";
                }
                return (
                  <button
                    key={i}
                    onClick={() => {
                      playTick();
                      submitAnswer(i);
                    }}
                    disabled={answered}
                    className={`rounded-xl border-2 px-4 py-4 text-left text-base font-medium transition disabled:cursor-default ${style}`}
                  >
                    {choice}
                  </button>
                );
              })}
            </div>
            {state.myAnswerCorrect !== null && (
              <p className="text-center text-sm font-medium text-zinc-500">
                {state.myAnswerCorrect ? t("correctFeedback") : t("wrongFeedback")}{" "}
                {t("waitingForResult")}
              </p>
            )}
          </section>
        )}

        {state.phase === "reveal" && (
          <section className="flex flex-col gap-4">
            <div className="rounded-2xl border border-zinc-200/80 bg-white p-5 text-center shadow-xl shadow-indigo-950/5 dark:border-zinc-800 dark:bg-zinc-900">
              <p className="text-sm text-zinc-500">
                {t("correctAnswer")}:{" "}
                <b className="text-emerald-600 dark:text-emerald-400">
                  {state.question?.choices[state.correctChoice ?? -1]}
                </b>
              </p>
              {revealQuip && <p className="mt-1 text-xs italic text-zinc-400">{revealQuip}</p>}
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
          <section className="flex flex-col gap-4">
            <p className="text-center text-3xl font-extrabold">{t("gameOver")}</p>
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
                    className={`flex items-center justify-between rounded-xl border px-4 py-3 shadow-sm ${
                      isFirst
                        ? "border-amber-300 bg-gradient-to-r from-amber-50 to-white dark:border-amber-800 dark:from-amber-950/40 dark:to-zinc-900"
                        : "border-zinc-200/80 bg-white dark:border-zinc-800 dark:bg-zinc-900"
                    }`}
                  >
                    <span className="flex items-center gap-2.5">
                      <Avatar avatar={avatarFor(id, state.roster)} size={32} />
                      <span className="flex flex-col">
                        <span className="font-medium">
                          #{i + 1} {nameFor(id, state.roster, state.selfId, t("you"))}
                        </span>
                        <span className="text-xs text-zinc-500">
                          {title.emoji} {title.title}
                        </span>
                      </span>
                    </span>
                    <span className="text-lg font-bold">{score}</span>
                  </li>
                );
              })}
            </ul>

            <div className="mt-2 flex flex-col gap-2 sm:flex-row">
              <button
                className="w-full rounded-xl bg-gradient-to-br from-indigo-600 to-fuchsia-600 px-4 py-3 font-semibold text-white shadow-lg shadow-indigo-600/25 transition hover:brightness-110 active:scale-[0.98] disabled:opacity-50"
                onClick={handleRematch}
                disabled={rematching}
              >
                {t("playAgain")}
              </button>
              <button
                className="w-full rounded-xl border border-zinc-300 px-4 py-3 font-medium transition hover:bg-zinc-100 active:scale-[0.98] dark:border-zinc-700 dark:hover:bg-zinc-800"
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
          className="flex items-center justify-between rounded-xl border border-zinc-200/80 bg-white px-4 py-3 shadow-sm dark:border-zinc-800 dark:bg-zinc-900"
        >
          <span className="flex items-center gap-2.5">
            <Avatar avatar={avatarFor(id, roster)} size={30} />
            <span className="font-medium">
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
      fallback={<p className="p-10 text-center text-zinc-500">Yükleniyor...</p>}
    >
      <RoomScreen />
    </Suspense>
  );
}
