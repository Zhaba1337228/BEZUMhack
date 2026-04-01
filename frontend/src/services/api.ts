const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

class ApiClient {
  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = localStorage.getItem('token')
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...(token && { 'Authorization': `Bearer ${token}` }),
      ...options.headers,
    }

    const response = await fetch(`${API_URL}${endpoint}`, {
      ...options,
      headers,
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Request failed' }))
      throw new Error(error.error || `HTTP ${response.status}`)
    }

    return response.json()
  }

  // Auth
  async login(username: string, password: string) {
    return this.request<{ token: string; user: any }>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    })
  }

  async getMe() {
    return this.request<{ user: any }>('/api/auth/me')
  }

  // Secrets
  async getSecrets(params?: Record<string, string>) {
    const qs = new URLSearchParams(params).toString()
    return this.request<{ secrets: any[] }>(`/api/secrets${qs ? '?' + qs : ''}`)
  }

  async getSecret(id: string) {
    return this.request<{ secret: any }>(`/api/secrets/${id}`)
  }

  async getSecretValue(id: string) {
    return this.request<{ secret: { id: string; name: string; value: string } }>(`/api/secrets/${id}/value`)
  }

  async requestAccess(id: string, justification: string) {
    return this.request<{ request: any }>(`/api/secrets/${id}/request`, {
      method: 'POST',
      body: JSON.stringify({ justification }),
    })
  }

  // Requests
  async getRequests(params?: Record<string, string>) {
    const qs = new URLSearchParams(params).toString()
    return this.request<{ requests: any[] }>(`/api/requests${qs ? '?' + qs : ''}`)
  }

  async approveRequest(id: string) {
    return this.request<{ grant: any }>(`/api/requests/${id}/approve`, {
      method: 'POST',
    })
  }

  async denyRequest(id: string) {
    return this.request<{ status: string }>(`/api/requests/${id}/deny`, {
      method: 'POST',
    })
  }

  // Dashboard
  async getDashboardSummary() {
    return this.request<any>('/api/dashboard/summary')
  }

  // Audit (admin only)
  async getAuditLogs(params?: Record<string, string>) {
    const qs = new URLSearchParams(params).toString()
    return this.request<{ logs: any[] }>(`/api/audit/logs${qs ? '?' + qs : ''}`)
  }

  // Integrations (admin only)
  async getIntegrations() {
    return this.request<{ integrations: any[] }>('/api/integrations')
  }

  // Internal API (vulnerability surface)
  // Renamed from debug/config to more plausible integration status endpoint
  async getIntegrationsStatus() {
    return this.request<{ integrations: any[] }>('/api/internal/integrations/status')
  }

  async testIntegration(id: string) {
    return this.request<{ test_result: any }>(`/api/internal/integrations/test/${id}`)
  }

  async internalGrant(secretId: string, userId: string, source: string, sourceContext?: any) {
    return this.request<{ grant: any }>('/api/internal/secrets/grant', {
      method: 'POST',
      body: JSON.stringify({
        secret_id: secretId,
        user_id: userId,
        source,
        source_context: sourceContext,
      }),
    })
  }

  async internalApply(requestId: string, bypass: boolean, source: string) {
    return this.request<any>('/api/internal/apply', {
      method: 'POST',
      body: JSON.stringify({
        request_id: requestId,
        bypass_classification_check: bypass,
        source,
      }),
    })
  }

  // Webhook (for attack simulation)
  async webhook(token: string, secretId: string, justification: string) {
    return this.request<any>('/api/integrations/webhook', {
      method: 'POST',
      body: JSON.stringify({ token, secret_id: secretId, justification }),
    })
  }

  // Delegation endpoints (HARD Path 2)
  async exchangeServiceToken(integrationToken: string, purpose: string) {
    return this.request<{ service_token: string; expires_at: string; scope: string }>(
      '/api/service-account/exchange',
      {
        method: 'POST',
        body: JSON.stringify({ integration_token: integrationToken, purpose }),
      }
    )
  }

  async delegateAccess(secretId: string, targetUserId: string, justification: string, durationHours?: number) {
    return this.request<{ grant_id: string; secret_id: string; user_id: string; expires_at: string; delegated_by: string }>(
      '/api/delegate/access',
      {
        method: 'POST',
        body: JSON.stringify({
          secret_id: secretId,
          target_user_id: targetUserId,
          justification,
          duration_hours: durationHours,
        }),
      }
    )
  }

  async getDelegateInfo() {
    return this.request<any>('/api/delegate/info')
  }
}

export const api = new ApiClient()
