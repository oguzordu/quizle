// Turkish flavor text: nickname generator, waiting-room quips, and
// end-of-game joke titles. All computed client-side from data already in
// GameState — no extra backend calls needed.

const ADJECTIVES = [
  "Öfkeli",
  "Gizemli",
  "Uykulu",
  "Efsanevi",
  "Sinsi",
  "Coşkulu",
  "Tembel",
  "Vahşi",
  "Kararsız",
  "Bilge",
];

const NOUNS = [
  "Simit",
  "Patates",
  "Ahtapot",
  "Baykuş",
  "Zürafa",
  "Semaver",
  "Kirpi",
  "Papağan",
  "Sincap",
  "Vatoz",
];

export function randomNickname(): string {
  const a = ADJECTIVES[Math.floor(Math.random() * ADJECTIVES.length)];
  const n = NOUNS[Math.floor(Math.random() * NOUNS.length)];
  return `${a} ${n}`;
}

export const EMOJI_CHOICES = [
  "🦊", "🐙", "🦄", "🐸", "🦉", "🐢", "🐼", "🦁", "🐧", "🦖",
  "👻", "🤖", "🎃", "🍕", "⚡", "🔥",
];

export const COLOR_CHOICES = [
  "#EF4444", "#F97316", "#EAB308", "#22C55E",
  "#06B6D4", "#3B82F6", "#8B5CF6", "#EC4899",
];

export function randomAvatar(): string {
  const emoji = EMOJI_CHOICES[Math.floor(Math.random() * EMOJI_CHOICES.length)];
  const color = COLOR_CHOICES[Math.floor(Math.random() * COLOR_CHOICES.length)];
  return `${emoji}|${color}`;
}

export function parseAvatar(avatar: string | undefined): { emoji: string; color: string } {
  if (!avatar || !avatar.includes("|")) {
    return { emoji: "🙂", color: "#94A3B8" };
  }
  const [emoji, color] = avatar.split("|");
  return { emoji, color };
}

export const LOBBY_WAITING_QUIPS = [
  "Herkes hazır olana kadar biraz gerinin.",
  "Şampiyonluk burada başlıyor (ya da rezillik, göreceğiz).",
  "Telefonunu değil, sorulara odaklan.",
  "Rakiplerini gözünle süz, korkut biraz.",
];

export function randomLobbyQuip(): string {
  return LOBBY_WAITING_QUIPS[Math.floor(Math.random() * LOBBY_WAITING_QUIPS.length)];
}

export const EVERYONE_WRONG_QUIPS = [
  "Kimse bilemedi! Bu soruyu kim yazdı ki?",
  "Hepiniz aynı anda yanlış bildiniz, tebrikler sayılır mı bilmiyorum.",
  "Google açık olsaydı hepiniz kazanırdınız muhtemelen.",
];

export function randomEveryoneWrongQuip(): string {
  return EVERYONE_WRONG_QUIPS[Math.floor(Math.random() * EVERYONE_WRONG_QUIPS.length)];
}

export interface FinalTitle {
  emoji: string;
  title: string;
}

/**
 * Computes a lighthearted title for each player from data already visible
 * at game end: their rank by score and whether they won the fastest-player
 * award. No extra per-question stats are needed, keeping this entirely
 * derived from GameState.
 */
export function finalTitleFor(
  playerId: string,
  rank: number,
  totalPlayers: number,
  score: number,
  isFastestPlayer: boolean
): FinalTitle {
  if (rank === 1) return { emoji: "🏆", title: "Quiz Ustası" };
  if (isFastestPlayer) return { emoji: "⚡", title: "Yıldırım Refleks" };
  if (score === 0) return { emoji: "💤", title: "Işıklar Açık Kimse Yok" };
  if (rank === totalPlayers) return { emoji: "🐢", title: "Son Ama Onurlu" };
  return { emoji: "🎯", title: "Fena Değil" };
}
