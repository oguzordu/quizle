"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { config } from "@/lib/config";
import { randomAvatar, randomNickname } from "@/lib/quips";
import { AvatarPicker } from "@/components/AvatarPicker";

const NAME_KEY = "quizle:profile:name";
const AVATAR_KEY = "quizle:profile:avatar";

export default function Home() {
  const router = useRouter();
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
      setError("Önce ismini yaz.");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      const res = await fetch(`${config.apiBase}/rooms`, { method: "POST" });
      if (!res.ok) throw new Error("Oda oluşturulamadı");
      const data = (await res.json()) as { code: string };
      router.push(`/room/${data.code}?${profileQuery()}`);
    } catch {
      setError("Oda oluşturulamadı. Sunucu çalışıyor mu?");
      setBusy(false);
    }
  }

  function joinRoom() {
    if (!name.trim()) {
      setError("Önce ismini yaz.");
      return;
    }
    if (!joinCode.trim()) {
      setError("Bir oda kodu gir.");
      return;
    }
    router.push(`/room/${joinCode.trim().toUpperCase()}?${profileQuery()}`);
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-zinc-50 px-4 dark:bg-black">
      <div className="w-full max-w-sm space-y-6">
        <h1 className="text-center text-4xl font-semibold tracking-tight text-zinc-900 dark:text-zinc-50">
          Quizle
        </h1>
        <p className="text-center text-sm text-zinc-500 dark:text-zinc-400">
          Arkadaşlarınla çok oyunculu bilgi yarışması
        </p>

        <AvatarPicker value={avatar} onChange={saveAvatar} />

        <div className="flex gap-2">
          <input
            className="w-full rounded-lg border border-zinc-300 px-4 py-3 text-base dark:border-zinc-700 dark:bg-zinc-900"
            placeholder="İsmin"
            value={name}
            onChange={(e) => saveName(e.target.value)}
            maxLength={20}
          />
          <button
            type="button"
            title="Rastgele takma ad"
            className="rounded-lg border border-zinc-300 px-3 text-lg dark:border-zinc-700"
            onClick={() => saveName(randomNickname())}
          >
            🎲
          </button>
        </div>

        <button
          className="w-full rounded-lg bg-zinc-900 px-4 py-3 text-base font-medium text-white transition hover:bg-zinc-700 disabled:opacity-50 dark:bg-zinc-50 dark:text-zinc-900"
          onClick={createRoom}
          disabled={busy}
        >
          Yeni Oda Oluştur
        </button>

        <div className="flex items-center gap-3 text-xs text-zinc-400">
          <div className="h-px flex-1 bg-zinc-200 dark:bg-zinc-800" />
          veya bir odaya katıl
          <div className="h-px flex-1 bg-zinc-200 dark:bg-zinc-800" />
        </div>

        <div className="flex gap-2">
          <input
            className="w-full rounded-lg border border-zinc-300 px-4 py-3 text-base uppercase tracking-widest dark:border-zinc-700 dark:bg-zinc-900"
            placeholder="ORNKOD"
            value={joinCode}
            onChange={(e) => setJoinCode(e.target.value)}
            maxLength={6}
          />
          <button
            className="rounded-lg border border-zinc-300 px-5 py-3 text-base font-medium transition hover:bg-zinc-100 dark:border-zinc-700 dark:hover:bg-zinc-900"
            onClick={joinRoom}
          >
            Katıl
          </button>
        </div>

        {error && <p className="text-center text-sm text-red-600">{error}</p>}
      </div>
    </div>
  );
}
