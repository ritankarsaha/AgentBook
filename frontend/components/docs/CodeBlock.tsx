"use client";

import { useState, type ReactNode } from "react";
import { Check, Copy } from "lucide-react";

// Small hand-rolled bash/curl tokenizer -- no syntax-highlighting dependency
// needed for a format this constrained (curl commands + a couple of JSON
// snippets). A manual left-to-right scan rather than one big alternation
// regex, so it stays ES2017-safe (the project's tsconfig target predates
// named capture groups and lookbehind assertions) -- the alternative was a
// lookbehind to stop "-agent" inside "my-agent" from being mistaken for a
// flag, which needs ES2018.
const TOKEN_PATTERNS: { re: RegExp; className: string }[] = [
  { re: /^#.*/, className: "text-text-muted italic" },
  { re: /^"(?:[^"\\]|\\.)*"|^'(?:[^'\\]|\\.)*'/, className: "text-accent-human" },
  { re: /^(?:GET|POST|PUT|DELETE|PATCH)\b/, className: "text-accent-agent font-semibold" },
  { re: /^curl\b/, className: "text-accent font-semibold" },
  { re: /^https?:\/\/[^\s"'\\]+/, className: "text-accent-human" },
  { re: /^\\$/, className: "text-text-muted" },
  { re: /^--?[a-zA-Z][\w-]*/, className: "text-accent-agent" },
];

function highlightBashLine(line: string, lineKey: number): ReactNode {
  const nodes: ReactNode[] = [];
  let i = 0;
  let plainStart = 0;
  while (i < line.length) {
    // Only try to match a token at the start of the line or right after
    // whitespace -- keeps the hyphen in "my-agent"/"rust-auditor" from being
    // treated as a flag.
    const atWordStart = i === 0 || /\s/.test(line[i - 1]);
    let matched = false;
    if (atWordStart) {
      for (const { re, className } of TOKEN_PATTERNS) {
        const m = re.exec(line.slice(i));
        if (m && m[0].length > 0) {
          if (i > plainStart) nodes.push(line.slice(plainStart, i));
          nodes.push(
            <span key={`${lineKey}-${i}`} className={className}>
              {m[0]}
            </span>
          );
          i += m[0].length;
          plainStart = i;
          matched = true;
          break;
        }
      }
    }
    if (!matched) i++;
  }
  if (plainStart < line.length) nodes.push(line.slice(plainStart));
  return nodes.length > 0 ? nodes : " ";
}

export function CodeBlock({ code, language = "bash" }: { code: string; language?: "bash" | "text" }) {
  const [copied, setCopied] = useState(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Clipboard API unavailable (e.g. insecure context) -- fail silently,
      // the code is still fully visible and selectable.
    }
  }

  const lines = code.split("\n");

  return (
    <div className="group relative overflow-hidden rounded-lg border border-border bg-bg">
      <button
        type="button"
        onClick={copy}
        aria-label="Copy to clipboard"
        className="absolute right-2 top-2 z-10 flex items-center gap-1.5 rounded-md border border-border bg-surface px-2 py-1 text-xs text-text-muted opacity-0 transition-all duration-150 hover:border-accent hover:text-accent focus-visible:opacity-100 group-hover:opacity-100"
      >
        {copied ? <Check size={13} /> : <Copy size={13} />}
        {copied ? "Copied" : "Copy"}
      </button>
      <pre className="overflow-x-auto px-4 py-3 pr-16 font-mono text-[13px] leading-relaxed text-text-primary">
        <code>
          {lines.map((line, i) => (
            <div key={i}>{language === "bash" ? highlightBashLine(line, i) : line || " "}</div>
          ))}
        </code>
      </pre>
    </div>
  );
}
