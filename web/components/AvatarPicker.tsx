"use client";

import { EMOJI_CHOICES, COLOR_CHOICES, randomAvatar } from "@/lib/quips";
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
        <div
          key={value}
          className="flex items-center justify-center rounded-full"
          style={{ animation: "pop 0.3s ease-out" }}
        >
          <Avatar avatar={value} size={72} />
        </div>
        <div className="flex flex-col gap-1">
          <span className="text-sm font-semibold text-purple-800">{t("pickAvatar")}</span>
          <button
            type="button"
            title={t("randomAvatar")}
            className="w-fit rounded-full border-2 border-purple-100 px-3 py-1 text-base transition hover:bg-purple-50 active:scale-90"
            onClick={() => onChange(randomAvatar())}
          >
            🎲
          </button>
        </div>
      </div>
      <div className="grid grid-cols-6 gap-2 sm:grid-cols-8">
        {EMOJI_CHOICES.map((e) => (
          <button
            key={e}
            type="button"
            onClick={() => onChange(`${e}|${color}`)}
            className={`flex aspect-square items-center justify-center rounded-full border-2 text-xl transition active:scale-90 ${
              e === emoji ? "border-purple-500" : "border-transparent hover:border-purple-200"
            }`}
            style={e === emoji ? { backgroundColor: `${color}33` } : undefined}
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
            className="h-7 w-7 rounded-full transition active:scale-90"
            style={{
              backgroundColor: c,
              ...(c === color
                ? { boxShadow: "0 0 0 2px white, 0 0 0 4px " + c, animation: "ringPulse 1.2s ease-out" }
                : { boxShadow: "0 0 0 2px white, 0 0 0 4px transparent" }),
            }}
          />
        ))}
      </div>
    </div>
  );
}
