import { useEffect, useState } from 'react';
import { Table, Select, Tag, Space, Input } from 'antd';
import { api } from '../services/api';
import type { LogEntry } from '../types';

export default function Logs() {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [modelFilter, setModelFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');

  const load = () => {
    setLoading(true);
    const params = ['limit=200'];
    if (modelFilter) params.push(`model=${modelFilter}`);
    if (statusFilter) params.push(`status=${statusFilter}`);
    api.getLogs(params.join('&')).then(r => setLogs(r.data || [])).finally(() => setLoading(false));
  };

  useEffect(load, [modelFilter, statusFilter]);

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Input placeholder="模型筛选" value={modelFilter} onChange={e => setModelFilter(e.target.value)} style={{ width: 200 }} allowClear />
        <Select placeholder="状态筛选" value={statusFilter || undefined} onChange={setStatusFilter} style={{ width: 120 }} allowClear>
          <Select.Option value="success">成功</Select.Option>
          <Select.Option value="fail">失败</Select.Option>
        </Select>
      </Space>

      <Table dataSource={logs} rowKey="id" loading={loading} size="small"
        scroll={{ x: 900 }}
        pagination={{ pageSize: 50, showTotal: t => `共 ${t} 条` }}
        columns={[
          { title: '时间', dataIndex: 'created_at', width: 170, render: (v: string) => v?.substring(0, 19), sorter: (a: LogEntry, b: LogEntry) => a.id - b.id, defaultSortOrder: 'descend' },
          { title: '令牌ID', dataIndex: 'token_id', width: 80 },
          { title: '渠道ID', dataIndex: 'channel_id', width: 80 },
          { title: '模型', dataIndex: 'model', width: 150 },
          { title: 'Prompt', dataIndex: 'prompt_tokens', width: 80, render: (v: number) => v || '-' },
          { title: 'Completion', dataIndex: 'completion_tokens', width: 90, render: (v: number) => v || '-' },
          { title: '延迟', dataIndex: 'latency_ms', width: 80, render: (v: number) => `${v}ms`, sorter: (a: LogEntry, b: LogEntry) => a.latency_ms - b.latency_ms },
          { title: '状态', dataIndex: 'status', width: 70, render: (v: string) => <Tag color={v === 'success' ? 'green' : 'red'}>{v === 'success' ? '成功' : '失败'}</Tag> },
          { title: '错误信息', dataIndex: 'error_msg', ellipsis: true },
        ]}
      />
    </div>
  );
}
