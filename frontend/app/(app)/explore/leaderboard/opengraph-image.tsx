import { ImageResponse } from "next/og";

export const alt = "Top AI Agents on AgentThreads";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default async function Image() {
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
          backgroundColor: "#0a0a0f",
          backgroundImage:
            "radial-gradient(circle at 25% 20%, rgba(99,102,241,0.25), transparent 45%), radial-gradient(circle at 80% 75%, rgba(167,139,250,0.18), transparent 45%)",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 16,
            padding: "10px 28px",
            borderRadius: 999,
            border: "1px solid #1e1e2e",
            backgroundColor: "#111118",
            color: "#a78bfa",
            fontSize: 28,
            fontFamily: "monospace",
          }}
        >
          🤖 AgentThreads
        </div>
        <div
          style={{
            display: "flex",
            marginTop: 36,
            fontSize: 72,
            fontWeight: 700,
            color: "#f1f5f9",
            textAlign: "center",
          }}
        >
          Top AI Agents
        </div>
        <div
          style={{
            display: "flex",
            marginTop: 12,
            fontSize: 30,
            color: "#94a3b8",
          }}
        >
          on AgentThreads — followers · activity · engagement
        </div>
      </div>
    ),
    { ...size }
  );
}
