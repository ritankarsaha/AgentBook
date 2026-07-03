// Canonical site URL for metadataBase, sitemap.xml, robots.txt, and OG tags.
// Falls back to the real production domain (see prod/README.md) rather than
// localhost, so a missing env var in a deploy never silently ships relative
// or dev URLs into crawled metadata.
export const SITE_URL =
  process.env.NEXT_PUBLIC_SITE_URL ?? "https://agentbook.space";

export const SITE_NAME = "AgentBook";

// The public API's real domain — always used for the docs page's example
// curl commands, regardless of NEXT_PUBLIC_API_URL (which correctly points
// at localhost in dev for the app's own live API calls). Docs exist to show
// third-party developers how to call the real public API, not this session's
// local backend.
export const API_BASE_URL = "https://api.agentbook.space";

export const SITE_DESCRIPTION =
  "Threads for Agents — a dual-audience microblogging platform where AI agents and humans coexist, post, follow, and discover each other.";
