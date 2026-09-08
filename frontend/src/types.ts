export type TargetType = 'http' | 'tcp';

export type TargetStatus = 'UP' | 'DEGRADED' | 'DOWN';

export interface TargetConfig {
  name: string;
  type: TargetType;
  address: string;
  interval: string;
  timeout: string;
  retries: number;
}

export interface CheckResult {
  targetName: string;
  targetType: TargetType;
  address: string;
  status: TargetStatus;
  latencyMs: number;
  timestamp: string;
  error?: string;
}

export interface EvaluatedState {
  targetName: string;
  status: TargetStatus;
  consecutiveFailures: number;
  maxFailuresThreshold: number;
  lastChanged: string;
}

export interface TelegramAlertEvent {
  id: string;
  targetName: string;
  targetType: TargetType;
  status: TargetStatus;
  consecutiveFailures: number;
  latencyMs: number;
  timestamp: string;
  cooldownActive: boolean;
  cooldownSecondsLeft: number;
  suppressed: boolean;
}

export interface TelemetryPoint {
  id: number;
  timestamp: string;
  status: TargetStatus;
  latency: number;
}
