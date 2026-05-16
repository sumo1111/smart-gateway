import { useEffect, useState } from 'react';
import { Card, Table, Tag, Select, Button, Space, Statistic, Row, Col, Spin, message, Progress } from 'antd';
import { ReloadOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { api } from '../services/api';
import type { AutoStatus, ModelStat } from '../types';

export default function Auto() {
  const [status, setStatus] = useState<AutoStatus | null>(null);
  const [loading, setLoading] = useState(false);
  const [probing, setProbing] = useState(false);

  const load = () => {
    setLoading(true);
    api.getAutoStatus().then(setStatus).finally(() => setLoading(false));
  };
  useEffect(load, []);

  const handleStrategy = async (strategy: string) => {
    await api.setAutoStrategy(strategy);
    load();
    message.success('策略已更新');
  };

  const handleRefresh = async () => {
    await api.refreshAutoScores();
    load();
    message.success('评分已刷新');
  };

  const handleProbe = async () => {
    setProbing(true);
    try {
      await api.probeAllChannels();
      message.success('健康检测完成');
      load();
    } catch { message.error('检测失败'); }
    setProbing(false);
  };

  if (loading || !status) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;

  return (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card><Statistic title="当前策略" value={status.strategy} prefix={<ThunderboltOutlined />} /></Card>
        </Col>
        <Col span={6}>
          <Card><Statistic title="模型路由数" value={status.model_stats?.length || 0} /></Card>
        </Col>
        <Col span={6}>
          <Card><Statistic title="屏蔽阈值" value={status.fail_ban_count} suffix="次" /></Card>
        </Col>
        <Col span={6}>
          <Card><Statistic title="屏蔽时长" value={status.fail_ban_duration} suffix="秒" /></Card>
        </Col>
      </Row>

      <Card style={{ marginBottom: 16 }}>
        <Space>
          <span>路由策略:</span>
          <Select value={status.strategy} style={{ width: 160 }} onChange={handleStrategy}
            options={[
              { value: 'weighted', label: '加权随机 (weighted)' },
              { value: 'lowest_latency', label: '最低延迟优先' },
              { value: 'round_robin', label: '轮询 (round_robin)' },
            ]}
          />
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>刷新评分</Button>
          <Button loading={probing} onClick={handleProbe}>全渠道健康检测</Button>
        </Space>
      </Card>

      <Card title="模型路由表">
        <Table dataSource={status.model_stats || []} rowKey={(r: ModelStat) => `${r.model}-${r.channel_id}`} size="small"
          pagination={{ pageSize: 20 }}
          columns={[
            { title: '模型', dataIndex: 'model', width: 180 },
            { title: '渠道ID', dataIndex: 'channel_id', width: 80 },
            { title: '评分', dataIndex: 'score', width: 120, render: (v: number) => <Progress percent={Math.round(v)} size="small" /> },
            { title: '成功率', width: 100, render: (_: unknown, r: ModelStat) => r.total_calls ? `${Math.round(r.success_calls / r.total_calls * 100)}%` : '-' },
            { title: '平均延迟', dataIndex: 'avg_latency_ms', width: 100, render: (v: number) => v ? `${Math.round(v)}ms` : '-' },
            { title: '连续失败', dataIndex: 'consecutive_fails', width: 90, render: (v: number) => v > 0 ? <Tag color="red">{v}</Tag> : <Tag color="green">0</Tag> },
            { title: '状态', width: 100, render: (_: unknown, r: ModelStat) => {
              if (r.banned_until) return <Tag color="red">已屏蔽</Tag>;
              return <Tag color="green">正常</Tag>;
            }},
            { title: '上次成功', dataIndex: 'last_success_at', width: 170, render: (v: string) => v?.substring(0, 19) || '-' },
            { title: '上次失败', dataIndex: 'last_fail_at', width: 170, render: (v: string) => v?.substring(0, 19) || '-' },
          ]}
        />
      </Card>
    </div>
  );
}
