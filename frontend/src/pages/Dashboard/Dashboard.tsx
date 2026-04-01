import { useState, useEffect } from 'react'
import { api } from '../../services/api'
import { useAuth } from '../../context/AuthContext'
import { Key, Clock, CheckCircle, Shield, ArrowRight, AlertCircle } from 'lucide-react'
import { Link } from 'react-router-dom'

export default function Dashboard() {
  const { user } = useAuth()
  const [summary, setSummary] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api.getDashboardSummary()
      .then((data) => {
        console.log('Dashboard summary:', data)
        setSummary(data)
        setError('')
      })
      .catch((err) => {
        console.error('Failed to load dashboard:', err)
        setError('Failed to load dashboard data')
      })
      .finally(() => setLoading(false))
  }, [])

  const isAdmin = user?.role === 'team_lead' || user?.role === 'security_admin'

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-brand-200 border-t-brand-600 rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-surface-500">Loading dashboard...</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <AlertCircle className="w-12 h-12 text-red-500 mx-auto mb-4" />
          <p className="text-surface-900 font-medium">{error}</p>
          <p className="text-surface-500 text-sm mt-1">Make sure the backend server is running</p>
        </div>
      </div>
    )
  }

  const stats = [
    {
      label: 'Active Grants',
      value: summary?.my_active_grants || 0,
      icon: CheckCircle,
      color: 'text-green-600',
      bgColor: 'bg-green-50',
    },
    {
      label: 'My Pending Requests',
      value: summary?.my_pending_requests || 0,
      icon: Clock,
      color: 'text-yellow-600',
      bgColor: 'bg-yellow-50',
    },
    ...(isAdmin ? [{
      label: 'Awaiting My Approval',
      value: summary?.pending_approvals || 0,
      icon: Shield,
      color: 'text-purple-600',
      bgColor: 'bg-purple-50',
    }] : []),
    {
      label: 'Total Secrets',
      value: summary?.total_secrets || summary?.secrets_by_classification?.reduce((sum: any, c: any) => sum + c.count, 0) || 0,
      icon: Key,
      color: 'text-brand-600',
      bgColor: 'bg-brand-50',
    },
  ]

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
      case 'CRITICAL': return 'bg-red-50 border-red-200'
      case 'HIGH': return 'bg-orange-50 border-orange-200'
      case 'MEDIUM': return 'bg-yellow-50 border-yellow-200'
      default: return 'bg-green-50 border-green-200'
    }
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-surface-900">Dashboard</h1>
          <p className="text-surface-500 mt-1">Overview of your secrets access</p>
        </div>
        <Link
          to="/secrets"
          className="flex items-center gap-2 px-4 py-2 bg-brand-600 text-white rounded-lg hover:bg-brand-700 transition-colors btn-press"
        >
          <Key className="w-4 h-4" />
          <span>Browse Secrets</span>
          <ArrowRight className="w-4 h-4" />
        </Link>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {stats.map((stat, index) => {
          const Icon = stat.icon
          return (
            <div
              key={index}
              className="bg-white rounded-xl shadow-soft p-6 card-hover border border-surface-100"
            >
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-surface-500">{stat.label}</p>
                  <p className={`text-3xl font-bold mt-2 ${stat.color}`}>{stat.value}</p>
                </div>
                <div className={`p-3 rounded-lg ${stat.bgColor}`}>
                  <Icon className={`w-6 h-6 ${stat.color}`} />
                </div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Classification Breakdown */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Secrets by Classification */}
        <div className="bg-white rounded-xl shadow-soft p-6 border border-surface-100">
          <div className="flex items-center gap-2 mb-6">
            <Shield className="w-5 h-5 text-surface-400" />
            <h2 className="text-lg font-semibold text-surface-900">Secrets by Classification</h2>
          </div>
          <div className="space-y-3">
            {summary?.secrets_by_classification?.map((row: any) => {
              const classification = row.classification || row.Classification || 'LOW'
              const count = row.count ?? row.Count ?? 0
              return (
              <div
                key={classification}
                className={`flex items-center justify-between p-3 rounded-lg border ${getClassBg(classification)}`}
              >
                <div className="flex items-center gap-3">
                  <div className={`w-2 h-2 rounded-full ${
                    classification === 'CRITICAL' ? 'bg-red-500' :
                    classification === 'HIGH' ? 'bg-orange-500' :
                    classification === 'MEDIUM' ? 'bg-yellow-500' :
                    'bg-green-500'
                  }`}></div>
                  <span className={`text-sm font-medium ${getClassColor(classification)}`}>
                    {classification}
                  </span>
                </div>
                <span className="text-lg font-bold text-surface-900">{count}</span>
              </div>
            )})}
          </div>
        </div>

        {/* Quick Info */}
        <div className="bg-white rounded-xl shadow-soft p-6 border border-surface-100">
          <div className="flex items-center gap-2 mb-6">
            <AlertCircle className="w-5 h-5 text-surface-400" />
            <h2 className="text-lg font-semibold text-surface-900">Security Notice</h2>
          </div>
          <div className="space-y-4">
            <div className="flex items-start gap-3 p-3 bg-blue-50 rounded-lg border border-blue-100">
              <Shield className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
              <div>
                <p className="text-sm text-blue-800 font-medium">Access is monitored</p>
                <p className="text-xs text-blue-600 mt-1">All secret access is logged and audited for security compliance.</p>
              </div>
            </div>
            <div className="flex items-start gap-3 p-3 bg-amber-50 rounded-lg border border-amber-100">
              <Clock className="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
              <div>
                <p className="text-sm text-amber-800 font-medium">Time-limited access</p>
                <p className="text-xs text-amber-600 mt-1">Grants automatically expire after 24 hours. Request renewal if needed.</p>
              </div>
            </div>
            <div className="flex items-start gap-3 p-3 bg-purple-50 rounded-lg border border-purple-100">
              <CheckCircle className="w-5 h-5 text-purple-600 flex-shrink-0 mt-0.5" />
              <div>
                <p className="text-sm text-purple-800 font-medium">Approval required</p>
                <p className="text-xs text-purple-600 mt-1">HIGH and CRITICAL secrets require approval from authorized personnel.</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
