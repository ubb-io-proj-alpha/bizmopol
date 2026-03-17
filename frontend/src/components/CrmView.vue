<script setup>
import { ref, reactive, onMounted } from "vue"
import { useRouter } from "vue-router"
import { useTableQuery } from "../composables/useTableQuery.js"

const props = defineProps({
    authFetch: { type: Function, required: true }
})

const router = useRouter()

const contacts = ref([])
const total = ref(0)
const totalPages = ref(1)
const loading = ref(false)
const error = ref("")

const { search, filterStatus, sortBy, sortDir, page, pageSize, pushQuery, buildParams } = useTableQuery({
    sortBy: "created_at", sortDir: "desc", page: 1, pageSize: 20
})

const showModal = ref(false)
const editingId = ref(null)
const form = reactive({ name: "", email: "", phone: "", company: "", status: "lead", notes: "" })
const formError = ref("")

const showMergeModal = ref(false)
const mergeForm = reactive({ groupName: "", company: "", status: "customer" })
const mergeError = ref("")

const selectedIds = ref(new Set())

const statuses = ["lead", "prospect", "customer", "inactive"]

async function loadContacts() {
    loading.value = true
    error.value = ""
    pushQuery()
    try {
        const res = await props.authFetch("/api/v1/contacts/?" + buildParams().toString())
        if (!res.ok) throw new Error("Błąd pobierania kontaktów")
        const data = await res.json()
        contacts.value = data.data || []
        total.value = data.total || 0
        totalPages.value = data.total_pages || 1
    } catch (e) {
        error.value = e.message
    } finally {
        loading.value = false
    }
}

onMounted(loadContacts)

function onSearch() {
    page.value = 1
    loadContacts()
}

function toggleSelect(id) {
    const next = new Set(selectedIds.value)
    if (next.has(id)) {
        next.delete(id)
    } else {
        next.add(id)
    }
    selectedIds.value = next
}

function toggleSelectAll() {
    if (selectedIds.value.size === contacts.value.length) {
        selectedIds.value = new Set()
    } else {
        selectedIds.value = new Set(contacts.value.map(c => c.id))
    }
}

function openMergeModal() {
    mergeForm.groupName = ""
    mergeForm.company = ""
    mergeForm.status = "customer"
    mergeError.value = ""
    showMergeModal.value = true
}

async function doMerge() {
    mergeError.value = ""
    if (!mergeForm.groupName.trim()) {
        mergeError.value = "Nazwa grupy jest wymagana."
        return
    }
    try {
        const res = await props.authFetch("/api/v1/contacts/merge", {
            method: "POST",
            body: JSON.stringify({
                contact_ids: Array.from(selectedIds.value),
                group_name: mergeForm.groupName,
                company: mergeForm.company,
                status: mergeForm.status,
            }),
        })
        if (!res.ok) {
            const d = await res.json()
            throw new Error(d.error || "Błąd scalania")
        }
        const group = await res.json()
        showMergeModal.value = false
        selectedIds.value = new Set()
        await loadContacts()
        router.push({ name: "contactDetail", params: { id: group.id } })
    } catch (e) {
        mergeError.value = e.message
    }
}

function openCreate() {
    editingId.value = null
    Object.assign(form, { name: "", email: "", phone: "", company: "", status: "lead", notes: "" })
    formError.value = ""
    showModal.value = true
}

function openEdit(c) {
    editingId.value = c.id
    Object.assign(form, { name: c.name, email: c.email, phone: c.phone, company: c.company, status: c.status, notes: c.notes })
    formError.value = ""
    showModal.value = true
}

function openDetail(id) {
    router.push({ name: "contactDetail", params: { id } })
}

async function saveContact() {
    formError.value = ""
    if (!form.name.trim()) {
        formError.value = "Nazwa jest wymagana."
        return
    }
    try {
        let res
        if (editingId.value) {
            res = await props.authFetch("/api/v1/contacts/" + editingId.value, {
                method: "PUT",
                body: JSON.stringify(form),
            })
        } else {
            res = await props.authFetch("/api/v1/contacts/", {
                method: "POST",
                body: JSON.stringify(form),
            })
        }
        if (!res.ok) {
            const d = await res.json()
            throw new Error(d.error || "Błąd zapisu")
        }
        showModal.value = false
        await loadContacts()
    } catch (e) {
        formError.value = e.message
    }
}

async function deleteContact(id) {
    if (!confirm("Usunąć kontakt?")) return
    try {
        const res = await props.authFetch("/api/v1/contacts/" + id, { method: "DELETE" })
        if (!res.ok && res.status !== 204) throw new Error("Błąd usuwania")
        selectedIds.value.delete(id)
        await loadContacts()
    } catch (e) {
        error.value = e.message
    }
}

function onSort(col) {
    if (sortBy.value === col) {
        sortDir.value = sortDir.value === "asc" ? "desc" : "asc"
    } else {
        sortBy.value = col
        sortDir.value = "asc"
    }
    page.value = 1
    loadContacts()
}

function goToPage(p) {
    if (p < 1 || p > totalPages.value) return
    page.value = p
    loadContacts()
}

function statusLabel(s) {
    const map = { lead: "Lead", prospect: "Prospect", customer: "Klient", inactive: "Nieaktywny" }
    return map[s] || s
}

function statusClass(s) {
    const map = { lead: "badge-lead", prospect: "badge-prospect", customer: "badge-customer", inactive: "badge-inactive" }
    return map[s] || ""
}

function scoreColor(score) {
    if (score >= 70) return "#22c55e"
    if (score >= 40) return "#f59e0b"
    return "#ef4444"
}

function scoreLabel(score) {
    if (score >= 70) return "Wysoki"
    if (score >= 40) return "Średni"
    return "Niski"
}
</script>

<template>
    <div class="crm-wrapper">
        <div class="crm-header">
            <div>
                <h1>CRM — Kontakty</h1>
                <p class="subtitle">Łącznie: {{ total }} kontaktów</p>
            </div>
            <div class="header-actions">
                <button
                    v-if="selectedIds.size >= 2"
                    class="btn-merge"
                    @click="openMergeModal"
                >
                    <span class="material-icons">merge</span>
                    Scal zaznaczone ({{ selectedIds.size }})
                </button>
                <button class="btn-primary" @click="openCreate">+ Nowy kontakt</button>
            </div>
        </div>

        <div class="crm-filters">
            <input
                v-model="search"
                class="filter-input"
                placeholder="Szukaj po nazwie, emailu, firmie..."
                @input="onSearch"
            />
            <select v-model="filterStatus" class="filter-select" @change="onSearch">
                <option value="">Wszystkie statusy</option>
                <option v-for="s in statuses" :key="s" :value="s">{{ statusLabel(s) }}</option>
            </select>
            <select v-model="sortBy" class="filter-select" @change="() => { page = 1; loadContacts() }">
                <option value="created_at">Data dodania</option>
                <option value="name">Nazwa</option>
                <option value="company">Firma</option>
                <option value="status">Status</option>
                <option value="lead_score">Lead Score</option>
            </select>
            <button class="btn-sort" @click="sortDir = sortDir === 'asc' ? 'desc' : 'asc'; page = 1; loadContacts()">
                <span class="material-icons sort-icon">{{ sortDir === "asc" ? "arrow_upward" : "arrow_downward" }}</span>
                {{ sortDir === "asc" ? "Rosnąco" : "Malejąco" }}
            </button>
            <select v-model="pageSize" class="filter-select" @change="() => { page = 1; loadContacts() }">
                <option :value="10">10 / str.</option>
                <option :value="20">20 / str.</option>
                <option :value="50">50 / str.</option>
                <option :value="100">100 / str.</option>
            </select>
        </div>

        <p v-if="error" class="err-msg">{{ error }}</p>

        <div v-if="!loading && contacts.length > 0" class="table-wrap">
            <table class="crud-table">
                <thead>
                    <tr>
                        <th class="col-check">
                            <input
                                type="checkbox"
                                :checked="selectedIds.size === contacts.length && contacts.length > 0"
                                :indeterminate="selectedIds.size > 0 && selectedIds.size < contacts.length"
                                @change="toggleSelectAll"
                            />
                        </th>
                        <th>Nazwa</th>
                        <th>Email</th>
                        <th>Telefon</th>
                        <th>Firma</th>
                        <th>Status</th>
                        <th>Lead Score</th>
                        <th>Akcje</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="row in contacts" :key="row.id" :class="{ selected: selectedIds.has(row.id) }">
                        <td class="col-check">
                            <input
                                type="checkbox"
                                :checked="selectedIds.has(row.id)"
                                @change="toggleSelect(row.id)"
                            />
                        </td>
                        <td class="name-cell">
                            <span v-if="row.is_group" class="group-icon material-icons" title="Grupa/Firma">corporate_fare</span>
                            {{ row.name }}
                        </td>
                        <td>{{ row.email }}</td>
                        <td>{{ row.phone }}</td>
                        <td>{{ row.company }}</td>
                        <td><span :class="['badge', statusClass(row.status)]">{{ statusLabel(row.status) }}</span></td>
                        <td>
                            <div class="score-cell">
                                <div class="score-bar-wrap">
                                    <div class="score-bar" :style="{ width: row.lead_score + '%', background: scoreColor(row.lead_score) }"></div>
                                </div>
                                <span class="score-value" :style="{ color: scoreColor(row.lead_score) }">{{ row.lead_score }}</span>
                            </div>
                        </td>
                        <td class="actions-cell">
                            <button class="btn-action btn-detail" @click="openDetail(row.id)" title="Szczegóły">
                                <span class="material-icons">visibility</span>
                            </button>
                            <button class="btn-action btn-edit" @click="openEdit(row)" title="Edytuj">
                                <span class="material-icons">edit</span>
                            </button>
                            <button class="btn-action btn-delete" @click="deleteContact(row.id)" title="Usuń">
                                <span class="material-icons">delete</span>
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
        <div v-else-if="loading" class="loading">Ładowanie...</div>
        <div v-else class="empty">Brak kontaktów.</div>

        <div v-if="totalPages > 1" class="pagination">
            <button class="page-btn" :disabled="page <= 1" @click="goToPage(page - 1)">
                <span class="material-icons">chevron_left</span>
            </button>
            <template v-for="p in totalPages" :key="p">
                <button
                    v-if="p === 1 || p === totalPages || Math.abs(p - page) <= 2"
                    class="page-btn"
                    :class="{ active: p === page }"
                    @click="goToPage(p)"
                >{{ p }}</button>
                <span v-else-if="Math.abs(p - page) === 3" class="page-ellipsis">…</span>
            </template>
            <button class="page-btn" :disabled="page >= totalPages" @click="goToPage(page + 1)">
                <span class="material-icons">chevron_right</span>
            </button>
        </div>

        <div v-if="showMergeModal" class="modal-overlay" @click.self="showMergeModal = false">
            <div class="modal">
                <h2>Scal kontakty w grupę</h2>
                <p class="merge-info">Scalasz {{ selectedIds.size }} kontaktów. Zostaną połączone jako firma/grupa. Historia i dane każdego kontaktu będą widoczne w widoku szczegółów.</p>
                <div class="form-group">
                    <label>Nazwa grupy / firmy *</label>
                    <input v-model="mergeForm.groupName" placeholder="np. ABC Corp" />
                </div>
                <div class="form-group">
                    <label>Firma (opcjonalnie)</label>
                    <input v-model="mergeForm.company" placeholder="Nazwa firmy" />
                </div>
                <div class="form-group">
                    <label>Status grupy</label>
                    <select v-model="mergeForm.status">
                        <option v-for="s in statuses" :key="s" :value="s">{{ statusLabel(s) }}</option>
                    </select>
                </div>
                <p v-if="mergeError" class="err-msg">{{ mergeError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showMergeModal = false">Anuluj</button>
                    <button class="btn-primary" @click="doMerge">Scal</button>
                </div>
            </div>
        </div>

        <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
            <div class="modal">
                <h2>{{ editingId ? "Edytuj kontakt" : "Nowy kontakt" }}</h2>
                <div class="form-group">
                    <label>Nazwa *</label>
                    <input v-model="form.name" placeholder="Jan Kowalski" />
                </div>
                <div class="form-group">
                    <label>Email</label>
                    <input v-model="form.email" type="email" placeholder="jan@firma.pl" />
                </div>
                <div class="form-group">
                    <label>Telefon</label>
                    <input v-model="form.phone" placeholder="+48 000 000 000" />
                </div>
                <div class="form-group">
                    <label>Firma</label>
                    <input v-model="form.company" placeholder="Nazwa firmy" />
                </div>
                <div class="form-group">
                    <label>Status</label>
                    <select v-model="form.status">
                        <option v-for="s in statuses" :key="s" :value="s">{{ statusLabel(s) }}</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Notatki</label>
                    <textarea v-model="form.notes" rows="3" placeholder="Dodatkowe informacje..."></textarea>
                </div>
                <p v-if="formError" class="err-msg">{{ formError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showModal = false">Anuluj</button>
                    <button class="btn-primary" @click="saveContact">{{ editingId ? "Zapisz" : "Dodaj" }}</button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.crm-wrapper { width: 100%; }
.crm-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; }
.crm-header h1 { font-size: 2rem; color: #f8fafc; margin-bottom: 4px; }
.subtitle { color: #94a3b8; font-size: 0.95rem; }
.header-actions { display: flex; gap: 10px; align-items: center; }

.crm-filters { display: flex; gap: 12px; margin-bottom: 20px; flex-wrap: wrap; }
.filter-input {
    flex: 1; min-width: 200px; padding: 10px 14px;
    background: #1e293b; border: 1px solid #334155; border-radius: 8px;
    color: #f1f5f9; outline: none;
}
.filter-input:focus { border-color: #38bdf8; }
.filter-select {
    padding: 10px 14px; background: #1e293b; border: 1px solid #334155;
    border-radius: 8px; color: #f1f5f9; outline: none; cursor: pointer;
}
.btn-sort {
    padding: 10px 16px; background: #1e293b; border: 1px solid #334155;
    border-radius: 8px; color: #f1f5f9; cursor: pointer; transition: 0.2s;
    display: flex; align-items: center; gap: 6px;
}
.btn-sort:hover { border-color: #38bdf8; color: #38bdf8; }
.sort-icon { font-size: 1rem; }

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
.btn-merge {
    padding: 10px 18px; background: #7c3aed; border: none;
    border-radius: 8px; color: #ede9fe; font-weight: bold; cursor: pointer; transition: 0.2s;
    display: flex; align-items: center; gap: 6px;
}
.btn-merge:hover { background: #6d28d9; }
.btn-merge .material-icons { font-size: 1.1rem; }

.err-msg { color: #f87171; font-size: 0.9rem; margin-bottom: 12px; }

.loading, .empty { color: #94a3b8; padding: 40px; text-align: center; }
.table-wrap { overflow-x: auto; border-radius: 12px; border: 1px solid #334155; }
.crud-table { width: 100%; border-collapse: collapse; }
.crud-table th {
    background: #1e293b; padding: 12px 16px; text-align: left;
    color: #94a3b8; font-size: 0.85rem; font-weight: 600; text-transform: uppercase;
    border-bottom: 1px solid #334155;
}
.crud-table td { padding: 14px 16px; border-bottom: 1px solid #1e293b; color: #e2e8f0; font-size: 0.95rem; }
.crud-table tr:last-child td { border-bottom: none; }
.crud-table tr:hover td { background: #1e293b44; }
.crud-table tr.selected td { background: #1e293b88; }
.col-check { width: 40px; padding: 12px 8px !important; }
.col-check input[type="checkbox"] { cursor: pointer; width: 16px; height: 16px; accent-color: #38bdf8; }
.name-cell { display: flex; align-items: center; gap: 6px; }
.group-icon { font-size: 1rem; color: #7c3aed; }
.actions-cell { display: flex; gap: 8px; }

.badge { padding: 4px 10px; border-radius: 20px; font-size: 0.78rem; font-weight: 600; }
.badge-lead { background: #1d4ed8; color: #bfdbfe; }
.badge-prospect { background: #7c3aed; color: #ede9fe; }
.badge-customer { background: #065f46; color: #a7f3d0; }
.badge-inactive { background: #374151; color: #9ca3af; }

.score-cell { display: flex; align-items: center; gap: 8px; min-width: 100px; }
.score-bar-wrap { flex: 1; height: 6px; background: #334155; border-radius: 4px; overflow: hidden; }
.score-bar { height: 100%; border-radius: 4px; transition: width 0.3s; }
.score-value { font-size: 0.8rem; font-weight: 700; min-width: 26px; text-align: right; }

.btn-action {
    background: none; border: none; cursor: pointer;
    padding: 4px 8px; border-radius: 6px; transition: 0.15s;
    display: flex; align-items: center; color: #94a3b8;
}
.btn-action .material-icons { font-size: 1.1rem; }
.btn-detail:hover { background: #1e3a5f; color: #38bdf8; }
.btn-edit:hover { background: #1e3a5f; color: #38bdf8; }
.btn-delete:hover { background: #3f1414; color: #f87171; }

.pagination {
    display: flex; align-items: center; justify-content: center;
    gap: 6px; margin-top: 20px; flex-wrap: wrap;
}
.page-btn {
    min-width: 36px; height: 36px; padding: 0 10px;
    background: #1e293b; border: 1px solid #334155; border-radius: 8px;
    color: #f1f5f9; cursor: pointer; transition: 0.15s;
    display: flex; align-items: center; justify-content: center; font-size: 0.9rem;
}
.page-btn:hover:not(:disabled) { border-color: #38bdf8; color: #38bdf8; }
.page-btn.active { background: #38bdf8; color: #0f172a; border-color: #38bdf8; font-weight: bold; }
.page-btn:disabled { opacity: 0.4; cursor: default; }
.page-ellipsis { color: #94a3b8; padding: 0 4px; line-height: 36px; }

.modal-overlay {
    position: fixed; inset: 0; background: rgba(0,0,0,0.6);
    display: flex; align-items: center; justify-content: center; z-index: 100;
}
.modal {
    background: #1e293b; border: 1px solid #334155; border-radius: 16px;
    padding: 32px; width: 100%; max-width: 480px; max-height: 90vh; overflow-y: auto;
}
.modal h2 { color: #38bdf8; margin-bottom: 12px; font-size: 1.4rem; }
.merge-info { color: #94a3b8; font-size: 0.9rem; margin-bottom: 20px; line-height: 1.5; }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; margin-bottom: 6px; color: #94a3b8; font-size: 0.88rem; }
.form-group input, .form-group select, .form-group textarea {
    width: 100%; padding: 10px 12px; background: #0f172a;
    border: 1px solid #334155; border-radius: 8px; color: #f1f5f9; outline: none; font-size: 0.95rem;
    resize: vertical;
}
.form-group input:focus, .form-group select:focus, .form-group textarea:focus { border-color: #38bdf8; }
.modal-actions { display: flex; gap: 12px; justify-content: flex-end; margin-top: 20px; }
</style>
