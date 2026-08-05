"use client";

import { useEffect, useState } from "react";
import { config } from "@/lib/config";
import { useLocale } from "@/lib/i18n";
import { MIXED, categoryEmoji, categoryLabel } from "@/lib/categories";

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
      <div className="flex max-h-40 flex-wrap gap-1.5 overflow-y-auto pr-1">
        {options.map((slug) => {
          const selected = slug === value;
          return (
            <button
              key={slug}
              type="button"
              onClick={() => onChange(slug)}
              className={`flex items-center gap-1 rounded-full border-2 px-3 py-1.5 text-xs font-bold transition active:scale-95 ${
                selected
                  ? "border-purple-600 bg-purple-600 text-white"
                  : "border-purple-100 bg-purple-50/60 text-purple-700 hover:border-purple-300"
              }`}
            >
              <span aria-hidden>{categoryEmoji(slug)}</span>
              {categoryLabel(slug, locale)}
            </button>
          );
        })}
      </div>
    </div>
  );
}
