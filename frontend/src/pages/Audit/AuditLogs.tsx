import { useEffect, useMemo, useState, Fragment } from 'react'
import { api } from '../../services/api'
import {
  ScrollText,
  Search,
  Shield,
  Key,
  UserCheck,
  UserX,
  LogIn,
  AlertCircle,
  Download,
  Clock3,
  RefreshCw,
  Filter,
  Activity,
  ChevronDown,
  Check,
  XCircle,
} from 'lucide-react'
import { Menu, Transition } from '@headlessui/react'

export default function AuditLogs() {
  const [logs, setLogs] = useState<any[]>([])
  const [timeline, setTimeline] = useState<any[]>([])
  const [timelineSecret, setTimelineSecret] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [timelineLoading, setTimelineLoading] = useState(false)
  const [error, setError] = useState('')

  const [search, setSearch] = useState('')
  const [actionFilter, setActionFilter] = useState('')
  const [secretIdFilter, setSecretIdFilter] = useState('')
  const [timelineSecretId, setTimelineSecretId] = useState('')
  const [riskyOnly, setRiskyOnly] = useState(false)

  // Action type options for dropdown
  type ActionOption = { value: string; label: string; icon?: any; color?: string }

  const actionOptions: ActionOption[] = [
    { value: '', label: 'All Actions', icon: Filter },
    { value: 'access_grant_created', label: 'Grant Created', icon: UserCheck, color: 'text-green-600' },
    { value: 'access_request_approved', label: 'Request Approved', icon: Check, color: 'text-blue-600' },
    { value: 'access_request_denied', label: 'Request Denied', icon: XCircle, color: 'text-red-600' },
    { value: 'secret_value_revealed', label: 'Secret Revealed', icon: Key, color: 'text-violet-600' },
    { value: 'login_success', label: 'Login Success', icon: LogIn, color: 'text-emerald-600' },
    { value: 'login_failure', label: 'Login Failure', icon: AlertCircle, color: 'text-amber-600' },
    { value: 'internal_api_call', label: 'Internal API', icon: Shield, color: 'text-amber-600' },
    { value: 'integration_token_used', label: 'Token Used', icon: Shield, color: 'text-red-600' },
  ]

  // Reusable Dropdown component for filters
  const FilterDropdown = ({
    options,
    value,
    onChange,
    label,
    width = 'w-56'
  }: {
    options: ActionOption[]
    value: string
    onChange: (val: string) => void
    label: string
    width?: string
  }) => {
    const selectedOption = options.find(opt => opt.value === value)
    const IconComponent = selectedOption?.icon

    return (
      <div className="flex items-center gap-2">
        <label className="text-sm font-medium text-surface-600">{label}:</label>
        <Menu as="div" className={`relative ${width}`}>
          <Menu.Button className="w-full flex items-center justify-between gap-2 px-4 py-2.5 bg-white border border-surface-200 rounded-xl text-sm font-medium text-surface-700 hover:bg-surface-50 hover:border-surface-300 focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent transition-all shadow-sm">
            <span className="flex items-center gap-2">
              {IconComponent && <IconComponent className={`w-4 h-4 ${selectedOption?.color || 'text-surface-500'}`} />}
              <span className={selectedOption?.color || ''}>{selectedOption?.label}</span>
            </span>
            <ChevronDown className="w-4 h-4 text-surface-400 transition-transform" />
          </Menu.Button>

          <Transition
            as={Fragment}
            enter="transition ease-out duration-100"
            enterFrom="transform opacity-0 scale-95"
            enterTo="transform opacity-100 scale-100"
            leave="transition ease-in duration-75"
            leaveFrom="transform opacity-100 scale-100"
            leaveTo="transform opacity-0 scale-95"
          >
            <Menu.Items className="absolute z-50 mt-2 w-full bg-white rounded-xl shadow-lg border border-surface-100 focus:outline-none overflow-hidden">
              <div className="p-1">
                {options.map((option) => {
                  const OptionIcon = option.icon
                  return (
                    <Menu.Item key={option.value}>
                      {({ active }) => (
                        <button
                          onClick={() => onChange(option.value)}
                          className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-colors ${
                            active ? 'bg-brand-50' : ''
                          } ${
                            value === option.value ? 'bg-brand-100 text-brand-900' : 'text-surface-700'
                          }`}
                        >
                          {OptionIcon && (
                            <OptionIcon className={`w-4 h-4 ${option.color || 'text-surface-500'}`} />
                          )}
                          <span className="flex-1 text-left">{option.label}</span>
                          {value === option.value && (
                            <Check className="w-4 h-4 text-brand-600" />
                          )}
                        </button>
                      )}
                    </Menu.Item>
                  )
                })}
              </div>
            </Menu.Items>
          </Transition>
        </Menu>
      </div>
    )
  }

  const loadLogs = async () => {
    try {
      setError('')
      setLoading(true)
      const params: Record<string, string> = { limit: '300' }
      if (actionFilter.trim()) params.action = actionFilter.trim()
      if (secretIdFilter.trim()) params.secret_id = secretIdFilter.trim()
      if (riskyOnly) params.risky = 'true'
      const data = await api.getAuditLogs(params)
      setLogs(data.logs || [])
    } catch (err: any) {
      setError(err.message || 'Failed to load audit logs')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadLogs()
  }, [])

  useEffect(() => {
    // Auto-apply risky toggle for better UX
    loadLogs()
  }, [riskyOnly])

  const filteredLogs = useMemo(() => {
    if (!search.trim()) return logs
    const q = search.toLowerCase()
    return logs.filter((log) =>
      log.action?.toLowerCase().includes(q) ||
      log.resource_type?.toLowerCase().includes(q) ||
      log.username?.toLowerCase().includes(q) ||
      log.purpose?.toLowerCase().includes(q) ||
      log.secret_id?.toLowerCase().includes(q)
    )
  }, [logs, search])

  const loadTimeline = async () => {
    if (!timelineSecretId.trim()) return
    try {
      setError('')
      setTimelineLoading(true)
      setTimelineSecret(null)
      setTimeline([])
      const data = await api.getAuditTimeline(timelineSecretId.trim(), { limit: '400' })
      setTimelineSecret(data.secret)
      setTimeline(data.timeline || [])
    } catch (err: any) {
      setError(err.message || 'Failed to load timeline')
    } finally {
      setTimelineLoading(false)
    }
  }

  const exportCsv = async () => {
    try {
      setError('')
      const params: Record<string, string> = { limit: '5000' }
      if (actionFilter.trim()) params.action = actionFilter.trim()
      if (secretIdFilter.trim()) params.secret_id = secretIdFilter.trim()
      if (riskyOnly) params.risky = 'true'
      const a = document.createElement('a')
      a.href = await api.exportAuditCsv(params)
      a.download = `audit_logs_${new Date().toISOString().replace(/[:.]/g, '-')}.csv`
      document.body.appendChild(a)
      a.click()
      a.remove()
    } catch (err: any) {
      setError(err.message || 'Failed to export CSV')
    }
  }

  const getActionIcon = (action: string) => {
    if (action.includes('login')) return { icon: LogIn, color: 'text-green-600', bg: 'bg-green-50' }
    if (action.includes('approved')) return { icon: UserCheck, color: 'text-blue-600', bg: 'bg-blue-50' }
    if (action.includes('denied')) return { icon: UserX, color: 'text-red-600', bg: 'bg-red-50' }
    if (action.includes('secret')) return { icon: Key, color: 'text-violet-600', bg: 'bg-violet-50' }
    if (action.includes('grant')) return { icon: Shield, color: 'text-emerald-600', bg: 'bg-emerald-50' }
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

  return (
    <div className="space-y-6 animate-fade-in">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-surface-900">Audit Center</h1>
          <p className="text-surface-500 mt-1">Полная история действий: кто, когда, зачем и с каким результатом</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={loadLogs}
            className="flex items-center gap-2 px-4 py-2 bg-white border border-surface-200 rounded-lg text-sm font-medium text-surface-600 hover:bg-surface-50"
          >
            <RefreshCw className="w-4 h-4" />
            Refresh
          </button>
          <button
            onClick={exportCsv}
            className="flex items-center gap-2 px-4 py-2 bg-brand-600 border border-brand-600 rounded-lg text-sm font-medium text-white hover:bg-brand-700"
          >
            <Download className="w-4 h-4" />
            Export CSV
          </button>
        </div>
      </div>

      {error && (
        <div className="p-4 border border-red-200 bg-red-50 rounded-lg text-red-700 text-sm">
          {error}
        </div>
      )}

      {/* Filters */}
      <div className="bg-white rounded-xl shadow-soft p-5 border border-slate-200">
        <div className="flex items-center gap-2 mb-4">
          <Filter className="w-4 h-4 text-slate-700" />
          <h2 className="text-sm font-semibold text-slate-900">Filters</h2>
        </div>
        <div className="flex flex-wrap items-center gap-4">
          {/* Search */}
          <div className="relative flex-1 min-w-64">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
            <input
              type="text"
              placeholder="Search by action, user, purpose..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-10 pr-4 py-2.5 border border-slate-300 rounded-xl text-sm text-slate-900 placeholder:text-slate-500 hover:border-slate-400 focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent transition-all"
            />
          </div>

          {/* Action Filter */}
          <FilterDropdown
            label="Action"
            options={actionOptions}
            value={actionFilter}
            onChange={(val) => setActionFilter(val)}
            width="w-52"
          />

          {/* Risky Toggle */}
          <label
            className={`group flex items-center gap-3 px-3.5 py-2.5 rounded-xl border cursor-pointer select-none transition-all ${
              riskyOnly
                ? 'bg-red-50 border-red-300 text-red-700'
                : 'bg-white border-slate-300 text-slate-800 hover:bg-slate-50'
            }`}
          >
            <input
              type="checkbox"
              checked={riskyOnly}
              onChange={(e) => {
                setRiskyOnly(e.target.checked)
                setTimeout(loadLogs, 50)
              }}
              className="sr-only peer"
            />
            <span
              className={`w-5 h-5 rounded-md border flex items-center justify-center transition-all ${
                riskyOnly
                  ? 'bg-red-600 border-red-600 text-white'
                  : 'bg-white border-slate-400 text-transparent group-hover:border-slate-500'
              } peer-focus:ring-2 peer-focus:ring-brand-500 peer-focus:ring-offset-1`}
            >
              <Check className="w-3.5 h-3.5" />
            </span>
            <Shield className={`w-4 h-4 ${riskyOnly ? 'text-red-600' : 'text-slate-500'}`} />
            <span className="text-sm font-medium">Risky only</span>
          </label>

          {/* Apply & Clear */}
          <div className="flex items-center gap-2 ml-auto">
            <button
              onClick={loadLogs}
              className="flex items-center gap-2 px-4 py-2.5 bg-brand-600 text-white rounded-xl text-sm font-medium hover:bg-brand-700 transition-all"
            >
              <Filter className="w-4 h-4" />
              Apply
            </button>
            <button
              onClick={() => {
                setActionFilter('')
                setSecretIdFilter('')
                setSearch('')
                setRiskyOnly(false)
                loadLogs()
              }}
              className="flex items-center gap-2 px-4 py-2.5 text-sm text-slate-700 hover:text-red-600 hover:bg-red-50 rounded-xl transition-all font-medium"
            >
              <XCircle className="w-4 h-4" />
              Clear
            </button>
          </div>
        </div>

        {/* Second row - Secret ID filters */}
        <div className="flex flex-wrap items-center gap-4 mt-4 pt-4 border-t border-slate-200">
          <div className="flex items-center gap-2">
            <label className="text-sm font-medium text-slate-800">Secret ID (logs):</label>
            <input
              type="text"
              placeholder="Enter secret UUID"
              value={secretIdFilter}
              onChange={(e) => setSecretIdFilter(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') loadLogs()
              }}
              className="w-64 px-4 py-2.5 border border-slate-300 rounded-xl text-sm text-slate-900 placeholder:text-slate-500 hover:border-slate-400 focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent transition-all"
            />
          </div>

          <div className="flex items-center gap-2">
            <label className="text-sm font-medium text-slate-800">Secret ID (timeline):</label>
            <input
              type="text"
              placeholder="Enter secret UUID"
              value={timelineSecretId}
              onChange={(e) => setTimelineSecretId(e.target.value)}
              className="w-64 px-4 py-2.5 border border-slate-300 rounded-xl text-sm text-slate-900 placeholder:text-slate-500 hover:border-slate-400 focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent transition-all"
            />
            <button
              onClick={loadTimeline}
              disabled={timelineLoading || !timelineSecretId.trim()}
              className="flex items-center gap-2 px-4 py-2.5 bg-white border border-slate-300 rounded-xl text-sm font-medium text-slate-900 hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed transition-all"
            >
              <Activity className="w-4 h-4" />
              Build Timeline
            </button>
          </div>
        </div>
      </div>

      {timelineSecret && (
        <div className="bg-white rounded-xl border border-surface-100 overflow-hidden">
          <div className="px-5 py-4 border-b border-surface-100 flex items-center justify-between">
            <div>
              <h2 className="font-semibold text-surface-900">Timeline: {timelineSecret.name}</h2>
              <p className="text-xs text-surface-500 mt-1">{timelineSecret.id} • {timelineSecret.classification} • {timelineSecret.environment}</p>
            </div>
            {timelineLoading && <span className="text-xs text-surface-500">Loading...</span>}
          </div>
          <div className="max-h-80 overflow-auto">
            {timeline.length === 0 && !timelineLoading ? (
              <div className="p-6 text-sm text-surface-500">No timeline events found for this secret.</div>
            ) : (
              <div className="divide-y divide-surface-100">
                {timeline.map((ev, idx) => (
                  <div key={`${ev.timestamp}-${idx}`} className="px-5 py-3">
                    <div className="flex items-center justify-between gap-3">
                      <div className="text-sm font-medium text-surface-900">{ev.event_type}</div>
                      <div className="text-xs text-surface-500 flex items-center gap-1"><Clock3 className="w-3 h-3" />{new Date(ev.timestamp).toLocaleString()}</div>
                    </div>
                    <div className="text-xs text-surface-600 mt-1">
                      Actor: {ev.actor_username || ev.actor_id || 'system'}
                      {ev.target_user ? ` • Target: ${ev.target_user}` : ''}
                      {ev.status ? ` • Status: ${ev.status}` : ''}
                      {ev.justification ? ` • Why: ${ev.justification}` : ''}
                      {ev.source ? ` • Source: ${ev.source}` : ''}
                      {ev.ip_address ? ` • IP: ${ev.ip_address}` : ''}
                    </div>
                    {ev.details && <div className="text-xs text-surface-500 mt-1">{ev.details}</div>}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )}

      <div className="bg-white rounded-xl shadow-soft border border-surface-100 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-surface-100">
            <thead className="bg-surface-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-semibold text-slate-700 uppercase">Timestamp</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-slate-700 uppercase">Action</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-slate-700 uppercase">Who</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-slate-700 uppercase">Resource</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-slate-700 uppercase">Why / Purpose</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-slate-700 uppercase">Risk</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-100">
              {filteredLogs.map((log) => {
                const IconInfo = getActionIcon(log.action)
                return (
                  <tr key={log.id} className="hover:bg-surface-50 transition-colors align-top">
                    <td className="px-4 py-3 text-sm text-slate-800 whitespace-nowrap">{new Date(log.timestamp).toLocaleString()}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <div className={`p-1.5 rounded-md ${IconInfo.bg}`}><IconInfo.icon className={`w-4 h-4 ${IconInfo.color}`} /></div>
                        <span className="text-sm font-medium text-surface-900">{log.action}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-sm text-slate-800">{log.username || log.user_id || 'system'}</td>
                    <td className="px-4 py-3 text-sm text-slate-700">
                      <div>{log.resource_type || '-'}</div>
                      <div className="font-mono text-xs text-slate-500 break-all">{log.resource_id || '-'}</div>
                      {log.secret_id && <div className="font-mono text-brand-700 mt-1">secret: {log.secret_id}</div>}
                    </td>
                    <td className="px-4 py-3 text-sm text-slate-800">
                      <div className="max-w-md break-words">{log.purpose || '-'}</div>
                      {log.details && (
                        <details className="mt-1">
                          <summary className="cursor-pointer text-brand-700 font-medium">Full details</summary>
                          <pre className="mt-1 p-2 bg-surface-50 rounded max-w-xl overflow-auto text-xs text-slate-700 whitespace-pre-wrap break-words">{JSON.stringify(log.details, null, 2)}</pre>
                        </details>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {log.risky ? (
                        <span className="inline-flex items-center px-2 py-1 rounded text-xs font-semibold bg-red-100 text-red-700">RISKY</span>
                      ) : (
                        <span className="inline-flex items-center px-2 py-1 rounded text-xs font-semibold bg-emerald-100 text-emerald-700">normal</span>
                      )}
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
            <p className="text-surface-400 text-sm mt-1">Try changing filters or search query</p>
          </div>
        )}
      </div>
    </div>
  )
}
