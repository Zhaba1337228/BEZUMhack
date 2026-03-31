import { NavLink } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import {
  LayoutDashboard,
  Key,
  FileText,
  CheckCircle,
  ScrollText,
  Settings,
  Shield,
} from 'lucide-react'

const navItems = [
  { path: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { path: '/secrets', label: 'Secrets', icon: Key },
  { path: '/requests', label: 'My Requests', icon: FileText },
]

const adminNavItems = [
  { path: '/requests/approvals', label: 'Approvals', icon: CheckCircle, roles: ['team_lead', 'security_admin'] },
  { path: '/audit', label: 'Audit Logs', icon: ScrollText, roles: ['security_admin'] },
  { path: '/settings/integrations', label: 'Integrations', icon: Settings, roles: ['security_admin'] },
]

export default function Sidebar({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { user } = useAuth()

  const canAccess = (roles: string[]) => {
    if (!user) return false
    return roles.includes(user.role)
  }

  return (
    <>
      {/* Mobile overlay */}
      {open && (
        <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-20 lg:hidden" onClick={onClose} />
      )}

      {/* Sidebar */}
      <aside className={`
        fixed lg:static inset-y-0 left-0 z-30
        w-72 bg-surface-900 text-white
        transform transition-transform duration-300 ease-out
        ${open ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
        flex flex-col
      `}>
        {/* Logo */}
        <div className="p-6 border-b border-surface-800">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-brand-600 rounded-lg shadow-glow">
              <Shield className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="text-lg font-bold tracking-tight">SecretFlow</h1>
              <p className="text-xs text-surface-400">Enterprise Security</p>
            </div>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 p-4 space-y-1 overflow-y-auto">
          {navItems.map((item) => {
            const Icon = item.icon
            return (
              <NavLink
                key={item.path}
                to={item.path}
                className={({ isActive }) =>
                  `flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium transition-all
                  ${isActive
                    ? 'bg-brand-600 text-white shadow-lg shadow-brand-900/20'
                    : 'text-surface-300 hover:bg-surface-800 hover:text-white'
                  }`
                }
              >
                <Icon className="w-5 h-5" />
                {item.label}
              </NavLink>
            )
          })}

          {(user?.role === 'team_lead' || user?.role === 'security_admin') && (
            <>
              <div className="pt-6 pb-2">
                <div className="px-4 text-xs font-semibold text-surface-500 uppercase tracking-wider">
                  Administration
                </div>
              </div>
              {adminNavItems.map((item) => {
                if (!canAccess(item.roles)) return null
                const Icon = item.icon
                return (
                  <NavLink
                    key={item.path}
                    to={item.path}
                    className={({ isActive }) =>
                      `flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium transition-all
                      ${isActive
                        ? 'bg-brand-600 text-white shadow-lg shadow-brand-900/20'
                        : 'text-surface-300 hover:bg-surface-800 hover:text-white'
                      }`
                    }
                  >
                    <Icon className="w-5 h-5" />
                    {item.label}
                  </NavLink>
                )
              })}
            </>
          )}
        </nav>

        {/* User profile */}
        <div className="p-4 border-t border-surface-800 bg-surface-900/50">
          <div className="flex items-center gap-3 px-4 py-3 rounded-lg bg-surface-800/50">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-brand-500 to-brand-700 flex items-center justify-center font-semibold text-sm">
              {user?.username.charAt(0).toUpperCase()}
            </div>
            <div className="flex-1 min-w-0">
              <div className="font-medium text-sm truncate">{user?.username}</div>
              <div className="text-xs text-surface-400 capitalize">{user?.role?.replace('_', ' ')}</div>
            </div>
          </div>
        </div>
      </aside>
    </>
  )
}
