<script setup>
import { ref, reactive, onMounted } from "vue"
import { useRouter } from "vue-router"
import CrudTable from "./CrudTable.vue"
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

const statuses = ["lead", "prospect", "customer", "inactive"]

const columns = [
    { key: "name", label: "Nazwa", sortable: true, cellClass: "name-cell" },
    { key: "email", label: "Email" },
    { key: "phone", label: "Telefon" },
    { key: "company", label: "Firma", sortable: true },
    { key: "status", label: "Status", sortable: true },
]

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
</script>

<template>
    <div class="crm-wrapper">
        <div class="crm-header">
            <div>
                <h1>CRM — Kontakty</h1>
                <p class="subtitle">Łącznie: {{ total }} kontaktów</p>
            </div>
            <button class="btn-primary" @click="openCreate">+ Nowy kontakt</button>
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

        <CrudTable
            :columns="columns"
            :rows="contacts"
            :sort-by="sortBy"
            :sort-dir="sortDir"
            :loading="loading"
            empty-text="Brak kontaktów."
            @sort="onSort"
        >
            <template #cell-status="{ row }">
                <span :class="['badge', statusClass(row.status)]">{{ statusLabel(row.status) }}</span>
            </template>
            <template #actions="{ row }">
                <button class="btn-action btn-detail" @click="openDetail(row.id)" title="Szczegóły">
                    <span class="material-icons">visibility</span>
                </button>
                <button class="btn-action btn-edit" @click="openEdit(row)" title="Edytuj">
                    <span class="material-icons">edit</span>
                </button>
                <button class="btn-action btn-delete" @click="deleteContact(row.id)" title="Usuń">
                    <span class="material-icons">delete</span>
                </button>
            </template>
        </CrudTable>

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

.err-msg { color: #f87171; font-size: 0.9rem; margin-bottom: 12px; }

.badge { padding: 4px 10px; border-radius: 20px; font-size: 0.78rem; font-weight: 600; }
.badge-lead { background: #1d4ed8; color: #bfdbfe; }
.badge-prospect { background: #7c3aed; color: #ede9fe; }
.badge-customer { background: #065f46; color: #a7f3d0; }
.badge-inactive { background: #374151; color: #9ca3af; }

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
.modal h2 { color: #38bdf8; margin-bottom: 20px; font-size: 1.4rem; }
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
