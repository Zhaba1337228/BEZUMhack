import { useState, useEffect, Fragment } from 'react'
import { api } from '../../services/api'
import {
  FileText,
  Key,
  Clock,
  CheckCircle,
  XCircle,
  Shield,
  Calendar,
  Zap,
  AlertCircle,
  ChevronDown,
  Filter,
  Check,
} from 'lucide-react'
import { Menu, Transition } from '@headlessui/react'

export default function MyRequests() {
  const [requests, setRequests] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [filters, setFilters] = useState({
    status: 'all', // all, pending, approved, denied
    classification: 'all', // all, CRITICAL, HIGH, MEDIUM, LOW
    sortBy: 'newest', // newest, oldest, classification
  })

  useEffect(() => {
    loadRequests()
  }, [])

  const loadRequests = async () => {
    try {
      setLoading(true)
      const data = await api.getRequests()
      setRequests(data.requests || [])
    } catch (error) {
      console.error('Failed to load requests:', error)
    } finally {
      setLoading(false)
    }
  }

  // Filter and sort requests
  const filteredRequests = requests
    .filter(req => {
      if (filters.status !== 'all' && req.status !== filters.status) return false
      if (filters.classification !== 'all' && req.secret?.classification !== filters.classification) return false
      return true
    })
    .sort((a, b) => {
      switch (filters.sortBy) {
        case 'newest':
          return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
        case 'oldest':
          return new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
        case 'classification':
          const order = { CRITICAL: 0, HIGH: 1, MEDIUM: 2, LOW: 3 }
          return (order[a.secret?.classification as keyof typeof order] || 4) -
                 (order[b.secret?.classification as keyof typeof order] || 4)
        default:
          return 0
      }
    })

  const getStatusInfo = (status: string) => {
    switch (status) {
      case 'approved':
        return {
          icon: CheckCircle,
          color: 'text-green-600',
          bg: 'bg-green-50',
          border: 'border-green-200',
          label: 'Approved',
        }
      case 'denied':
        return {
          icon: XCircle,
          color: 'text-red-600',
          bg: 'bg-red-50',
          border: 'border-red-200',
          label: 'Denied',
        }
      default:
        return {
          icon: Clock,
          color: 'text-yellow-600',
          bg: 'bg-yellow-50',
          border: 'border-yellow-200',
          label: 'Pending',
        }
    }
  }

  const getClassBg = (classification: string) => {
    switch (classification) {
      case 'CRITICAL': return 'bg-red-100 border-red-200'
      case 'HIGH': return 'bg-orange-100 border-orange-200'
      case 'MEDIUM': return 'bg-yellow-100 border-yellow-200'
      default: return 'bg-green-100 border-green-200'
    }
  }

  // Dropdown option types
  type FilterOption = { value: string; label: string; color?: string; icon?: any }

  const statusOptions: FilterOption[] = [
    { value: 'all', label: 'All Statuses', icon: FileText },
    { value: 'pending', label: 'Pending', icon: Clock, color: 'text-yellow-600' },
    { value: 'approved', label: 'Approved', icon: CheckCircle, color: 'text-green-600' },
    { value: 'denied', label: 'Denied', icon: XCircle, color: 'text-red-600' },
  ]

  const classificationOptions: FilterOption[] = [
    { value: 'all', label: 'All Classifications' },
    { value: 'CRITICAL', label: 'Critical', color: 'text-red-700 bg-red-50' },
    { value: 'HIGH', label: 'High', color: 'text-orange-700 bg-orange-50' },
    { value: 'MEDIUM', label: 'Medium', color: 'text-yellow-700 bg-yellow-50' },
    { value: 'LOW', label: 'Low', color: 'text-green-700 bg-green-50' },
  ]

  const sortOptions: FilterOption[] = [
    { value: 'newest', label: 'Newest First' },
    { value: 'oldest', label: 'Oldest First' },
    { value: 'classification', label: 'By Classification' },
  ]

  // Reusable Dropdown component
  const FilterDropdown = ({
    options,
    value,
    onChange,
    label,
    width = 'w-48'
  }: {
    options: FilterOption[]
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
                          {!OptionIcon && option.color && (
                            <span className={`w-4 h-4 rounded ${option.color.split(' ')[1]}`} />
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

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-brand-200 border-t-brand-600 rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-surface-500">Loading your requests...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-surface-900">My Access Requests</h1>
          <p className="text-surface-500 mt-1">Track and manage your secret access requests</p>
        </div>
        <div className="flex items-center gap-2 text-sm text-surface-500">
          <FileText className="w-4 h-4" />
          <span>{filteredRequests.length} of {requests.length} requests</span>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-xl shadow-soft p-5 border border-surface-100">
        <div className="flex items-center gap-2 mb-4">
          <Filter className="w-4 h-4 text-surface-500" />
          <h2 className="text-sm font-semibold text-surface-700">Filters</h2>
        </div>
        <div className="flex flex-wrap items-center gap-4">
          {/* Status Filter */}
          <FilterDropdown
            label="Status"
            options={statusOptions}
            value={filters.status}
            onChange={(val) => setFilters({ ...filters, status: val })}
            width="w-44"
          />

          {/* Classification Filter */}
          <FilterDropdown
            label="Classification"
            options={classificationOptions}
            value={filters.classification}
            onChange={(val) => setFilters({ ...filters, classification: val })}
            width="w-48"
          />

          {/* Sort Filter */}
          <FilterDropdown
            label="Sort"
            options={sortOptions}
            value={filters.sortBy}
            onChange={(val) => setFilters({ ...filters, sortBy: val })}
            width="w-44"
          />

          {/* Clear Filters */}
          <button
            onClick={() => setFilters({ status: 'all', classification: 'all', sortBy: 'newest' })}
            className="ml-auto flex items-center gap-1.5 px-4 py-2.5 text-sm text-surface-600 hover:text-red-600 hover:bg-red-50 rounded-xl transition-all font-medium"
          >
            <XCircle className="w-4 h-4" />
            Clear all
          </button>
        </div>
      </div>

      {/* Requests List */}
      {requests.length === 0 ? (
        <div className="bg-white rounded-xl shadow-soft p-12 text-center border border-surface-100">
          <FileText className="w-12 h-12 text-surface-300 mx-auto mb-4" />
          <p className="text-surface-600 font-medium">No requests yet</p>
          <p className="text-surface-400 text-sm mt-1">
            Request access to secrets to see your requests here
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4">
          {requests.map((req) => {
            const StatusInfo = getStatusInfo(req.status)
            const Icon = StatusInfo.icon
            return (
              <div
                key={req.id}
                className="bg-white rounded-xl shadow-soft border border-surface-100 p-6 hover:border-brand-200 transition-all"
              >
                <div className="flex items-center justify-between gap-6">
                  {/* Left side - Request details */}
                  <div className="flex-1 flex items-center gap-4">
                    {/* Status Icon */}
                    <div className={`p-3 rounded-xl ${StatusInfo.bg} border ${StatusInfo.border}`}>
                      <Icon className={`w-6 h-6 ${StatusInfo.color}`} />
                    </div>

                    {/* Info */}
                    <div className="flex-1 space-y-2">
                      {/* Secret name and classification */}
                      <div className="flex items-center gap-3">
                        <Key className="w-4 h-4 text-surface-400" />
                        <span className="font-semibold text-surface-900">
                          {req.secret?.name || 'Unknown Secret'}
                        </span>
                        {req.secret?.classification && (
                          <span className={`px-2.5 py-1 rounded-lg text-xs font-semibold border ${getClassBg(req.secret.classification)}`}>
                            <span className={
                              req.secret.classification === 'CRITICAL' ? 'text-red-700' :
                              req.secret.classification === 'HIGH' ? 'text-orange-700' :
                              req.secret.classification === 'MEDIUM' ? 'text-yellow-700' :
                              'text-green-700'
                            }>
                              {req.secret.classification}
                            </span>
                          </span>
                        )}
                      </div>

                      {/* Justification */}
                      <div className="flex items-center gap-2 text-sm text-surface-500">
                        <AlertCircle className="w-4 h-4" />
                        <span className="truncate max-w-lg italic">"{req.justification}"</span>
                      </div>

                      {/* Meta info */}
                      <div className="flex items-center gap-4 text-xs text-surface-400">
                        <span className="flex items-center gap-1.5">
                          <Calendar className="w-3.5 h-3.5" />
                          {new Date(req.created_at).toLocaleDateString('en-US', {
                            year: 'numeric',
                            month: 'short',
                            day: 'numeric',
                          })}
                        </span>
                        <span className="flex items-center gap-1.5">
                          <Shield className="w-3.5 h-3.5" />
                          Source: <code className="font-mono">{req.source}</code>
                        </span>
                        {req.auto_approved && (
                          <span className="flex items-center gap-1.5 px-2 py-0.5 bg-green-50 text-green-700 rounded border border-green-100">
                            <Zap className="w-3 h-3" />
                            Auto-approved
                          </span>
                        )}
                      </div>
                    </div>
                  </div>

                  {/* Right side - Status badge */}
                  <div className="flex-shrink-0">
                    <span className={`flex items-center gap-2 px-4 py-2.5 rounded-lg text-sm font-semibold border ${StatusInfo.bg} ${StatusInfo.border} ${StatusInfo.color}`}>
                      <Icon className="w-4 h-4" />
                      {StatusInfo.label}
                    </span>
                  </div>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
