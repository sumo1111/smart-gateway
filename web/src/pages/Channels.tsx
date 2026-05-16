import { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, Input, Select, InputNumber, Switch, Tag, message, Space, Popconfirm } from 'antd';
import { PlusOutlined, ReloadOutlined, ExperimentOutlined } from '@ant-design/icons';
import { api } from '../services/api';
import type { Channel } from '../types';

const channelTypes = [
  { value: 'openai', label: 'OpenAI 兼容' },
  { value: 'anthropic', label: 'Anthropic Claude' },
  { value: 'custom', label: '自定义端点' },
];

export default function Channels() {
  const [channels, setChannels] = useState<Channel[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editItem, setEditItem] = useState<Channel | null>(null);
  const [testing, setTesting] = useState<number | null>(null);
  const [form] = Form.useForm();

  const load = () => { setLoading(true); api.getChannels().then(r => setChannels(r.data || [])).finally(() => setLoading(false)); };
  useEffect(load, []);

  const handleSave = async () => {
    const values = await form.validateFields();
    values.models = Array.isArray(values.models) ? values.models.join(',') : values.models;
    if (editItem) {
      await api.updateChannel({ ...values, id: editItem.id });
    } else {
      await api.addChannel(values);
    }
    setModalOpen(false);
    setEditItem(null);
    load();
    message.success('保存成功');
  };

  const handleTest = async (id: number) => {
    setTesting(id);
    try {
      const r = await api.testChannel(id);
      if (r.ok) message.success(`连通正常 (${r.latency_ms}ms)`);
      else message.error(`连接失败: ${r.message}`);
    } catch { message.error('测试请求失败'); }
    setTesting(null);
  };

  const handleDelete = async (id: number) => {
    await api.deleteChannel(id);
    load();
    message.success('已删除');
  };

  const openEdit = (ch: Channel) => {
    setEditItem(ch);
    form.setFieldsValue({ ...ch, models: ch.models ? ch.models.split(',') : [] });
    setModalOpen(true);
  };

  const openCreate = () => {
    setEditItem(null);
    form.resetFields();
    setModalOpen(true);
  };

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>添加渠道</Button>
        <Button icon={<ReloadOutlined />} onClick={load}>刷新</Button>
      </Space>

      <Table dataSource={channels} rowKey="id" loading={loading}
        columns={[
          { title: '名称', dataIndex: 'name', width: 150 },
          { title: '类型', dataIndex: 'type', width: 120, render: (v: string) => <Tag>{v}</Tag> },
          { title: 'Base URL', dataIndex: 'base_url', width: 250, ellipsis: true },
          { title: '状态', dataIndex: 'status', width: 80, render: (v: number) => <Tag color={v ? 'green' : 'red'}>{v ? '启用' : '禁用'}</Tag> },
          { title: '优先级', dataIndex: 'priority', width: 80 },
          { title: '模型数', render: (_: unknown, r: Channel) => r.models ? r.models.split(',').length : 0, width: 80 },
          {
            title: '操作', width: 220,
            render: (_: unknown, r: Channel) => (
              <Space>
                <Button size="small" onClick={() => openEdit(r)}>编辑</Button>
                <Button size="small" icon={<ExperimentOutlined />} loading={testing === r.id} onClick={() => handleTest(r.id)}>测试</Button>
                <Popconfirm title="确认删除?" onConfirm={() => handleDelete(r.id)}>
                  <Button size="small" danger>删除</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />

      <Modal title={editItem ? '编辑渠道' : '添加渠道'} open={modalOpen} onOk={handleSave} onCancel={() => setModalOpen(false)} width={560}>
        <Form form={form} labelCol={{ span: 5 }}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}> <Input /> </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}> <Select options={channelTypes} /> </Form.Item>
          <Form.Item name="base_url" label="Base URL" rules={[{ required: true }]}> <Input placeholder="https://api.openai.com/v1" /> </Form.Item>
          <Form.Item name="api_key" label="API Key" rules={[{ required: true }]}> <Input.Password /> </Form.Item>
          <Form.Item name="models" label="模型列表"> <Select mode="tags" placeholder="输入模型名回车添加" /> </Form.Item>
          <Form.Item name="priority" label="优先级" initialValue={5}> <InputNumber min={1} max={10} /> </Form.Item>
          <Form.Item name="weight" label="权重" initialValue={50}> <InputNumber min={1} max={100} /> </Form.Item>
          <Form.Item name="status" label="启用" valuePropName="checked" initialValue={true}>
            <Switch checkedChildren="开" unCheckedChildren="关" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
