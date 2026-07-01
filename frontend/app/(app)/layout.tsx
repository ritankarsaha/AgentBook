import { Sidebar } from "@/components/layout/Sidebar";
import { RightPanel } from "@/components/layout/RightPanel";
import { MobileNav } from "@/components/layout/MobileNav";
import { TopBar } from "@/components/layout/TopBar";
import { getCurrentUser } from "@/lib/auth";

export default async function AppLayout({ children }: { children: React.ReactNode }) {
  const user = await getCurrentUser();

  return (
    <div className="mx-auto flex w-full max-w-6xl">
      {/* Left sidebar — sticky, does not scroll with the feed */}
      <aside className="hidden lg:flex lg:w-64 shrink-0 flex-col sticky top-0 h-screen border-r border-border bg-bg self-start">
        <Sidebar user={user} />
      </aside>

      {/* Center feed — the only column that scrolls */}
      <div className="min-h-screen w-full flex-1 border-r border-border pb-16 lg:max-w-2xl lg:pb-0">
        <TopBar user={user} />
        {children}
      </div>

      {/* Right panel — sticky, does not scroll with the feed */}
      <aside className="hidden xl:flex xl:w-80 shrink-0 flex-col sticky top-0 h-screen bg-bg self-start">
        <RightPanel />
      </aside>

      <MobileNav userHandle={user?.handle ?? null} />
    </div>
  );
}
