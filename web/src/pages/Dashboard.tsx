import { useEffect, useState } from 'react';
import { Card, Col, Row, Statistic, Table, Tag, Spin } from 'antd';
import { ArrowUpOutlined, ThunderboltOutlined, ApiOutlined, ClockCircleOutlined } from '@ant-design/icons';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { api } from '../services/api';
import type { DashboardStats, LogEntry } from '../types';

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([api.getStats(7), api.getLogs('limit=10')])
      .then(([s, l]) => { setStats(s); setLogs(l.data || []); })
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;

  return (
    <div>
      <Row gutter={[16, 16]}>
        <Col span={6}>
          <Card><Statistic title="总请求数" value={stats?.total_requests || 0} prefix={<ThunderboltOutlined />} /></Card>
        </Col>
        <Col span={6}>
          <Card><Statistic title="成功率" value={stats?.success_rate || 0} suffix="%" precision={1} valueStyle={{ color: (stats?.success_rate || 0) > 90 ? '#3f8600' : '#cf1322' }} prefix={<ArrowUpOutlined />} /></Card>
        </Col>
        <Col span={6}>
          <Card><Statistic title="平均延迟" value={Math.round(stats?.avg_latency_ms || 0)} suffix="ms" prefix={<ClockCircleOutlined />} /></Card>
        </Col>
        <Col span={6}>
          <Card><Statistic title="活跃渠道" value={stats?.active_channels || 0} suffix={`/ ${stats?.total_models || 0} 模型`} prefix={<ApiOutlined />} /></Card>
        </Col>
      </Row>

      <Card title="近7日请求趋势" style={{ marginTop: 16 }}>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={stats?.daily || []}>
            <CartesianGrid strokeDasharray="3 3" stroke="#333" />
            <XAxis dataKey="day" stroke="#999" />
            <YAxis stroke="#999" />
            <Tooltip />
            <Line type="monotone" dataKey="total" stroke="#1677ff" name="总请求" />
            <Line type="monotone" dataKey="success" stroke="#52c41a" name="成功" />
          </LineChart>
        </ResponsiveContainer>
      </Card>

      <Card title="最近请求" style={{ marginTop: 16 }}>
        <Table dataSource={logs} rowKey="id" size="small" pagination={false}
          columns={[
            { title: '时间', dataIndex: 'created_at', width: 170, render: (v: string) => v?.substring(0, 19) },
            { title: '模型', dataIndex: 'model', width: 150 },
            { title: 'Token数', render: (_: unknown, r: LogEntry) => `${r.prompt_tokens}+${r.completion_tokens}`, width: 120 },
            { title: '延迟', dataIndex: 'latency_ms', width: 80, render: (v: number) => `${v}ms` },
            { title: '状态', dataIndex: 'status', width: 80, render: (v: string) => <Tag color={v === 'success' ? 'green' : 'red'}>{v}</Tag> },
            { title: '错误', dataIndex: 'error_msg', ellipsis: true },
          ]}
        />
      </Card>
    </div>
  );
}
