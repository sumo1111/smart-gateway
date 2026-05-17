export interface Channel {
  id: number;
  name: string;
  type: string;
  base_url: string;
  api_key: string;
  models: string;
  status: number;
  priority: number;
  weight: number;
  max_qps: number;
  created_at: string;
  updated_at: string;
}

export interface Token {
  id: number;
  name: string;
  key: string;
  models: string;
  quota_used: number;
  quota_limit: number;
  status: number;
  created_at: string;
}

export interface LogEntry {
  id: number;
  token_id: number;
  channel_id: number;
  model: string;
  prompt_tokens: number;
  completion_tokens: number;
  cost: number;
  latency_ms: number;
  status: string;
  error_msg: string;
  created_at: string;
}

export interface ModelStat {
  model: string;
  channel_id: number;
  total_calls: number;
  success_calls: number;
  avg_latency_ms: number;
  last_success_at: string | null;
  last_fail_at: string | null;
  consecutive_fails: number;
  banned_until: string | null;
  score: number;
  race_rank: number;
  race_role: string;
  avg_ttft_ms: number;
  avg_tps: number;
  race_success_rate: number;
}

export interface DashboardStats {
  total_requests: number;
  success_rate: number;
  avg_latency_ms: number;
  active_channels: number;
  total_models: number;
  daily: Array<{ day: string; total: number; success: number; avg_latency: number }>;
}

export interface AutoStatus {
  strategy: string;
  model_stats: ModelStat[];
  fail_ban_count: number;
  fail_ban_duration: number;
}

// 竞速相关类型
export interface Racer {
  model: string;
  channel_id: number;
  channel_name: string;
  rank: number;
  role: string; // champion / hot_standby / cold_standby / benched
  score: number;
  avg_ttft_ms: number;
  avg_total_ms: number;
  avg_tps: number;
  success_rate: number;
  champion_since: string | null;
  last_test_at: string;
  streak: number;
  fail_streak: number;
}

export interface RaceTrack {
  model: string;
  champion: Racer | null;
  hot_spare: Racer | null;
  racers: Racer[];
  updated_at: string;
}

export interface RaceLeaderboard {
  strategy: string;
  race_interval: number;
  tracks: RaceTrack[];
}

export interface RaceResult {
  model: string;
  channel_id: number;
  channel_name: string;
  ttft_ms: number;
  ttft_total_ms: number;
  tps: number;
  success: boolean;
  error_msg: string;
  tested_at: string;
}
