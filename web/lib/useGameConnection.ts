"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { config } from "./config";
import {
  AnswerAcceptedPayload,
  GameFinishedPayload,
  GameState,
  JoinedPayload,
  PlayerJoinedPayload,
  QuestionRevealedPayload,
  QuestionStartedPayload,
  ServerMessage,
  initialGameState,
} from "./gameTypes";

function tokenKey(code: string) {
  return `quizle:token:${code}`;
}

// How long to keep showing the question screen (with the player's own
// correct/wrong button already colored in) after the round closes, before
// switching to the shared reveal screen. Without this, the last person to
// answer never gets to see their own feedback — the screen jumps straight
// to results the instant they click.
const REVEAL_TRANSITION_DELAY_MS = 1300;

function applyMessage(state: GameState, msg: ServerMessage): GameState {
  switch (msg.type) {
    case "joined": {
      const p = msg.payload as JoinedPayload;
      return {
        ...state,
        connected: true,
        selfId: p.player_id,
        phase: state.phase === "connecting" ? "lobby" : state.phase,
      };
    }
    case "player_joined": {
      const p = msg.payload as PlayerJoinedPayload;
      if (state.roster.some((r) => r.id === p.player_id)) return state;
      return {
        ...state,
        roster: [...state.roster, { id: p.player_id, name: p.name, avatar: p.avatar }],
      };
    }
    case "question_started": {
      const p = msg.payload as QuestionStartedPayload;
      return {
        ...state,
        phase: "question",
        question: p,
        myAnswerChoice: null,
        myAnswerCorrect: null,
        correctChoice: null,
        pointsAwarded: {},
      };
    }
    case "answer_accepted": {
      const p = msg.payload as AnswerAcceptedPayload;
      if (p.player_id !== state.selfId) return state;
      return { ...state, myAnswerCorrect: p.correct };
    }
    case "question_revealed": {
      const p = msg.payload as QuestionRevealedPayload;
      // Only the answer data lands here — correctChoice becomes visible
      // right away so the correct button can turn green even before
      // whoever just answered. The phase itself flips to "reveal" a beat
      // later (see useGameConnection), so that green highlight is visible
      // on the question screen instead of being skipped straight past.
      return {
        ...state,
        correctChoice: p.correct_choice,
        pointsAwarded: p.points_awarded,
        scores: p.scores,
      };
    }
    case "__enter_reveal_phase": {
      return { ...state, phase: "reveal" };
    }
    case "game_finished": {
      const p = msg.payload as GameFinishedPayload;
      return {
        ...state,
        phase: "finished",
        scores: p.final_scores,
        hasFastestPlayer: p.has_fastest_player,
        fastestPlayerId: p.fastest_player_id ?? null,
      };
    }
    case "game_reset": {
      return {
        ...state,
        phase: "lobby",
        question: null,
        myAnswerChoice: null,
        myAnswerCorrect: null,
        correctChoice: null,
        pointsAwarded: {},
        scores: {},
        hasFastestPlayer: false,
        fastestPlayerId: null,
      };
    }
    default:
      return state;
  }
}

export function useGameConnection(code: string, name: string, avatar: string) {
  const [state, setState] = useState<GameState>(initialGameState);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!code) return;
    let cancelled = false;
    let retryTimer: ReturnType<typeof setTimeout> | undefined;
    const pendingTimers = new Set<ReturnType<typeof setTimeout>>();

    function connect() {
      if (cancelled) return;
      const token = sessionStorage.getItem(tokenKey(code)) ?? "";
      const params = new URLSearchParams({ code, name, avatar });
      if (token) params.set("token", token);
      const ws = new WebSocket(`${config.wsBase}/ws?${params.toString()}`);
      wsRef.current = ws;

      ws.onmessage = (evt) => {
        const msg = JSON.parse(evt.data) as ServerMessage;

        if (msg.type === "joined") {
          // The server never echoes our own join as a player_joined
          // broadcast (it only notifies other, already-connected clients),
          // so we add ourselves to the roster here using the name we
          // already know from this connection's own parameters.
          const p = msg.payload as JoinedPayload;
          sessionStorage.setItem(tokenKey(code), p.token);
          setState((prev) => {
            const next = applyMessage(prev, msg);
            if (next.roster.some((r) => r.id === p.player_id)) return next;
            return { ...next, roster: [...next.roster, { id: p.player_id, name, avatar }] };
          });
          return;
        }

        if (msg.type === "question_revealed") {
          // Apply the answer data (including which choice was correct)
          // immediately, so the correct button can turn green right away —
          // but hold the phase transition a beat longer so whoever just
          // answered sees that feedback instead of being yanked straight to
          // the results screen.
          setState((prev) => applyMessage(prev, msg));
          const timer = setTimeout(() => {
            pendingTimers.delete(timer);
            if (cancelled) return;
            setState((prev) => applyMessage(prev, { type: "__enter_reveal_phase", payload: null }));
          }, REVEAL_TRANSITION_DELAY_MS);
          pendingTimers.add(timer);
          return;
        }

        setState((prev) => applyMessage(prev, msg));
      };
      ws.onclose = () => {
        if (cancelled) return;
        setState((prev) => ({ ...prev, connected: false }));
        retryTimer = setTimeout(connect, 1500);
      };
      ws.onerror = () => {
        ws.close();
      };
    }

    // Deferred rather than called synchronously: React's development-mode
    // Strict Mode mounts every effect, cleans it up, then mounts it again in
    // the same tick to catch missing cleanup logic. Calling `new
    // WebSocket(...)` synchronously here would open a real connection (and
    // let the server register a real player) for that throwaway first
    // mount. Deferring by a tick means the throwaway mount's cleanup sets
    // `cancelled` before this ever runs, so only the real, final mount
    // actually opens a socket.
    const initialConnectTimer = setTimeout(connect, 0);

    return () => {
      cancelled = true;
      clearTimeout(initialConnectTimer);
      if (retryTimer) clearTimeout(retryTimer);
      pendingTimers.forEach(clearTimeout);
      wsRef.current?.close();
    };
  }, [code, name, avatar]);

  const submitAnswer = useCallback((choice: number) => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    setState((prev) => ({ ...prev, myAnswerChoice: choice }));
    ws.send(JSON.stringify({ type: "submit_answer", choice }));
  }, []);

  const rematch = useCallback(async () => {
    await fetch(`${config.apiBase}/rooms/${code}/rematch`, { method: "POST" });
  }, [code]);

  return { state, submitAnswer, rematch };
}
