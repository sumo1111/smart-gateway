import { useEffect, useState } from 'react';
import { Table, Button, Modal, Form, Input, InputNumber, Switch, Tag, message, Space, Popconfirm, Typography } from 'antd';
import { PlusOutlined, CopyOutlined, ReloadOutlined } from '@ant-design/icons';
import { api } from '../services/api';
import type { Token } from '../types';

const { Text } = Typography;

export default function Tokens() {
  const [tokens, setTokens] = useState<Token[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalOpen, setModalOpen] = useState(false);
  const [editItem, setEditItem] = useState<Token | null>(null);
  const [form] = Form.useForm();

  const load = () => { setLoading(true); api.getTokens().then(r => setTokens(r.data || [])).finally(() => setLoading(false)); };
  useEffect(load, []);

  const handleSave = async () => {
    const values = await form.validateFields();
    if (editItem) {
      await api.updateToken({ ...values, id: editItem.id });
    } else {
      await api.addToken(values);
    }
    setModalOpen(false);
    setEditItem(null);
    load();
    message.success('保存成功');
  };

  const handleDelete = async (id: number) => {
    await api.deleteToken(id);
    load();
    message.success('已删除');
  };

  const handleRefreshKey = async (id: number) => {
    const r = await api.refreshTokenKey(id);
    message.success(`新Key: ${r.key}`);
    load();
  };

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key);
    message.success('已复制');
  };

  const openEdit = (t: Token) => {
    setEditItem(t);
    form.setFieldsValue(t);
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
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>创建令牌</Button>
        <Button icon={<ReloadOutlined />} onClick={load}>刷新</Button>
      </Space>

      <Table dataSource={tokens} rowKey="id" loading={loading}
        columns={[
          { title: '名称', dataIndex: 'name', width: 150 },
          {
            title: 'Key', dataIndex: 'key', width: 300,
            render: (v: string) => (
              <Space>
                <Text code>{v.substring(0, 12)}...{v.substring(v.length - 4)}</Text>
                <Button size="small" icon={<CopyOutlined />} onClick={() => copyKey(v)} />
              </Space>
            ),
          },
          { title: '已用/额度', render: (_: unknown, r: Token) => `${r.quota_used}/${r.quota_limit || '∞'}`, width: 120 },
          { title: '状态', dataIndex: 'status', width: 80, render: (v: number) => <Tag color={v ? 'green' : 'red'}>{v ? '启用' : '禁用'}</Tag> },
          { title: '创建时间', dataIndex: 'created_at', width: 170, render: (v: string) => v?.substring(0, 19) },
          {
            title: '操作', width: 220,
            render: (_: unknown, r: Token) => (
              <Space>
                <Button size="small" onClick={() => openEdit(r)}>编辑</Button>
                <Button size="small" onClick={() => handleRefreshKey(r.id)}>刷新Key</Button>
                <Popconfirm title="确认删除?" onConfirm={() => handleDelete(r.id)}>
                  <Button size="small" danger>删除</Button>
                </Popconfirm>
              </Space>
            ),
          },
        ]}
      />

      <Modal title={editItem ? '编辑令牌' : '创建令牌'} open={modalOpen} onOk={handleSave} onCancel={() => setModalOpen(false)}>
        <Form form={form} labelCol={{ span: 5 }}>
          <Form.Item name="name" label="名称" rules={[{ required: true }]}> <Input /> </Form.Item>
          <Form.Item name="quota_limit" label="额度限制" initialValue={0} extra="0 = 不限制">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label="启用" valuePropName="checked" initialValue={true}>
            <Switch checkedChildren="开" unCheckedChildren="关" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}
