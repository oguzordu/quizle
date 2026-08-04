/**
 * Converts a 2-letter ISO country code into its flag emoji using Unicode
 * regional indicator symbols — no image assets or libraries needed.
 * "TR" -> 🇹🇷
 */
export function flagEmoji(isoCode: string): string {
  const code = isoCode.toUpperCase();
  if (code.length !== 2) return "🏳️";
  const base = 0x1f1e6; // regional indicator "A"
  const chars = [...code].map((c) => base + (c.charCodeAt(0) - 65));
  return String.fromCodePoint(...chars);
}

/** Parses a question's opaque `image` hint (e.g. "flag:TR") into an emoji, or null if not a flag. */
export function flagFromImageHint(image?: string): string | null {
  if (!image || !image.startsWith("flag:")) return null;
  return flagEmoji(image.slice("flag:".length));
}
