import React from 'react';
import { motion, MotionValue } from 'motion/react';
import { ChevronDown, Github, ArrowUpRight } from 'lucide-react';

interface SectionHeroProps {
  isActive: boolean;
  slideTransition: { duration: number; ease: number[] };
  bgParallaxX: MotionValue<number>;
  bgParallaxY: MotionValue<number>;
  onMouseMove: (e: React.MouseEvent<HTMLDivElement>) => void;
  onMouseLeave: () => void;
  onJumpToArchitecture: () => void;
}

export const SectionHero: React.FC<SectionHeroProps> = ({
  isActive,
  slideTransition,
  bgParallaxX,
  bgParallaxY,
  onMouseMove,
  onMouseLeave,
  onJumpToArchitecture,
}) => {
  return (
    <section
      id="section-1"
      className="h-[100dvh] w-full shrink-0 flex flex-col justify-center items-center relative px-4 sm:px-6 md:px-8 pt-10 pb-6 sm:py-12 overflow-hidden select-none"
      onMouseMove={onMouseMove}
      onMouseLeave={onMouseLeave}
    >
      {/* Background Image with Cinematic Dark Vignette & Parallax */}
      <motion.div
        className="absolute inset-0 z-0 scale-105 pointer-events-none select-none"
        style={{ x: bgParallaxX, y: bgParallaxY }}
      >
        <img
          src="/hero-bg.jpg"
          alt="Watchtower Sentinel Infrastructure"
          loading="eager"
          decoding="async"
          fetchPriority="high"
          className="w-full h-full object-cover object-center opacity-40 sm:opacity-55"
        />
        <div className="absolute inset-0 bg-gradient-to-b from-[#08090C]/85 via-[#08090C]/55 to-[#08090C]" />
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-transparent via-[#08090C]/40 to-[#08090C]" />
      </motion.div>

      {/* Subtle Ambient Radial Glow */}
      <div className="pointer-events-none absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[320px] sm:w-[520px] h-[320px] sm:h-[520px] bg-emerald-950/20 blur-[150px] rounded-full z-0" />

      {/* Central Content (Pure, Minimal, Uncluttered) */}
      <motion.div
        className="w-full max-w-2xl flex flex-col items-center text-center space-y-4 sm:space-y-6 relative z-10 origin-center"
        animate={{
          opacity: isActive ? 1 : 0.35,
          scale: isActive ? 1 : 0.98,
        }}
        transition={slideTransition}
      >
        {/* Main Title & Single Crisp Value Proposition */}
        <div className="space-y-2.5 sm:space-y-4 px-2">
          <h1 className="text-4xl sm:text-6xl md:text-8xl font-bold tracking-tight text-white leading-none">
            Watchtower
          </h1>
          <p className="text-sm sm:text-lg md:text-xl text-zinc-300 font-normal max-w-md sm:max-w-lg mx-auto leading-relaxed">
            Lightweight production-oriented uptime &amp; health monitoring service written in Go.
          </p>
        </div>

        {/* Clean Minimal CTAs */}
        <div className="flex items-center gap-3 pt-2">
          <button
            onClick={onJumpToArchitecture}
            className="px-5 py-2.5 rounded-lg bg-zinc-100 hover:bg-white text-zinc-950 font-medium text-xs sm:text-sm transition-all shadow-md cursor-pointer flex items-center gap-2"
          >
            <span>Architecture</span>
            <ChevronDown className="w-3.5 h-3.5" />
          </button>
          <a
            href="https://github.com/mori-dv/watchtower"
            target="_blank"
            rel="noopener noreferrer"
            className="px-5 py-2.5 rounded-lg bg-zinc-900/80 hover:bg-zinc-800 text-zinc-300 hover:text-white border border-zinc-800 text-xs sm:text-sm font-medium transition-all flex items-center gap-2 group"
          >
            <Github className="w-4 h-4 text-zinc-400 group-hover:text-white transition-colors" />
            <span>GitHub</span>
            <ArrowUpRight className="w-3.5 h-3.5 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 transition-transform text-zinc-400 group-hover:text-zinc-200" />
          </a>
        </div>
      </motion.div>

      {/* Minimal Bottom Indicator without text */}
      <button
        onClick={onJumpToArchitecture}
        className="absolute bottom-3 sm:bottom-6 left-1/2 -translate-x-1/2 flex items-center justify-center text-zinc-500 hover:text-zinc-300 transition-colors cursor-pointer p-2 focus:outline-none z-10"
        aria-label="Next Section"
      >
        <ChevronDown className="w-4 h-4 animate-bounce" />
      </button>
    </section>
  );
};
