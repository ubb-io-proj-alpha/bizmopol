<template>
  <div class="pipeline-container">
    <div class="pipeline-header">
      <h1>{{ pipeline.name }}</h1>
      <button class="btn-add-stage" @click="showAddStageModal = true">+ Dodaj Stage</button>
    </div>

    <div class="stages-container">
      <div
        v-for="stage in pipeline.stages"
        :key="stage.id"
        class="stage-column"
      >
        <div class="stage-header">
          <h3>{{ stage.name }}</h3>
          <button class="btn-delete" @click="deleteStage(stage.id)">×</button>
        </div>

        <div class="leads-list">
          <div
            v-for="lead in stage.leads"
            :key="lead.id"
            class="lead-card"
          >
            <div class="lead-info" @click="selectLead(lead)">
              <div class="lead-name">{{ lead.name }}</div>
              <div class="lead-email" v-if="lead.email">{{ lead.email }}</div>
              <div class="lead-phone" v-if="lead.phone">{{ lead.phone }}</div>
            </div>
            <div class="lead-actions">              <button class="btn-edit-lead" @click.stop="openEditLeadModal(lead)" title="Edytuj">✏️</button>              <select class="move-select" @change="(e) => moveLeadToStage(lead.id, parseInt(e.target.value), stage.id)" :value="stage.id">
                <option :value="stage.id" disabled>Przenieś do...</option>
                <option v-for="s in pipeline.stages" :key="s.id" :value="s.id">{{ s.name }}</option>
              </select>
              <button class="btn-delete-lead" @click="deleteLead(lead.id, stage.id)">×</button>
            </div>
          </div>
        </div>

        <button class="btn-add-lead" @click="showAddLeadModal(stage.id)">+ Dodaj Lead</button>
      </div>
    </div>

    <!-- Add Stage Modal -->
    <div v-if="showAddStageModal" class="modal">
      <div class="modal-content">
        <h2>Nowy Stage</h2>
        <input v-model="newStageName" type="text" placeholder="Nazwa stage'a" />
        <div class="modal-buttons">
          <button @click="addStage" class="btn-primary">Dodaj</button>
          <button @click="showAddStageModal = false" class="btn-secondary">Anuluj</button>
        </div>
      </div>
    </div>

    <!-- Add Lead Modal -->
    <div v-if="showAddLeadModalActive" class="modal">
      <div class="modal-content">
        <h2>Nowy Lead</h2>
        <input v-model="newLeadName" type="text" placeholder="Imię/Nazwa" />
        <input v-model="newLeadEmail" type="email" placeholder="Email" />
        <input v-model="newLeadPhone" type="tel" placeholder="Telefon" />
        <div class="modal-buttons">
          <button @click="addLead" class="btn-primary">Dodaj</button>
          <button @click="showAddLeadModalActive = false" class="btn-secondary">Anuluj</button>
        </div>
      </div>
    </div>

    <!-- Edit Lead Modal -->
    <div v-if="showEditLeadModalActive && editingLead" class="modal">
      <div class="modal-content">
        <h2>Edytuj Lead</h2>
        <input v-model="editingLead.name" type="text" placeholder="Imię/Nazwa" />
        <input v-model="editingLead.email" type="email" placeholder="Email" />
        <input v-model="editingLead.phone" type="tel" placeholder="Telefon" />
        <div class="modal-buttons">
          <button @click="updateLead" class="btn-primary">Zapisz</button>
          <button @click="showEditLeadModalActive = false" class="btn-secondary">Anuluj</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { pipelineAPI } from '@/services/api'

const props = defineProps({
  pipelineId: {
    type: Number,
    required: true,
  },
})

const pipeline = ref({ name: '', stages: [] })
const showAddStageModal = ref(false)
const showAddLeadModalActive = ref(false)
const showEditLeadModalActive = ref(false)
const newStageName = ref('')
const newLeadName = ref('')
const newLeadEmail = ref('')
const newLeadPhone = ref('')
const activeStageId = ref(null)
const editingLead = ref(null)

onMounted(async () => {
  await loadPipeline()
})

const loadPipeline = async () => {
  try {
    const response = await pipelineAPI.getPipeline(props.pipelineId)
    pipeline.value = response.data
  } catch (error) {
    console.error('Error loading pipeline:', error)
  }
}

const addStage = async () => {
  if (!newStageName.value.trim()) {
    alert('Wpisz nazwę stage\'a')
    return
  }

  try {
    const response = await pipelineAPI.createStage(
      props.pipelineId,
      newStageName.value.trim(),
      pipeline.value.stages.length
    )
    
    if (response.data) {
      const newStage = {
        id: response.data.id,
        name: response.data.name,
        position: response.data.position || pipeline.value.stages.length,
        leads: response.data.leads || []
      }
      pipeline.value.stages.push(newStage)
      newStageName.value = ''
      showAddStageModal.value = false
    }
  } catch (error) {
    console.error('Error adding stage:', error)
    console.log('Response:', error.response?.data)
    alert('Błąd: ' + (error.response?.data?.error || error.message))
  }
}

const deleteStage = async (stageId) => {
  if (!confirm('Na pewno usunąć stage?')) return

  try {
    await pipelineAPI.deleteStage(props.pipelineId, stageId)
    pipeline.value.stages = pipeline.value.stages.filter(s => s.id !== stageId)
  } catch (error) {
    console.error('Error deleting stage:', error)
  }
}

const showAddLeadModal = (stageId) => {
  activeStageId.value = stageId
  showAddLeadModalActive.value = true
}

const addLead = async () => {
  if (!newLeadName.value) return

  try {
    const response = await pipelineAPI.createLead(
      props.pipelineId,
      newLeadName.value,
      newLeadEmail.value,
      newLeadPhone.value,
      activeStageId.value
    )
    const stage = pipeline.value.stages.find(s => s.id === activeStageId.value)
    if (stage) {
      stage.leads.push(response.data)
    }
    newLeadName.value = ''
    newLeadEmail.value = ''
    newLeadPhone.value = ''
    showAddLeadModalActive.value = false
  } catch (error) {
    console.error('Error adding lead:', error)
  }
}

const moveLeadToStage = async (leadId, toStageId, fromStageId) => {
  if (toStageId === fromStageId) return

  try {
    const fromStage = pipeline.value.stages.find(s => s.id === fromStageId)
    const toStage = pipeline.value.stages.find(s => s.id === toStageId)
    
    const lead = fromStage.leads.find(l => l.id === leadId)
    if (lead && toStage) {
      // Optimistic update
      fromStage.leads = fromStage.leads.filter(l => l.id !== leadId)
      toStage.leads.push(lead)
      
      // Backend update
      await pipelineAPI.moveLeadToStage(props.pipelineId, leadId, toStageId, toStage.leads.length - 1)
    }
  } catch (error) {
    console.error('Error moving lead:', error)
    await loadPipeline() // Reload on error
  }
}

const deleteLead = async (leadId, stageId) => {
  if (!confirm('Na pewno usunąć lead?')) return

  try {
    await pipelineAPI.deleteLead(props.pipelineId, leadId)
    const stage = pipeline.value.stages.find(s => s.id === stageId)
    if (stage) {
      stage.leads = stage.leads.filter(l => l.id !== leadId)
    }
  } catch (error) {
    console.error('Error deleting lead:', error)
  }
}

const openEditLeadModal = (lead) => {
  editingLead.value = {
    ...lead,
    originalStageId: pipeline.value.stages.find(s => s.leads.some(l => l.id === lead.id))?.id
  }
  showEditLeadModalActive.value = true
}

const updateLead = async () => {
  if (!editingLead.value.name.trim()) {
    alert('Wpisz nazwę')
    return
  }

  try {
    await pipelineAPI.updateLead(
      props.pipelineId,
      editingLead.value.id,
      editingLead.value.name.trim(),
      editingLead.value.email,
      editingLead.value.phone
    )

    // Update local state
    const stage = pipeline.value.stages.find(s => s.id === editingLead.value.originalStageId)
    if (stage) {
      const lead = stage.leads.find(l => l.id === editingLead.value.id)
      if (lead) {
        lead.name = editingLead.value.name
        lead.email = editingLead.value.email
        lead.phone = editingLead.value.phone
      }
    }

    showEditLeadModalActive.value = false
    editingLead.value = null
  } catch (error) {
    console.error('Error updating lead:', error)
    alert('Błąd podczas edycji')
  }
}

const selectLead = (lead) => {
  // Placeholder for lead detail modal
  console.log('Selected lead:', lead)
}
</script>

<style scoped>
.pipeline-container {
  padding: 40px;
  background: #2c3e50;
  height: 100%;
  overflow-y: auto;
  color: #f1f5f9;
}

.pipeline-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 40px;
}

.pipeline-header h1 {
  font-size: 2.5rem;
  margin: 0;
  color: #38bdf8;
}

.btn-add-stage {
  background: #38bdf8;
  border: none;
  padding: 12px 24px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: bold;
  color: #0f172a;
}

.stages-container {
  display: flex;
  gap: 24px;
  overflow-x: auto;
  padding-bottom: 20px;
}

.stage-column {
  min-width: 350px;
  background: #1e293b;
  border-radius: 12px;
  padding: 16px;
  border: 1px solid #334155;
}

.stage-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.stage-header h3 {
  margin: 0;
  font-size: 1.3rem;
  color: #38bdf8;
}

.btn-delete {
  background: transparent;
  border: none;
  color: #ef4444;
  font-size: 1.5rem;
  cursor: pointer;
}

.leads-list {
  min-height: 100px;
  margin-bottom: 16px;
}

.lead-card {
  background: #334155;
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 10px;
  border-left: 4px solid #38bdf8;
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.lead-card:hover {
  background: #475569;
}

.lead-info {
  flex: 1;
  cursor: pointer;
}

.lead-name {
  font-weight: bold;
  color: #f1f5f9;
}

.lead-email {
  font-size: 0.85rem;
  color: #cbd5e1;
  margin-top: 4px;
}

.lead-phone {
  font-size: 0.85rem;
  color: #cbd5e1;
  margin-top: 2px;
}

.btn-delete-lead {
  position: absolute;
  top: 4px;
  right: 4px;
  background: transparent;
  border: none;
  color: #ef4444;
  font-size: 1.2rem;
  cursor: pointer;
  display: none;
}

.lead-card:hover .btn-delete-lead {
  display: block;
}

.lead-actions {
  display: flex;
  gap: 8px;
  margin-left: 8px;
}

.btn-edit-lead {
  background: transparent;
  border: none;
  color: #38bdf8;
  font-size: 1rem;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.2s;
}

.btn-edit-lead:hover {
  background: rgba(56, 189, 248, 0.1);
  color: #7dd3fc;
}

.move-select {
  padding: 6px 8px;
  background: #1e293b;
  border: 1px solid #38bdf8;
  border-radius: 4px;
  color: #f1f5f9;
  cursor: pointer;
  font-size: 0.85rem;
}

.move-select option {
  background: #0f172a;
  color: #f1f5f9;
}

.btn-add-lead {
  width: 100%;
  background: #38bdf8;
  border: none;
  padding: 10px;
  border-radius: 6px;
  cursor: pointer;
  color: #0f172a;
  font-weight: bold;
}

.modal {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.modal-content {
  background: linear-gradient(135deg, #1e293b 0%, #263549 100%);
  padding: 40px;
  border-radius: 16px;
  min-width: 420px;
  border: 1px solid #38bdf8;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.8), 0 0 40px rgba(56, 189, 248, 0.1);
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.modal-content h2 {
  margin-top: 0;
  margin-bottom: 24px;
  color: #38bdf8;
  font-size: 1.8rem;
}

.modal-content input {
  width: 100%;
  padding: 12px 16px;
  margin: 12px 0;
  border: 1px solid #334155;
  border-radius: 8px;
  background: #0f172a;
  color: #f1f5f9;
  box-sizing: border-box;
  font-size: 1rem;
  transition: all 0.3s;
}

.modal-content input:focus {
  outline: none;
  border-color: #38bdf8;
  box-shadow: 0 0 0 3px rgba(56, 189, 248, 0.1);
  background: #1a2537;
}

.modal-content input::placeholder {
  color: #64748b;
}

.modal-buttons {
  display: flex;
  gap: 12px;
  margin-top: 28px;
}

.btn-primary {
  flex: 1;
  background: linear-gradient(135deg, #38bdf8 0%, #0ea5e9 100%);
  border: none;
  padding: 12px 24px;
  border-radius: 8px;
  cursor: pointer;
  color: #0f172a;
  font-weight: bold;
  font-size: 1rem;
  transition: all 0.3s;
  box-shadow: 0 4px 15px rgba(56, 189, 248, 0.3);
}

.btn-primary:hover {
  background: linear-gradient(135deg, #0ea5e9 0%, #06b6d4 100%);
  box-shadow: 0 6px 20px rgba(56, 189, 248, 0.4);
  transform: translateY(-2px);
}

.btn-primary:active {
  transform: translateY(0);
}

.btn-secondary {
  flex: 1;
  background: transparent;
  border: 2px solid #334155;
  padding: 10px 24px;
  border-radius: 8px;
  cursor: pointer;
  color: #94a3b8;
  font-weight: bold;
  font-size: 1rem;
  transition: all 0.3s;
}

.btn-secondary:hover {
  border-color: #64748b;
  color: #cbd5e1;
  background: rgba(100, 116, 139, 0.1);
}
</style>
