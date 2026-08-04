/**
 * Renders a country flag from a question's opaque `image` hint (e.g.
 * "flag:TR") using the bundled `flag-icons` SVGs — real images, not emoji
 * glyphs, since flag emoji don't render as flags on every platform (Windows
 * shows the raw two-letter code instead of a flag).
 */
export function FlagImage({ image, size = 72 }: { image?: string; size?: number }) {
  if (!image || !image.startsWith("flag:")) return null;
  const code = image.slice("flag:".length).toLowerCase();
  return (
    <div className="flex justify-center">
      <span
        className={`fi fi-${code} rounded-md shadow-md ring-1 ring-black/10`}
        style={{ width: size * 1.4, height: size }}
      />
    </div>
  );
}
