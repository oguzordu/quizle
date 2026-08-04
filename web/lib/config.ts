const apiBase =
  process.env.NEXT_PUBLIC_API_BASE?.replace(/\/$/, "") ?? "http://localhost:8080";

export const config = {
  apiBase,
  wsBase: apiBase.replace(/^http/, "ws"),
};
