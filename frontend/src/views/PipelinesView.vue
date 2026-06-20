<script setup>
import { ref, reactive, onMounted, inject } from "vue"
import { useRouter } from "vue-router"

const authFetch = inject("authFetch")
const confirm = inject("confirm")
const router = useRouter()

const pipelines = ref([])
const loading = ref(false)
const error = ref("")

const showModal = ref(false)
const editingId = ref(null)
const form = reactive({ name: "", description: "", stages: [] })
const formError = ref("")
const newStageName = ref("")
const newStageColor = ref("#38bdf8")

async function load() {
    loading.value = true
    error.value = ""
    try {
        const res = await authFetch("/api/v1/pipelines/")
        if (!res.ok) throw new Error("Błąd pobierania")
        pipelines.value = await res.json()
    } catch (e) {
        error.value = e.message
    } finally {
        loading.value = false
    }
}

onMounted(load)

function openCreate() {
    editingId.value = null
    Object.assign(form, { name: "", description: "", stages: [] })
    formError.value = ""
    newStageName.value = ""
    newStageColor.value = "#38bdf8"
    showModal.value = true
}

function openEdit(p) {
    editingId.value = p.id
    form.name = p.name
    form.description = p.description
    form.stages = []
    formError.value = ""
    showModal.value = true
}

function addStage() {
    if (!newStageName.value.trim()) return
    form.stages.push({ name: newStageName.value.trim(), color: newStageColor.value, sort_order: form.stages.length })
    newStageName.value = ""
    newStageColor.value = "#38bdf8"
}

function removeStage(idx) {
    form.stages.splice(idx, 1)
}

async function save() {
    formError.value = ""
    if (!form.name.trim()) {
        formError.value = "Nazwa jest wymagana."
        return
    }
    try {
        let res
        if (editingId.value) {
            res = await authFetch("/api/v1/pipelines/" + editingId.value, {
                method: "PUT",
                body: JSON.stringify({ name: form.name, description: form.description }),
            })
        } else {
            res = await authFetch("/api/v1/pipelines/", {
                method: "POST",
                body: JSON.stringify({ name: form.name, description: form.description, stages: form.stages }),
            })
        }
        if (!res.ok) {
            const d = await res.json()
            throw new Error(d.error || "Błąd zapisu")
        }
        showModal.value = false
        await load()
    } catch (e) {
        formError.value = e.message
    }
}

async function deletePipeline(id) {
    if (!await confirm({ title: "Usuń pipeline", message: "Pipeline i wszystkie etapy zostaną trwale usunięte.", confirmLabel: "Usuń", variant: "danger" })) return
    try {
        await authFetch("/api/v1/pipelines/" + id, { method: "DELETE" })
        await load()
    } catch (e) {
        error.value = e.message
    }
}

function openKanban(id) {
    router.push({ name: "kanban", params: { id } })
}
</script>

<template>
    <section class="view-section">
        <div class="header">
            <div>
                <h1>Pipelines</h1>
                <p class="subtitle">Zarządzaj etapami sprzedaży</p>
            </div>
            <button class="btn-primary" @click="openCreate">+ Nowy pipeline</button>
        </div>

        <p v-if="error" class="err-msg">{{ error }}</p>
        <div v-if="loading" class="loading">Ładowanie...</div>
        <div v-else-if="pipelines.length === 0" class="empty">Brak pipeline'ów. Utwórz pierwszy.</div>
        <div v-else class="pipelines-grid">
            <div v-for="p in pipelines" :key="p.id" class="pipeline-card">
                <div class="pipeline-top">
                    <div>
                        <h3>{{ p.name }}</h3>
                        <p class="pipeline-desc">{{ p.description || "Brak opisu" }}</p>
                    </div>
                    <div class="card-actions">
                        <button class="btn-action btn-kanban" @click="openKanban(p.id)" title="Kanban">
                            <span class="material-icons">view_kanban</span>
                        </button>
                        <button class="btn-action btn-edit" @click="openEdit(p)" title="Edytuj">
                            <span class="material-icons">edit</span>
                        </button>
                        <button class="btn-action btn-delete" @click="deletePipeline(p.id)" title="Usuń">
                            <span class="material-icons">delete</span>
                        </button>
                    </div>
                </div>
                <div class="stages-row">
                    <div v-for="s in (p.stages || [])" :key="s.id" class="stage-chip" :style="{ borderColor: s.color, color: s.color }">
                        {{ s.name }}
                    </div>
                    <span v-if="!p.stages || p.stages.length === 0" class="no-stages">Brak etapów</span>
                </div>
                <button class="btn-open-kanban" @click="openKanban(p.id)">
                    <span class="material-icons">view_kanban</span> Otwórz Kanban
                </button>
            </div>
        </div>

        <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
            <div class="modal">
                <h2>{{ editingId ? "Edytuj pipeline" : "Nowy pipeline" }}</h2>
                <div class="form-group">
                    <label>Nazwa *</label>
                    <input v-model="form.name" placeholder="np. Sprzedaż B2B" />
                </div>
                <div class="form-group">
                    <label>Opis</label>
                    <textarea v-model="form.description" rows="2" placeholder="Opis procesu..."></textarea>
                </div>
                <div v-if="!editingId" class="form-group">
                    <label>Etapy (opcjonalnie)</label>
                    <div class="stage-add-row">
                        <input v-model="newStageName" placeholder="Nazwa etapu" class="stage-input" @keyup.enter="addStage" />
                        <input type="color" v-model="newStageColor" class="color-picker" />
                        <button class="btn-add-stage" type="button" @click="addStage">+</button>
                    </div>
                    <div class="stages-list">
                        <div v-for="(s, idx) in form.stages" :key="idx" class="stage-item">
                            <span class="stage-dot" :style="{ background: s.color }"></span>
                            <span class="stage-name">{{ s.name }}</span>
                            <button class="btn-remove-stage" @click="removeStage(idx)">×</button>
                        </div>
                    </div>
                </div>
                <p v-if="formError" class="err-msg">{{ formError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showModal = false">Anuluj</button>
                    <button class="btn-primary" @click="save">{{ editingId ? "Zapisz" : "Utwórz" }}</button>
                </div>
            </div>
        </div>
    </section>
</template>

<style scoped>
.view-section { width: 100%; max-width: 1200px; margin: 0 auto; }
.header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 28px; }
.header h1 { font-size: 2rem; color: #f8fafc; margin-bottom: 4px; }
.subtitle { color: #94a3b8; font-size: 0.95rem; }
.loading, .empty { color: #94a3b8; padding: 40px; text-align: center; }
.err-msg { color: #f87171; font-size: 0.9rem; margin-bottom: 12px; }

.pipelines-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 20px; }
.pipeline-card {
    background: #1e293b; border: 1px solid #334155; border-radius: 14px; padding: 20px;
    display: flex; flex-direction: column; gap: 14px;
}
.pipeline-top { display: flex; justify-content: space-between; align-items: flex-start; }
.pipeline-card h3 { color: #f1f5f9; font-size: 1.1rem; margin-bottom: 4px; }
.pipeline-desc { color: #64748b; font-size: 0.85rem; }
.card-actions { display: flex; gap: 6px; }
.stages-row { display: flex; flex-wrap: wrap; gap: 6px; }
.stage-chip {
    padding: 3px 10px; border-radius: 20px; font-size: 0.78rem; font-weight: 600;
    border: 1px solid; background: transparent;
}
.no-stages { color: #475569; font-size: 0.82rem; }
.btn-open-kanban {
    width: 100%; padding: 10px; background: #0f172a; border: 1px solid #334155;
    border-radius: 8px; color: #38bdf8; cursor: pointer; font-size: 0.9rem; font-weight: 600;
    display: flex; align-items: center; justify-content: center; gap: 6px; transition: 0.2s;
}
.btn-open-kanban:hover { background: #1e3a5f; border-color: #38bdf8; }
.btn-open-kanban .material-icons { font-size: 1.1rem; }

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
.btn-action {
    background: none; border: none; cursor: pointer; padding: 4px 6px;
    border-radius: 6px; transition: 0.15s; display: flex; align-items: center; color: #94a3b8;
}
.btn-action .material-icons { font-size: 1.1rem; }
.btn-kanban:hover { background: #1e3a5f; color: #38bdf8; }
.btn-edit:hover { background: #1e3a5f; color: #38bdf8; }
.btn-delete:hover { background: #3f1414; color: #f87171; }

.modal-overlay {
    position: fixed; inset: 0; background: rgba(0,0,0,0.6);
    display: flex; align-items: center; justify-content: center; z-index: 100;
}
.modal {
    background: #1e293b; border: 1px solid #334155; border-radius: 16px;
    padding: 32px; width: 100%; max-width: 480px; max-height: 90vh; overflow-y: auto;
}
.modal h2 { color: #38bdf8; margin-bottom: 20px; font-size: 1.4rem; }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; margin-bottom: 6px; color: #94a3b8; font-size: 0.88rem; }
.form-group input, .form-group textarea {
    width: 100%; padding: 10px 12px; background: #0f172a;
    border: 1px solid #334155; border-radius: 8px; color: #f1f5f9; outline: none; font-size: 0.95rem; resize: vertical;
}
.form-group input:focus, .form-group textarea:focus { border-color: #38bdf8; }
.stage-add-row { display: flex; gap: 8px; margin-bottom: 10px; }
.stage-input { flex: 1; padding: 8px 10px; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #f1f5f9; outline: none; font-size: 0.9rem; }
.stage-input:focus { border-color: #38bdf8; }
.color-picker { width: 40px; height: 36px; padding: 2px; border-radius: 6px; border: 1px solid #334155; background: #0f172a; cursor: pointer; }
.btn-add-stage {
    padding: 8px 14px; background: #38bdf8; border: none; border-radius: 8px;
    color: #0f172a; font-weight: bold; cursor: pointer; font-size: 1rem;
}
.stages-list { display: flex; flex-direction: column; gap: 6px; }
.stage-item {
    display: flex; align-items: center; gap: 8px;
    background: #0f172a; border: 1px solid #334155; border-radius: 8px; padding: 6px 10px;
}
.stage-dot { width: 10px; height: 10px; border-radius: 50%; flex-shrink: 0; }
.stage-name { flex: 1; color: #e2e8f0; font-size: 0.9rem; }
.btn-remove-stage {
    background: none; border: none; color: #ef4444; cursor: pointer; font-size: 1.1rem;
    padding: 0 4px; line-height: 1;
}
.modal-actions { display: flex; gap: 12px; justify-content: flex-end; margin-top: 20px; }
</style>