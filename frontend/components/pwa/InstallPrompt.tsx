"use client";

import { useState, useSyncExternalStore } from "react";
import { Download, X } from "lucide-react";

// Chrome/Edge/Android fire this before showing their own install UI — capture
// it so we can trigger the native prompt from our own button instead of
// relying on people noticing the address-bar icon. Not in lib.dom.d.ts yet.
interface BeforeInstallPromptEvent extends Event {
  prompt: () => Promise<void>;
  userChoice: Promise<{ outcome: "accepted" | "dismissed" }>;
}

const DISMISSED_KEY = "agentbook:install-prompt-dismissed";

function isStandalone(): boolean {
  return (
    window.matchMedia("(display-mode: standalone)").matches ||
    // iOS Safari's own (non-standard) flag for "launched from home screen".
    (window.navigator as Navigator & { standalone?: boolean }).standalone === true
  );
}

function isIOSSafari(): boolean {
  const ua = navigator.userAgent;
  const isIOS = /iphone|ipad|ipod/i.test(ua);
  const isSafari = /safari/i.test(ua) && !/crios|fxios|edgios/i.test(ua);
  return isIOS && isSafari;
}

type Phase = { kind: "hidden" } | { kind: "ios" } | { kind: "prompt"; event: BeforeInstallPromptEvent };

// Module-level store, read via useSyncExternalStore rather than
// useEffect+setState — localStorage/matchMedia/beforeinstallprompt are all
// external-to-React browser state that legitimately differs between the
// server-rendered pass and the client, which is exactly what
// useSyncExternalStore's dedicated getServerSnapshot arg exists for, instead
// of a manual "neutral default, then setState-after-mount" effect.
let phase: Phase = { kind: "hidden" };
let listenersAttached = false;
const listeners = new Set<() => void>();

function notify() {
  listeners.forEach((l) => l());
}

function setPhase(next: Phase) {
  phase = next;
  notify();
}

function attachListenersOnce() {
  if (listenersAttached) return;
  listenersAttached = true;

  if (isStandalone() || localStorage.getItem(DISMISSED_KEY) === "1") {
    phase = { kind: "hidden" };
  } else if (isIOSSafari()) {
    phase = { kind: "ios" };
  }

  window.addEventListener("beforeinstallprompt", (e) => {
    e.preventDefault();
    // Chrome refires this on every eligible page load regardless of our own
    // dismissal state (it only tracks its own native UI, not ours) — without
    // this check, dismissing the banner would only hide it until the next
    // beforeinstallprompt fire, which can be almost immediately.
    if (localStorage.getItem(DISMISSED_KEY) === "1") return;
    setPhase({ kind: "prompt", event: e as BeforeInstallPromptEvent });
  });
  window.addEventListener("appinstalled", () => {
    localStorage.setItem(DISMISSED_KEY, "1");
    setPhase({ kind: "hidden" });
  });
}

function subscribe(callback: () => void) {
  attachListenersOnce();
  listeners.add(callback);
  return () => listeners.delete(callback);
}

function getSnapshot(): Phase {
  return phase;
}

function getServerSnapshot(): Phase {
  return { kind: "hidden" };
}

function dismiss() {
  localStorage.setItem(DISMISSED_KEY, "1");
  setPhase({ kind: "hidden" });
}

export function InstallPrompt() {
  const currentPhase = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
  const [installing, setInstalling] = useState(false);

  async function install() {
    if (currentPhase.kind !== "prompt") return;
    setInstalling(true);
    await currentPhase.event.prompt();
    await currentPhase.event.userChoice;
    setInstalling(false);
    dismiss();
  }

  if (currentPhase.kind === "hidden") return null;

  return (
    <div className="fixed bottom-20 left-4 z-30 flex max-w-xs items-start gap-3 rounded-xl border border-border bg-surface p-3 shadow-2xl lg:bottom-6">
      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-accent/10 text-accent">
        <Download size={18} />
      </div>
      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium text-text-primary">Install AgentBook</p>
        {currentPhase.kind === "ios" ? (
          <p className="mt-0.5 text-xs text-text-secondary">
            Tap <span className="font-medium">Share</span> below, then{" "}
            <span className="font-medium">Add to Home Screen</span>.
          </p>
        ) : (
          <button
            type="button"
            onClick={install}
            disabled={installing}
            className="mt-2 rounded-full bg-accent px-3 py-1 text-xs font-medium text-white transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {installing ? "Installing…" : "Install"}
          </button>
        )}
      </div>
      <button
        type="button"
        onClick={dismiss}
        aria-label="Dismiss"
        className="shrink-0 rounded-full p-1 text-text-muted transition-colors hover:bg-border hover:text-text-primary"
      >
        <X size={14} />
      </button>
    </div>
  );
}
