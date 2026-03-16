<script setup>
import { ref, onMounted, inject } from "vue"
import { useRoute, useRouter } from "vue-router"

const authFetch = inject("authFetch")

const route = useRoute()
const router = useRouter()

const contact = ref(null)
const loading = ref(false)
const error = ref("")

const history = ref([])
const historyLoading = ref(false)
const historyError = ref("")

const showAddNote = ref(false)
const noteForm = ref({ action: "note", description: "" })
const noteError = ref("")

const actionLabels = {
    created: "Utworzono",
    updated: "Zaktualizowano",
    note: "Notatka",
    call: "Telefon",
    email: "Email",
    meeting: "Spotkanie",
}

onMounted(async () => {
    loading.value = true
    try {
        const res = await authFetch("/api/v1/contacts/" + route.params.id)
        if (!res.ok) throw new Error("Błąd pobierania kontaktu")
        contact.value = await res.json()
    } catch (e) {
        error.value = e.message
    } finally {
        loading.value = false
    }
    loadHistory()
})

async function loadHistory() {
    historyLoading.value = true
    historyError.value = ""
    try {
        const res = await authFetch("/api/v1/contacts/" + route.params.id + "/history")
        if (!res.ok) throw new Error("Błąd pobierania historii")
        history.value = await res.json()
    } catch (e) {
        historyError.value = e.message
    } finally {
        historyLoading.value = false
    }
}

async function addNote() {
    noteError.value = ""
    if (!noteForm.value.description.trim()) {
        noteError.value = "Treść jest wymagana."
        return
    }
    try {
        const res = await authFetch("/api/v1/contacts/" + route.params.id + "/history", {
            method: "POST",
            body: JSON.stringify(noteForm.value),
        })
        if (!res.ok) throw new Error("Błąd zapisu")
        showAddNote.value = false
        noteForm.value = { action: "note", description: "" }
        await loadHistory()
    } catch (e) {
        noteError.value = e.message
    }
}

function statusLabel(s) {
    const map = { lead: "Lead", prospect: "Prospect", customer: "Klient", inactive: "Nieaktywny" }
    return map[s] || s
}

function statusClass(s) {
    const map = { lead: "badge-lead", prospect: "badge-prospect", customer: "badge-customer", inactive: "badge-inactive" }
    return map[s] || ""
}

function formatDate(d) {
    return new Date(d).toLocaleString("pl-PL")
}

function actionIcon(a) {
    const map = { created: "add_circle", updated: "edit", note: "sticky_note_2", call: "call", email: "email", meeting: "event" }
    return map[a] || "history"
}
</script>

<template>
    <div class="detail-wrapper">
        <button class="btn-back" @click="router.push({ name: 'contacts' })">← Wróć do listy</button>

        <div v-if="loading" class="loading">Ładowanie...</div>
        <p v-else-if="error" class="err-msg">{{ error }}</p>

        <div v-else-if="contact" class="detail-card">
            <div class="detail-header">
                <div>
                    <h1>{{ contact.name }}</h1>
                    <span :class="['badge', statusClass(contact.status)]">{{ statusLabel(contact.status) }}</span>
                </div>
            </div>

            <div class="detail-grid">
                <div class="detail-item">
                    <span class="label">Email</span>
                    <span class="value">{{ contact.email || "—" }}</span>
                </div>
                <div class="detail-item">
                    <span class="label">Telefon</span>
                    <span class="value">{{ contact.phone || "—" }}</span>
                </div>
                <div class="detail-item">
                    <span class="label">Firma</span>
                    <span class="value">{{ contact.company || "—" }}</span>
                </div>
                <div class="detail-item">
                    <span class="label">Data dodania</span>
                    <span class="value">{{ formatDate(contact.created_at) }}</span>
                </div>
                <div class="detail-item">
                    <span class="label">Ostatnia aktualizacja</span>
                    <span class="value">{{ formatDate(contact.updated_at) }}</span>
                </div>
            </div>

            <div v-if="contact.notes" class="detail-notes">
                <span class="label">Notatki</span>
                <p>{{ contact.notes }}</p>
            </div>
        </div>

        <div class="history-section">
            <div class="history-header">
                <h2>Historia</h2>
                <button class="btn-primary" @click="showAddNote = !showAddNote">+ Dodaj wpis</button>
            </div>

            <div v-if="showAddNote" class="add-note-form">
                <div class="form-group">
                    <label>Typ</label>
                    <select v-model="noteForm.action">
                        <option value="note">Notatka</option>
                        <option value="call">Telefon</option>
                        <option value="email">Email</option>
                        <option value="meeting">Spotkanie</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Opis *</label>
                    <textarea v-model="noteForm.description" rows="3" placeholder="Opis zdarzenia..."></textarea>
                </div>
                <p v-if="noteError" class="err-msg">{{ noteError }}</p>
                <div class="form-actions">
                    <button class="btn-secondary" @click="showAddNote = false">Anuluj</button>
                    <button class="btn-primary" @click="addNote">Zapisz</button>
                </div>
            </div>

            <div v-if="historyLoading" class="loading">Ładowanie historii...</div>
            <p v-else-if="historyError" class="err-msg">{{ historyError }}</p>
            <div v-else-if="history.length === 0" class="empty-history">Brak wpisów w historii.</div>
            <div v-else class="timeline">
                <div v-for="h in history" :key="h.id" class="timeline-item">
                    <div class="timeline-icon">
                        <span class="material-icons">{{ actionIcon(h.action) }}</span>
                    </div>
                    <div class="timeline-body">
                        <div class="timeline-meta">
                            <span class="action-label">{{ actionLabels[h.action] || h.action }}</span>
                            <span class="timeline-date">{{ formatDate(h.created_at) }}</span>
                        </div>
                        <p v-if="h.description" class="timeline-desc">{{ h.description }}</p>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.detail-wrapper { width: 100%; max-width: 800px; margin: 0 auto; }
.btn-back {
    background: transparent; border: 1px solid #334155; color: #94a3b8;
    padding: 8px 16px; border-radius: 8px; cursor: pointer; margin-bottom: 24px;
    transition: 0.2s; font-size: 0.9rem;
}
.btn-back:hover { border-color: #38bdf8; color: #38bdf8; }
.loading { color: #94a3b8; padding: 40px; text-align: center; }
.err-msg { color: #f87171; font-size: 0.9rem; }
.detail-card {
    background: #1e293b; border: 1px solid #334155; border-radius: 16px; padding: 32px;
}
.detail-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 28px; }
.detail-header h1 { font-size: 1.8rem; color: #f8fafc; margin-bottom: 10px; }
.detail-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 20px; margin-bottom: 24px; }
.detail-item { display: flex; flex-direction: column; gap: 4px; }
.label { font-size: 0.8rem; color: #94a3b8; text-transform: uppercase; font-weight: 600; }
.value { color: #e2e8f0; font-size: 1rem; }
.detail-notes { border-top: 1px solid #334155; padding-top: 20px; }
.detail-notes p { color: #e2e8f0; margin-top: 8px; line-height: 1.6; white-space: pre-wrap; }
.badge { padding: 4px 10px; border-radius: 20px; font-size: 0.78rem; font-weight: 600; }
.badge-lead { background: #1d4ed8; color: #bfdbfe; }
.badge-prospect { background: #7c3aed; color: #ede9fe; }
.badge-customer { background: #065f46; color: #a7f3d0; }
.badge-inactive { background: #374151; color: #9ca3af; }

.history-section { margin-top: 32px; }
.history-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.history-header h2 { color: #f8fafc; font-size: 1.3rem; }

.btn-primary {
    padding: 8px 16px; background: #38bdf8; border: none;
    border-radius: 8px; color: #0f172a; font-weight: bold; cursor: pointer; transition: 0.2s; font-size: 0.9rem;
}
.btn-primary:hover { background: #7dd3fc; }
.btn-secondary {
    padding: 8px 16px; background: transparent; border: 1px solid #475569;
    border-radius: 8px; color: #94a3b8; cursor: pointer; transition: 0.2s; font-size: 0.9rem;
}
.btn-secondary:hover { border-color: #f1f5f9; color: #f1f5f9; }

.add-note-form {
    background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 20px; margin-bottom: 20px;
}
.form-group { margin-bottom: 14px; }
.form-group label { display: block; margin-bottom: 6px; color: #94a3b8; font-size: 0.85rem; }
.form-group select, .form-group textarea {
    width: 100%; padding: 10px 12px; background: #0f172a;
    border: 1px solid #334155; border-radius: 8px; color: #f1f5f9; outline: none; font-size: 0.95rem; resize: vertical;
}
.form-group select:focus, .form-group textarea:focus { border-color: #38bdf8; }
.form-actions { display: flex; gap: 10px; justify-content: flex-end; margin-top: 12px; }

.empty-history { color: #94a3b8; padding: 24px; text-align: center; }

.timeline { display: flex; flex-direction: column; gap: 0; }
.timeline-item { display: flex; gap: 16px; padding: 16px 0; border-bottom: 1px solid #1e293b; }
.timeline-item:last-child { border-bottom: none; }
.timeline-icon {
    width: 36px; height: 36px; border-radius: 50%; background: #1e293b; border: 1px solid #334155;
    display: flex; align-items: center; justify-content: center; flex-shrink: 0;
}
.timeline-icon .material-icons { font-size: 1rem; color: #38bdf8; }
.timeline-body { flex: 1; }
.timeline-meta { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.action-label { font-size: 0.85rem; font-weight: 600; color: #38bdf8; }
.timeline-date { font-size: 0.78rem; color: #64748b; }
.timeline-desc { color: #e2e8f0; font-size: 0.9rem; line-height: 1.5; white-space: pre-wrap; }
</style>
