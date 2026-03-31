import { Bell, LogOut, Menu } from 'lucide-react'
import { useAuth } from '../../context/AuthContext'

export default function TopBar({ onMenuClick }: { onMenuClick: () => void }) {
  const { user, logout } = useAuth()

  return (
    <header className="bg-white/80 backdrop-blur-md border-b border-surface-200 sticky top-0 z-10">
      <div className="flex items-center justify-between px-6 py-3">
        {/* Mobile menu button */}
        <button
          onClick={onMenuClick}
          className="lg:hidden p-2 text-surface-500 hover:bg-surface-100 rounded-lg transition-colors"
        >
          <Menu className="w-5 h-5" />
        </button>

        {/* Page title placeholder - can be customized per page */}
        <div className="hidden lg:block">
          <h2 className="text-sm font-medium text-surface-600">
            Secure Secrets Management
          </h2>
        </div>

        {/* Right side actions */}
        <div className="flex items-center gap-4">
          {/* Notifications placeholder */}
          <button className="relative p-2 text-surface-400 hover:bg-surface-100 rounded-lg transition-colors">
            <Bell className="w-5 h-5" />
            <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full border-2 border-white"></span>
          </button>

          {/* Divider */}
          <div className="w-px h-6 bg-surface-200" />

          {/* User info */}
          <div className="hidden sm:flex items-center gap-3">
            <div className="text-right">
              <div className="text-sm font-medium text-surface-900">{user?.username}</div>
              <div className="text-xs text-surface-500 capitalize">{user?.role?.replace('_', ' ')}</div>
            </div>
            <div className="w-9 h-9 rounded-full bg-gradient-to-br from-brand-500 to-brand-700 flex items-center justify-center font-semibold text-sm text-white shadow-md">
              {user?.username.charAt(0).toUpperCase()}
            </div>
          </div>

          {/* Logout button */}
          <button
            onClick={logout}
            className="flex items-center gap-2 px-4 py-2 text-sm text-surface-600 hover:text-red-600 hover:bg-red-50 rounded-lg transition-all"
          >
            <LogOut className="w-4 h-4" />
            <span className="hidden sm:inline">Logout</span>
          </button>
        </div>
      </div>
    </header>
  )
}
