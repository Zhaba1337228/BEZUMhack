import { useState, useEffect, useRef } from 'react'
import Cookies from 'js-cookie'
import { Bell, Check, ChevronRight, AlertCircle, Info, Shield, Key } from 'lucide-react'
import { api } from '../../services/api'
import { useAuth } from '../../context/AuthContext'

export interface Notification {
  id: string
  type: 'request' | 'approval' | 'grant' | 'alert' | 'info'
  title: string
  message: string
  timestamp: string
  read: boolean
  actionUrl?: string
  secretId?: string
  requestId?: string
}

const READ_NOTIFICATIONS_COOKIE = 'sf_read_notifications'
const COOKIE_OPTIONS: Cookies.CookieAttributes = {
  expires: 7,
  secure: window.location.protocol === 'https:',
  sameSite: 'strict',
  path: '/',
}

export default function Notifications() {
  const { user } = useAuth()
  const [isOpen, setIsOpen] = useState(false)
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [loading, setLoading] = useState(false)
  const [connected, setConnected] = useState(false)
  const eventSourceRef = useRef<EventSource | null>(null)

  // Get read notification IDs from secure cookie
  const getReadNotificationIds = (): Set<string> => {
    try {
      const stored = Cookies.get(READ_NOTIFICATIONS_COOKIE)
      return stored ? new Set(JSON.parse(stored)) : new Set()
    } catch {
      return new Set()
    }
  }

  // Save read notification IDs to secure cookie
  const saveReadNotificationIds = (ids: Set<string>) => {
    try {
      Cookies.set(READ_NOTIFICATIONS_COOKIE, JSON.stringify(Array.from(ids)), COOKIE_OPTIONS)
    } catch (e) {
      console.error('Failed to save read notifications:', e)
    }
  }

  // Setup SSE connection for real-time notifications
  useEffect(() => {
    connectToEventStream()
    return () => disconnectFromEventStream()
  }, [])

  const connectToEventStream = () => {
    const token = localStorage.getItem('token')
    if (!token || !user) return

    const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
    const streamUrl = `${API_URL}/api/events/stream?token=${encodeURIComponent(token)}`
    const eventSource = new EventSource(streamUrl)

    eventSource.onopen = () => {
      setConnected(true)
      console.log('Connected to notification stream')
    }

    eventSource.onerror = (error) => {
      setConnected(false)
      console.error('Notification stream error:', error)
      // Reconnect after 5 seconds
      setTimeout(() => {
        if (eventSource.readyState === EventSource.CLOSED) {
          connectToEventStream()
        }
      }, 5000)
    }

    eventSource.addEventListener('notification', (event) => {
      try {
        const data = JSON.parse(event.data)
        addNotification({
          id: data.id || `evt-${Date.now()}`,
          type: mapEventType(data.type),
          title: data.title || 'Notification',
          message: data.message || '',
          timestamp: new Date().toLocaleString(),
          read: false,
        })
      } catch (e) {
        console.error('Failed to parse notification:', e)
      }
    })

    eventSourceRef.current = eventSource
  }

  const disconnectFromEventStream = () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close()
      eventSourceRef.current = null
    }
    setConnected(false)
  }

  const mapEventType = (type: string): Notification['type'] => {
    switch (type) {
      case 'grant': return 'grant'
      case 'alert': return 'alert'
      case 'approval': return 'approval'
      case 'request': return 'request'
      default: return 'info'
    }
  }

  const addNotification = (notification: Notification) => {
    setNotifications(prev => [notification, ...prev].slice(0, 10))
  }

  // Load notifications on mount and every 30 seconds
  useEffect(() => {
    loadNotifications()
    const interval = setInterval(loadNotifications, 30000)
    return () => clearInterval(interval)
  }, [])

  const loadNotifications = async () => {
    try {
      setLoading(true)
      const readIds = getReadNotificationIds()

      // Get recent requests for current user
      const requestsData = await api.getRequests({ mine: 'true' })

      // Get audit logs for recent activity (security_admin only)
      let auditLogs: any[] = []
      if (user?.role === 'security_admin') {
        try {
          const logsData = await api.getAuditLogs({ limit: '10' })
          auditLogs = logsData.logs || []
        } catch (e) {
          // Ignore: audit logs might be temporarily unavailable
        }
      }

      const newNotifications: Notification[] = []

      // Add recent requests as notifications (pending/approved/denied)
      requestsData.requests?.slice(0, 5).forEach((req: any) => {
        const id = `req-${req.id}`
        const status = req.status || 'pending'
        const type: Notification['type'] = status === 'approved'
          ? 'grant'
          : status === 'denied'
            ? 'alert'
            : 'request'
        const statusMessage = status === 'approved'
          ? 'Request approved'
          : status === 'denied'
            ? 'Request denied'
            : 'Pending approval'

        newNotifications.push({
          id,
          type,
          title: req.secret?.name || 'Access Request',
          message: `${statusMessage}: ${req.justification}`,
          timestamp: new Date(req.created_at).toLocaleString(),
          read: readIds.has(id),
          actionUrl: '/requests',
          requestId: req.id,
        })
      })

      // Add recent grants as notifications
      auditLogs
        .filter(log => log.action === 'access_grant_created')
        .slice(0, 3)
        .forEach((log) => {
          const id = `grant-${log.id}`
          newNotifications.push({
            id,
            type: 'grant',
            title: 'Access Granted',
            message: log.details?.message || 'New access grant created',
            timestamp: new Date(log.timestamp).toLocaleString(),
            read: readIds.has(id),
            actionUrl: '/secrets',
            secretId: log.details?.secret_id,
          })
        })

      // Add alert for critical secrets accessed
      auditLogs
        .filter(log => log.details?.classification === 'CRITICAL' && log.action === 'access_grant_created')
        .slice(0, 2)
        .forEach((log) => {
          const id = `alert-${log.id}`
          newNotifications.push({
            id,
            type: 'alert',
            title: 'Critical Secret Accessed',
            message: `${log.details?.secret_name || 'Secret'} access granted via ${log.details?.reason || 'automation'}`,
            timestamp: new Date(log.timestamp).toLocaleString(),
            read: readIds.has(id),
            actionUrl: '/audit',
          })
        })

      setNotifications(newNotifications.slice(0, 10))
    } catch (error) {
      console.error('Failed to load notifications:', error)
    } finally {
      setLoading(false)
    }
  }

  const markAsRead = (id: string) => {
    const readIds = getReadNotificationIds()
    readIds.add(id)
    saveReadNotificationIds(readIds)

    setNotifications(prev =>
      prev.map(n => n.id === id ? { ...n, read: true } : n)
    )
  }

  const markAllAsRead = () => {
    const allIds = notifications.map(n => n.id)
    const readIds = getReadNotificationIds()
    allIds.forEach(id => readIds.add(id))
    saveReadNotificationIds(readIds)

    setNotifications(prev => prev.map(n => ({ ...n, read: true })))
  }

  const handleNotificationClick = (notification: Notification) => {
    // Mark as read immediately
    markAsRead(notification.id)

    // Navigate after a short delay to ensure state update
    if (notification.actionUrl) {
      const url = notification.actionUrl
      setTimeout(() => {
        window.location.href = url
      }, 50)
    }
  }

  const unreadCount = notifications.filter(n => !n.read).length

  const getIcon = (type: Notification['type']) => {
    switch (type) {
      case 'request':
        return <AlertCircle className="w-5 h-5 text-amber-500" />
      case 'approval':
        return <Check className="w-5 h-5 text-emerald-500" />
      case 'grant':
        return <Key className="w-5 h-5 text-blue-500" />
      case 'alert':
        return <Shield className="w-5 h-5 text-red-500" />
      default:
        return <Info className="w-5 h-5 text-slate-500" />
    }
  }

  return (
    <div className="relative z-50">
      {/* Notification Bell Button */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="relative p-2 text-surface-400 hover:bg-surface-100 hover:text-surface-600 rounded-lg transition-all group cursor-pointer"
        aria-label="Notifications"
      >
        <Bell className="w-5 h-5" />
        {/* Connection status indicator */}
        <span className={`absolute -top-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-white ${
          connected ? 'bg-green-500' : 'bg-amber-500'
        }`} title={connected ? 'Live updates enabled' : 'Reconnecting...'} />
        {unreadCount > 0 && (
          <>
            <span className="absolute top-1 right-1 min-w-5 h-5 px-1 flex items-center justify-center bg-red-500 text-white text-xs font-bold rounded-full border-2 border-white shadow-sm">
              {unreadCount > 9 ? '9+' : unreadCount}
            </span>
            <span className="absolute inset-0 bg-red-400/20 rounded-lg animate-ping opacity-75" />
          </>
        )}
      </button>

      {/* Notifications Dropdown */}
      {isOpen && (
        <>
          {/* Backdrop - click outside to close */}
          <div
            className="fixed inset-0"
            onClick={() => setIsOpen(false)}
            aria-hidden="true"
          />

          {/* Dropdown Panel */}
          <div className="absolute right-0 mt-2 w-96 max-h-[500px] bg-white rounded-xl shadow-2xl border border-surface-200 overflow-hidden flex flex-col" style={{ zIndex: 9999 }}>
            {/* Header */}
            <div className="flex items-center justify-between px-4 py-3 border-b border-surface-100 bg-gradient-to-r from-surface-50 to-white">
              <div className="flex items-center gap-2">
                <Bell className="w-5 h-5 text-surface-600" />
                <h3 className="font-semibold text-surface-800">Notifications</h3>
                {unreadCount > 0 && (
                  <span className="px-2 py-0.5 bg-red-500 text-white text-xs font-bold rounded-full">
                    {unreadCount} new
                  </span>
                )}
              </div>
              {unreadCount > 0 && (
                <button
                  onClick={markAllAsRead}
                  className="text-xs text-brand-600 hover:text-brand-700 font-medium"
                >
                  Mark all read
                </button>
              )}
            </div>

            {/* Notifications List */}
            <div className="flex-1 overflow-y-auto">
              {loading && (
                <div className="flex items-center justify-center py-8">
                  <div className="w-6 h-6 border-2 border-brand-500 border-t-transparent rounded-full animate-spin" />
                </div>
              )}

              {!loading && notifications.length === 0 && (
                <div className="flex flex-col items-center justify-center py-12 text-surface-400">
                  <Bell className="w-12 h-12 mb-3 opacity-20" />
                  <p className="text-sm">No notifications</p>
                </div>
              )}

              {notifications.map((notification) => (
                <div
                  key={notification.id}
                  onClick={() => handleNotificationClick(notification)}
                  className={`flex items-start gap-3 px-4 py-3 border-b border-surface-50 cursor-pointer transition-all ${
                    !notification.read
                      ? 'bg-blue-50/50 hover:bg-blue-50'
                      : 'hover:bg-surface-50'
                  }`}
                >
                  {/* Icon */}
                  <div className="flex-shrink-0 mt-0.5">
                    {getIcon(notification.type)}
                  </div>

                  {/* Content */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-start justify-between gap-2">
                      <h4
                        className={`text-sm font-medium truncate ${
                          !notification.read ? 'text-surface-900' : 'text-surface-600'
                        }`}
                      >
                        {notification.title}
                      </h4>
                      {!notification.read && (
                        <span className="w-2 h-2 bg-blue-500 rounded-full flex-shrink-0 mt-1" />
                      )}
                    </div>
                    <p className="text-xs text-surface-500 mt-0.5 line-clamp-2">
                      {notification.message}
                    </p>
                    <div className="flex items-center gap-2 mt-1.5">
                      <span className="text-xs text-surface-400">{notification.timestamp}</span>
                      {notification.actionUrl && (
                        <span className="flex items-center gap-0.5 text-xs text-brand-600 font-medium">
                          Open <ChevronRight className="w-3 h-3" />
                        </span>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Footer */}
            {notifications.length > 0 && (
              <div className="px-4 py-2 border-t border-surface-100 bg-surface-50">
                <button
                  onClick={() => window.location.href = '/requests'}
                  className="w-full text-center text-xs text-surface-500 hover:text-surface-700 font-medium"
                >
                  View all activity {'->'}
                </button>
              </div>
            )}
          </div>
        </>
      )}
    </div>
  )
}


