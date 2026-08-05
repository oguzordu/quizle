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
};

/** The pseudo-category for "draw from everything", offered first. */
export const MIXED = "mixed";

export const CATEGORY_META: Record<string, CategoryMeta> = {
  [MIXED]: { emoji: "🎲", tr: "Karışık", en: "Mixed" },

  general: { emoji: "🧠", tr: "Genel Kültür", en: "General Knowledge" },
  movies: { emoji: "🎬", tr: "Film", en: "Film" },
  music: { emoji: "🎵", tr: "Müzik", en: "Music" },
  tv: { emoji: "📺", tr: "Dizi & TV", en: "Television" },
  videogames: { emoji: "🎮", tr: "Oyun", en: "Video Games" },
  celebrities: { emoji: "🌟", tr: "Ünlüler", en: "Celebrities" },
  geography: { emoji: "🌍", tr: "Coğrafya", en: "Geography" },
  history: { emoji: "🏛️", tr: "Tarih", en: "History" },
  science: { emoji: "🔬", tr: "Bilim", en: "Science & Nature" },
  sports: { emoji: "⚽", tr: "Spor", en: "Sports" },
  animals: { emoji: "🐾", tr: "Hayvanlar", en: "Animals" },
  arts: { emoji: "🎨", tr: "Sanat", en: "Art" },
  flags: { emoji: "🚩", tr: "Bayraklar", en: "Flags" },
  capitals: { emoji: "🏙️", tr: "Başkentler", en: "Capitals" },
  books: { emoji: "📚", tr: "Kitaplar", en: "Books" },
  anime: { emoji: "🎌", tr: "Anime & Manga", en: "Anime & Manga" },
  comics: { emoji: "💥", tr: "Çizgi Roman", en: "Comics" },
  cartoons: { emoji: "🧸", tr: "Çizgi Film", en: "Cartoons" },
  mythology: { emoji: "🐉", tr: "Mitoloji", en: "Mythology" },
  computers: { emoji: "💻", tr: "Bilgisayar", en: "Computers" },
  math: { emoji: "🔢", tr: "Matematik", en: "Mathematics" },
  politics: { emoji: "🗳️", tr: "Siyaset", en: "Politics" },
  vehicles: { emoji: "🚗", tr: "Araçlar", en: "Vehicles" },
  boardgames: { emoji: "♟️", tr: "Kutu Oyunları", en: "Board Games" },
  theatre: { emoji: "🎭", tr: "Tiyatro & Müzikal", en: "Musicals & Theatres" },
  gadgets: { emoji: "📱", tr: "Teknoloji", en: "Gadgets" },
};

/** Label for a slug in the given locale, falling back to the slug itself. */
export function categoryLabel(slug: string, locale: "tr" | "en"): string {
  return CATEGORY_META[slug]?.[locale] ?? slug;
}

/** Emoji for a slug, with a neutral fallback for slugs we don't know yet. */
export function categoryEmoji(slug: string): string {
  return CATEGORY_META[slug]?.emoji ?? "❓";
}
