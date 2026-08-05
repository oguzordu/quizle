"use client";

import { createContext, useContext, useEffect, useState } from "react";

export type Locale = "tr" | "en";

const dict = {
  tr: {
    appTitle: "Quizle",
    tagline: "Arkadaşlarını topla, kim daha hızlı cevaplayacak?",
    pickAvatar: "Avatarını seç",
    randomAvatar: "Rastgele avatar",
    pickCategory: "Kategori seç",
    speedBonus: "Hız Bonusu",
    namePlaceholder: "İsmin",
    randomName: "Rastgele takma ad",
    createRoom: "Yeni Oda Oluştur",
    orJoin: "veya bir odaya katıl",
    roomCodePlaceholder: "ORNKOD",
    join: "Katıl",
    needName: "Önce ismini yaz.",
    needCode: "Bir oda kodu gir.",
    createFailed: "Oda oluşturulamadı. Sunucu çalışıyor mu?",
    roomCode: "Oda kodu",
    connected: "bağlı",
    connecting: "bağlanıyor...",
    connectingToRoom: "Odaya bağlanılıyor...",
    shareCode: "Bu kodu paylaş, herkes katılınca oyunu başlat:",
    startGame: "Oyunu Başlat",
    you: "sen",
    correctAnswer: "Doğru cevap",
    waitingForResult: "Sonuç bekleniyor...",
    correctFeedback: "✅ Doğru!",
    wrongFeedback: "❌ Yanlış.",
    gameOver: "🏁 Oyun bitti!",
    question: "Soru",
    playAgain: "🔁 Tekrar Oyna",
    backToHome: "🏠 Ana Menü",
  },
  en: {
    appTitle: "Quizle",
    tagline: "Gather your friends — who'll answer first?",
    pickAvatar: "Pick your avatar",
    randomAvatar: "Random avatar",
    pickCategory: "Pick a category",
    speedBonus: "Speed Bonus",
    namePlaceholder: "Your name",
    randomName: "Random nickname",
    createRoom: "Create New Room",
    orJoin: "or join a room",
    roomCodePlaceholder: "CODE",
    join: "Join",
    needName: "Enter your name first.",
    needCode: "Enter a room code.",
    createFailed: "Couldn't create the room. Is the server running?",
    roomCode: "Room code",
    connected: "connected",
    connecting: "connecting...",
    connectingToRoom: "Connecting to room...",
    shareCode: "Share this code, then start once everyone's in:",
    startGame: "Start Game",
    you: "you",
    correctAnswer: "Correct answer",
    waitingForResult: "Waiting for results...",
    correctFeedback: "✅ Correct!",
    wrongFeedback: "❌ Wrong.",
    gameOver: "🏁 Game over!",
    question: "Question",
    playAgain: "🔁 Play Again",
    backToHome: "🏠 Home",
  },
} as const;

export type TranslationKey = keyof typeof dict.tr;

const LOCALE_KEY = "quizle:locale";

const LocaleContext = createContext<{
  locale: Locale;
  setLocale: (l: Locale) => void;
  t: (key: TranslationKey) => string;
}>({
  locale: "tr",
  setLocale: () => {},
  t: (key) => dict.tr[key],
});

export function LocaleProvider({ children }: { children: React.ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>("tr");

  useEffect(() => {
    const saved = localStorage.getItem(LOCALE_KEY);
    if (saved === "tr" || saved === "en") {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setLocaleState(saved);
    }
  }, []);

  function setLocale(l: Locale) {
    setLocaleState(l);
    localStorage.setItem(LOCALE_KEY, l);
  }

  function t(key: TranslationKey) {
    return dict[locale][key];
  }

  return (
    <LocaleContext.Provider value={{ locale, setLocale, t }}>{children}</LocaleContext.Provider>
  );
}

export function useLocale() {
  return useContext(LocaleContext);
}
