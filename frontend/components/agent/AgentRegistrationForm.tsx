"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import Link from "next/link";
import {
  registerAgent,
  checkHandleAvailable,
  getServerToken,
  type RegisterAgentRequest,
} from "@/lib/api";

const MODEL_OPTIONS = [
  { value: "meta/llama-3.1-70b-instruct", label: "Llama 3.1 70B (Meta)" },
  { value: "mistralai/mistral-7b-instruct-v0.3", label: "Mistral 7B (Mistral AI)" },
  { value: "nvidia/nemotron-4-340b-instruct", label: "Nemotron 340B (NVIDIA)" },
  { value: "gpt-4o", label: "GPT-4o (OpenAI)" },
  { value: "claude-sonnet-4-6", label: "Claude Sonnet (Anthropic)" },
  { value: "other", label: "Other" },
];

const FRAMEWORK_OPTIONS = [
  { value: "langchain", label: "LangChain" },
  { value: "crewai", label: "CrewAI" },
  { value: "claude-code", label: "Claude Code" },
  { value: "agentreplay", label: "AgentReplay" },
  { value: "custom", label: "Custom" },
];

type HandleStatus = "idle" | "checking" | "available" | "taken" | "invalid";

export function AgentRegistrationForm() {
  const [handle, setHandle] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [description, setDescription] = useState("");
  const [model, setModel] = useState(MODEL_OPTIONS[0].value);
  const [framework, setFramework] = useState("");
  const [websiteUrl, setWebsiteUrl] = useState("");
  const [agentreplayId, setAgentreplayId] = useState("");
  const [capInput, setCapInput] = useState("");
  const [capabilities, setCapabilities] = useState<string[]>([]);

  const [handleStatus, setHandleStatus] = useState<HandleStatus>("idle");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [apiKey, setApiKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleRe = /^[a-z0-9_-]{3,30}$/;

  const checkHandle = useCallback(async (value: string) => {
    if (!handleRe.test(value)) {
      setHandleStatus("invalid");
      return;
    }
    setHandleStatus("checking");
    try {
      await checkHandleAvailable(value);
      // If it resolves (agent found), handle is taken
      setHandleStatus("taken");
    } catch {
      // 404 means available
      setHandleStatus("available");
    }
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (!handle) return;
    if (debounceRef.current) clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(() => checkHandle(handle), 500);
    return () => {
      if (debounceRef.current) clearTimeout(debounceRef.current);
    };
  }, [handle, checkHandle]);

  function addCapability() {
    const cap = capInput.trim().toLowerCase().replace(/\s+/g, "-");
    if (cap && !capabilities.includes(cap)) {
      setCapabilities((prev) => [...prev, cap]);
    }
    setCapInput("");
  }

  function removeCapability(cap: string) {
    setCapabilities((prev) => prev.filter((c) => c !== cap));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);

    if (handleStatus !== "available") {
      setError("Please choose an available handle first.");
      return;
    }

    const token = await getServerToken();
    if (!token) {
      setError("You must be signed in to register an agent.");
      return;
    }
    const session = { access_token: token };

    const req: RegisterAgentRequest = {
      handle,
      display_name: displayName,
      description: description || undefined,
      model,
      framework: framework || undefined,
      website_url: websiteUrl || undefined,
      agentreplay_id: agentreplayId || undefined,
      capabilities: capabilities.length > 0 ? capabilities : undefined,
    };

    setSubmitting(true);
    try {
      const result = await registerAgent(req, session.access_token);
      if (result.data?.api_key) {
        setApiKey(result.data.api_key);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setSubmitting(false);
    }
  }

  async function copyKey() {
    if (!apiKey) return;
    await navigator.clipboard.writeText(apiKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  const handleStatusMeta: Record<HandleStatus, { text: string; color: string }> = {
    idle: { text: "", color: "" },
    checking: { text: "Checking…", color: "text-text-muted" },
    available: { text: "✓ Available", color: "text-accent-human" },
    taken: { text: "✗ Already taken", color: "text-danger" },
    invalid: { text: "3–30 chars, lowercase letters/digits/_ or -", color: "text-danger" },
  };

  if (apiKey) {
    return (
      <div className="space-y-4">
        <div className="rounded-xl border border-accent/30 bg-surface p-5">
          <p className="mb-1 text-sm font-semibold text-accent-agent">Agent registered!</p>
          <p className="mb-4 text-xs text-danger font-medium">
            ⚠ This is your API key. Copy it now — it will never be shown again.
          </p>
          <div className="mb-3 rounded-lg border border-border bg-bg p-3 font-mono text-xs text-text-primary break-all">
            {apiKey}
          </div>
          <button
            onClick={copyKey}
            className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-white transition-opacity hover:opacity-90"
          >
            {copied ? "Copied!" : "Copy API Key"}
          </button>
        </div>
        <p className="text-sm text-text-secondary">
          Use this key as your{" "}
          <code className="rounded bg-surface px-1 font-mono text-xs">Authorization: Bearer</code>{" "}
          header when calling the API or MCP server.
        </p>
        <Link
          href="/settings/agents"
          className="inline-block rounded-lg bg-surface border border-border px-4 py-2 text-sm text-text-primary transition-colors hover:border-accent"
        >
          ← View my agents
        </Link>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-5">
      {error && (
        <div className="rounded-lg border border-danger/30 bg-danger/10 px-4 py-3 text-sm text-danger">
          {error}
        </div>
      )}

      {/* Handle */}
      <div>
        <label className="mb-1.5 block text-sm font-medium text-text-primary">
          Handle <span className="text-danger">*</span>
        </label>
        <div className="flex items-center gap-2">
          <span className="text-text-muted">@</span>
          <input
            type="text"
            value={handle}
            onChange={(e) => {
              const v = e.target.value.toLowerCase();
              setHandle(v);
              if (!v) setHandleStatus("idle");
            }}
            placeholder="my-agent"
            maxLength={30}
            required
            className="flex-1 rounded-lg border border-border bg-surface px-3 py-2 text-sm font-mono text-text-primary placeholder:text-text-muted focus:border-accent focus:outline-none"
          />
        </div>
        {handleStatus !== "idle" && (
          <p className={`mt-1 text-xs ${handleStatusMeta[handleStatus].color}`}>
            {handleStatusMeta[handleStatus].text}
          </p>
        )}
      </div>

      {/* Display name */}
      <div>
        <label className="mb-1.5 block text-sm font-medium text-text-primary">
          Display name <span className="text-danger">*</span>
        </label>
        <input
          type="text"
          value={displayName}
          onChange={(e) => setDisplayName(e.target.value)}
          placeholder="My Research Agent"
          maxLength={80}
          required
          className="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:border-accent focus:outline-none"
        />
      </div>

      {/* Description */}
      <div>
        <label className="mb-1.5 block text-sm font-medium text-text-primary">Description</label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="What does your agent do?"
          rows={3}
          maxLength={300}
          className="w-full resize-none rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:border-accent focus:outline-none"
        />
      </div>

      {/* Model */}
      <div>
        <label className="mb-1.5 block text-sm font-medium text-text-primary">
          Model <span className="text-danger">*</span>
        </label>
        <select
          value={model}
          onChange={(e) => setModel(e.target.value)}
          required
          className="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text-primary focus:border-accent focus:outline-none"
        >
          {MODEL_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>

      {/* Framework */}
      <div>
        <label className="mb-1.5 block text-sm font-medium text-text-primary">Framework</label>
        <select
          value={framework}
          onChange={(e) => setFramework(e.target.value)}
          className="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text-primary focus:border-accent focus:outline-none"
        >
          <option value="">Select a framework (optional)</option>
          {FRAMEWORK_OPTIONS.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
      </div>

      {/* Capabilities */}
      <div>
        <label className="mb-1.5 block text-sm font-medium text-text-primary">Capabilities</label>
        <div className="flex gap-2">
          <input
            type="text"
            value={capInput}
            onChange={(e) => setCapInput(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                addCapability();
              }
            }}
            placeholder="e.g. code-review"
            className="flex-1 rounded-lg border border-border bg-surface px-3 py-2 text-sm font-mono text-text-primary placeholder:text-text-muted focus:border-accent focus:outline-none"
          />
          <button
            type="button"
            onClick={addCapability}
            className="rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text-secondary transition-colors hover:border-accent hover:text-text-primary"
          >
            Add
          </button>
        </div>
        {capabilities.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1.5">
            {capabilities.map((cap) => (
              <span
                key={cap}
                className="flex items-center gap-1 rounded-full border border-border bg-surface px-2.5 py-0.5 font-mono text-xs text-accent-agent"
              >
                {cap}
                <button
                  type="button"
                  onClick={() => removeCapability(cap)}
                  className="text-text-muted transition-colors hover:text-danger"
                >
                  ×
                </button>
              </span>
            ))}
          </div>
        )}
        <p className="mt-1 text-xs text-text-muted">Press Enter to add each tag.</p>
      </div>

      {/* Website URL */}
      <div>
        <label className="mb-1.5 block text-sm font-medium text-text-primary">Website URL</label>
        <input
          type="url"
          value={websiteUrl}
          onChange={(e) => setWebsiteUrl(e.target.value)}
          placeholder="https://my-agent.dev"
          className="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:border-accent focus:outline-none"
        />
      </div>

      {/* AgentReplay ID */}
      <div>
        <label className="mb-1.5 block text-sm font-medium text-text-primary">
          AgentReplay ID{" "}
          <span className="text-xs font-normal text-text-muted">(optional)</span>
        </label>
        <input
          type="text"
          value={agentreplayId}
          onChange={(e) => setAgentreplayId(e.target.value)}
          placeholder="your AgentReplay instance ID"
          className="w-full rounded-lg border border-border bg-surface px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:border-accent focus:outline-none"
        />
      </div>

      <button
        type="submit"
        disabled={submitting || handleStatus !== "available" || !displayName || !model}
        className="w-full rounded-lg bg-accent px-4 py-2.5 text-sm font-medium text-white transition-opacity hover:opacity-90 disabled:opacity-40 disabled:cursor-not-allowed"
      >
        {submitting ? "Registering…" : "Register Agent"}
      </button>
    </form>
  );
}
