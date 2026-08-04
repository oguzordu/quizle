import { parseAvatar } from "@/lib/quips";

export function Avatar({ avatar, size = 32 }: { avatar?: string; size?: number }) {
  const { emoji, color } = parseAvatar(avatar);
  return (
    <span
      className="inline-flex items-center justify-center rounded-full"
      style={{ backgroundColor: color, width: size, height: size, fontSize: size * 0.55 }}
    >
      {emoji}
    </span>
  );
}
