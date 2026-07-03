// Parses the backend's /llms-full.txt into structured sections so the docs
// page can render real HTML (not a <pre> dump) and synthesize a runnable
// curl example for every documented endpoint. The source format is a small,
// known convention (not general markdown): "#"/"##"/"###" headings, "---"
// dividers, plain paragraph lines, and 2-space-indented lines used as code
// blocks — there are no fenced code blocks anywhere in the source.

export type DocLine = { text: string; indented: boolean };

export type DocChunk = {
  heading: string;
  endpoint?: { method: string; path: string; authTag?: string };
  lines: DocLine[];
};

export type DocSection = {
  title: string;
  // Lines appearing directly under the "##" heading, before any "###" sub-
  // heading (e.g. "## MCP Server" has two plain lines and no "###" at all).
  intro: DocLine[];
  chunks: DocChunk[];
};

export type ParsedDoc = {
  title: string;
  preamble: DocLine[];
  sections: DocSection[];
};

const ENDPOINT_RE = /^(GET|POST|PUT|DELETE|PATCH)\s+(\S+)\s*(?:\[([^\]]+)\])?\s*$/;

export function parseLlmsFullTxt(raw: string): ParsedDoc {
  const rawLines = raw.replace(/\r\n/g, "\n").split("\n");

  let title = "";
  const preamble: DocLine[] = [];
  const sections: DocSection[] = [];

  let currentSection: DocSection | null = null;
  let currentChunk: DocChunk | null = null;

  const pushChunk = () => {
    if (currentChunk && currentSection) currentSection.chunks.push(currentChunk);
    currentChunk = null;
  };
  const pushSection = () => {
    pushChunk();
    if (currentSection) sections.push(currentSection);
    currentSection = null;
  };

  for (const raw of rawLines) {
    const line = raw.replace(/\s+$/, "");
    if (line.trim() === "---") continue; // dividers are purely visual in the source
    if (line.trim() === "") continue;

    if (line.startsWith("# ")) {
      title = line.slice(2).trim();
      continue;
    }
    if (line.startsWith("## ")) {
      pushSection();
      currentSection = { title: line.slice(3).trim(), intro: [], chunks: [] };
      continue;
    }
    if (line.startsWith("### ")) {
      pushChunk();
      const heading = line.slice(4).trim();
      const m = heading.match(ENDPOINT_RE);
      currentChunk = {
        heading,
        endpoint: m ? { method: m[1], path: m[2], authTag: m[3]?.trim() } : undefined,
        lines: [],
      };
      continue;
    }

    const indented = /^ {2,}/.test(line);
    const text = indented ? line.replace(/^ +/, "") : line.trim();
    if (currentChunk) {
      currentChunk.lines.push({ text, indented });
    } else if (currentSection) {
      currentSection.intro.push({ text, indented });
    } else {
      preamble.push({ text, indented });
    }
  }
  pushSection();

  return { title, preamble, sections };
}

// ─── curl example synthesis ────────────────────────────────────────────────

const FIELD_EXAMPLES: Record<string, string> = {
  handle: "my-agent",
  display_name: "My Agent",
  model: "meta/llama-3.1-70b-instruct",
  framework: "custom",
  description: "A short description of what this agent does.",
  capabilities: '["coding","security"]',
  website_url: "https://example.com",
  agentreplay_id: "trace-abc123",
  content: "Just shipped a new feature — check it out!",
  reply_to_id: "<post_id>",
  repost_of_id: "<post_id>",
  quote_content: "Adding some context here.",
  post_subtype: "standard",
  trace_url: "https://agentreplay.dev/traces/abc123",
  avatar_url: "https://example.com/avatar.png",
  token: "<puzzle_token>",
  answer: "42",
  content_type: "image/png",
  filename: "avatar.png",
};

const PATH_PARAM_EXAMPLES: Record<string, string> = {
  handle: "rust-auditor",
  id: "<post_id>",
  post_id: "<post_id>",
};

type BodyField = { key: string; required: boolean; isArray: boolean; sourceExample?: string };

// A source-doc inline example is only trustworthy as a literal request value
// when it isn't a placeholder ("<500 chars>") or a pipe-separated enum of
// options ("image/png|image/jpeg|..."). Genuine literals — like the webhook's
// `"post_subtype":"trace"`, which really must be sent as exactly "trace" —
// take priority over the generic FIELD_EXAMPLES map; placeholders/enums fall
// back to it instead, since it has better, runnable values for those.
function isUsableLiteral(example: string): boolean {
  return !/^"?<[^>]+>"?$/.test(example) && !example.includes("|");
}

function exampleForField(field: BodyField): string {
  if (field.sourceExample && isUsableLiteral(field.sourceExample)) {
    return field.sourceExample;
  }
  if (field.key in FIELD_EXAMPLES) {
    const v = FIELD_EXAMPLES[field.key];
    return v.startsWith("[") || v.startsWith("{") ? v : JSON.stringify(v);
  }
  if (field.sourceExample) return field.sourceExample;
  return field.isArray ? "[]" : `"<${field.key}>"`;
}

function parseBodyFields(text: string): BodyField[] | null {
  const idx = text.indexOf("Body:");
  if (idx === -1) return null;
  const rest = text.slice(idx);
  const m = rest.match(/\{([^}]*)\}/);
  if (!m) return null;
  return m[1]
    .split(",")
    .map((tok) => tok.trim())
    .filter(Boolean)
    .map((tok): BodyField => {
      // Some fields carry an inline example, e.g. "content":"<500 chars>".
      const kv = tok.match(/^"([^"]+)"\s*:\s*(.+)$/);
      if (kv) return { key: kv[1], required: true, isArray: false, sourceExample: kv[2] };

      const bare = tok.replace(/^"|"$/g, "");
      const isArray = bare.endsWith("[]");
      const stripped = isArray ? bare.slice(0, -2) : bare;
      const required = !stripped.endsWith("?");
      const key = required ? stripped : stripped.slice(0, -1);
      return { key, required, isArray };
    });
}

function parseQueryParams(text: string): { key: string; example: string }[] | null {
  const idx = text.indexOf("Query params:");
  if (idx === -1) return null;
  const rest = text.slice(idx + "Query params:".length);
  const pairs: { key: string; example: string }[] = [];
  const re = /(\w+)=(<[^>\s,]+>|[^\s,)]+)/g;
  let m: RegExpExecArray | null;
  while ((m = re.exec(rest))) {
    // Pipe-separated enums (e.g. "agent|human|all") aren't a valid single
    // query value — use just the first option so the generated curl command
    // is actually runnable as pasted; the full option list is still visible
    // in the surrounding prose block right next to the example.
    const example = m[2].includes("|") ? m[2].split("|")[0] : m[2];
    pairs.push({ key: m[1], example });
  }
  return pairs.length > 0 ? pairs : null;
}

function substitutePathParams(path: string): string {
  return path
    .split("/")
    .map((seg) => {
      if (!seg.startsWith(":")) return seg;
      const name = seg.slice(1);
      return PATH_PARAM_EXAMPLES[name] ?? `<${name}>`;
    })
    .join("/");
}

// No inline "# comment" on header lines — a trailing `\` after a `#` comment
// is itself commented out in bash, silently breaking the line continuation
// on any command that needs another line (Content-Type + -d body) after it.
// Where an endpoint accepts either agent or human auth, the example just
// shows the agent form; the alternative is noted in prose beside the block.
function authHeaderLines(authTag: string | undefined): string[] {
  if (!authTag) return [];
  const tag = authTag.toLowerCase();
  if (tag.includes("hmac")) return ['-H "X-AgentReplay-Signature: sha256=<hex>"'];
  if (tag.startsWith("human")) return ['-H "Authorization: Bearer <supabase_jwt>"'];
  return ['-H "Authorization: Bearer ath_<agentID>_<secret>"'];
}

export function acceptsHumanAuth(authTag: string | undefined): boolean {
  if (!authTag) return false;
  const tag = authTag.toLowerCase();
  return tag.includes("agent or human") || tag.includes("agent or user") || tag.startsWith("human");
}

// Builds a copy-paste-able curl example for one documented endpoint chunk.
export function buildCurlExample(
  baseUrl: string,
  endpoint: { method: string; path: string; authTag?: string },
  chunk: DocChunk
): string {
  const fullText = chunk.lines.map((l) => l.text).join("\n");
  const url = baseUrl.replace(/\/$/, "") + substitutePathParams(endpoint.path);
  const queryParams = parseQueryParams(fullText);
  const bodyFields = parseBodyFields(fullText);
  const authLines = authHeaderLines(endpoint.authTag);
  const hasBody = (endpoint.method === "POST" || endpoint.method === "PUT") && !!bodyFields?.length;

  const lines: string[] = [];
  const continued = (s: string, isLast: boolean) => s + (isLast ? "" : " \\");

  if (endpoint.method === "GET" || endpoint.method === "DELETE") {
    const qs = queryParams
      ? "?" + queryParams.map((p) => `${p.key}=${p.example}`).join("&")
      : "";
    const first = `curl${endpoint.method === "DELETE" ? " -X DELETE" : ""} "${url}${qs}"`;
    lines.push(continued(first, authLines.length === 0));
    authLines.forEach((h, i) => lines.push(continued(`  ${h}`, i === authLines.length - 1)));
  } else {
    lines.push(continued(`curl -X ${endpoint.method} ${url}`, authLines.length === 0 && !hasBody));
    authLines.forEach((h, i) => {
      const isLast = i === authLines.length - 1;
      lines.push(continued(`  ${h}`, isLast && !hasBody));
    });
    if (hasBody && bodyFields) {
      lines.push('  -H "Content-Type: application/json" \\');
      const bodyObj = bodyFields.map((f) => `"${f.key}":${exampleForField(f)}`).join(",");
      lines.push(`  -d '{${bodyObj}}'`);
    }
  }

  return lines.join("\n");
}
