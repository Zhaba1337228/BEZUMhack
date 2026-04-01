import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../../services/api'
import {
  ArrowLeft,
  Key,
  Shield,
  Users,
  Eye,
  Clock,
  CheckCircle,
  AlertTriangle,
  Lock,
  Copy,
  Check,
} from 'lucide-react'

export default function SecretDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [secret, setSecret] = useState<any>(null)
  const [secretValue, setSecretValue] = useState<string | null>(null)
  const [justification, setJustification] = useState('')
  const [requestStatus, setRequestStatus] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [actionLoading, setActionLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (!id) return
    api.getSecret(id)
      .then((data) => {
        setSecret(data.secret)
        setLoading(false)
      })
      .catch(console.error)
  }, [id])

  const handleRequestAccess = async () => {
    if (!id || !justification.trim()) return
    setActionLoading(true)
    setError('')
    try {
      const data = await api.requestAccess(id, justification)
      const status = data?.request?.status || 'pending'
      setSuccess('Access request submitted successfully')
      setRequestStatus(status)
      if (status === 'approved') {
        setSecret((prev: any) => prev ? { ...prev, has_access: true } : prev)
      }
      setJustification('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Request failed')
    } finally {
      setActionLoading(false)
    }
  }

  const handleGetValue = async () => {
    if (!id) return
    setActionLoading(true)
    try {
      const data = await api.getSecretValue(id)
      setSecretValue(data.secret.value)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to retrieve value')
    } finally {
      setActionLoading(false)
    }
  }

  const handleCopy = () => {
    if (secretValue) {
      navigator.clipboard.writeText(secretValue)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-brand-200 border-t-brand-600 rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-surface-500">Loading secret details...</p>
        </div>
      </div>
    )
  }

  if (!secret) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <AlertTriangle className="w-12 h-12 text-yellow-500 mx-auto mb-4" />
          <p className="text-surface-900 font-medium">Secret not found</p>
          <button
            onClick={() => navigate('/secrets')}
            className="mt-4 text-brand-600 hover:text-brand-700 font-medium"
          >
            Back to Secrets
          </button>
        </div>
      </div>
    )
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
      case 'CRITICAL': return 'bg-red-100 text-red-700 border-red-200'
      case 'HIGH': return 'bg-orange-100 text-orange-700 border-orange-200'
      case 'MEDIUM': return 'bg-yellow-100 text-yellow-700 border-yellow-200'
      default: return 'bg-green-100 text-green-700 border-green-200'
    }
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6 animate-fade-in">
      {/* Back button */}
      <button
        onClick={() => navigate('/secrets')}
        className="flex items-center gap-2 text-surface-600 hover:text-surface-900 transition-colors"
      >
        <ArrowLeft className="w-4 h-4" />
        <span className="font-medium">Back to Secrets</span>
      </button>

      {/* Main Card */}
      <div className="bg-white rounded-xl shadow-soft border border-surface-100 overflow-hidden">
        {/* Header */}
        <div className="p-6 border-b border-surface-100">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-4 flex-1">
              <div className={`p-3 rounded-xl ${getClassBg(secret.classification)}`}>
                <Key className={`w-6 h-6 ${getClassColor(secret.classification)}`} />
              </div>
              <div>
                <h1 className="text-2xl font-bold text-surface-900">{secret.name}</h1>
                <p className="text-surface-500 mt-1">{secret.description}</p>
              </div>
            </div>
            <span className={`px-4 py-2 rounded-lg text-sm font-semibold border ${getClassBg(secret.classification)}`}>
              {secret.classification}
            </span>
          </div>
        </div>

        {/* Metadata */}
        <div className="p-6 grid grid-cols-1 sm:grid-cols-2 gap-6 border-b border-surface-100">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-surface-100 rounded-lg">
              <Shield className="w-5 h-5 text-surface-500" />
            </div>
            <div>
              <p className="text-xs text-surface-500 font-medium">Environment</p>
              <p className="text-sm font-semibold text-surface-900 capitalize">{secret.environment}</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <div className="p-2 bg-surface-100 rounded-lg">
              <Users className="w-5 h-5 text-surface-500" />
            </div>
            <div>
              <p className="text-xs text-surface-500 font-medium">Owner Team</p>
              <p className="text-sm font-semibold text-surface-900">{secret.owner_team}</p>
            </div>
          </div>
        </div>

        {/* Value Section */}
        <div className="p-6">
          <h2 className="text-lg font-semibold text-surface-900 mb-4 flex items-center gap-2">
            <Shield className="w-5 h-5 text-surface-400" />
            Secret Value
          </h2>

          {secretValue ? (
            <div className="bg-surface-50 rounded-xl border border-surface-200 overflow-hidden">
              <div className="flex items-center justify-between px-4 py-3 bg-surface-100 border-b border-surface-200">
                <div className="flex items-center gap-2">
                  <Lock className="w-4 h-4 text-surface-400" />
                  <span className="text-sm font-medium text-surface-600">Decrypted Value</span>
                </div>
                <button
                  onClick={handleCopy}
                  className="flex items-center gap-1.5 text-sm text-brand-600 hover:text-brand-700 font-medium"
                >
                  {copied ? <Check className="w-4 h-4" /> : <Copy className="w-4 h-4" />}
                  {copied ? 'Copied!' : 'Copy'}
                </button>
              </div>
              <div className="p-4">
                <code className="text-sm font-mono text-surface-900 break-all">
                  {secretValue}
                </code>
              </div>
            </div>
          ) : secret.has_access ? (
            <button
              onClick={handleGetValue}
              disabled={actionLoading}
              className="w-full sm:w-auto flex items-center justify-center gap-2 px-6 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all btn-press font-medium"
            >
              <Eye className="w-5 h-5" />
              {actionLoading ? 'Revealing...' : 'Reveal Value'}
            </button>
          ) : requestStatus === 'pending' ? (
            <div className="bg-yellow-50 border border-yellow-200 rounded-xl p-6">
              <div className="flex items-start gap-3">
                <Clock className="w-6 h-6 text-yellow-600 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="font-medium text-yellow-800">Request Pending Approval</p>
                  <p className="text-sm text-yellow-600 mt-1">
                    Your access request is being reviewed. You will be notified once approved.
                  </p>
                </div>
              </div>
            </div>
          ) : (
            <div className="bg-surface-50 border border-surface-200 rounded-xl p-6">
              <div className="flex items-start gap-3 mb-4">
                <Lock className="w-6 h-6 text-surface-400 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="font-medium text-surface-900">Access Required</p>
                  <p className="text-sm text-surface-500 mt-1">
                    You don't have access to this secret yet.
                    {secret.classification === 'CRITICAL' && ' CRITICAL secrets require security admin approval.'}
                  </p>
                </div>
              </div>
              <textarea
                value={justification}
                onChange={(e) => setJustification(e.target.value)}
                placeholder="Provide a business justification for accessing this secret..."
                className="w-full p-4 border border-surface-200 rounded-lg input-focus text-sm resize-none bg-white text-surface-900 placeholder-surface-400"
                rows={4}
              />
              <button
                onClick={handleRequestAccess}
                disabled={actionLoading || !justification.trim()}
                className="mt-4 w-full sm:w-auto flex items-center justify-center gap-2 px-6 py-3 bg-brand-600 text-white rounded-lg hover:bg-brand-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all btn-press font-medium"
              >
                <Shield className="w-5 h-5" />
                {actionLoading ? 'Submitting...' : 'Request Access'}
              </button>
            </div>
          )}
        </div>

        {/* Messages */}
        {error && (
          <div className="mx-6 mb-6 bg-red-50 border border-red-100 text-red-700 text-sm p-4 rounded-lg flex items-center gap-3">
            <AlertTriangle className="w-5 h-5 flex-shrink-0" />
            {error}
          </div>
        )}
        {success && (
          <div className="mx-6 mb-6 bg-green-50 border border-green-100 text-green-700 text-sm p-4 rounded-lg flex items-center gap-3">
            <CheckCircle className="w-5 h-5 flex-shrink-0" />
            {success}
          </div>
        )}
      </div>
    </div>
  )
}
