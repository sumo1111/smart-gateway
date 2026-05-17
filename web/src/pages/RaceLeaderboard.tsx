import { useEffect, useState } from 'react';
import { Card, Table, Tag, Button, Row, Col, Spin, message, Progress, Space, Statistic } from 'antd';
import { TrophyOutlined, FireOutlined, ThunderboltOutlined, ReloadOutlined } from '@ant-design/icons';
import { api } from '../services/api';
import type { RaceLeaderboard, RaceTrack, Racer } from '../types';

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
    case 'cold_standby': return <ThunderboltOutlined />;
    case 'benched': return <span>❌</span>;
    default: return null;
  }
};

const roleLabel = (role: string) => {
  switch (role) {
    case 'champion': return '👑 冠军上岗';
    case 'hot_standby': return '🔥 热备';
    case 'cold_standby': return '❄️ 冷备';
    case 'benched': return '🪑 坐板凳';
    default: return role;
  }
};

export default function RaceLeaderboardPage() {
  const [data, setData] = useState<RaceLeaderboard | null>(null);
  const [loading, setLoading] = useState(false);
  const [triggering, setTriggering] = useState(false);

  const load = () => {
    setLoading(true);
    api.getRaceLeaderboard().then(setData).finally(() => setLoading(false));
  };
  useEffect(load, []);

  const handleTrigger = async () => {
    setTriggering(true);
    try {
      await api.triggerRace();
      message.success('竞速已触发，等待结果...');
      setTimeout(load, 5000); // 5秒后刷新
    } catch { message.error('触发失败'); }
    setTriggering(false);
  };

  if (loading || !data) return <Spin size="large" style={{ display: 'block', margin: '100px auto' }} />;

  const totalChampions = data.tracks?.filter(t => t.champion).length || 0;
  const totalRacers = data.tracks?.reduce((sum, t) => sum + (t.racers?.length || 0), 0) || 0;

  return (
    <div>
      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col span={6}>
          <Card><Statistic title="竞速间隔" value={data.race_interval} suffix="秒" prefix={<ThunderboltOutlined />} /></Card>
        </Col>
        <Col span={6}>
          <Card><Statistic title="赛道数" value={data.tracks?.length || 0} suffix="个模型" /></Card>
        </Col>
        <Col span={6}>
          <Card><Statistic title="在岗冠军" value={totalChampions} prefix={<TrophyOutlined />} /></Card>
        </Col>
        <Col span={6}>
          <Card><Statistic title="参赛选手" value={totalRacers} prefix={<FireOutlined />} /></Card>
        </Col>
      </Row>

      <Card style={{ marginBottom: 16 }}>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={load}>刷新排行榜</Button>
          <Button type="primary" loading={triggering} onClick={handleTrigger}>
            🏁 立即竞速
          </Button>
          <span style={{ color: '#999' }}>自动竞速间隔: {data.race_interval}秒</span>
        </Space>
      </Card>

      {data.tracks?.map(track => (
        <Card
          key={track.model}
          title={
            <Space>
              <span style={{ fontSize: 16 }}>🏁 {track.model}</span>
              {track.champion && (
                <Tag color="gold" icon={<TrophyOutlined />}>
                  冠军: {track.champion.channel_name} ({track.champion.score.toFixed(1)}分)
                </Tag>
              )}
              {track.hot_spare && (
                <Tag color="volcano" icon={<FireOutlined />}>
                  热备: {track.hot_spare.channel_name}
                </Tag>
              )}
            </Space>
          }
          style={{ marginBottom: 16 }}
          extra={<span style={{ color: '#666', fontSize: 12 }}>更新: {track.updated_at}</span>}
        >
          <Table
            dataSource={track.racers || []}
            rowKey={(r: Racer) => `${r.channel_id}`}
            size="small"
            pagination={false}
            columns={[
              {
                title: '排名', width: 70, render: (_: unknown, r: Racer) => {
                  if (r.rank === 1) return <span style={{ fontSize: 18 }}>🥇</span>;
                  if (r.rank === 2) return <span style={{ fontSize: 18 }}>🥈</span>;
                  if (r.rank === 3) return <span style={{ fontSize: 18 }}>🥉</span>;
                  return <span style={{ color: '#999' }}>#{r.rank}</span>;
                }
              },
              {
                title: '岗位', dataIndex: 'role', width: 120, render: (v: string) => (
                  <Tag color={roleColor(v)} icon={roleIcon(v)}>{roleLabel(v)}</Tag>
                )
              },
              { title: '渠道', dataIndex: 'channel_name', width: 150 },
              {
                title: '评分', dataIndex: 'score', width: 140, render: (v: number) => (
                  <Progress
                    percent={Math.round(v)}
                    size="small"
                    strokeColor={v >= 70 ? '#52c41a' : v >= 40 ? '#faad14' : '#ff4d4f'}
                    format={(p) => `${p}分`}
                  />
                )
              },
              {
                title: 'TTFT', dataIndex: 'avg_ttft_ms', width: 100, render: (v: number) => v ? (
                  <span style={{ color: v < 500 ? '#52c41a' : v < 1500 ? '#faad14' : '#ff4d4f', fontWeight: v < 500 ? 'bold' : 'normal' }}>
                    {v}ms
                  </span>
                ) : '-'
              },
              {
                title: 'TPS', dataIndex: 'avg_tps', width: 90, render: (v: number) => v ? (
                  <span style={{ color: v > 50 ? '#52c41a' : v > 20 ? '#faad14' : '#ff4d4f', fontWeight: v > 50 ? 'bold' : 'normal' }}>
                    {v.toFixed(1)}/s
                  </span>
                ) : '-'
              },
              {
                title: '成功率', dataIndex: 'success_rate', width: 90, render: (v: number) => (
                  <span style={{ color: v >= 0.9 ? '#52c41a' : v >= 0.5 ? '#faad14' : '#ff4d4f' }}>
                    {v ? `${Math.round(v * 100)}%` : '-'}
                  </span>
                )
              },
              {
                title: '连胜', dataIndex: 'streak', width: 70, render: (v: number) => v >= 3 ? (
                  <Tag color="green">🔥{v}</Tag>
                ) : v > 0 ? <Tag>{v}</Tag> : '-'
              },
              {
                title: '连败', dataIndex: 'fail_streak', width: 70, render: (v: number) => v > 0 ? (
                  <Tag color="red">⚠️{v}</Tag>
                ) : '-'
              },
              {
                title: '卫冕', width: 100, render: (_: unknown, r: Racer) => r.champion_since ? (
                  <span style={{ color: '#faad14' }}>👑 {new Date(r.champion_since).toLocaleTimeString()}</span>
                ) : '-'
              },
            ]}
          />
        </Card>
      ))}

      {(!data.tracks || data.tracks.length === 0) && (
        <Card>
          <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
            <p style={{ fontSize: 48 }}>🏁</p>
            <p>暂无竞速数据</p>
            <p>请确保渠道已配置且策略为 "race"</p>
            <Button type="primary" onClick={handleTrigger}>立即竞速</Button>
          </div>
        </Card>
      )}
    </div>
  );
}
