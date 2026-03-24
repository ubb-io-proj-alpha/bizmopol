import axios from 'axios'

// W dev mode: proxy to backendu, w production: relative path
const API_BASE_URL = '/api/v1'

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Pipeline API
export const pipelineAPI = {
  createPipeline: (name) => apiClient.post('/pipelines', { name }),
  getPipelines: () => apiClient.get('/pipelines'),
  getPipeline: (id) => apiClient.get(`/pipelines/${id}`),
  updatePipeline: (id, name) => apiClient.put(`/pipelines/${id}`, { name }),
  deletePipeline: (id) => apiClient.delete(`/pipelines/${id}`),

  // Stages
  createStage: (pipelineId, name, position = 0) =>
    apiClient.post(`/pipelines/${pipelineId}/stages`, { name, pipeline_id: pipelineId, position }),
  updateStage: (pipelineId, stageId, name) =>
    apiClient.put(`/pipelines/${pipelineId}/stages/${stageId}`, { name }),
  updateStagePosition: (pipelineId, stageId, position) =>
    apiClient.put(`/pipelines/${pipelineId}/stages/${stageId}/position`, { position }),
  deleteStage: (pipelineId, stageId) =>
    apiClient.delete(`/pipelines/${pipelineId}/stages/${stageId}`),

  // Leads
  createLead: (pipelineId, name, email, phone, stageId = null) =>
    apiClient.post(`/pipelines/${pipelineId}/leads`, { name, email, phone, stage_id: stageId }),
  updateLead: (pipelineId, leadId, name, email, phone) =>
    apiClient.put(`/pipelines/${pipelineId}/leads/${leadId}`, { name, email, phone }),
  moveLeadToStage: (pipelineId, leadId, stageId, position) =>
    apiClient.put(`/pipelines/${pipelineId}/leads/${leadId}/move`, { stage_id: stageId, position }),
  deleteLead: (pipelineId, leadId) =>
    apiClient.delete(`/pipelines/${pipelineId}/leads/${leadId}`),
}

export default apiClient
