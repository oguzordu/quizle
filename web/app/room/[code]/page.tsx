"use client";

import { Suspense, useState } from "react";
import { useParams, useSearchParams } from "next/navigation";
import { useGameConnection } from "@/lib/useGameConnection";
import { config } from "@/lib/config";

function nameFor(id: string | null, roster: { id: string; name: string }[], selfId: string | null) {
  if (!id) return "?";
  if (id === selfId) return "Sen";
  return roster.find((p) => p.id === id)?.name ?? id.slice(0, 6);
}

function RoomScreen() {
  const params = useParams<{ code: string }>();
  const searchParams = useSearchParams();
  const name = searchParams.get("name") ?? "Oyuncu";
  const code = params.code.toUpperCase();
  const { state, submitAnswer } = useGameConnection(code, name);
  const [starting, setStarting] = useState(false);

  async function startGame() {
    setStarting(true);
    try {
      await fetch(`${config.apiBase}/rooms/${code}/start`, { method: "POST" });
    } finally {
      setStarting(false);
    }
  }

  const sortedScores = Object.entries(state.scores).sort(([, a], [, b]) => b - a);

  return (
    <div className="mx-auto flex min-h-screen max-w-lg flex-col gap-6 px-4 py-10">
      <header className="flex items-center justify-between">
        <span className="text-sm text-zinc-500">
          Oda kodu <b className="tracking-widest text-zinc-900 dark:text-zinc-50">{code}</b>
        </span>
        <span className={`text-xs ${state.connected ? "text-green-600" : "text-amber-600"}`}>
          {state.connected ? "● bağlı" : "○ bağlanıyor..."}
        </span>
      </header>

      {state.phase === "connecting" && (
        <p className="text-center text-zinc-500">Odaya bağlanılıyor...</p>
      )}

      {state.phase === "lobby" && (
        <section className="flex flex-col items-center gap-4 rounded-xl border border-zinc-200 p-6 dark:border-zinc-800">
          <p className="text-sm text-zinc-500">
            Bu kodu paylaş, herkes katılınca oyunu başlat:
          </p>
          <p className="text-3xl font-bold tracking-[0.3em]">{code}</p>
          <ul className="flex flex-wrap justify-center gap-2">
            {state.roster.map((p) => (
              <li
                key={p.id}
                className="rounded-full bg-zinc-100 px-3 py-1 text-sm dark:bg-zinc-800"
              >
                {p.id === state.selfId ? `${p.name} (sen)` : p.name}
              </li>
            ))}
          </ul>
          <button
            className="w-full rounded-lg bg-zinc-900 px-4 py-3 font-medium text-white transition hover:bg-zinc-700 disabled:opacity-50 dark:bg-zinc-50 dark:text-zinc-900"
            onClick={startGame}
            disabled={starting}
          >
            Oyunu Başlat
          </button>
        </section>
      )}

      {state.phase === "question" && state.question && (
        <section className="flex flex-col gap-4">
          <p className="text-lg font-medium">{state.question.question_id}</p>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {state.question.choices.map((choice, i) => {
              const picked = state.myAnswerChoice === i;
              return (
                <button
                  key={i}
                  onClick={() => submitAnswer(i)}
                  disabled={state.myAnswerChoice !== null}
                  className={`rounded-lg border px-4 py-4 text-left text-base transition disabled:opacity-60 ${
                    picked
                      ? "border-zinc-900 bg-zinc-900 text-white dark:border-zinc-50 dark:bg-zinc-50 dark:text-zinc-900"
                      : "border-zinc-300 hover:bg-zinc-100 dark:border-zinc-700 dark:hover:bg-zinc-900"
                  }`}
                >
                  {choice}
                </button>
              );
            })}
          </div>
          {state.myAnswerCorrect !== null && (
            <p className="text-center text-sm text-zinc-500">
              {state.myAnswerCorrect ? "✅ Doğru! Sonuç bekleniyor..." : "❌ Yanlış. Sonuç bekleniyor..."}
            </p>
          )}
        </section>
      )}

      {state.phase === "reveal" && (
        <section className="flex flex-col gap-4">
          <p className="text-center text-sm text-zinc-500">
            Doğru cevap: <b>{state.question?.choices[state.correctChoice ?? -1]}</b>
          </p>
          <ScoreTable
            sortedScores={sortedScores}
            pointsAwarded={state.pointsAwarded}
            roster={state.roster}
            selfId={state.selfId}
          />
        </section>
      )}

      {state.phase === "finished" && (
        <section className="flex flex-col gap-4">
          <p className="text-center text-2xl font-semibold">🏁 Oyun bitti!</p>
          {state.hasFastestPlayer && (
            <p className="text-center text-sm text-amber-600">
              ⚡ En Hızlı Oyuncu: {nameFor(state.fastestPlayerId, state.roster, state.selfId)}
            </p>
          )}
          <ScoreTable
            sortedScores={sortedScores}
            pointsAwarded={{}}
            roster={state.roster}
            selfId={state.selfId}
          />
        </section>
      )}
    </div>
  );
}

function ScoreTable({
  sortedScores,
  pointsAwarded,
  roster,
  selfId,
}: {
  sortedScores: [string, number][];
  pointsAwarded: Record<string, number>;
  roster: { id: string; name: string }[];
  selfId: string | null;
}) {
  return (
    <ul className="flex flex-col gap-2">
      {sortedScores.map(([id, score], i) => (
        <li
          key={id}
          className="flex items-center justify-between rounded-lg bg-zinc-100 px-4 py-3 dark:bg-zinc-900"
        >
          <span>
            #{i + 1} {nameFor(id, roster, selfId)}
          </span>
          <span className="font-medium">
            {score}
            {pointsAwarded[id] ? (
              <span className="ml-2 text-sm text-green-600">+{pointsAwarded[id]}</span>
            ) : null}
          </span>
        </li>
      ))}
    </ul>
  );
}

export default function RoomPage() {
  return (
    <Suspense fallback={<p className="p-10 text-center text-zinc-500">Yükleniyor...</p>}>
      <RoomScreen />
    </Suspense>
  );
}
