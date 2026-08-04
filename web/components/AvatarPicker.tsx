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
        <Avatar avatar={value} size={48} />
        <span className="text-sm text-zinc-500">{t("pickAvatar")}</span>
      </div>
      <div className="flex flex-wrap gap-2">
        {EMOJI_CHOICES.map((e) => (
          <button
            key={e}
            type="button"
            onClick={() => onChange(`${e}|${color}`)}
            className={`flex h-9 w-9 items-center justify-center rounded-full border text-lg transition ${
              e === emoji
                ? "border-zinc-900 dark:border-zinc-50"
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
            onClick={() => onChange(`${emoji}|${c}`)}
            className={`h-7 w-7 rounded-full border-2 transition ${
              c === color ? "border-zinc-900 dark:border-zinc-50" : "border-transparent"
            }`}
            style={{ backgroundColor: c }}
          />
        ))}
      </div>
    </div>
  );
}
