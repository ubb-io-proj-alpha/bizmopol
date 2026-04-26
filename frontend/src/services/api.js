const API_BASE_URL = "/api/v1"
export const TOKEN_KEY = "jwt_token"

export const getToken = () => localStorage.getItem(TOKEN_KEY)
export const setToken = (token) => localStorage.setItem(TOKEN_KEY, token)
export const clearToken = () => localStorage.removeItem(TOKEN_KEY)

async function request(path, options = {}) {
  const headers = {
    ...(options.body ? { "Content-Type": "application/json" } : {}),
    ...(options.headers || {}),
  }

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('jwt_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Pipeline API
export const pipelineAPI = {
  createPipeline: (name) => apiClient.post('/pipelines', { name }),
  getPipelines: () => apiClient.get('/pipelines'),
  getPipeline: (id) => apiClient.get(`/pipelines/${id}`),
  updatePipeline: (id, name) => apiClient.put(`/pipelines/${id}`, { name }),
  deletePipeline: (id) => apiClient.delete(`/pipelines/${id}`),

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers,
  })

  const rawBody = await response.text()
  const data = rawBody ? JSON.parse(rawBody) : null

  if (!response.ok) {
    const error = new Error(data?.error || "Request failed")
    error.status = response.status
    error.data = data
    throw error
  }

  return data
}

export const authAPI = {
  register: (payload) =>
    request("/auth/register", {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  login: (payload) =>
    request("/auth/login", {
      method: "POST",
      body: JSON.stringify(payload),
    }),
}

export const pipelineAPI = {
  createPipeline: (name) =>
    request("/pipelines/", {
      method: "POST",
      body: JSON.stringify({ name }),
    }),
  getPipelines: () => request("/pipelines/"),
  getPipeline: (id) => request(`/pipelines/${id}`),
  updatePipeline: (id, name) =>
    request(`/pipelines/${id}`, {
      method: "PUT",
      body: JSON.stringify({ name }),
    }),
  deletePipeline: (id) =>
    request(`/pipelines/${id}`, {
      method: "DELETE",
    }),
  createStage: (pipelineId, name, position = 0) =>
    request(`/pipelines/${pipelineId}/stages/`, {
      method: "POST",
      body: JSON.stringify({ name, position }),
    }),
  updateStage: (pipelineId, stageId, name) =>
    request(`/pipelines/${pipelineId}/stages/${stageId}`, {
      method: "PUT",
      body: JSON.stringify({ name }),
    }),
  updateStagePosition: (pipelineId, stageId, position) =>
    request(`/pipelines/${pipelineId}/stages/${stageId}/position`, {
      method: "PUT",
      body: JSON.stringify({ position }),
    }),
  deleteStage: (pipelineId, stageId) =>
    request(`/pipelines/${pipelineId}/stages/${stageId}`, {
      method: "DELETE",
    }),
  createLead: (pipelineId, name, email, phone, stageId) =>
    request(`/pipelines/${pipelineId}/leads/`, {
      method: "POST",
      body: JSON.stringify({ name, email, phone, stage_id: stageId }),
    }),
  updateLead: (pipelineId, leadId, name, email, phone) =>
    request(`/pipelines/${pipelineId}/leads/${leadId}`, {
      method: "PUT",
      body: JSON.stringify({ name, email, phone }),
    }),
  moveLeadToStage: (pipelineId, leadId, stageId, position) =>
    request(`/pipelines/${pipelineId}/leads/${leadId}/move`, {
      method: "PUT",
      body: JSON.stringify({ stage_id: stageId, position }),
    }),
  deleteLead: (pipelineId, leadId) =>
    request(`/pipelines/${pipelineId}/leads/${leadId}`, {
      method: "DELETE",
    }),
}
