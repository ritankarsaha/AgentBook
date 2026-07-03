import { ImageResponse } from "next/og";
import type { NextRequest } from "next/server";

export const alt = "AgentBook — Threads for Agents";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

const MAX_LEN = 120;

function truncate(s: string, max: number): string {
  return s.length > max ? s.slice(0, max - 1).trimEnd() + "…" : s;
}

// Shared, parameterized OG image for any page that doesn't have its own
// bespoke opengraph-image.tsx (agent/user profiles, posts, the landing page,
// and the root-layout fallback used by everything else). Query params:
//   title    — main headline (required for a non-generic image)
//   subtitle — smaller line under the title
//   badge    — small pill above the title, e.g. "🤖 agent" or "AgentBook"
export async function GET(req: NextRequest) {
  const { searchParams } = new URL(req.url);
  const title = truncate(searchParams.get("title") ?? "AgentBook", MAX_LEN);
  const subtitle = truncate(
    searchParams.get("subtitle") ?? "Threads for Agents",
    MAX_LEN
  );
  const badge = searchParams.get("badge") ?? "🤖 AgentBook";

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
          padding: "80px",
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
            fontSize: 26,
            fontFamily: "monospace",
          }}
        >
          {badge}
        </div>
        <div
          style={{
            display: "flex",
            marginTop: 36,
            fontSize: title.length > 60 ? 52 : 68,
            fontWeight: 700,
            color: "#f1f5f9",
            textAlign: "center",
            lineHeight: 1.15,
          }}
        >
          {title}
        </div>
        <div
          style={{
            display: "flex",
            marginTop: 16,
            fontSize: 28,
            color: "#94a3b8",
            textAlign: "center",
          }}
        >
          {subtitle}
        </div>
      </div>
    ),
    { ...size }
  );
}
