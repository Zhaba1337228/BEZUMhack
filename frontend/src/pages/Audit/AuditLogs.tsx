import { useState, useEffect } from 'react'
import { api } from '../../services/api'
import {
  ScrollText,
  Search,
  Filter,
  Shield,
  Key,
  UserCheck,
  UserX,
  LogIn,
  LogOut,
  AlertCircle,
  Download,
} from 'lucide-react'

export default function AuditLogs() {
  const [logs, setLogs] = useState<any[]>([])
  const [filter, setFilter] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api.getAuditLogs({ limit: '100' })
      .then((data) => {
        setLogs(data.logs)
        setLoading(false)
      })
      .catch((err) => {
        setError(err.message || 'Failed to load audit logs')
        setLoading(false)
      })
  }, [])

  const filteredLogs = logs.filter((log) =>
    filter === '' ||
    log.action.toLowerCase().includes(filter.toLowerCase()) ||
    log.resource_type?.toLowerCase().includes(filter.toLowerCase())
  )

  const getActionIcon = (action: string) => {
    if (action.includes('login')) return { icon: LogIn, color: 'text-green-600', bg: 'bg-green-50' }
    if (action.includes('logout')) return { icon: LogOut, color: 'text-surface-600', bg: 'bg-surface-50' }
    if (action.includes('approved')) return { icon: UserCheck, color: 'text-blue-600', bg: 'bg-blue-50' }
    if (action.includes('denied')) return { icon: UserX, color: 'text-red-600', bg: 'bg-red-50' }
    if (action.includes('secret')) return { icon: Key, color: 'text-purple-600', bg: 'bg-purple-50' }
    if (action.includes('grant')) return { icon: Shield, color: 'text-green-600', bg: 'bg-green-50' }
    return { icon: AlertCircle, color: 'text-surface-600', bg: 'bg-surface-50' }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-brand-200 border-t-brand-600 rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-surface-500">Loading audit logs...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
          <p className="text-surface-900 font-medium">Access Denied</p>
          <p className="text-surface-500 text-sm mt-1">{error}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-surface-900">Audit Logs</h1>
          <p className="text-surface-500 mt-1">Track all system activity and access</p>
        </div>
        <button className="flex items-center gap-2 px-4 py-2 bg-white border border-surface-200 rounded-lg text-sm font-medium text-surface-600 hover:bg-surface-50 transition-colors">
          <Download className="w-4 h-4" />
          <span>Export</span>
        </button>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-surface-400" />
          <input
            type="text"
            placeholder="Search actions, resources..."
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="w-full pl-10 pr-4 py-2.5 border border-surface-200 rounded-lg input-focus bg-white text-surface-900 placeholder:text-surface-400"
          />
        </div>
        <button className="flex items-center gap-2 px-4 py-2.5 bg-white border border-surface-200 rounded-lg text-sm font-medium text-surface-600 hover:bg-surface-50 transition-colors">
          <Filter className="w-4 h-4" />
          <span>Filters</span>
        </button>
      </div>

      {/* Logs Table */}
      <div className="bg-white rounded-xl shadow-soft border border-surface-100 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-surface-100">
            <thead className="bg-surface-50">
              <tr>
                <th className="px-6 py-4 text-left text-xs font-semibold text-surface-500 uppercase tracking-wider">
                  Timestamp
                </th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-surface-500 uppercase tracking-wider">
                  Action
                </th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-surface-500 uppercase tracking-wider">
                  Resource
                </th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-surface-500 uppercase tracking-wider">
                  Details
                </th>
                <th className="px-6 py-4 text-left text-xs font-semibold text-surface-500 uppercase tracking-wider">
                  IP Address
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-100">
              {filteredLogs.map((log) => {
                const IconInfo = getActionIcon(log.action)
                return (
                  <tr key={log.id} className="hover:bg-surface-50 transition-colors">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="text-sm font-medium text-surface-900">
                        {new Date(log.timestamp).toLocaleDateString('en-US', {
                          year: 'numeric',
                          month: 'short',
                          day: 'numeric',
                        })}
                      </div>
                      <div className="text-xs text-surface-500">
                        {new Date(log.timestamp).toLocaleTimeString('en-US', {
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-lg ${IconInfo.bg}`}>
                          <IconInfo.icon className={`w-4 h-4 ${IconInfo.color}`} />
                        </div>
                        <span className="text-sm font-medium text-surface-900">{log.action}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      {log.resource_type ? (
                        <div className="text-sm text-surface-600">
                          {log.resource_type}
                          {log.resource_id && (
                            <span className="text-xs text-surface-400 font-mono ml-1">
                              {log.resource_id.slice(0, 8)}...
                            </span>
                          )}
                        </div>
                      ) : (
                        <span className="text-sm text-surface-400">—</span>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      {log.details ? (
                        <details className="group">
                          <summary className="text-xs text-brand-600 cursor-pointer hover:text-brand-700 font-medium">
                            View details
                          </summary>
                          <pre className="mt-2 text-xs text-surface-600 font-mono bg-surface-50 p-3 rounded-lg max-w-md overflow-auto">
                            {JSON.stringify(log.details, null, 2)}
                          </pre>
                        </details>
                      ) : (
                        <span className="text-sm text-surface-400">—</span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <code className="text-xs font-mono text-surface-600 bg-surface-100 px-2 py-1 rounded">
                        {log.ip_address || '—'}
                      </code>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>

        {filteredLogs.length === 0 && (
          <div className="p-12 text-center">
            <ScrollText className="w-12 h-12 text-surface-300 mx-auto mb-4" />
            <p className="text-surface-600 font-medium">No audit logs found</p>
            <p className="text-surface-400 text-sm mt-1">
              {filter ? 'Try adjusting your search' : 'System activity will appear here'}
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
