import { useState, useEffect } from 'react'
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
} from 'lucide-react'

export default function MyRequests() {
  const [requests, setRequests] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.getRequests()
      .then((data) => {
        setRequests(data.requests)
        setLoading(false)
      })
      .catch(console.error)
  }, [])

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
          <span>{requests.length} requests</span>
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
