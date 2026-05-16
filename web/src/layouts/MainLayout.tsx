import { useState } from 'react';
import { Layout, Menu, Input, Button, Typography } from 'antd';
import {
  DashboardOutlined, ApiOutlined, KeyOutlined,
  FileTextOutlined, ThunderboltOutlined, SettingOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';

const { Sider, Header, Content } = Layout;
const { Title } = Typography;

const menuItems = [
  { key: '/', icon: <DashboardOutlined />, label: '仪表盘' },
  { key: '/channels', icon: <ApiOutlined />, label: '渠道管理' },
  { key: '/tokens', icon: <KeyOutlined />, label: '令牌管理' },
  { key: '/logs', icon: <FileTextOutlined />, label: '请求日志' },
  { key: '/auto', icon: <ThunderboltOutlined />, label: 'Auto 路由' },
];

export default function MainLayout({ children }: { children: React.ReactNode }) {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const [password, setPassword] = useState(localStorage.getItem('admin_password') || '');
  const [loggedIn, setLoggedIn] = useState(!!localStorage.getItem('admin_password'));

  const handleLogin = () => {
    localStorage.setItem('admin_password', password);
    setLoggedIn(true);
  };

  if (!loggedIn) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh', background: '#141414' }}>
        <div style={{ width: 320, padding: 32, background: '#1f1f1f', borderRadius: 8 }}>
          <Title level={3} style={{ textAlign: 'center', color: '#fff' }}>🚀 Smart Gateway</Title>
          <Input.Password
            placeholder="管理员密码"
            value={password}
            onChange={e => setPassword(e.target.value)}
            onPressEnter={handleLogin}
            style={{ marginBottom: 16 }}
          />
          <Button type="primary" block onClick={handleLogin}>登录</Button>
        </div>
      </div>
    );
  }

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider collapsible collapsed={collapsed} onCollapse={setCollapsed}
        style={{ background: '#001529' }}>
        <div style={{ height: 48, margin: 12, textAlign: 'center' }}>
          <Title level={4} style={{ color: '#fff', margin: 0 }}>
            {collapsed ? '🚀' : '🚀 Smart GW'}
          </Title>
        </div>
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
        />
      </Sider>
      <Layout>
        <Header style={{ background: '#141414', padding: '0 24px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Title level={4} style={{ color: '#fff', margin: 0 }}>
            {menuItems.find(m => m.key === location.pathname)?.label || 'Smart Gateway'}
          </Title>
          <Button type="text" icon={<SettingOutlined />} style={{ color: '#999' }}
            onClick={() => { localStorage.removeItem('admin_password'); setLoggedIn(false); }}>
            退出
          </Button>
        </Header>
        <Content style={{ margin: 24, padding: 24, background: '#1f1f1f', borderRadius: 8, minHeight: 360 }}>
          {children}
        </Content>
      </Layout>
    </Layout>
  );
}
