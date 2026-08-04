"use client";

import { EMOJI_CHOICES, COLOR_CHOICES } from "@/lib/quips";
import { useLocale } from "@/lib/i18n";
import { Avatar } from "./Avatar";

export function AvatarPicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (avatar: string) => void;
}) {
  const { t } = useLocale();
  const [emoji, color] = value.includes("|") ? value.split("|") : ["🙂", "#94A3B8"];

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center gap-3">
        <Avatar avatar={value} size={52} />
        <span className="text-sm font-medium text-zinc-600 dark:text-zinc-300">
          {t("pickAvatar")}
        </span>
      </div>
      <div className="grid grid-cols-6 gap-2 sm:grid-cols-8">
        {EMOJI_CHOICES.map((e) => (
          <button
            key={e}
            type="button"
            onClick={() => onChange(`${e}|${color}`)}
            className={`flex aspect-square items-center justify-center rounded-full border-2 text-lg transition active:scale-90 ${
              e === emoji
                ? "border-indigo-500 bg-indigo-50 dark:bg-indigo-950"
                : "border-transparent hover:border-zinc-300 dark:hover:border-zinc-700"
            }`}
          >
            {e}
          </button>
        ))}
      </div>
      <div className="flex flex-wrap gap-2">
        {COLOR_CHOICES.map((c) => (
          <button
            key={c}
            type="button"
            aria-label={c}
            onClick={() => onChange(`${emoji}|${c}`)}
            className={`h-7 w-7 rounded-full ring-2 ring-offset-2 transition active:scale-90 dark:ring-offset-zinc-900 ${
              c === color ? "ring-indigo-500" : "ring-transparent"
            }`}
            style={{ backgroundColor: c }}
          />
        ))}
      </div>
    </div>
  );
}
