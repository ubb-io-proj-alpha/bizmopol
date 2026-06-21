<script setup>
import { ref, onMounted, inject } from "vue"
import { useRoute, useRouter } from "vue-router"

const authFetch = inject("authFetch")
const confirm = inject("confirm")
const route = useRoute()
const router = useRouter()

const kanban = ref(null)
const loading = ref(false)
const error = ref("")

const dragging = ref(null)
const dragOverStageId = ref(null)

const showStageModal = ref(false)
const editingStageId = ref(null)
const stageForm = ref({ name: "", color: "#38bdf8", sort_order: 0 })
const stageFormError = ref("")

const showAddContact = ref(false)
const allContacts = ref([])
const selectedContactId = ref("")
const selectedStageId = ref("")
const addContactError = ref("")

async function load() {
    loading.value = true
    error.value = ""
    try {
        const res = await authFetch("/api/v1/pipelines/" + route.params.id + "/kanban")
        if (!res.ok) throw new Error("Błąd pobierania Kanban")
        kanban.value = await res.json()
    } catch (e) {
        error.value = e.message
    } finally {
        loading.value = false
    }
}

async function loadContacts() {
    try {
        const res = await authFetch("/api/v1/contacts/?page_size=200")
        if (res.ok) {
            const data = await res.json()
            allContacts.value = data.data || []
        }
    } catch (_) {}
}

onMounted(() => {
    load()
    loadContacts()
})

function onDragStart(contact, stageId) {
    dragging.value = { contact, stageId }
}

function onDragOver(e, stageId) {
    e.preventDefault()
    dragOverStageId.value = stageId
}

function onDragLeave() {
    dragOverStageId.value = null
}

async function onDrop(e, targetStageId) {
    e.preventDefault()
    dragOverStageId.value = null
    if (!dragging.value) return
    const { contact, stageId } = dragging.value
    dragging.value = null
    if (stageId === targetStageId) return
    await moveContact(contact.id, targetStageId)
}

async function moveContact(contactId, stageId) {
    try {
        const res = await authFetch("/api/v1/pipelines/" + route.params.id + "/move", {
            method: "POST",
            body: JSON.stringify({ contact_id: contactId, stage_id: stageId }),
        })
        if (!res.ok) throw new Error("Błąd przenoszenia")
        await load()
    } catch (e) {
        error.value = e.message
    }
}

function openAddStage() {
    editingStageId.value = null
    stageForm.value = { name: "", color: "#38bdf8", sort_order: 0 }
    stageFormError.value = ""
    showStageModal.value = true
}

function openEditStage(stage) {
    editingStageId.value = stage.id
    stageForm.value = { name: stage.name, color: stage.color, sort_order: stage.sort_order }
    stageFormError.value = ""
    showStageModal.value = true
}

async function saveStage() {
    stageFormError.value = ""
    if (!stageForm.value.name.trim()) {
        stageFormError.value = "Nazwa jest wymagana."
        return
    }
    try {
        let res
        if (editingStageId.value) {
            res = await authFetch("/api/v1/pipelines/" + route.params.id + "/stages/" + editingStageId.value, {
                method: "PUT",
                body: JSON.stringify(stageForm.value),
            })
        } else {
            res = await authFetch("/api/v1/pipelines/" + route.params.id + "/stages", {
                method: "POST",
                body: JSON.stringify(stageForm.value),
            })
        }
        if (!res.ok) {
            const d = await res.json()
            throw new Error(d.error || "Błąd zapisu")
        }
        showStageModal.value = false
        await load()
    } catch (e) {
        stageFormError.value = e.message
    }
}

async function deleteStage(stageId) {
    if (!await confirm({ title: "Usuń etap", message: "Etap i przypisane kontakty zostaną usunięte z pipeline.", confirmLabel: "Usuń", variant: "danger" })) return
    try {
        await authFetch("/api/v1/pipelines/" + route.params.id + "/stages/" + stageId, { method: "DELETE" })
        await load()
    } catch (e) {
        error.value = e.message
    }
}

function openAddContact(stageId) {
    selectedStageId.value = stageId
    selectedContactId.value = ""
    addContactError.value = ""
    showAddContact.value = true
}

async function addContactToStage() {
    addContactError.value = ""
    if (!selectedContactId.value) {
        addContactError.value = "Wybierz kontakt."
        return
    }
    await moveContact(selectedContactId.value, selectedStageId.value)
    showAddContact.value = false
}

function scoreColor(score) {
    if (score >= 70) return "#22c55e"
    if (score >= 40) return "#f59e0b"
    return "#ef4444"
}

function statusLabel(s) {
    const map = { lead: "Lead", prospect: "Prospect", customer: "Klient", inactive: "Nieaktywny" }
    return map[s] || s
}

function statusClass(s) {
    const map = { lead: "badge-lead", prospect: "badge-prospect", customer: "badge-customer", inactive: "badge-inactive" }
    return map[s] || ""
}

function goToContact(id) {
    router.push({ name: "contactDetail", params: { id } })
}
</script>

<template>
    <div class="kanban-wrapper">
        <div class="kanban-header">
            <div class="header-left">
                <button class="btn-back" @click="router.push({ name: 'pipelines' })">← Pipelines</button>
                <div v-if="kanban">
                    <h1>{{ kanban.pipeline.name }}</h1>
                    <p class="subtitle">{{ kanban.pipeline.description }}</p>
                </div>
            </div>
            <button class="btn-primary" @click="openAddStage">+ Dodaj etap</button>
        </div>

        <p v-if="error" class="err-msg">{{ error }}</p>
        <div v-if="loading" class="loading">Ładowanie...</div>

        <div v-else-if="kanban" class="kanban-board">
            <div
                v-for="col in kanban.columns"
                :key="col.stage.id"
                class="kanban-column"
                :class="{ 'drag-over': dragOverStageId === col.stage.id }"
                @dragover="onDragOver($event, col.stage.id)"
                @dragleave="onDragLeave"
                @drop="onDrop($event, col.stage.id)"
            >
                <div class="column-header" :style="{ borderTopColor: col.stage.color }">
                    <div class="column-title">
                        <span class="stage-dot" :style="{ background: col.stage.color }"></span>
                        <span class="stage-name">{{ col.stage.name }}</span>
                        <span class="contact-count">{{ col.contacts.length }}</span>
                    </div>
                    <div class="col-actions">
                        <button class="btn-col-action" @click="openAddContact(col.stage.id)" title="Dodaj kontakt">
                            <span class="material-icons">add</span>
                        </button>
                        <button class="btn-col-action" @click="openEditStage(col.stage)" title="Edytuj etap">
                            <span class="material-icons">edit</span>
                        </button>
                        <button class="btn-col-action btn-col-delete" @click="deleteStage(col.stage.id)" title="Usuń etap">
                            <span class="material-icons">delete</span>
                        </button>
                    </div>
                </div>

                <div class="column-body">
                    <div
                        v-for="contact in col.contacts"
                        :key="contact.id"
                        class="kanban-card"
                        draggable="true"
                        @dragstart="onDragStart(contact, col.stage.id)"
                        @dragend="dragging = null"
                    >
                        <div class="card-drag-handle">
                            <span class="material-icons drag-icon">drag_indicator</span>
                        </div>
                        <div class="card-body" @click="goToContact(contact.id)">
                            <div class="card-name">{{ contact.name }}</div>
                            <div v-if="contact.company" class="card-company">{{ contact.company }}</div>
                            <div class="card-footer">
                                <span :class="['badge', statusClass(contact.status)]">{{ statusLabel(contact.status) }}</span>
                                <span class="card-score" :style="{ color: scoreColor(contact.lead_score) }">{{ contact.lead_score }}</span>
                            </div>
                        </div>
                    </div>
                    <div v-if="col.contacts.length === 0" class="column-empty">
                        Przeciągnij kontakt tutaj
                    </div>
                </div>
            </div>

            <div v-if="kanban.columns.length === 0" class="no-stages">
                Brak etapów. Dodaj pierwszy etap.
            </div>
        </div>

        <div v-if="showStageModal" class="modal-overlay" @click.self="showStageModal = false">
            <div class="modal">
                <h2>{{ editingStageId ? "Edytuj etap" : "Nowy etap" }}</h2>
                <div class="form-group">
                    <label>Nazwa *</label>
                    <input v-model="stageForm.name" placeholder="np. Kwalifikacja" />
                </div>
                <div class="form-group">
                    <label>Kolor</label>
                    <div class="color-row">
                        <input type="color" v-model="stageForm.color" class="color-picker" />
                        <span class="stage-chip" :style="{ borderColor: stageForm.color, color: stageForm.color }">{{ stageForm.name || "Podgląd" }}</span>
                    </div>
                </div>
                <div class="form-group">
                    <label>Kolejność</label>
                    <input v-model.number="stageForm.sort_order" type="number" min="0" />
                </div>
                <p v-if="stageFormError" class="err-msg">{{ stageFormError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showStageModal = false">Anuluj</button>
                    <button class="btn-primary" @click="saveStage">{{ editingStageId ? "Zapisz" : "Dodaj" }}</button>
                </div>
            </div>
        </div>

        <div v-if="showAddContact" class="modal-overlay" @click.self="showAddContact = false">
            <div class="modal">
                <h2>Dodaj kontakt do etapu</h2>
                <div class="form-group">
                    <label>Kontakt</label>
                    <select v-model="selectedContactId">
                        <option value="">-- Wybierz --</option>
                        <option v-for="c in allContacts" :key="c.id" :value="c.id">{{ c.name }} ({{ c.company || c.email }})</option>
                    </select>
                </div>
                <p v-if="addContactError" class="err-msg">{{ addContactError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showAddContact = false">Anuluj</button>
                    <button class="btn-primary" @click="addContactToStage">Dodaj</button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.kanban-wrapper { width: 100%; height: 100%; display: flex; flex-direction: column; }
.kanban-header {
    display: flex; justify-content: space-between; align-items: flex-start;
    margin-bottom: 24px; flex-shrink: 0;
}
.header-left { display: flex; align-items: flex-start; gap: 16px; }
.kanban-header h1 { font-size: 1.8rem; color: #f8fafc; margin-bottom: 2px; }
.subtitle { color: #94a3b8; font-size: 0.9rem; }
.btn-back {
    background: transparent; border: 1px solid #334155; color: #94a3b8;
    padding: 8px 14px; border-radius: 8px; cursor: pointer; transition: 0.2s;
    font-size: 0.88rem; white-space: nowrap; margin-top: 4px;
}
.btn-back:hover { border-color: #38bdf8; color: #38bdf8; }

.loading, .empty { color: #94a3b8; padding: 40px; text-align: center; }
.err-msg { color: #f87171; font-size: 0.9rem; margin-bottom: 12px; }

.kanban-board {
    display: flex; gap: 16px; overflow-x: auto; flex: 1;
    align-items: flex-start; padding-bottom: 16px;
}

.kanban-column {
    min-width: 280px; max-width: 280px; background: #1e293b;
    border-radius: 12px; border: 1px solid #334155;
    display: flex; flex-direction: column; max-height: calc(100vh - 200px);
    transition: 0.15s;
}
.kanban-column.drag-over { border-color: #38bdf8; background: #1e3a5f; }

.column-header {
    padding: 14px 12px 10px; border-top: 3px solid transparent;
    border-radius: 12px 12px 0 0; display: flex; justify-content: space-between; align-items: center;
    flex-shrink: 0;
}
.column-title { display: flex; align-items: center; gap: 8px; }
.stage-dot { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; }
.stage-name { color: #f1f5f9; font-weight: 600; font-size: 0.95rem; }
.contact-count {
    background: #334155; color: #94a3b8; font-size: 0.75rem; font-weight: 600;
    padding: 2px 7px; border-radius: 12px; min-width: 22px; text-align: center;
}
.col-actions { display: flex; gap: 4px; }
.btn-col-action {
    background: none; border: none; cursor: pointer; padding: 3px 5px;
    border-radius: 5px; color: #64748b; transition: 0.15s; display: flex; align-items: center;
}
.btn-col-action .material-icons { font-size: 1rem; }
.btn-col-action:hover { background: #334155; color: #f1f5f9; }
.btn-col-delete:hover { background: #3f1414; color: #f87171; }

.column-body { padding: 8px; overflow-y: auto; flex: 1; display: flex; flex-direction: column; gap: 8px; }

.kanban-card {
    background: #0f172a; border: 1px solid #334155; border-radius: 10px;
    cursor: grab; transition: 0.15s; display: flex;
}
.kanban-card:hover { border-color: #475569; box-shadow: 0 2px 8px rgba(0,0,0,0.3); }
.kanban-card:active { cursor: grabbing; }

.card-drag-handle {
    display: flex; align-items: center; padding: 8px 4px 8px 8px; color: #334155;
}
.drag-icon { font-size: 1rem; }
.kanban-card:hover .drag-icon { color: #475569; }

.card-body { flex: 1; padding: 10px 10px 10px 4px; cursor: pointer; }
.card-body:hover .card-name { color: #38bdf8; }
.card-name { color: #f1f5f9; font-weight: 600; font-size: 0.92rem; margin-bottom: 2px; transition: color 0.15s; }
.card-company { color: #64748b; font-size: 0.8rem; margin-bottom: 8px; }
.card-footer { display: flex; align-items: center; justify-content: space-between; }
.card-score { font-size: 0.78rem; font-weight: 700; }

.column-empty {
    color: #334155; font-size: 0.82rem; text-align: center;
    padding: 20px; border: 1px dashed #334155; border-radius: 8px;
}

.no-stages { color: #94a3b8; padding: 40px; text-align: center; }

.badge { padding: 2px 8px; border-radius: 20px; font-size: 0.72rem; font-weight: 600; }
.badge-lead { background: #1d4ed8; color: #bfdbfe; }
.badge-prospect { background: #7c3aed; color: #ede9fe; }
.badge-customer { background: #065f46; color: #a7f3d0; }
.badge-inactive { background: #374151; color: #9ca3af; }

.btn-primary {
    padding: 10px 20px; background: #38bdf8; border: none;
    border-radius: 8px; color: #0f172a; font-weight: bold; cursor: pointer; transition: 0.2s;
}
.btn-primary:hover { background: #7dd3fc; }
.btn-secondary {
    padding: 10px 20px; background: transparent; border: 1px solid #475569;
    border-radius: 8px; color: #94a3b8; cursor: pointer; transition: 0.2s;
}
.btn-secondary:hover { border-color: #f1f5f9; color: #f1f5f9; }

.modal-overlay {
    position: fixed; inset: 0; background: rgba(0,0,0,0.6);
    display: flex; align-items: center; justify-content: center; z-index: 100;
}
.modal {
    background: #1e293b; border: 1px solid #334155; border-radius: 16px;
    padding: 32px; width: 100%; max-width: 420px; max-height: 90vh; overflow-y: auto;
}
.modal h2 { color: #38bdf8; margin-bottom: 20px; font-size: 1.4rem; }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; margin-bottom: 6px; color: #94a3b8; font-size: 0.88rem; }
.form-group input, .form-group select {
    width: 100%; padding: 10px 12px; background: #0f172a;
    border: 1px solid #334155; border-radius: 8px; color: #f1f5f9; outline: none; font-size: 0.95rem;
}
.form-group input:focus, .form-group select:focus { border-color: #38bdf8; }
.color-row { display: flex; align-items: center; gap: 12px; }
.color-picker { width: 48px; height: 40px; padding: 2px; border-radius: 6px; border: 1px solid #334155; background: #0f172a; cursor: pointer; }
.stage-chip { padding: 4px 12px; border-radius: 20px; font-size: 0.85rem; font-weight: 600; border: 1px solid; }
.modal-actions { display: flex; gap: 12px; justify-content: flex-end; margin-top: 20px; }
</style>