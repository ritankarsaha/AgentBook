import { CodeBlock } from "./CodeBlock";
import type { DocLine } from "@/lib/llmsDoc";

// Renders a chunk's raw lines faithfully: consecutive indented lines become
// one code block, consecutive plain lines each become their own paragraph
// line (these are short, standalone doc sentences — "Body: ...", "Returns:
// ..." — not prose meant to be run together).
export function DocLinesView({ lines }: { lines: DocLine[] }) {
  const groups: { indented: boolean; lines: string[] }[] = [];
  for (const line of lines) {
    const last = groups[groups.length - 1];
    if (last && last.indented === line.indented) {
      last.lines.push(line.text);
    } else {
      groups.push({ indented: line.indented, lines: [line.text] });
    }
  }

  return (
    <div className="flex flex-col gap-2">
      {groups.map((g, i) =>
        g.indented ? (
          <CodeBlock key={i} code={g.lines.join("\n")} language="text" />
        ) : (
          <div key={i} className="flex flex-col gap-1">
            {g.lines.map((l, j) => (
              <p key={j} className="text-sm leading-relaxed text-text-secondary">
                {l}
              </p>
            ))}
          </div>
        )
      )}
    </div>
  );
}
