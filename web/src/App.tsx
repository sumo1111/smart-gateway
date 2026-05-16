import { Routes, Route, Navigate } from 'react-router-dom';
import MainLayout from './layouts/MainLayout';
import Dashboard from './pages/Dashboard';
import Channels from './pages/Channels';
import Tokens from './pages/Tokens';
import Logs from './pages/Logs';
import Auto from './pages/Auto';

export default function App() {
  return (
    <MainLayout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/channels" element={<Channels />} />
        <Route path="/tokens" element={<Tokens />} />
        <Route path="/logs" element={<Logs />} />
        <Route path="/auto" element={<Auto />} />
        <Route path="*" element={<Navigate to="/" />} />
      </Routes>
    </MainLayout>
  );
}
