import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../../services/api'
import { Key, Search, Shield, AlertCircle, CheckCircle, Clock, Lock } from 'lucide-react'

export default function SecretsList() {
  const [secrets, setSecrets] = useState<any[]>([])
  const [filter, setFilter] = useState('')
  const [classFilter, setClassFilter] = useState('all')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.getSecrets()
      .then((data) => {
        setSecrets(data.secrets)
        setLoading(false)
      })
      .catch(console.error)
  }, [])

  const filteredSecrets = secrets.filter((s) => {
    const matchesSearch = s.name.toLowerCase().includes(filter.toLowerCase()) ||
      s.description?.toLowerCase().includes(filter.toLowerCase())
    const matchesClass = classFilter === 'all' || s.classification === classFilter
    return matchesSearch && matchesClass
  })

  const getClassColor = (classification: string) => {
    switch (classification) {
      case 'CRITICAL': return 'text-secret-critical'
      case 'HIGH': return 'text-secret-high'
      case 'MEDIUM': return 'text-secret-medium'
      default: return 'text-secret-low'
    }
  }

  const getClassBg = (classification: string) => {
    switch (classification) {
      case 'CRITICAL': return 'bg-red-100 text-red-700 border-red-200'
      case 'HIGH': return 'bg-orange-100 text-orange-700 border-orange-200'
      case 'MEDIUM': return 'bg-yellow-100 text-yellow-700 border-yellow-200'
      default: return 'bg-green-100 text-green-700 border-green-200'
    }
  }

  const getEnvBadge = (env: string) => {
    switch (env) {
      case 'production':
        return <span className="px-2 py-1 rounded-md text-xs font-medium bg-red-50 text-red-700 border border-red-100">Production</span>
      case 'staging':
        return <span className="px-2 py-1 rounded-md text-xs font-medium bg-yellow-50 text-yellow-700 border border-yellow-100">Staging</span>
      default:
        return <span className="px-2 py-1 rounded-md text-xs font-medium bg-surface-100 text-surface-600 border border-surface-200">Development</span>
    }
  }

  const AccessBadge = ({ secret }: { secret: any }) => {
    if (secret.has_access) {
      return (
        <span className="flex items-center gap-1.5 text-xs font-medium text-green-700 bg-green-50 px-2.5 py-1.5 rounded-md border border-green-100">
          <CheckCircle className="w-3.5 h-3.5" />
          Access Granted
        </span>
      )
    }
    if (secret.pending_request) {
      return (
        <span className="flex items-center gap-1.5 text-xs font-medium text-yellow-700 bg-yellow-50 px-2.5 py-1.5 rounded-md border border-yellow-100">
          <Clock className="w-3.5 h-3.5" />
          Pending Approval
        </span>
      )
    }
    return (
      <span className="flex items-center gap-1.5 text-xs font-medium text-surface-600 bg-surface-100 px-2.5 py-1.5 rounded-md border border-surface-200">
        <Lock className="w-3.5 h-3.5" />
        Request Access
      </span>
    )
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-brand-200 border-t-brand-600 rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-surface-500">Loading secrets catalog...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-surface-900">Secrets Catalog</h1>
          <p className="text-surface-500 mt-1">Browse and request access to secrets</p>
        </div>
        <div className="flex items-center gap-2 text-sm text-surface-500">
          <Shield className="w-4 h-4" />
          <span>{filteredSecrets.length} secrets available</span>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-surface-400" />
          <input
            type="text"
            placeholder="Search secrets by name or description..."
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="w-full pl-10 pr-4 py-2.5 border border-surface-200 rounded-lg input-focus bg-white text-surface-900 placeholder:text-surface-400"
          />
        </div>
        <div className="flex gap-2">
          {['all', 'LOW', 'MEDIUM', 'HIGH', 'CRITICAL'].map((cls) => (
            <button
              key={cls}
              onClick={() => setClassFilter(cls)}
              className={`px-4 py-2.5 rounded-lg text-sm font-medium transition-all ${
                classFilter === cls
                  ? 'bg-brand-600 text-white shadow-md'
                  : 'bg-white text-surface-600 border border-surface-200 hover:bg-surface-50'
              }`}
            >
              {cls === 'all' ? 'All' : cls}
            </button>
          ))}
        </div>
      </div>

      {/* Secrets Grid */}
      {filteredSecrets.length === 0 ? (
        <div className="bg-white rounded-xl shadow-soft p-12 text-center border border-surface-100">
          <Key className="w-12 h-12 text-surface-300 mx-auto mb-4" />
          <p className="text-surface-600 font-medium">No secrets found</p>
          <p className="text-surface-400 text-sm mt-1">Try adjusting your search or filters</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {filteredSecrets.map((secret) => (
            <Link
              key={secret.id}
              to={`/secrets/${secret.id}`}
              className="group bg-white rounded-xl shadow-soft p-5 border border-surface-100 hover:border-brand-200 hover:shadow-lg transition-all card-hover"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-2 flex-1 min-w-0">
                  <div className={`p-2 rounded-lg ${getClassBg(secret.classification)}`}>
                    <Key className={`w-4 h-4 ${getClassColor(secret.classification)}`} />
                  </div>
                  <h3 className="font-semibold text-surface-900 truncate flex-1">{secret.name}</h3>
                </div>
              </div>

              <p className="text-sm text-surface-500 mb-4 line-clamp-2 min-h-[2.5rem]">
                {secret.description || 'No description provided'}
              </p>

              <div className="flex items-center justify-between pt-4 border-t border-surface-100">
                <div className="flex items-center gap-2">
                  {getEnvBadge(secret.environment)}
                  <span className={`px-2 py-1 rounded-md text-xs font-medium border ${getClassBg(secret.classification)}`}>
                    {secret.classification}
                  </span>
                </div>
                <AccessBadge secret={secret} />
              </div>

              {/* Hover indicator */}
              <div className="mt-3 flex items-center gap-1 text-sm text-brand-600 opacity-0 group-hover:opacity-100 transition-opacity">
                <span>View details</span>
                <AlertCircle className="w-4 h-4" />
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
