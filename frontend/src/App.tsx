import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './context/AuthContext'
import Layout from './components/Layout/Layout'
import Login from './pages/Login/Login'
import Dashboard from './pages/Dashboard/Dashboard'
import SecretsList from './pages/Secrets/SecretsList'
import SecretDetail from './pages/Secrets/SecretDetail'
import MyRequests from './pages/Requests/MyRequests'
import Approvals from './pages/Requests/Approvals'
import AuditLogs from './pages/Audit/AuditLogs'
import Integrations from './pages/Settings/Integrations'

function ProtectedRoute({ children, allowedRoles }: { children: React.ReactNode, allowedRoles?: string[] }) {
  const { user, loading } = useAuth()

  if (loading) {
    return <div className="flex items-center justify-center h-screen">Loading...</div>
  }

  if (!user) {
    return <Navigate to="/login" replace />
  }

  if (allowedRoles && !allowedRoles.includes(user.role)) {
    return <Navigate to="/dashboard" replace />
  }

  return <Layout>{children}</Layout>
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/dashboard" element={
        <ProtectedRoute><Dashboard /></ProtectedRoute>
      } />
      <Route path="/secrets" element={
        <ProtectedRoute><SecretsList /></ProtectedRoute>
      } />
      <Route path="/secrets/:id" element={
        <ProtectedRoute><SecretDetail /></ProtectedRoute>
      } />
      <Route path="/requests" element={
        <ProtectedRoute><MyRequests /></ProtectedRoute>
      } />
      <Route path="/requests/approvals" element={
        <ProtectedRoute allowedRoles={['team_lead', 'security_admin']}><Approvals /></ProtectedRoute>
      } />
      <Route path="/audit" element={
        <ProtectedRoute allowedRoles={['security_admin']}><AuditLogs /></ProtectedRoute>
      } />
      <Route path="/settings/integrations" element={
        <ProtectedRoute allowedRoles={['security_admin']}><Integrations /></ProtectedRoute>
      } />
      <Route path="/" element={<Navigate to="/dashboard" replace />} />
    </Routes>
  )
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <AppRoutes />
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App
