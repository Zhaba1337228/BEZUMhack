import { useState, useEffect } from 'react'
import { api } from '../../services/api'
import {
  Puzzle,
  CheckCircle,
  XCircle,
  Calendar,
  Folder,
  Settings2,
  ExternalLink,
  Activity,
} from 'lucide-react'

export default function Integrations() {
  const [integrations, setIntegrations] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.getIntegrations()
      .then((data) => {
        setIntegrations(data.integrations)
        setLoading(false)
      })
      .catch(console.error)
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-brand-200 border-t-brand-600 rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-surface-500">Loading integrations...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-surface-900">Integrations</h1>
          <p className="text-surface-500 mt-1">Manage external service integrations</p>
        </div>
        <div className="flex items-center gap-2 text-sm text-surface-500">
          <Activity className="w-4 h-4" />
          <span>{integrations.length} integrations</span>
        </div>
      </div>

      {/* Integrations Grid */}
      {integrations.length === 0 ? (
        <div className="bg-white rounded-xl shadow-soft p-12 text-center border border-surface-100">
          <Puzzle className="w-12 h-12 text-surface-300 mx-auto mb-4" />
          <p className="text-surface-600 font-medium">No integrations configured</p>
          <p className="text-surface-400 text-sm mt-1">Connect external services to enable automated access</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {integrations.map((integration) => (
            <div
              key={integration.id}
              className="bg-white rounded-xl shadow-soft border border-surface-100 overflow-hidden card-hover"
            >
              {/* Card Header */}
              <div className="p-6 border-b border-surface-100">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-4">
                    <div className="p-3 bg-gradient-to-br from-brand-500 to-brand-600 rounded-xl shadow-md">
                      <Puzzle className="w-6 h-6 text-white" />
                    </div>
                    <div>
                      <h2 className="text-lg font-semibold text-surface-900">{integration.name}</h2>
                      <p className="text-sm text-surface-500 capitalize">Provider: {integration.provider}</p>
                    </div>
                  </div>
                  <span
                    className={`px-3 py-1.5 rounded-lg text-xs font-semibold border ${
                      integration.enabled
                        ? 'bg-green-50 text-green-700 border-green-200'
                        : 'bg-red-50 text-red-700 border-red-200'
                    }`}
                  >
                    {integration.enabled ? (
                      <span className="flex items-center gap-1.5">
                        <CheckCircle className="w-3.5 h-3.5" />
                        Active
                      </span>
                    ) : (
                      <span className="flex items-center gap-1.5">
                        <XCircle className="w-3.5 h-3.5" />
                        Disabled
                      </span>
                    )}
                  </span>
                </div>
              </div>

              {/* Card Body */}
              <div className="p-6 space-y-4">
                {/* Project */}
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-surface-100 rounded-lg">
                    <Folder className="w-4 h-4 text-surface-500" />
                  </div>
                  <div>
                    <p className="text-xs text-surface-500 font-medium">Project</p>
                    <p className="text-sm font-semibold text-surface-900">
                      {integration.project_name || 'Not configured'}
                    </p>
                  </div>
                </div>

                {/* Last Updated */}
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-surface-100 rounded-lg">
                    <Calendar className="w-4 h-4 text-surface-500" />
                  </div>
                  <div>
                    <p className="text-xs text-surface-500 font-medium">Last Updated</p>
                    <p className="text-sm font-semibold text-surface-900">
                      {new Date(integration.updated_at).toLocaleDateString('en-US', {
                        year: 'numeric',
                        month: 'short',
                        day: 'numeric',
                      })}
                    </p>
                  </div>
                </div>

                {/* Configuration */}
                {integration.config && (
                  <div className="pt-4 border-t border-surface-100">
                    <div className="flex items-center gap-2 mb-3">
                      <Settings2 className="w-4 h-4 text-surface-400" />
                      <p className="text-xs font-medium text-surface-500">Configuration</p>
                    </div>
                    <pre className="bg-surface-50 p-4 rounded-lg text-xs font-mono text-surface-700 overflow-auto max-h-48 border border-surface-200">
                      {JSON.stringify(integration.config, null, 2)}
                    </pre>
                  </div>
                )}

                {/* Webhook URL if available */}
                {integration.config?.webhook_url && (
                  <div className="flex items-center justify-between pt-4 border-t border-surface-100">
                    <div className="flex items-center gap-2">
                      <ExternalLink className="w-4 h-4 text-surface-400" />
                      <p className="text-xs text-surface-500">Webhook Endpoint</p>
                    </div>
                    <code className="text-xs font-mono text-brand-600 bg-brand-50 px-2 py-1 rounded">
                      {integration.config.webhook_url}
                    </code>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
