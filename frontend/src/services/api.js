const API_BASE_URL = "/api/v1"
export const TOKEN_KEY = "jwt_token"

export const getToken = () => localStorage.getItem(TOKEN_KEY)
export const setToken = (token) => localStorage.setItem(TOKEN_KEY, token)
export const clearToken = () => localStorage.removeItem(TOKEN_KEY)

const parseResponseData = (rawBody) => {
  if (!rawBody) {
    return null
  }

  try {
    return JSON.parse(rawBody)
  } catch {
    return { error: rawBody }
  }
}

async function request(path, options = {}) {
  const headers = {
    ...(options.body ? { "Content-Type": "application/json" } : {}),
    ...(options.headers || {}),
  }

  const token = getToken()
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  let response
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...options,
      headers,
    })
  } catch (error) {
    const networkError = new Error("Network request failed")
    networkError.status = 0
    networkError.data = null
    networkError.cause = error
    throw networkError
  }

  const rawBody = await response.text()
  const data = parseResponseData(rawBody)

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
