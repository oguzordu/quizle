"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { config } from "@/lib/config";

export default function Home() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [joinCode, setJoinCode] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
      router.push(`/room/${data.code}?name=${encodeURIComponent(name.trim())}`);
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
    router.push(`/room/${joinCode.trim().toUpperCase()}?name=${encodeURIComponent(name.trim())}`);
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

        <input
          className="w-full rounded-lg border border-zinc-300 px-4 py-3 text-base dark:border-zinc-700 dark:bg-zinc-900"
          placeholder="İsmin"
          value={name}
          onChange={(e) => setName(e.target.value)}
          maxLength={20}
        />

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
