"use client";

import { useEffect, useState } from "react";
import { config } from "@/lib/config";
import { useLocale } from "@/lib/i18n";
import { MIXED, categoryEmoji, categoryGradient, categoryLabel } from "@/lib/categories";

type ApiCategory = { slug: string; count: number };

/**
 * Lets the room creator pick which category the game draws from. The list
 * comes from the server, which only advertises categories it can fill a full
 * round with in the current language — so Turkish and English can legitimately
 * offer different lists, and neither can offer a category that would fail.
 */
export function CategoryPicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (slug: string) => void;
}) {
  const { locale, t } = useLocale();
  const [slugs, setSlugs] = useState<string[]>([]);

  useEffect(() => {
    let cancelled = false;
    fetch(`${config.apiBase}/categories?lang=${locale}`)
      .then((r) => (r.ok ? r.json() : Promise.reject(new Error("bad status"))))
      .then((data: { categories: ApiCategory[] }) => {
        if (!cancelled) setSlugs(data.categories.map((c) => c.slug));
      })
      // A failed lookup isn't fatal: the player can still start a mixed
      // game, which needs no category list.
      .catch(() => {
        if (!cancelled) setSlugs([]);
      });
    return () => {
      cancelled = true;
    };
  }, [locale]);

  const options = [MIXED, ...slugs];

  return (
    <div className="flex flex-col gap-2">
      <span className="text-sm font-semibold text-purple-800">{t("pickCategory")}</span>
      <div className="grid max-h-64 grid-cols-3 gap-2 overflow-y-auto p-1 sm:grid-cols-4">
        {options.map((slug) => {
          const selected = slug === value;
          return (
            <button
              key={slug}
              type="button"
              onClick={() => onChange(slug)}
              className={`group relative flex aspect-square flex-col items-center justify-center gap-1 overflow-hidden rounded-2xl bg-gradient-to-br p-2 text-center shadow-md transition active:scale-95 ${categoryGradient(
                slug
              )} ${
                selected
                  ? "ring-4 ring-purple-700 ring-offset-2"
                  : "opacity-90 hover:opacity-100 hover:shadow-lg"
              }`}
            >
              <span
                className="text-3xl drop-shadow sm:text-4xl"
                style={selected ? { animation: "pop 0.3s ease-out" } : undefined}
                aria-hidden
              >
                {categoryEmoji(slug)}
              </span>
              <span className="line-clamp-2 text-[11px] font-bold leading-tight text-white drop-shadow-sm sm:text-xs">
                {categoryLabel(slug, locale)}
              </span>
              {selected && (
                <span className="absolute right-1.5 top-1.5 flex h-5 w-5 items-center justify-center rounded-full bg-white text-xs font-bold text-purple-700 shadow">
                  ✓
                </span>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
