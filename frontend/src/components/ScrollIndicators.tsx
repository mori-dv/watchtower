import React from 'react';

interface ScrollIndicatorsProps {
  activeSection: number;
  onSelectSection: (index: number) => void;
}

export const ScrollIndicators: React.FC<ScrollIndicatorsProps> = ({
  activeSection,
  onSelectSection,
}) => {
  return (
    <div className="fixed right-2.5 sm:right-6 md:right-8 top-1/2 -translate-y-1/2 z-50 flex flex-col gap-2.5 sm:gap-3.5 pointer-events-auto">
      {[0, 1, 2].map((idx) => {
        const isActive = activeSection === idx;
        return (
          <button
            key={idx}
            onClick={() => onSelectSection(idx)}
            className="cursor-pointer p-1.5 focus:outline-none focus-visible:ring-1 focus-visible:ring-emerald-400 rounded-full transition-transform"
            aria-label={`Jump to Section ${idx + 1}`}
          >
            <span
              className={`block w-1.5 sm:w-2 h-1.5 sm:h-2 rounded-full transition-all duration-300 ${
                isActive
                  ? 'bg-emerald-400 scale-125 shadow-[0_0_10px_rgba(52,211,153,0.8)]'
                  : 'bg-zinc-600/60 hover:bg-zinc-400 hover:scale-110'
              }`}
            />
          </button>
        );
      })}
    </div>
  );
};
