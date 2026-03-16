<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from "vue-router"

const props = defineProps({
    authFetch: { type: Function, required: true }
})

const router = useRouter()

const contacts = ref([])
const total = ref(0)
const loading = ref(false)
const error = ref("")

const search = ref("")
const filterStatus = ref("")
const sortBy = ref("created_at")
const sortDir = ref("desc")

const showModal = ref(false)
const editingId = ref(null)
const form = reactive({ name: "", email: "", phone: "", company: "", status: "lead", notes: "" })
const formError = ref("")

const statuses = ["lead", "prospect", "customer", "inactive"]

const sortOptions = [
    { value: "created_at", label: "Data dodania" },
    { value: "name", label: "Nazwa" },
    { value: "company", label: "Firma" },
    { value: "status", label: "Status" },
]

async function loadContacts() {
    loading.value = true
    error.value = ""
    const params = new URLSearchParams()
    if (search.value) params.set("search", search.value)
    if (filterStatus.value) params.set("status", filterStatus.value)
    params.set("sort_by", sortBy.value)
    params.set("sort_dir", sortDir.value)
    try {
        const res = await props.authFetch("/api/v1/contacts/?" + params.toString())
        if (!res.ok) throw new Error("Błąd pobierania kontaktów")
        const data = await res.json()
        contacts.value = data.data || []
        total.value = data.total || 0
    } catch (e) {
        error.value = e.message
    } finally {
        loading.value = false
    }
}

onMounted(loadContacts)

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

function toggleSort(col) {
    if (sortBy.value === col) {
        sortDir.value = sortDir.value === "asc" ? "desc" : "asc"
    } else {
        sortBy.value = col
        sortDir.value = "asc"
    }
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
                @input="loadContacts"
            />
            <select v-model="filterStatus" class="filter-select" @change="loadContacts">
                <option value="">Wszystkie statusy</option>
                <option v-for="s in statuses" :key="s" :value="s">{{ statusLabel(s) }}</option>
            </select>
            <select v-model="sortBy" class="filter-select" @change="loadContacts">
                <option v-for="o in sortOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
            </select>
            <button class="btn-sort" @click="sortDir = sortDir === 'asc' ? 'desc' : 'asc'; loadContacts()">
                <span class="material-icons sort-icon">{{ sortDir === 'asc' ? 'arrow_upward' : 'arrow_downward' }}</span>
                {{ sortDir === 'asc' ? 'Rosnąco' : 'Malejąco' }}
            </button>
        </div>

        <p v-if="error" class="err-msg">{{ error }}</p>

        <div v-if="loading" class="loading">Ładowanie...</div>

        <div v-else-if="contacts.length === 0" class="empty">Brak kontaktów.</div>

        <div v-else class="contact-table-wrap">
            <table class="contact-table">
                <thead>
                    <tr>
                        <th @click="toggleSort('name')" class="sortable">
                            Nazwa
                            <span v-if="sortBy === 'name'" class="material-icons th-sort-icon">{{ sortDir === 'asc' ? 'arrow_upward' : 'arrow_downward' }}</span>
                        </th>
                        <th>Email</th>
                        <th>Telefon</th>
                        <th @click="toggleSort('company')" class="sortable">
                            Firma
                            <span v-if="sortBy === 'company'" class="material-icons th-sort-icon">{{ sortDir === 'asc' ? 'arrow_upward' : 'arrow_downward' }}</span>
                        </th>
                        <th @click="toggleSort('status')" class="sortable">
                            Status
                            <span v-if="sortBy === 'status'" class="material-icons th-sort-icon">{{ sortDir === 'asc' ? 'arrow_upward' : 'arrow_downward' }}</span>
                        </th>
                        <th>Akcje</th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="c in contacts" :key="c.id">
                        <td class="name-cell">{{ c.name }}</td>
                        <td>{{ c.email }}</td>
                        <td>{{ c.phone }}</td>
                        <td>{{ c.company }}</td>
                        <td><span :class="['badge', statusClass(c.status)]">{{ statusLabel(c.status) }}</span></td>
                        <td class="actions-cell">
                            <button class="btn-action btn-detail" @click="openDetail(c.id)" title="Szczegóły">
                                <span class="material-icons">visibility</span>
                            </button>
                            <button class="btn-action btn-edit" @click="openEdit(c)" title="Edytuj">
                                <span class="material-icons">edit</span>
                            </button>
                            <button class="btn-action btn-delete" @click="deleteContact(c.id)" title="Usuń">
                                <span class="material-icons">delete</span>
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <div v-if="showModal" class="modal-overlay" @click.self="showModal = false">
            <div class="modal">
                <h2>{{ editingId ? 'Edytuj kontakt' : 'Nowy kontakt' }}</h2>
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
                    <button class="btn-primary" @click="saveContact">{{ editingId ? 'Zapisz' : 'Dodaj' }}</button>
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

.loading, .empty { color: #94a3b8; padding: 40px; text-align: center; }
.err-msg { color: #f87171; font-size: 0.9rem; margin-bottom: 12px; }

.contact-table-wrap { overflow-x: auto; border-radius: 12px; border: 1px solid #334155; }
.contact-table { width: 100%; border-collapse: collapse; }
.contact-table th {
    background: #1e293b; padding: 12px 16px; text-align: left;
    color: #94a3b8; font-size: 0.85rem; font-weight: 600; text-transform: uppercase;
    border-bottom: 1px solid #334155;
}
.contact-table th.sortable { cursor: pointer; user-select: none; }
.contact-table th.sortable:hover { color: #38bdf8; }
.th-sort-icon { font-size: 0.9rem; vertical-align: middle; margin-left: 4px; }
.contact-table td { padding: 14px 16px; border-bottom: 1px solid #1e293b; color: #e2e8f0; font-size: 0.95rem; }
.contact-table tr:last-child td { border-bottom: none; }
.contact-table tr:hover td { background: #1e293b44; }
.name-cell { font-weight: 500; color: #f8fafc; }

.badge { padding: 4px 10px; border-radius: 20px; font-size: 0.78rem; font-weight: 600; }
.badge-lead { background: #1d4ed8; color: #bfdbfe; }
.badge-prospect { background: #7c3aed; color: #ede9fe; }
.badge-customer { background: #065f46; color: #a7f3d0; }
.badge-inactive { background: #374151; color: #9ca3af; }

.actions-cell { display: flex; gap: 8px; }
.btn-action {
    background: none; border: none; cursor: pointer;
    padding: 4px 8px; border-radius: 6px; transition: 0.15s;
    display: flex; align-items: center; color: #94a3b8;
}
.btn-action .material-icons { font-size: 1.1rem; }
.btn-detail:hover { background: #1e3a5f; color: #38bdf8; }
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
.form-group input, .form-group select, .form-group textarea {
    width: 100%; padding: 10px 12px; background: #0f172a;
    border: 1px solid #334155; border-radius: 8px; color: #f1f5f9; outline: none; font-size: 0.95rem;
    resize: vertical;
}
.form-group input:focus, .form-group select:focus, .form-group textarea:focus { border-color: #38bdf8; }
.modal-actions { display: flex; gap: 12px; justify-content: flex-end; margin-top: 20px; }
</style>
