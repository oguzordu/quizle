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
      return { ...state, roster: [...state.roster, { id: p.player_id, name: p.name }] };
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
      return {
        ...state,
        phase: "reveal",
        correctChoice: p.correct_choice,
        pointsAwarded: p.points_awarded,
        scores: p.scores,
      };
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
    default:
      return state;
  }
}

export function useGameConnection(code: string, name: string) {
  const [state, setState] = useState<GameState>(initialGameState);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!code) return;
    let cancelled = false;
    let retryTimer: ReturnType<typeof setTimeout> | undefined;

    function connect() {
      if (cancelled) return;
      const token = sessionStorage.getItem(tokenKey(code)) ?? "";
      const params = new URLSearchParams({ code, name });
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
            return { ...next, roster: [...next.roster, { id: p.player_id, name }] };
          });
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
      wsRef.current?.close();
    };
  }, [code, name]);

  const submitAnswer = useCallback((choice: number) => {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    setState((prev) => ({ ...prev, myAnswerChoice: choice }));
    ws.send(JSON.stringify({ type: "submit_answer", choice }));
  }, []);

  return { state, submitAnswer };
}
