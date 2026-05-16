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
