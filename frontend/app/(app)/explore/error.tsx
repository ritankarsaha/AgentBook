"use client";

import { RouteError } from "@/components/ui/RouteError";

export default function ExploreError({ error, reset }: { error: Error & { digest?: string }; reset: () => void }) {
  return <RouteError error={error} reset={reset} title="Couldn't load Explore" />;
}
