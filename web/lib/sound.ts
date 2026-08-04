"use client";

// Tiny synthesized sound effects via the Web Audio API — no audio files to
// ship or license. A shared AudioContext is created lazily on first use
// since browsers block audio until a user gesture has happened anyway.
let ctx: AudioContext | null = null;

function getContext(): AudioContext | null {
  if (typeof window === "undefined") return null;
  if (!ctx) {
    const AudioCtx = window.AudioContext ?? (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
    ctx = new AudioCtx();
  }
  if (ctx.state === "suspended") ctx.resume();
  return ctx;
}

function tone(freq: number, startOffset: number, duration: number, type: OscillatorType = "sine", gain = 0.15) {
  const audio = getContext();
  if (!audio) return;
  const osc = audio.createOscillator();
  const g = audio.createGain();
  osc.type = type;
  osc.frequency.value = freq;
  const startAt = audio.currentTime + startOffset;
  g.gain.setValueAtTime(gain, startAt);
  g.gain.exponentialRampToValueAtTime(0.001, startAt + duration);
  osc.connect(g);
  g.connect(audio.destination);
  osc.start(startAt);
  osc.stop(startAt + duration);
}

/** A bright two-note rising chime for a correct answer. */
export function playCorrect() {
  tone(660, 0, 0.12, "sine", 0.18);
  tone(880, 0.1, 0.18, "sine", 0.18);
}

/** A short low buzz for a wrong answer. */
export function playWrong() {
  tone(180, 0, 0.25, "sawtooth", 0.12);
}

/** A quiet click for UI interactions (e.g. picking an answer). */
export function playTick() {
  tone(440, 0, 0.05, "square", 0.06);
}
