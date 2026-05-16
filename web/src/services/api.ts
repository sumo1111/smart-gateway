const BASE = '';

function headers(): Record<string, string> {
  const h: Record<string, string> = { 'Content-Type': 'application/json' };
  const pw = localStorage.getItem('admin_password');
  if (pw) h['X-Admin-Password'] = pw;
  return h;
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + url, { ...options, headers: { ...headers(), ...(options?.headers || {}) } });
  if (!res.ok) throw new Error(`HTTP ${res.status}: ${await res.text()}`);
  return res.json();
}

export const api = {
  // Channels
  getChannels: () => request<{ data: Channel[] }>('/api/channel'),
  getChannel: (id: number) => request<{ data: Channel }>(`/api/channel/${id}`),
  addChannel: (ch: Partial<Channel>) => request<{ data: Channel }>('/api/channel', { method: 'POST', body: JSON.stringify(ch) }),
  updateChannel: (ch: Partial<Channel>) => request<{ data: Channel }>('/api/channel', { method: 'PUT', body: JSON.stringify(ch) }),
  deleteChannel: (id: number) => request(`/api/channel/${id}`, { method: 'DELETE' }),
  testChannel: (id: number) => request<{ ok: boolean; message: string; latency_ms: number }>(`/api/channel/test/${id}`),

  // Tokens
  getTokens: () => request<{ data: Token[] }>('/api/token'),
  addToken: (t: Partial<Token>) => request<{ data: Token }>('/api/token', { method: 'POST', body: JSON.stringify(t) }),
  updateToken: (t: Partial<Token>) => request<{ data: Token }>('/api/token', { method: 'PUT', body: JSON.stringify(t) }),
  deleteToken: (id: number) => request(`/api/token/${id}`, { method: 'DELETE' }),
  refreshTokenKey: (id: number) => request<{ key: string }>(`/api/token/refresh/${id}`, { method: 'POST' }),

  // Logs
  getLogs: (params?: string) => request<{ data: LogEntry[] }>(`/api/log${params ? '?' + params : ''}`),

  // Stats
  getStats: (days = 7) => request<DashboardStats>(`/api/stats?days=${days}`),

  // Auto
  getAutoStatus: () => request<AutoStatus>('/api/auto/status'),
  setAutoStrategy: (strategy: string) => request('/api/auto/strategy', { method: 'POST', body: JSON.stringify({ strategy }) }),
  refreshAutoScores: () => request('/api/auto/refresh', { method: 'POST' }),
  probeAllChannels: () => request('/api/auto/probe', { method: 'POST' }),
};

// 需要import类型
import type { Channel, Token, LogEntry, DashboardStats, AutoStatus } from '../types';
