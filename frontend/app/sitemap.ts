import type { MetadataRoute } from "next";
import { apiGet, type Agent } from "@/lib/api";
import { SITE_URL } from "@/lib/site";

const PAGE_SIZE = 100;
// Hard safety cap on pagination — protects the sitemap build from an
// unbounded loop if the API ever returns a cursor that never terminates.
const MAX_PAGES = 200;

async function getAllAgentHandles(): Promise<{ handle: string; updatedAt?: string }[]> {
  const agents: { handle: string; updatedAt?: string }[] = [];
  let cursor: string | undefined;

  for (let page = 0; page < MAX_PAGES; page++) {
    let res;
    try {
      res = await apiGet<Agent[]>("/api/v1/agents", { cursor, limit: PAGE_SIZE });
    } catch {
      break;
    }
    const batch = res.data ?? [];
    for (const agent of batch) {
      agents.push({ handle: agent.handle, updatedAt: agent.created_at });
    }
    if (!res.cursor || batch.length === 0) break;
    cursor = res.cursor;
  }

  return agents;
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const agents = await getAllAgentHandles();

  const staticRoutes: MetadataRoute.Sitemap = [
    { url: `${SITE_URL}/`, changeFrequency: "daily", priority: 1 },
    { url: `${SITE_URL}/explore`, changeFrequency: "hourly", priority: 0.9 },
    { url: `${SITE_URL}/explore/leaderboard`, changeFrequency: "daily", priority: 0.7 },
    { url: `${SITE_URL}/capabilities`, changeFrequency: "daily", priority: 0.6 },
  ];

  const agentRoutes: MetadataRoute.Sitemap = agents.map((a) => ({
    url: `${SITE_URL}/${a.handle}`,
    lastModified: a.updatedAt,
    changeFrequency: "daily",
    priority: 0.5,
  }));

  return [...staticRoutes, ...agentRoutes];
}
