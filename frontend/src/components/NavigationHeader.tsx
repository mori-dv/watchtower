import React from 'react';
import { Github } from 'lucide-react';

export const NavigationHeader: React.FC = () => {
  return (
    <nav className="fixed top-0 left-0 right-0 z-50 px-4 sm:px-8 md:px-10 py-3 sm:py-5 flex items-center justify-between pointer-events-none bg-gradient-to-b from-[#08090C]/90 via-[#08090C]/50 to-transparent backdrop-blur-[2px]">
      <div className="flex items-center gap-2 pointer-events-auto">
        <span className="text-xs sm:text-sm font-mono font-medium tracking-wider text-zinc-400">
          INFRA // <strong className="text-zinc-200 font-semibold">Watchtower</strong>
        </span>
      </div>

      <a
        href="https://github.com/mori-dv/watchtower"
        target="_blank"
        rel="noopener noreferrer"
        className="pointer-events-auto p-1.5 text-zinc-400 hover:text-white transition-colors"
        title="GitHub Repository"
        aria-label="GitHub Repository"
      >
        <Github className="w-4 h-4" />
      </a>
    </nav>
  );
};
