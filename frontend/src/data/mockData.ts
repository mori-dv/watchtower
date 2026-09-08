import { TargetConfig, TelemetryPoint } from '../types';

export const INITIAL_TARGETS: TargetConfig[] = [
  {
    name: 'google',
    type: 'http',
    address: 'https://google.com',
    interval: '10s',
    timeout: '5s',
    retries: 2,
  },
  {
    name: 'cloudflare-dns',
    type: 'tcp',
    address: '1.1.1.1:53',
    interval: '5s',
    timeout: '3s',
    retries: 1,
  },
  {
    name: 'divar',
    type: 'http',
    address: 'https://divar.ir',
    interval: '15s',
    timeout: '5s',
    retries: 2,
  },
  {
    name: 'internal-api-gateway',
    type: 'http',
    address: 'https://api.internal.local:8443',
    interval: '5s',
    timeout: '2s',
    retries: 2,
  },
];

// Generate 32 initial telemetry slots for the mini uptime bar chart
export const GENERATE_INITIAL_TELEMETRY = (): TelemetryPoint[] => {
  const points: TelemetryPoint[] = [];
  const now = Date.now();
  for (let i = 31; i >= 0; i--) {
    // Mostly healthy, with a tiny degraded slice in the past
    const isDegraded = i === 12;
    points.push({
      id: 32 - i,
      timestamp: new Date(now - i * 15000).toISOString(),
      status: isDegraded ? 'DEGRADED' : 'UP',
      latency: isDegraded ? 245 : Math.floor(12 + Math.random() * 18),
    });
  }
  return points;
};
