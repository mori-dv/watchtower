import React from 'react';
import { motion } from 'motion/react';
import { Github, BookOpen, MessageSquare, ArrowUpRight } from 'lucide-react';

interface SectionObservabilityProps {
  isActive: boolean;
  slideTransition: { duration: number; ease: number[] };
}

export const SectionObservability: React.FC<SectionObservabilityProps> = ({
  isActive,
  slideTransition,
}) => {
  return (
    <section 
      id="section-3"
      className="h-[100dvh] w-full shrink-0 flex flex-col justify-between items-center relative px-3 sm:px-6 md:px-8 pt-10 pb-3 sm:py-10 select-none overflow-hidden"
    >
      {/* Spacer */}
      <div className="h-1 sm:h-4" />

      {/* Center Content */}
      <motion.div 
        className="w-full max-w-5xl space-y-2.5 sm:space-y-6 md:space-y-8 my-auto origin-center"
        animate={{ 
          opacity: isActive ? 1 : 0.35, 
          scale: isActive ? 1 : 0.98 
        }}
        transition={slideTransition}
      >
        <div className="text-center space-y-1 sm:space-y-1.5 px-2">
          <h2 className="text-lg sm:text-2xl md:text-4xl font-semibold tracking-tight text-white leading-tight">
            Developer Resources &amp; Ecosystem
          </h2>
          <p className="text-[9px] sm:text-xs md:text-sm text-zinc-400 max-w-sm sm:max-w-md mx-auto leading-normal font-sans">
            Explore the open-source Go backend codebase, inspect telemetry dashboards, and contribute.
          </p>
        </div>

        {/* Link Cards (Compact rows on mobile, 3-Col on Desktop) */}
        <div className="flex flex-col sm:grid sm:grid-cols-3 gap-2 sm:gap-3.5 md:gap-5 font-sans">
          
          {/* Card 1: GitHub Repository */}
          <a
            href="https://github.com/mori-dv/watchtower"
            target="_blank"
            rel="noopener noreferrer"
            className="rounded-lg sm:rounded-xl p-2.5 sm:p-4 md:p-6 bg-[#0E1015]/80 border border-zinc-800/80 hover:border-zinc-700 hover:bg-[#12141C] transition-all flex flex-col justify-between space-y-2 group shadow-lg"
          >
            <div className="space-y-1 sm:space-y-1.5">
              <div className="flex items-center justify-between text-zinc-400">
                <Github className="w-3.5 h-3.5 text-zinc-300 group-hover:text-white transition-colors" />
                <ArrowUpRight className="w-3 h-3 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 group-hover:text-zinc-200 transition-all" />
              </div>
              <div className="font-semibold text-zinc-200 text-xs sm:text-sm group-hover:text-white transition-colors">
                GitHub Repository
              </div>
              <p className="text-[9px] sm:text-xs text-zinc-400 leading-relaxed font-normal">
                Explore Go source code, scheduler worker pools, Docker configs, and modular packages.
              </p>
            </div>

            <div className="text-[8px] sm:text-[10px] font-mono text-zinc-500 group-hover:text-zinc-400 pt-1.5 border-t border-zinc-800/60 transition-colors">
              mori-dv/watchtower // open-source
            </div>
          </a>

          {/* Card 2: Architecture & Metrics */}
          <a
            href="https://github.com/mori-dv/watchtower#readme"
            target="_blank"
            rel="noopener noreferrer"
            className="rounded-lg sm:rounded-xl p-2.5 sm:p-4 md:p-6 bg-[#0E1015]/80 border border-zinc-800/80 hover:border-zinc-700 hover:bg-[#12141C] transition-all flex flex-col justify-between space-y-2 group shadow-lg"
          >
            <div className="space-y-1 sm:space-y-1.5">
              <div className="flex items-center justify-between text-zinc-400">
                <BookOpen className="w-3.5 h-3.5 text-zinc-300 group-hover:text-white transition-colors" />
                <ArrowUpRight className="w-3 h-3 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 group-hover:text-zinc-200 transition-all" />
              </div>
              <div className="font-semibold text-zinc-200 text-xs sm:text-sm group-hover:text-white transition-colors">
                Architecture &amp; Metrics
              </div>
              <p className="text-[9px] sm:text-xs text-zinc-400 leading-relaxed font-normal">
                Review Grafana dashboards, Prometheus PromQL queries, and Loki log analytics.
              </p>
            </div>

            <div className="text-[8px] sm:text-[10px] font-mono text-zinc-500 group-hover:text-zinc-400 pt-1.5 border-t border-zinc-800/60 transition-colors">
              Prometheus // Grafana // Loki
            </div>
          </a>

          {/* Card 3: Issues & Contributing */}
          <a
            href="https://github.com/mori-dv/watchtower/issues"
            target="_blank"
            rel="noopener noreferrer"
            className="rounded-lg sm:rounded-xl p-2.5 sm:p-4 md:p-6 bg-[#0E1015]/80 border border-zinc-800/80 hover:border-zinc-700 hover:bg-[#12141C] transition-all flex flex-col justify-between space-y-2 group shadow-lg"
          >
            <div className="space-y-1 sm:space-y-1.5">
              <div className="flex items-center justify-between text-zinc-400">
                <MessageSquare className="w-3.5 h-3.5 text-zinc-300 group-hover:text-white transition-colors" />
                <ArrowUpRight className="w-3 h-3 group-hover:translate-x-0.5 group-hover:-translate-y-0.5 group-hover:text-zinc-200 transition-all" />
              </div>
              <div className="font-semibold text-zinc-200 text-xs sm:text-sm group-hover:text-white transition-colors">
                Issues &amp; Contributing
              </div>
              <p className="text-[9px] sm:text-xs text-zinc-400 leading-relaxed font-normal">
                Submit bug reports, propose RFC improvements, or collaborate on probe engines.
              </p>
            </div>

            <div className="text-[8px] sm:text-[10px] font-mono text-zinc-500 group-hover:text-zinc-400 pt-1.5 border-t border-zinc-800/60 transition-colors">
              GitHub Issues // MIT License
            </div>
          </a>

        </div>

      </motion.div>

      {/* Minimal Footer */}
      <footer className="w-full max-w-5xl flex flex-row items-center justify-between text-[8px] sm:text-[10px] md:text-[11px] font-mono text-zinc-400 pt-2 sm:pt-4 border-t border-zinc-800/60 gap-1.5 pb-1">
        <div className="flex items-center gap-1.5">
          <span>© 2026 mori</span>
          <span className="hidden xs:inline">•</span>
          <span className="hidden xs:inline text-zinc-500">Watchtower Sentinel</span>
        </div>

        <div className="flex items-center gap-1.5 sm:gap-2 text-zinc-500 text-[8px] sm:text-[10px]">
          <span>Go 1.25</span>
          <span>•</span>
          <span>Redis 8.6</span>
          <span className="hidden sm:inline">•</span>
          <span>MIT License</span>
        </div>
      </footer>

    </section>
  );
};
