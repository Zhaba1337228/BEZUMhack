import { useState, useEffect } from 'react'
import { api } from '../../services/api'
import {
  CheckCircle,
  XCircle,
  Clock,
  Shield,
  Key,
  User,
  FileText,
  Zap,
} from 'lucide-react'

export default function Approvals() {
  const [requests, setRequests] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  useEffect(() => {
    api.getRequests({ pending: 'true' })
      .then((data) => {
        setRequests(data.requests)
        setLoading(false)
      })
      .catch(console.error)
  }, [])

  const handleApprove = async (id: string) => {
    setActionLoading(id)
    try {
      await api.approveRequest(id)
      setRequests((prev) => prev.filter((r) => r.id !== id))
    } catch (err) {
      console.error('Approval failed:', err)
    } finally {
      setActionLoading(null)
    }
  }

  const handleDeny = async (id: string) => {
    setActionLoading(id)
    try {
      await api.denyRequest(id)
      setRequests((prev) => prev.filter((r) => r.id !== id))
    } catch (err) {
      console.error('Denial failed:', err)
    } finally {
      setActionLoading(null)
    }
  }

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
          <p className="text-surface-500">Loading requests...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-surface-900">Pending Approvals</h1>
          <p className="text-surface-500 mt-1">Review and approve access requests</p>
        </div>
        <div className="flex items-center gap-2 px-4 py-2 bg-brand-50 border border-brand-200 rounded-lg">
          <Clock className="w-4 h-4 text-brand-600" />
          <span className="text-sm font-medium text-brand-700">{requests.length} pending</span>
        </div>
      </div>

      {/* Requests List */}
      {requests.length === 0 ? (
        <div className="bg-white rounded-xl shadow-soft p-12 text-center border border-surface-100">
          <CheckCircle className="w-12 h-12 text-green-500 mx-auto mb-4" />
          <p className="text-surface-600 font-medium">All caught up!</p>
          <p className="text-surface-400 text-sm mt-1">No pending requests requiring approval</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4">
          {requests.map((req) => (
            <div
              key={req.id}
              className="bg-white rounded-xl shadow-soft border border-surface-100 p-6 hover:border-brand-200 transition-all"
            >
              <div className="flex items-start justify-between gap-6">
                {/* Left side - Request info */}
                <div className="flex-1 space-y-4">
                  {/* User and Secret */}
                  <div className="flex items-center gap-4">
                    <div className="p-2 bg-surface-100 rounded-lg">
                      <User className="w-5 h-5 text-surface-500" />
                    </div>
                    <div>
                      <p className="text-sm font-semibold text-surface-900">
                        {req.user?.username || 'Unknown'}
                      </p>
                      <p className="text-xs text-surface-500 capitalize">
                        {req.user?.role?.replace('_', ' ')}
                      </p>
                    </div>
                    <div className="h-8 w-px bg-surface-200 mx-2" />
                    <div className="flex items-center gap-2">
                      <Key className="w-4 h-4 text-surface-400" />
                      <p className="text-sm font-medium text-surface-900">
                        {req.secret?.name || 'Unknown'}
                      </p>
                    </div>
                  </div>

                  {/* Classification and Justification */}
                  <div className="flex items-center gap-4">
                    <span className={`px-3 py-1.5 rounded-lg text-xs font-semibold border ${getClassBg(req.secret?.classification)}`}>
                      <span className={getClassColor(req.secret?.classification)}>
                        {req.secret?.classification}
                      </span>
                    </span>
                    <div className="flex items-center gap-2 text-sm text-surface-500">
                      <FileText className="w-4 h-4" />
                      <span className="truncate max-w-md">{req.justification}</span>
                    </div>
                  </div>

                  {/* Source info */}
                  <div className="flex items-center gap-2 text-xs text-surface-500">
                    <Shield className="w-3.5 h-3.5" />
                    <span>Source: <code className="font-mono">{req.source}</code></span>
                    {req.auto_approved && (
                      <span className="flex items-center gap-1 ml-3 px-2 py-0.5 bg-green-50 text-green-700 rounded border border-green-100">
                        <Zap className="w-3 h-3" />
                        Auto-approved
                      </span>
                    )}
                    {req.source_context && (
                      <span className="text-surface-400">
                        • Context: {JSON.stringify(req.source_context)}
                      </span>
                    )}
                  </div>
                </div>

                {/* Right side - Actions */}
                <div className="flex flex-col gap-2">
                  {!req.auto_approved && req.status === 'pending' ? (
                    <>
                      <button
                        onClick={() => handleApprove(req.id)}
                        disabled={actionLoading === req.id}
                        className="flex items-center gap-2 px-4 py-2.5 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all btn-press font-medium"
                      >
                        <CheckCircle className="w-4 h-4" />
                        Approve
                      </button>
                      <button
                        onClick={() => handleDeny(req.id)}
                        disabled={actionLoading === req.id}
                        className="flex items-center gap-2 px-4 py-2.5 bg-white border border-red-200 text-red-600 rounded-lg hover:bg-red-50 disabled:opacity-50 disabled:cursor-not-allowed transition-all btn-press font-medium"
                      >
                        <XCircle className="w-4 h-4" />
                        Deny
                      </button>
                    </>
                  ) : (
                    <div className="px-4 py-2.5 text-sm text-surface-500">
                      {req.status === 'approved' && (
                        <span className="flex items-center gap-2 text-green-600">
                          <CheckCircle className="w-4 h-4" />
                          Approved
                        </span>
                      )}
                      {req.status === 'denied' && (
                        <span className="flex items-center gap-2 text-red-600">
                          <XCircle className="w-4 h-4" />
                          Denied
                        </span>
                      )}
                      {req.auto_approved && (
                        <span className="flex items-center gap-2 text-green-600">
                          <Zap className="w-4 h-4" />
                          Auto-approved
                        </span>
                      )}
                    </div>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
