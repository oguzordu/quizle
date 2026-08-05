import { ImageResponse } from "next/og";

export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function OpengraphImage() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          background: "linear-gradient(to bottom, #6d28d9, #581c87, #3b0764)",
        }}
      >
        <div style={{ fontSize: 140, display: "flex" }}>🏆</div>
        <div style={{ fontSize: 96, fontWeight: 800, color: "white", display: "flex" }}>
          Quizle
        </div>
        <div style={{ fontSize: 36, color: "#e9d5ff", marginTop: 12, display: "flex" }}>
          Arkadaşlarını topla, kim daha hızlı cevaplayacak?
        </div>
      </div>
    ),
    size
  );
}
