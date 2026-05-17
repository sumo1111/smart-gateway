import { useEffect, useState } from 'react';
import { Card, Table, Tag, Select, Button, Space, Statistic, Row, Col, Spin, message, Progress, Tooltip } from 'antd';
import { ReloadOutlined, ThunderboltOutlined, TrophyOutlined, FireOutlined, SwapOutlined } from '@ant-design/icons';
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

  const roleColor = (role: string) => {
    switch (role) {
      case 'champion': return 'gold';
      case 'hot_standby': return 'volcano';
      case 'cold_standby': return 'blue';
      case 'benched': return 'red';
      default: return 'default';
    }
  };

  const roleIcon = (role: string) => {
    switch (role) {
      case 'champion': return <TrophyOutlined />;
      case 'hot_standby': return <FireOutlined />;
      case 'cold_standby': return <SwapOutlined />;
      case 'benched': return <span>❌</span>;
      default: return null;
    }
  };

  const roleLabel = (role: string) => {
    switch (role) {
      case 'champion': return '冠军上岗';
      case 'hot_standby': return '热备';
      case 'cold_standby': return '冷备';
      case 'benched': return '坐板凳';
      default: return role;
    }
  };

  return (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card><Statistic title="当前策略" value={status.strategy === 'race' ? '🏁 竞速上岗' : status.strategy} prefix={<ThunderboltOutlined />} /></Card>
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
          <Select value={status.strategy} style={{ width: 180 }} onChange={handleStrategy}
            options={[
              { value: 'race', label: '🏁 竞速上岗 (race)' },
              { value: 'weighted', label: '加权随机 (weighted)' },
              { value: 'lowest_latency', label: '最低延迟优先' },
              { value: 'round_robin', label: '轮询 (round_robin)' },
            ]}
          />
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>刷新评分</Button>
          <Button loading={probing} onClick={handleProbe}>全渠道健康检测</Button>
        </Space>
      </Card>

      <Card title={status.strategy === 'race' ? '🏁 竞速排行榜' : '模型路由表'}>
        <Table dataSource={status.model_stats || []} rowKey={(r: ModelStat) => `${r.model}-${r.channel_id}`} size="small"
          pagination={{ pageSize: 20 }}
          columns={[
            { title: '模型', dataIndex: 'model', width: 180, fixed: 'left' },
            { title: '岗位', dataIndex: 'race_role', width: 100, render: (v: string) => (
              <Tooltip title={roleLabel(v)}>
                <Tag color={roleColor(v)} icon={roleIcon(v)}>{roleLabel(v)}</Tag>
              </Tooltip>
            )},
            { title: '排名', dataIndex: 'race_rank', width: 70, render: (v: number) => {
              if (v === 1) return <span style={{ color: '#faad14', fontWeight: 'bold' }}>🥇 {v}</span>;
              if (v === 2) return <span style={{ color: '#ff7a45' }}>🥈 {v}</span>;
              if (v === 3) return <span style={{ color: '#69c0ff' }}>🥉 {v}</span>;
              return v || '-';
            }},
            { title: '评分', dataIndex: 'score', width: 130, render: (v: number) => <Progress percent={Math.round(v)} size="small"
              strokeColor={v >= 70 ? '#52c41a' : v >= 40 ? '#faad14' : '#ff4d4f'} /> },
            { title: 'TTFT', dataIndex: 'avg_ttft_ms', width: 90, render: (v: number) => v ? (
              <span style={{ color: v < 500 ? '#52c41a' : v < 1500 ? '#faad14' : '#ff4d4f' }}>{v}ms</span>
            ) : '-' },
            { title: 'TPS', dataIndex: 'avg_tps', width: 80, render: (v: number) => v ? (
              <span style={{ color: v > 50 ? '#52c41a' : v > 20 ? '#faad14' : '#ff4d4f' }}>{v.toFixed(1)}</span>
            ) : '-' },
            { title: '成功率', dataIndex: 'race_success_rate', width: 90, render: (v: number) => v ? `${Math.round(v * 100)}%` : (
              <span>{(() => { const r = status.model_stats?.find(s => true); return r && r.total_calls ? `${Math.round(r.success_calls / r.total_calls * 100)}%` : '-'; })()}</span>
            )},
            { title: '平均延迟', dataIndex: 'avg_latency_ms', width: 100, render: (v: number) => v ? `${Math.round(v)}ms` : '-' },
            { title: '连续失败', dataIndex: 'consecutive_fails', width: 90, render: (v: number) => v > 0 ? <Tag color="red">{v}</Tag> : <Tag color="green">0</Tag> },
            { title: '状态', width: 100, render: (_: unknown, r: ModelStat) => {
              if (r.banned_until) return <Tag color="red">已屏蔽</Tag>;
              if (r.race_role === 'benched') return <Tag color="red">坐板凳</Tag>;
              return <Tag color="green">正常</Tag>;
            }},
          ]}
        />
      </Card>
    </div>
  );
}
