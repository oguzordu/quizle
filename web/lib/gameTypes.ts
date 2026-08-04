export type PlayerId = string;

export interface ServerMessage<T = unknown> {
  type: string;
  payload: T;
}

export interface JoinedPayload {
  player_id: PlayerId;
  token: string;
}

export interface PlayerJoinedPayload {
  player_id: PlayerId;
  name: string;
}

export interface QuestionStartedPayload {
  question_id: string;
  choices: string[];
  deadline: string; // RFC3339
}

export interface AnswerAcceptedPayload {
  player_id: PlayerId;
  correct: boolean;
}

export interface QuestionRevealedPayload {
  correct_choice: number;
  points_awarded: Record<PlayerId, number>;
  scores: Record<PlayerId, number>;
}

export interface GameFinishedPayload {
  final_scores: Record<PlayerId, number>;
  has_fastest_player: boolean;
  fastest_player_id?: PlayerId;
}

export type GamePhase = "connecting" | "lobby" | "question" | "reveal" | "finished";

export interface Player {
  id: PlayerId;
  name: string;
}

export interface GameState {
  phase: GamePhase;
  connected: boolean;
  selfId: PlayerId | null;
  roster: Player[];
  question: QuestionStartedPayload | null;
  myAnswerChoice: number | null;
  myAnswerCorrect: boolean | null;
  correctChoice: number | null;
  pointsAwarded: Record<PlayerId, number>;
  scores: Record<PlayerId, number>;
  hasFastestPlayer: boolean;
  fastestPlayerId: PlayerId | null;
  error: string | null;
}

export const initialGameState: GameState = {
  phase: "connecting",
  connected: false,
  selfId: null,
  roster: [],
  question: null,
  myAnswerChoice: null,
  myAnswerCorrect: null,
  correctChoice: null,
  pointsAwarded: {},
  scores: {},
  hasFastestPlayer: false,
  fastestPlayerId: null,
  error: null,
};
