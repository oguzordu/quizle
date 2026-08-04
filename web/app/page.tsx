"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { config } from "@/lib/config";
import { randomAvatar, randomNickname } from "@/lib/quips";
import { AvatarPicker } from "@/components/AvatarPicker";
import { useLocale } from "@/lib/i18n";

const NAME_KEY = "quizle:profile:name";
const AVATAR_KEY = "quizle:profile:avatar";

export default function Home() {
  const router = useRouter();
  const { locale, setLocale, t } = useLocale();
  const [name, setName] = useState("");
  const [avatar, setAvatar] = useState("");
  const [joinCode, setJoinCode] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Reading a saved profile from localStorage is inherently client-only
    // (unavailable during server render), so this has to run after mount
    // rather than as a lazy useState initializer.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setName(localStorage.getItem(NAME_KEY) ?? randomNickname());
    setAvatar(localStorage.getItem(AVATAR_KEY) ?? randomAvatar());
  }, []);

  function saveName(v: string) {
    setName(v);
    localStorage.setItem(NAME_KEY, v);
  }

  function saveAvatar(v: string) {
    setAvatar(v);
    localStorage.setItem(AVATAR_KEY, v);
  }

  function profileQuery() {
    return `name=${encodeURIComponent(name.trim())}&avatar=${encodeURIComponent(avatar)}`;
  }

  async function createRoom() {
    if (!name.trim()) {
      setError(t("needName"));
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const res = await fetch(`${config.apiBase}/rooms`, { method: "POST" });
      if (!res.ok) throw new Error("create failed");
      const data = (await res.json()) as { code: string };
      router.push(`/room/${data.code}?${profileQuery()}`);
    } catch {
      setError(t("createFailed"));
      setBusy(false);
    }
  }

  function joinRoom() {
    if (!name.trim()) {
      setError(t("needName"));
      return;
    }
    if (!joinCode.trim()) {
      setError(t("needCode"));
      return;
    }
    router.push(`/room/${joinCode.trim().toUpperCase()}?${profileQuery()}`);
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gradient-to-b from-indigo-50 via-white to-white px-4 py-10 dark:from-zinc-950 dark:via-black dark:to-black">
      <div className="w-full max-w-sm">
        <div className="mb-3 flex justify-end gap-1 text-xs">
          <button
            onClick={() => setLocale("tr")}
            className={`rounded px-2 py-1 transition ${
              locale === "tr"
                ? "bg-indigo-600 font-semibold text-white"
                : "text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-300"
            }`}
          >
            TR
          </button>
          <button
            onClick={() => setLocale("en")}
            className={`rounded px-2 py-1 transition ${
              locale === "en"
                ? "bg-indigo-600 font-semibold text-white"
                : "text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-300"
            }`}
          >
            EN
          </button>
        </div>

        <div className="rounded-2xl border border-zinc-200/80 bg-white p-6 shadow-xl shadow-indigo-950/5 dark:border-zinc-800 dark:bg-zinc-900 sm:p-8">
          <h1 className="text-center text-4xl font-extrabold tracking-tight text-transparent bg-clip-text bg-gradient-to-br from-indigo-600 to-fuchsia-600">
            {t("appTitle")}
          </h1>
          <p className="mt-2 text-center text-sm text-zinc-500 dark:text-zinc-400">
            {t("tagline")}
          </p>

          <div className="mt-6">
            <AvatarPicker value={avatar} onChange={saveAvatar} />
          </div>

          <div className="mt-5 flex gap-2">
            <input
              className="w-full rounded-xl border border-zinc-300 px-4 py-3 text-base outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 dark:border-zinc-700 dark:bg-zinc-800"
              placeholder={t("namePlaceholder")}
              value={name}
              onChange={(e) => saveName(e.target.value)}
              maxLength={20}
            />
            <button
              type="button"
              title={t("randomName")}
              className="shrink-0 rounded-xl border border-zinc-300 px-3 text-lg transition hover:bg-zinc-100 active:scale-95 dark:border-zinc-700 dark:hover:bg-zinc-800"
              onClick={() => saveName(randomNickname())}
            >
              🎲
            </button>
          </div>

          <button
            className="mt-4 w-full rounded-xl bg-gradient-to-br from-indigo-600 to-fuchsia-600 px-4 py-3.5 text-base font-semibold text-white shadow-lg shadow-indigo-600/25 transition hover:brightness-110 active:scale-[0.98] disabled:opacity-50"
            onClick={createRoom}
            disabled={busy}
          >
            {t("createRoom")}
          </button>

          <div className="my-5 flex items-center gap-3 text-xs text-zinc-400">
            <div className="h-px flex-1 bg-zinc-200 dark:bg-zinc-800" />
            {t("orJoin")}
            <div className="h-px flex-1 bg-zinc-200 dark:bg-zinc-800" />
          </div>

          <div className="flex gap-2">
            <input
              className="w-full rounded-xl border border-zinc-300 px-4 py-3 text-base uppercase tracking-widest outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 dark:border-zinc-700 dark:bg-zinc-800"
              placeholder={t("roomCodePlaceholder")}
              value={joinCode}
              onChange={(e) => setJoinCode(e.target.value)}
              maxLength={6}
            />
            <button
              className="shrink-0 rounded-xl border border-zinc-300 px-5 py-3 text-base font-medium transition hover:bg-zinc-100 active:scale-95 dark:border-zinc-700 dark:hover:bg-zinc-800"
              onClick={joinRoom}
            >
              {t("join")}
            </button>
          </div>

          {error && <p className="mt-4 text-center text-sm text-red-600">{error}</p>}
        </div>
      </div>
    </div>
  );
}
