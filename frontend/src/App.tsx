import React, { useState, useRef, useEffect, useCallback } from 'react';
import { motion, useMotionValue, useSpring, useTransform } from 'motion/react';
import { NavigationHeader } from './components/NavigationHeader';
import { ScrollIndicators } from './components/ScrollIndicators';
import { SectionHero } from './components/SectionHero';
import { SectionMechanics } from './components/SectionMechanics';
import { SectionObservability } from './components/SectionObservability';

export default function App() {
  const [activeSection, setActiveSection] = useState(0);

  const containerRef = useRef<HTMLDivElement>(null);
  const isLockedRef = useRef(false);
  const lockTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const touchStartY = useRef<number | null>(null);

  // Snappy, high-impact cubic-bezier easing curve matching Gatekeeper
  const slideTransition = {
    duration: 0.42,
    ease: [0.16, 1, 0.3, 1],
  };

  const changeSection = useCallback((direction: 1 | -1) => {
    if (isLockedRef.current) return;

    setActiveSection((prev) => {
      const next = Math.max(0, Math.min(2, prev + direction));
      if (next !== prev) {
        isLockedRef.current = true;
        if (lockTimeoutRef.current) clearTimeout(lockTimeoutRef.current);
        lockTimeoutRef.current = setTimeout(() => {
          isLockedRef.current = false;
        }, 500);
        return next;
      }
      return prev;
    });
  }, []);

  const jumpToSection = useCallback((targetIndex: number) => {
    if (isLockedRef.current || targetIndex === activeSection) return;
    if (targetIndex < 0 || targetIndex > 2) return;

    isLockedRef.current = true;
    setActiveSection(targetIndex);
    if (lockTimeoutRef.current) clearTimeout(lockTimeoutRef.current);
    lockTimeoutRef.current = setTimeout(() => {
      isLockedRef.current = false;
    }, 500);
  }, [activeSection]);

  // Clean up lock timeout on unmount
  useEffect(() => {
    return () => {
      if (lockTimeoutRef.current) clearTimeout(lockTimeoutRef.current);
    };
  }, []);

  // Wheel Event & Touch Movement normalization to prevent rubber-banding
  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;

    let wheelAccumulator = 0;
    let wheelResetTimer: NodeJS.Timeout | null = null;

    const onWheel = (e: WheelEvent) => {
      e.preventDefault();
      if (isLockedRef.current) return;

      wheelAccumulator += e.deltaY;
      if (wheelResetTimer) clearTimeout(wheelResetTimer);
      wheelResetTimer = setTimeout(() => {
        wheelAccumulator = 0;
      }, 100);

      // Normalized threshold for discrete notched wheels & trackpad inertia
      if (wheelAccumulator > 25) {
        wheelAccumulator = 0;
        changeSection(1);
      } else if (wheelAccumulator < -25) {
        wheelAccumulator = 0;
        changeSection(-1);
      }
    };

    const onTouchMove = (e: TouchEvent) => {
      if (e.cancelable) e.preventDefault();
    };

    el.addEventListener('wheel', onWheel, { passive: false });
    el.addEventListener('touchmove', onTouchMove, { passive: false });

    return () => {
      el.removeEventListener('wheel', onWheel);
      el.removeEventListener('touchmove', onTouchMove);
      if (wheelResetTimer) clearTimeout(wheelResetTimer);
    };
  }, [changeSection]);

  // Touch swipe detection for mobile
  const handleTouchStart = (e: React.TouchEvent<HTMLDivElement>) => {
    if (e.touches.length > 0) {
      touchStartY.current = e.touches[0].clientY;
    }
  };

  const handleTouchEnd = (e: React.TouchEvent<HTMLDivElement>) => {
    if (touchStartY.current === null || isLockedRef.current) return;
    const touchEndY = e.changedTouches[0]?.clientY;
    if (touchEndY === undefined) return;

    const diff = touchStartY.current - touchEndY;
    touchStartY.current = null;

    if (diff > 45) {
      changeSection(1);
    } else if (diff < -45) {
      changeSection(-1);
    }
  };

  // Keyboard Navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (isLockedRef.current) return;

      if (e.key === 'ArrowDown' || e.key === 'PageDown' || e.key === ' ' || e.key === 'j') {
        e.preventDefault();
        changeSection(1);
      } else if (e.key === 'ArrowUp' || e.key === 'PageUp' || e.key === 'k') {
        e.preventDefault();
        changeSection(-1);
      } else if (e.key === 'Home') {
        e.preventDefault();
        jumpToSection(0);
      } else if (e.key === 'End') {
        e.preventDefault();
        jumpToSection(2);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [changeSection, jumpToSection]);

  // Interactive parallax spring for hero backdrop
  const mouseX = useMotionValue(0);
  const mouseY = useMotionValue(0);

  const springConfig = { stiffness: 100, damping: 20, mass: 0.5 };
  const bgParallaxX = useSpring(useTransform(mouseX, [-0.5, 0.5], [-16, 16]), springConfig);
  const bgParallaxY = useSpring(useTransform(mouseY, [-0.5, 0.5], [-16, 16]), springConfig);

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    const rect = e.currentTarget.getBoundingClientRect();
    const x = (e.clientX - rect.left) / rect.width - 0.5;
    const y = (e.clientY - rect.top) / rect.height - 0.5;
    mouseX.set(x);
    mouseY.set(y);
  };

  const handleMouseLeave = () => {
    mouseX.set(0);
    mouseY.set(0);
  };

  return (
    <div
      ref={containerRef}
      id="root-container"
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
      className="h-[100dvh] w-full overflow-hidden bg-[#08090C] text-zinc-100 antialiased selection:bg-emerald-950 selection:text-emerald-300 relative font-sans select-none overscroll-none touch-none"
    >
      {/* 1. Minimal Fixed Top Bar */}
      <NavigationHeader />

      {/* 2. Side Indicator Dots (Minimal glowing dots, no text labels) */}
      <ScrollIndicators
        activeSection={activeSection}
        onSelectSection={jumpToSection}
      />

      {/* 3. Hardware-Accelerated Snappy Slide Container */}
      <motion.div
        className="w-full h-full will-change-transform"
        animate={{ y: `-${activeSection * 100}%` }}
        transition={slideTransition}
        style={{ transform: `translate3d(0, -${activeSection * 100}%, 0)` }}
      >
        {/* SECTION 1: HERO & SENTINEL OVERVIEW */}
        <SectionHero
          isActive={activeSection === 0}
          slideTransition={slideTransition}
          bgParallaxX={bgParallaxX}
          bgParallaxY={bgParallaxY}
          onMouseMove={handleMouseMove}
          onMouseLeave={handleMouseLeave}
          onJumpToArchitecture={() => jumpToSection(1)}
        />

        {/* SECTION 2: MONITORING MECHANICS & STATE MACHINE */}
        <SectionMechanics
          isActive={activeSection === 1}
          slideTransition={slideTransition}
        />

        {/* SECTION 3: OBSERVABILITY STACK & ECOSYSTEM */}
        <SectionObservability
          isActive={activeSection === 2}
          slideTransition={slideTransition}
        />
      </motion.div>
    </div>
  );
}
