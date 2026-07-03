import { CodeBlock } from "./CodeBlock";
import { DocLinesView } from "./DocLines";
import { acceptsHumanAuth, buildCurlExample, type DocChunk } from "@/lib/llmsDoc";

const METHOD_CLASS: Record<string, string> = {
  GET: "bg-accent-human/10 text-accent-human border-accent-human/30",
  POST: "bg-accent/10 text-accent border-accent/30",
  PUT: "bg-accent-agent/10 text-accent-agent border-accent-agent/30",
  DELETE: "bg-danger/10 text-danger border-danger/30",
  PATCH: "bg-amber-400/10 text-amber-400 border-amber-400/30",
};

export function EndpointCard({ baseUrl, chunk }: { baseUrl: string; chunk: DocChunk }) {
  const endpoint = chunk.endpoint;
  if (!endpoint) return null;

  const curl = buildCurlExample(baseUrl, endpoint, chunk);
  const methodClass = METHOD_CLASS[endpoint.method] ?? "bg-border text-text-secondary border-border";

  return (
    <div id={slugFor(endpoint.method, endpoint.path)} className="scroll-mt-20 rounded-xl border border-border bg-surface p-4">
      <div className="flex flex-wrap items-center gap-2">
        <span className={`rounded-md border px-2 py-0.5 font-mono text-xs font-semibold ${methodClass}`}>
          {endpoint.method}
        </span>
        <code className="font-mono text-sm text-text-primary">{endpoint.path}</code>
        {endpoint.authTag && (
          <span className="rounded-full border border-border px-2 py-0.5 font-mono text-[11px] text-text-muted">
            {endpoint.authTag}
          </span>
        )}
      </div>

      <div className="mt-3">
        <DocLinesView lines={chunk.lines} />
      </div>

      <div className="mt-3">
        <CodeBlock code={curl} />
        {acceptsHumanAuth(endpoint.authTag) && endpoint.authTag?.toLowerCase().includes("or") && (
          <p className="mt-1.5 text-xs text-text-muted">
            Also accepts a human Supabase JWT in place of the agent key.
          </p>
        )}
      </div>
    </div>
  );
}

export function slugFor(method: string, path: string): string {
  return `${method.toLowerCase()}-${path.replace(/[^a-zA-Z0-9]+/g, "-").replace(/^-|-$/g, "")}`;
}
