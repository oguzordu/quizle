"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { config } from "@/lib/config";
import { randomAvatar, randomNickname } from "@/lib/quips";
import { AvatarPicker } from "@/components/AvatarPicker";
import { useLocale } from "@/lib/i18n";

const NAME_KEY = "quizle:profile:name";
const AVATAR_KEY = "quizle:profile:avatar";

const FLOATERS = [
  { emoji: "🎓", top: "6%", left: "8%", size: 40, delay: "0s", duration: "3.5s" },
  { emoji: "⭐", top: "12%", left: "82%", size: 28, delay: "0.4s", duration: "2.8s" },
  { emoji: "🪙", top: "24%", left: "88%", size: 32, delay: "0.8s", duration: "4.2s" },
  { emoji: "📚", top: "78%", left: "10%", size: 34, delay: "0.2s", duration: "3.1s" },
  { emoji: "🪙", top: "70%", left: "85%", size: 26, delay: "1s", duration: "3.9s" },
  { emoji: "✨", top: "40%", left: "4%", size: 24, delay: "0.6s", duration: "2.5s" },
];

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
    <div className="relative min-h-screen overflow-hidden bg-gradient-to-b from-violet-700 via-purple-800 to-purple-950 px-4 py-10">
      {FLOATERS.map((f, i) => (
        <span
          key={i}
          aria-hidden
          className="absolute select-none opacity-70"
          style={{
            top: f.top,
            left: f.left,
            fontSize: f.size,
            animation: `float ${f.duration} ease-in-out ${f.delay} infinite`,
          }}
        >
          {f.emoji}
        </span>
      ))}

      <div className="relative mx-auto w-full max-w-sm">
        <div className="mb-3 flex justify-end gap-1 text-xs">
          <button
            onClick={() => setLocale("tr")}
            className={`rounded-full px-2.5 py-1 transition ${
              locale === "tr" ? "bg-white font-semibold text-purple-800" : "text-purple-200"
            }`}
          >
            TR
          </button>
          <button
            onClick={() => setLocale("en")}
            className={`rounded-full px-2.5 py-1 transition ${
              locale === "en" ? "bg-white font-semibold text-purple-800" : "text-purple-200"
            }`}
          >
            EN
          </button>
        </div>

        <div className="flex flex-col items-center pb-4 pt-6 text-center">
          <span
            className="inline-block text-6xl"
            style={{ animation: "wobble 2.6s ease-in-out infinite" }}
          >
            🏆
          </span>
          <h1 className="mt-3 text-4xl font-extrabold text-white">
            {t("appTitle")} <span className="text-amber-300">Academy</span>
          </h1>
          <p className="mt-2 text-sm font-medium text-purple-200">{t("tagline")}</p>
        </div>

        <div
          style={{ animation: "fadeSlideIn 0.4s ease-out" }}
          className="rounded-3xl bg-white/95 p-6 shadow-2xl backdrop-blur sm:p-7"
        >
          <AvatarPicker value={avatar} onChange={saveAvatar} />

          <div className="mt-5 flex gap-2">
            <input
              className="w-full rounded-full border-2 border-purple-100 bg-purple-50/60 px-5 py-3 text-base text-purple-900 outline-none transition focus:border-purple-400"
              placeholder={t("namePlaceholder")}
              value={name}
              onChange={(e) => saveName(e.target.value)}
              maxLength={20}
            />
            <button
              type="button"
              title={t("randomName")}
              className="shrink-0 rounded-full border-2 border-purple-100 px-3.5 text-lg transition hover:bg-purple-50 active:scale-90"
              onClick={() => saveName(randomNickname())}
            >
              🎲
            </button>
          </div>

          <button
            className="mt-4 w-full rounded-full bg-gradient-to-r from-violet-600 to-purple-600 px-4 py-3.5 text-base font-bold text-white shadow-lg shadow-purple-600/30 transition hover:brightness-110 hover:[animation:wiggle_0.4s_ease-in-out] active:scale-[0.97] disabled:opacity-50"
            onClick={createRoom}
            disabled={busy}
          >
            {t("createRoom")}
          </button>

          <div className="my-5 flex items-center gap-3 text-xs font-medium text-purple-300">
            <div className="h-px flex-1 bg-purple-100" />
            {t("orJoin")}
            <div className="h-px flex-1 bg-purple-100" />
          </div>

          <div className="flex gap-2">
            <input
              className="w-full rounded-full border-2 border-purple-100 bg-purple-50/60 px-5 py-3 text-base uppercase tracking-widest text-purple-900 outline-none transition focus:border-purple-400"
              placeholder={t("roomCodePlaceholder")}
              value={joinCode}
              onChange={(e) => setJoinCode(e.target.value)}
              maxLength={6}
            />
            <button
              className="shrink-0 rounded-full border-2 border-purple-200 px-5 py-3 text-base font-bold text-purple-700 transition hover:bg-purple-50 hover:[animation:wiggle_0.4s_ease-in-out] active:scale-95"
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
