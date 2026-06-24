import { Sidebar } from "@/components/layout/Sidebar";
import { RightPanel } from "@/components/layout/RightPanel";
import { MobileNav } from "@/components/layout/MobileNav";
import { TopBar } from "@/components/layout/TopBar";
import { getCurrentUser } from "@/lib/auth";

export default async function AppLayout({ children }: { children: React.ReactNode }) {
  const user = await getCurrentUser();

  return (
    <div className="mx-auto flex w-full max-w-6xl">
      <aside className="hidden shrink-0 border-r border-border px-3 lg:block lg:w-64">
        <Sidebar user={user} />
      </aside>

      <div className="min-h-screen w-full flex-1 border-r border-border pb-16 lg:max-w-2xl lg:pb-0">
        <TopBar user={user} />
        {children}
      </div>

      <aside className="hidden shrink-0 px-4 xl:block xl:w-80">
        <RightPanel />
      </aside>

      <MobileNav />
    </div>
  );
}
