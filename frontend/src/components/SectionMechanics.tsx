import React, { useState, useEffect } from 'react';
import { motion } from 'motion/react';
import { 
  Activity, Server, CheckCircle2, AlertTriangle, 
  XCircle, Send, Terminal, Database, Clock, RefreshCw
} from 'lucide-react';

interface SectionMechanicsProps {
  isActive: boolean;
  slideTransition: { duration: number; ease: number[] };
}

type SimulationScenario = 'nominal' | 'degraded' | 'outage';

interface MonitorTarget {
  name: string;
  type: 'http' | 'tcp';
  endpoint: string;
  interval: string;
  latency: number;
  status: 'UP' | 'DEGRADED' | 'DOWN';
  statusText: string;
  lastChecked: string;
}

export const SectionMechanics: React.FC<SectionMechanicsProps> = ({
  isActive,
  slideTransition,
}) => {
  const [scenario, setScenario] = useState<SimulationScenario>('nominal');
  const [ticker, setTicker] = useState(0);

  // Targets state
  const [targets, setTargets] = useState<MonitorTarget[]>([
    {
      name: 'google',
      type: 'http',
      endpoint: 'https://google.com',
      interval: '10s',
      latency: 18,
      status: 'UP',
      statusText: '200 OK',
      lastChecked: '2s ago',
    },
    {
      name: 'cloudflare-dns',
      type: 'tcp',
      endpoint: '1.1.1.1:53',
      interval: '5s',
      latency: 8,
      status: 'UP',
      statusText: 'CONNECTED',
      lastChecked: '1s ago',
    },
    {
      name: 'divar',
      type: 'http',
      endpoint: 'https://divar.ir',
      interval: '15s',
      latency: 32,
      status: 'UP',
      statusText: '200 OK',
      lastChecked: '4s ago',
    },
    {
      name: 'internal-gateway',
      type: 'http',
      endpoint: 'https://api.internal:8443',
      interval: '5s',
      latency: 4,
      status: 'UP',
      statusText: '200 OK',
      lastChecked: 'just now',
    },
  ]);

  // Subtle real-time latency updates
  useEffect(() => {
    if (!isActive) return;
    const interval = setInterval(() => {
      setTicker((prev) => prev + 1);
      setTargets((prev) =>
        prev.map((t) => {
          if (scenario === 'outage' && t.name === 'google') {
            return {
              ...t,
              latency: 5000,
              status: 'DOWN',
              statusText: 'TIMEOUT',
              lastChecked: 'just now',
            };
          }
          if (scenario === 'degraded' && t.name === 'google') {
            return {
              ...t,
              latency: 380,
              status: 'DEGRADED',
              statusText: '502 BAD GW',
              lastChecked: 'just now',
            };
          }
          const base = t.name === 'cloudflare-dns' ? 8 : t.name === 'divar' ? 32 : t.name === 'internal-gateway' ? 4 : 18;
          const jitter = Math.floor(Math.random() * 5) - 2;
          return {
            ...t,
            latency: Math.max(2, base + jitter),
            status: 'UP',
            statusText: t.type === 'http' ? '200 OK' : 'CONNECTED',
            lastChecked: 'just now',
          };
        })
      );
    }, 3000);
    return () => clearInterval(interval);
  }, [isActive, scenario]);

  // Handle Scenario Switch
  const handleSelectScenario = (s: SimulationScenario) => {
    setScenario(s);
    setTargets((prev) =>
      prev.map((t) => {
        if (t.name !== 'google') return t;
        if (s === 'outage') {
          return { ...t, latency: 5000, status: 'DOWN', statusText: 'TIMEOUT', lastChecked: 'just now' };
        }
        if (s === 'degraded') {
          return { ...t, latency: 380, status: 'DEGRADED', statusText: '502 BAD GW', lastChecked: 'just now' };
        }
        return { ...t, latency: 18, status: 'UP', statusText: '200 OK', lastChecked: 'just now' };
      })
    );
  };

  // Dynamic values derived from scenario
  const consecutiveFailures = scenario === 'outage' ? 3 : scenario === 'degraded' ? 1 : 0;
  const targetStatus = scenario === 'outage' ? 'DOWN' : scenario === 'degraded' ? 'DEGRADED' : 'UP';
  const cooldownActive = scenario === 'outage';

  return (
    <section
      id="section-2"
      className="h-[100dvh] w-full shrink-0 flex flex-col justify-center items-center relative px-3 sm:px-6 md:px-8 py-4 sm:py-8 md:py-12 select-none overflow-hidden"
    >
      <motion.div
        className="w-full max-w-5xl space-y-3 sm:space-y-5 md:space-y-7 origin-center"
        animate={{
          opacity: isActive ? 1 : 0.35,
          scale: isActive ? 1 : 0.98,
        }}
        transition={slideTransition}
      >
        {/* Section Header */}
        <div className="text-center space-y-1 sm:space-y-1.5 px-2">
          <h2 className="text-lg sm:text-2xl md:text-4xl font-semibold tracking-tight text-white leading-tight">
            Synthetic Probes &amp; State Evaluator
          </h2>
          <p className="text-[9px] sm:text-xs md:text-sm text-zinc-400 font-mono max-w-xl mx-auto">
            Real-time HTTP/TCP sentinel probes with Redis atomic failure evaluation and Telegram anti-spam cooldowns.
          </p>
        </div>

        {/* 2-Column Responsive Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3 sm:gap-4 md:gap-6">
          
          {/* Card 1: Active Synthetic Monitor Console */}
          <div className="rounded-xl sm:rounded-2xl bg-[#0D1017]/95 border border-zinc-800/80 p-3 sm:p-5 md:p-6 flex flex-col justify-between space-y-3 shadow-xl backdrop-blur-sm">
            <div className="space-y-3">
              {/* Header with Live Heartbeat */}
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 text-xs font-mono text-zinc-400 uppercase tracking-wider">
                  <Activity className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                  <span>Sentinel Target Monitor</span>
                </div>
                <div className="flex items-center gap-1.5 text-[10px] font-mono text-zinc-400">
                  <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                  <span>PROBES ACTIVE ({targets.length})</span>
                </div>
              </div>

              {/* Targets Table / Matrix */}
              <div className="space-y-1.5 font-mono">
                {targets.map((t) => {
                  const isDegraded = t.status === 'DEGRADED';
                  const isDown = t.status === 'DOWN';
                  return (
                    <div
                      key={t.name}
                      className={`p-2 sm:p-2.5 rounded-lg border transition-all flex items-center justify-between text-[11px] sm:text-xs ${
                        isDown
                          ? 'bg-red-950/20 border-red-500/40 text-red-200 shadow-[0_0_12px_rgba(239,68,68,0.1)]'
                          : isDegraded
                          ? 'bg-amber-950/20 border-amber-500/40 text-amber-200'
                          : 'bg-[#08090C]/80 border-zinc-800/70 hover:border-zinc-700/80 text-zinc-200'
                      }`}
                    >
                      <div className="flex items-center gap-2 min-w-0">
                        <span
                          className={`w-2 h-2 rounded-full shrink-0 ${
                            isDown ? 'bg-red-400' : isDegraded ? 'bg-amber-400 animate-pulse' : 'bg-emerald-400'
                          }`}
                        />
                        <span className="font-semibold truncate">{t.name}</span>
                        <span className="text-[9px] uppercase px-1.5 py-0.2 rounded bg-zinc-800/80 text-zinc-400 border border-zinc-700/50">
                          {t.type}
                        </span>
                      </div>

                      <div className="flex items-center gap-2.5 shrink-0">
                        <span className="text-zinc-400 text-[10px]">
                          {t.latency} ms
                        </span>
                        <span
                          className={`px-1.5 py-0.5 rounded text-[9px] font-semibold border ${
                            isDown
                              ? 'bg-red-500/10 text-red-400 border-red-500/30'
                              : isDegraded
                              ? 'bg-amber-500/10 text-amber-400 border-amber-500/30'
                              : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
                          }`}
                        >
                          {t.statusText}
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>

              {/* Realistic Operational Simulation Modes */}
              <div className="pt-1">
                <div className="text-[9px] font-mono text-zinc-400 pb-1.5 flex items-center justify-between">
                  <span>TELEMETRY SCENARIO SIMULATOR:</span>
                  <span className="text-zinc-400">TARGET: google</span>
                </div>
                <div className="grid grid-cols-3 gap-1.5 font-mono text-[10px]">
                  <button
                    onClick={() => handleSelectScenario('nominal')}
                    className={`py-1.5 px-2 rounded border transition-all cursor-pointer text-center ${
                      scenario === 'nominal'
                        ? 'bg-emerald-500/15 border-emerald-500/60 text-emerald-300 font-semibold shadow-sm'
                        : 'bg-zinc-900/80 border-zinc-800 text-zinc-400 hover:text-zinc-200'
                    }`}
                  >
                    Nominal (UP)
                  </button>
                  <button
                    onClick={() => handleSelectScenario('degraded')}
                    className={`py-1.5 px-2 rounded border transition-all cursor-pointer text-center ${
                      scenario === 'degraded'
                        ? 'bg-amber-500/15 border-amber-500/60 text-amber-300 font-semibold shadow-sm'
                        : 'bg-zinc-900/80 border-zinc-800 text-zinc-400 hover:text-zinc-200'
                    }`}
                  >
                    Packet Loss
                  </button>
                  <button
                    onClick={() => handleSelectScenario('outage')}
                    className={`py-1.5 px-2 rounded border transition-all cursor-pointer text-center ${
                      scenario === 'outage'
                        ? 'bg-red-500/15 border-red-500/60 text-red-300 font-semibold shadow-sm'
                        : 'bg-zinc-900/80 border-zinc-800 text-zinc-400 hover:text-zinc-200'
                    }`}
                  >
                    Outage (Fail &ge;3)
                  </button>
                </div>
              </div>
            </div>

            {/* Footer Spec */}
            <div className="pt-2 sm:pt-3 border-t border-zinc-800/60 flex items-center justify-between text-[8px] sm:text-[10px] font-mono text-zinc-400">
              <span className="flex items-center gap-1.5">
                <Server className="w-3 h-3 text-emerald-400" />
                WORKER POOL: 5 THREADS
              </span>
              <span>SYNTHETIC ENGINE: GO 1.25</span>
            </div>
          </div>

          {/* Card 2: State Evaluator & Telegram Dispatcher */}
          <div className="rounded-xl sm:rounded-2xl bg-[#0D1017]/95 border border-zinc-800/80 p-3 sm:p-5 md:p-6 flex flex-col justify-between space-y-3 shadow-xl backdrop-blur-sm">
            <div className="space-y-3">
              {/* Header with Redis Key State */}
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 text-xs font-mono text-zinc-400 uppercase tracking-wider">
                  <Database className="w-3.5 h-3.5 text-cyan-400 shrink-0" />
                  <span>Redis State &amp; Alert Engine</span>
                </div>
                <span className="text-[10px] font-mono text-zinc-400 bg-zinc-900 px-2 py-0.5 rounded border border-zinc-800">
                  failures:google = <strong className="text-white">{consecutiveFailures}/3</strong>
                </span>
              </div>

              {/* 3 State Machine Transition Steps */}
              <div className="grid grid-cols-3 gap-1.5 font-mono text-center">
                <div className={`p-2 rounded-lg border transition-all ${
                  targetStatus === 'UP'
                    ? 'bg-emerald-950/30 border-emerald-500/60 text-emerald-300 shadow-[0_0_12px_rgba(16,185,129,0.12)]'
                    : 'bg-zinc-900/40 border-zinc-800/60 text-zinc-500'
                }`}>
                  <div className="flex justify-center pb-1">
                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                  </div>
                  <div className="text-[11px] font-semibold">HEALTHY</div>
                  <div className="text-[9px] text-zinc-400">0 Failures</div>
                </div>

                <div className={`p-2 rounded-lg border transition-all ${
                  targetStatus === 'DEGRADED'
                    ? 'bg-amber-950/30 border-amber-500/60 text-amber-300 shadow-[0_0_12px_rgba(245,158,11,0.12)]'
                    : 'bg-zinc-900/40 border-zinc-800/60 text-zinc-500'
                }`}>
                  <div className="flex justify-center pb-1">
                    <AlertTriangle className="w-3.5 h-3.5 text-amber-400" />
                  </div>
                  <div className="text-[11px] font-semibold">DEGRADED</div>
                  <div className="text-[9px] text-zinc-400">1-2 Failures</div>
                </div>

                <div className={`p-2 rounded-lg border transition-all ${
                  targetStatus === 'DOWN'
                    ? 'bg-red-950/30 border-red-500/60 text-red-300 shadow-[0_0_12px_rgba(239,68,68,0.18)]'
                    : 'bg-zinc-900/40 border-zinc-800/60 text-zinc-500'
                }`}>
                  <div className="flex justify-center pb-1">
                    <XCircle className="w-3.5 h-3.5 text-red-400" />
                  </div>
                  <div className="text-[11px] font-semibold">DOWN</div>
                  <div className="text-[9px] text-zinc-400">&ge; 3 Failures</div>
                </div>
              </div>

              {/* Realistic Structured JSON Log Snippet from Watchtower Go Backend */}
              <div className="p-2.5 rounded-lg bg-[#08090C] border border-zinc-800/80 font-mono text-[9px] sm:text-[10px] space-y-1">
                <div className="flex items-center justify-between text-zinc-400 pb-1 border-b border-zinc-800/50">
                  <span className="flex items-center gap-1.5">
                    <Terminal className="w-3 h-3 text-cyan-400" />
                    <span>Structured JSON Telemetry Log</span>
                  </span>
                  <span>level: {targetStatus === 'DOWN' ? 'error' : targetStatus === 'DEGRADED' ? 'warn' : 'info'}</span>
                </div>
                <div className="text-zinc-300 leading-relaxed overflow-x-auto whitespace-pre">
                  {targetStatus === 'DOWN'
                    ? `{"target":"google","status":"DOWN","consecutive_failures":3,"latency":"5002ms","msg":"check failed: timeout"}\n{"action":"telegram_alert","cooldown":"15m","dispatched":true}`
                    : targetStatus === 'DEGRADED'
                    ? `{"target":"google","status":"DEGRADED","consecutive_failures":1,"latency":"380ms","msg":"transient failure recorded"}`
                    : `{"target":"google","status":"UP","consecutive_failures":0,"latency":"18ms","msg":"check completed"}`}
                </div>
              </div>

              {/* Realistic Telegram Dark-Mode Notification Bubble */}
              <div className="p-2.5 rounded-lg bg-[#08090C] border border-zinc-800/80 font-mono space-y-1.5">
                <div className="flex items-center justify-between text-[10px] pb-1 border-b border-zinc-800/60 text-zinc-400">
                  <div className="flex items-center gap-1.5 text-zinc-300">
                    <Send className="w-3 h-3 text-cyan-400" />
                    <span>Telegram Alert Dispatcher</span>
                  </div>
                  <div className="flex items-center gap-1 text-[9px]">
                    <Clock className="w-3 h-3 text-zinc-400" />
                    <span className={cooldownActive ? 'text-amber-400 font-semibold' : 'text-zinc-400'}>
                      {cooldownActive ? 'ANTI-SPAM: 15M COOLDOWN' : 'ANTI-SPAM: STANDBY'}
                    </span>
                  </div>
                </div>

                <div className="text-[10px] text-zinc-300 flex items-start gap-2">
                  <span className="text-base leading-none">
                    {targetStatus === 'DOWN' ? '🚨' : targetStatus === 'DEGRADED' ? '⚠️' : '✅'}
                  </span>
                  <div className="space-y-0.5 min-w-0">
                    <span className="font-semibold block text-zinc-200">
                      {targetStatus === 'DOWN'
                        ? 'Watchtower Alert: google is DOWN'
                        : targetStatus === 'DEGRADED'
                        ? 'Degraded Performance: google (latency spike)'
                        : 'Watchtower Sentinel: All targets nominal'}
                    </span>
                    <p className="text-zinc-400 text-[9px] leading-tight">
                      {targetStatus === 'DOWN'
                        ? '3 consecutive failures recorded in Redis. Alert sent. Duplicate notifications suppressed for 15 minutes.'
                        : targetStatus === 'DEGRADED'
                        ? '1 failure detected. State set to DEGRADED without spamming alerts.'
                        : 'Health probes passing within configured SLA latency budgets.'}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            {/* Footer Spec */}
            <div className="pt-2 sm:pt-3 border-t border-zinc-800/60 flex items-center justify-between text-[8px] sm:text-[10px] font-mono text-zinc-400">
              <span>STORAGE: REDIS 8.6</span>
              <span>SLIDING COOLDOWN: 15 MIN</span>
            </div>
          </div>

        </div>

      </motion.div>
    </section>
  );
};
