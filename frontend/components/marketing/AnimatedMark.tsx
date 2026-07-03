import Image from "next/image";

export function AnimatedMark() {
  return (
    <div className="relative flex h-[520px] w-[520px] items-center justify-center">
      {/* Glow */}
      <div className="animate-glow-pulse absolute h-72 w-72 rounded-full bg-accent/30 blur-[100px]" />

      {/* Rotating rings */}
      <svg viewBox="0 0 520 520" className="animate-spin-slow absolute h-full w-full text-accent/50" fill="none">
        <circle cx="260" cy="260" r="248" stroke="currentColor" strokeWidth="1.5" strokeDasharray="2 14" />
      </svg>
      <svg
        viewBox="0 0 420 420"
        className="animate-spin-slow-reverse absolute h-[420px] w-[420px] text-accent-human/40"
        fill="none"
      >
        <circle cx="210" cy="210" r="200" stroke="currentColor" strokeWidth="1.5" strokeDasharray="1 10" />
      </svg>

      {/* Mark */}
      <div className="animate-float-slow relative">
        <Image
          src="/logo-mark.png"
          alt=""
          width={240}
          height={240}
          priority
          className="opacity-95 [filter:drop-shadow(0_0_40px_rgba(99,102,241,0.45))]"
        />
      </div>
    </div>
  );
}
