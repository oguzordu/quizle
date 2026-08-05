// Display metadata for question categories. The server decides *which*
// categories exist and are playable (it only advertises ones with enough
// questions in the requested language); this file only says how to draw
// them. A slug the server sends but we don't know here still renders, using
// the slug itself as its label, so adding a question pack never requires a
// frontend release.

export type CategoryMeta = {
  emoji: string;
  tr: string;
  en: string;
  /** Tailwind gradient stops, used as the card's "cover art" background. */
  gradient: string;
};

/** The pseudo-category for "draw from everything", offered first. */
export const MIXED = "mixed";

export const CATEGORY_META: Record<string, CategoryMeta> = {
  [MIXED]: { emoji: "🎲", tr: "Karışık", en: "Mixed", gradient: "from-fuchsia-500 to-purple-600" },

  general: { emoji: "🧠", tr: "Genel Kültür", en: "General Knowledge", gradient: "from-indigo-500 to-blue-600" },
  movies: { emoji: "🎬", tr: "Film", en: "Film", gradient: "from-rose-500 to-red-600" },
  music: { emoji: "🎵", tr: "Müzik", en: "Music", gradient: "from-pink-500 to-fuchsia-600" },
  tv: { emoji: "📺", tr: "Dizi & TV", en: "Television", gradient: "from-violet-500 to-indigo-600" },
  videogames: { emoji: "🎮", tr: "Oyun", en: "Video Games", gradient: "from-emerald-500 to-teal-600" },
  celebrities: { emoji: "🌟", tr: "Ünlüler", en: "Celebrities", gradient: "from-amber-400 to-orange-500" },
  geography: { emoji: "🌍", tr: "Coğrafya", en: "Geography", gradient: "from-sky-500 to-cyan-600" },
  history: { emoji: "🏛️", tr: "Tarih", en: "History", gradient: "from-yellow-600 to-amber-700" },
  science: { emoji: "🔬", tr: "Bilim", en: "Science & Nature", gradient: "from-lime-500 to-green-600" },
  sports: { emoji: "⚽", tr: "Spor", en: "Sports", gradient: "from-orange-500 to-red-500" },
  animals: { emoji: "🐾", tr: "Hayvanlar", en: "Animals", gradient: "from-teal-500 to-emerald-600" },
  arts: { emoji: "🎨", tr: "Sanat", en: "Art", gradient: "from-fuchsia-500 to-pink-600" },
  flags: { emoji: "🚩", tr: "Bayraklar", en: "Flags", gradient: "from-red-500 to-rose-600" },
  capitals: { emoji: "🏙️", tr: "Başkentler", en: "Capitals", gradient: "from-slate-500 to-gray-700" },
  books: { emoji: "📚", tr: "Kitaplar", en: "Books", gradient: "from-amber-600 to-yellow-700" },
  anime: { emoji: "🎌", tr: "Anime & Manga", en: "Anime & Manga", gradient: "from-pink-500 to-rose-500" },
  comics: { emoji: "💥", tr: "Çizgi Roman", en: "Comics", gradient: "from-yellow-500 to-orange-600" },
  cartoons: { emoji: "🧸", tr: "Çizgi Film", en: "Cartoons", gradient: "from-cyan-500 to-blue-500" },
  mythology: { emoji: "🐉", tr: "Mitoloji", en: "Mythology", gradient: "from-emerald-600 to-green-700" },
  computers: { emoji: "💻", tr: "Bilgisayar", en: "Computers", gradient: "from-blue-600 to-indigo-700" },
  math: { emoji: "🔢", tr: "Matematik", en: "Mathematics", gradient: "from-purple-500 to-violet-600" },
  politics: { emoji: "🗳️", tr: "Siyaset", en: "Politics", gradient: "from-slate-600 to-blue-800" },
  vehicles: { emoji: "🚗", tr: "Araçlar", en: "Vehicles", gradient: "from-gray-600 to-slate-700" },
  boardgames: { emoji: "♟️", tr: "Kutu Oyunları", en: "Board Games", gradient: "from-stone-600 to-neutral-700" },
  theatre: { emoji: "🎭", tr: "Tiyatro & Müzikal", en: "Musicals & Theatres", gradient: "from-red-600 to-purple-700" },
  gadgets: { emoji: "📱", tr: "Teknoloji", en: "Gadgets", gradient: "from-sky-600 to-indigo-600" },
};

const FALLBACK_GRADIENT = "from-purple-500 to-violet-600";

/** Gradient classes for a slug, with a neutral fallback for unknown slugs. */
export function categoryGradient(slug: string): string {
  return CATEGORY_META[slug]?.gradient ?? FALLBACK_GRADIENT;
}

/** Label for a slug in the given locale, falling back to the slug itself. */
export function categoryLabel(slug: string, locale: "tr" | "en"): string {
  return CATEGORY_META[slug]?.[locale] ?? slug;
}

/** Emoji for a slug, with a neutral fallback for slugs we don't know yet. */
export function categoryEmoji(slug: string): string {
  return CATEGORY_META[slug]?.emoji ?? "❓";
}
