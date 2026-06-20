<script setup>
import { ref, reactive, onMounted, computed } from "vue"
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

const allTags = ref([])
const customFields = ref([])
const visibleFields = computed(() => customFields.value.filter(f => f.visible))

const { search, filterStatus, sortBy, sortDir, page, pageSize, pushQuery, buildParams } = useTableQuery({
    sortBy: "created_at", sortDir: "desc", page: 1, pageSize: 20
})

const filterTags = ref([])

const showModal = ref(false)
const editingId = ref(null)
const form = reactive({ name: "", email: "", phone: "", company: "", status: "lead", notes: "", tagIds: [], customValues: {} })
const formError = ref("")

const showMergeModal = ref(false)
const mergeForm = reactive({ groupName: "", company: "", status: "customer" })
const mergeError = ref("")

const showDndModal = ref(false)
const dndContactId = ref(null)
const dndForm = reactive({ dnd_active: true, dnd_type: "permanent", dnd_reason: "", dnd_until: "" })
const dndError = ref("")

const selectedIds = ref(new Set())

const statuses = ["lead", "prospect", "customer", "inactive"]

async function loadMeta() {
    try {
        const [tr, fr] = await Promise.all([
            props.authFetch("/api/v1/tags/"),
            props.authFetch("/api/v1/custom-fields/"),
        ])
        if (tr.ok) allTags.value = await tr.json()
        if (fr.ok) customFields.value = await fr.json()
    } catch (_) {}
}

async function loadContacts() {
    loading.value = true
    error.value = ""
    pushQuery()
    try {
        const params = buildParams()
        if (filterTags.value.length > 0) {
            params.set("tags", filterTags.value.join(","))
        }
        const res = await props.authFetch("/api/v1/contacts/?" + params.toString())
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

onMounted(async () => {
    await loadMeta()
    loadContacts()
})

function onSearch() {
    page.value = 1
    loadContacts()
}

function toggleSelect(id) {
    const next = new Set(selectedIds.value)
    if (next.has(id)) next.delete(id)
    else next.add(id)
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
    Object.assign(form, { name: "", email: "", phone: "", company: "", status: "lead", notes: "", tagIds: [], customValues: {} })
    formError.value = ""
    showModal.value = true
}

function openEdit(c) {
    editingId.value = c.id
    const tagIds = (c.tags || []).map(t => t.id)
    const customValues = {}
    for (const cv of (c.custom_values || [])) {
        customValues[cv.field_id] = cv.value
    }
    Object.assign(form, { name: c.name, email: c.email, phone: c.phone, company: c.company, status: c.status, notes: c.notes, tagIds, customValues })
    formError.value = ""
    showModal.value = true
}

function openDetail(id) {
    router.push({ name: "contactDetail", params: { id } })
}

function openDndModal(c) {
    dndContactId.value = c.id
    dndForm.dnd_active = c.dnd_active || false
    dndForm.dnd_type = c.dnd_type || "permanent"
    dndForm.dnd_reason = c.dnd_reason || ""
    dndForm.dnd_until = c.dnd_until ? c.dnd_until.slice(0, 10) : ""
    dndError.value = ""
    showDndModal.value = true
}

async function saveDnd() {
    dndError.value = ""
    try {
        const payload = {
            dnd_active: dndForm.dnd_active,
            dnd_type: dndForm.dnd_type,
            dnd_reason: dndForm.dnd_reason,
            dnd_until: dndForm.dnd_type === "temporary" && dndForm.dnd_until ? new Date(dndForm.dnd_until).toISOString() : null,
        }
        const res = await props.authFetch("/api/v1/contacts/" + dndContactId.value + "/dnd", {
            method: "PUT",
            body: JSON.stringify(payload),
        })
        if (!res.ok) {
            const d = await res.json()
            throw new Error(d.error || "Błąd zapisu DND")
        }
        showDndModal.value = false
        await loadContacts()
    } catch (e) {
        dndError.value = e.message
    }
}

async function saveContact() {
    formError.value = ""
    if (!form.name.trim()) {
        formError.value = "Nazwa jest wymagana."
        return
    }
    try {
        const payload = {
            name: form.name,
            email: form.email,
            phone: form.phone,
            company: form.company,
            status: form.status,
            notes: form.notes,
            tag_ids: form.tagIds,
            custom_values: form.customValues,
        }
        let res
        if (editingId.value) {
            res = await props.authFetch("/api/v1/contacts/" + editingId.value, {
                method: "PUT",
                body: JSON.stringify(payload),
            })
        } else {
            res = await props.authFetch("/api/v1/contacts/", {
                method: "POST",
                body: JSON.stringify(payload),
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

function toggleTagFilter(tagName) {
    const idx = filterTags.value.indexOf(tagName)
    if (idx === -1) filterTags.value.push(tagName)
    else filterTags.value.splice(idx, 1)
    page.value = 1
    loadContacts()
}

function toggleFormTag(tagId) {
    const idx = form.tagIds.indexOf(tagId)
    if (idx === -1) form.tagIds.push(tagId)
    else form.tagIds.splice(idx, 1)
}

function getCustomValue(row, fieldId) {
    if (!row.custom_values) return ""
    const cv = row.custom_values.find(v => v.field_id === fieldId)
    return cv ? cv.value : ""
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

        <div v-if="allTags.length > 0" class="tag-filters">
            <span class="tag-filter-label">Filtruj po tagach:</span>
            <button
                v-for="t in allTags"
                :key="t.id"
                :class="['tag-filter-chip', { active: filterTags.includes(t.name) }]"
                :style="{ '--tag-color': t.color }"
                @click="toggleTagFilter(t.name)"
            >
                {{ t.name }}
            </button>
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
                        <th class="sortable" @click="onSort('name')">
                            Nazwa
                            <span v-if="sortBy === 'name'" class="material-icons th-sort-icon">{{ sortDir === "asc" ? "arrow_upward" : "arrow_downward" }}</span>
                        </th>
                        <th>Email</th>
                        <th>Telefon</th>
                        <th class="sortable" @click="onSort('company')">
                            Firma
                            <span v-if="sortBy === 'company'" class="material-icons th-sort-icon">{{ sortDir === "asc" ? "arrow_upward" : "arrow_downward" }}</span>
                        </th>
                        <th class="sortable" @click="onSort('status')">
                            Status
                            <span v-if="sortBy === 'status'" class="material-icons th-sort-icon">{{ sortDir === "asc" ? "arrow_upward" : "arrow_downward" }}</span>
                        </th>
                        <th>Tagi</th>
                        <th class="sortable" @click="onSort('lead_score')">
                            Lead Score
                            <span v-if="sortBy === 'lead_score'" class="material-icons th-sort-icon">{{ sortDir === "asc" ? "arrow_upward" : "arrow_downward" }}</span>
                        </th>
                        <th v-for="cf in visibleFields" :key="cf.id">
                            {{ cf.name }}
                        </th>
                        <th>Akcje</th>
                    </tr>
                </thead>
                <tbody>
                    <tr
                        v-for="row in contacts"
                        :key="row.id"
                        :class="{ selected: selectedIds.has(row.id), 'dnd-row': row.dnd_active }"
                    >
                        <td>
                            <div class="col-check">
                                <input
                                    type="checkbox"
                                    :checked="selectedIds.has(row.id)"
                                    @change="toggleSelect(row.id)"
                                />
                            </div>
                        </td>
                        <td>
                            <div class="name-cell">
                                <span v-if="row.is_group" class="group-icon material-icons" title="Grupa/Firma">corporate_fare</span>
                                <span v-if="row.dnd_active" class="dnd-icon material-icons" title="Do Not Disturb">do_not_disturb_on</span>
                                {{ row.name }}
                            </div>
                        </td>
                        <td>{{ row.email }}</td>
                        <td>{{ row.phone }}</td>
                        <td>{{ row.company }}</td>
                        <td><span :class="['badge', statusClass(row.status)]">{{ statusLabel(row.status) }}</span></td>
                        <td>
                            <div class="tags-cell">
                                <span
                                    v-for="t in (row.tags || [])"
                                    :key="t.id"
                                    class="tag-chip-small"
                                    :style="{ background: t.color + '33', color: t.color, borderColor: t.color }"
                                >{{ t.name }}</span>
                            </div>
                        </td>
                        <td>
                            <div class="score-cell">
                                <div class="score-bar-wrap">
                                    <div class="score-bar" :style="{ width: row.lead_score + '%', background: scoreColor(row.lead_score) }"></div>
                                </div>
                                <span class="score-value" :style="{ color: scoreColor(row.lead_score) }">{{ row.lead_score }}</span>
                            </div>
                        </td>
                        <td v-for="cf in visibleFields" :key="cf.id">
                            {{ getCustomValue(row, cf.id) }}
                        </td>
                        <td>
                            <div class="actions-cell">
                                <button class="btn-action btn-detail" @click="openDetail(row.id)" title="Szczegóły">
                                    <span class="material-icons">visibility</span>
                                </button>
                                <button class="btn-action btn-edit" @click="openEdit(row)" title="Edytuj">
                                    <span class="material-icons">edit</span>
                                </button>
                                <button
                                    class="btn-action"
                                    :class="row.dnd_active ? 'btn-dnd-active' : 'btn-dnd'"
                                    @click="openDndModal(row)"
                                    title="Do Not Disturb"
                                >
                                    <span class="material-icons">do_not_disturb_on</span>
                                </button>
                                <button class="btn-action btn-delete" @click="deleteContact(row.id)" title="Usuń">
                                    <span class="material-icons">delete</span>
                                </button>
                            </div>
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

        <div v-if="showDndModal" class="modal-overlay" @click.self="showDndModal = false">
            <div class="modal">
                <h2>Do Not Disturb</h2>
                <div class="form-group">
                    <label>Status DND</label>
                    <div class="toggle-row">
                        <label class="toggle-label">
                            <input type="checkbox" v-model="dndForm.dnd_active" class="toggle-input" />
                            <span class="toggle-track">
                                <span class="toggle-thumb"></span>
                            </span>
                            <span class="toggle-text">{{ dndForm.dnd_active ? "Aktywny (DND włączony)" : "Nieaktywny" }}</span>
                        </label>
                    </div>
                </div>
                <div v-if="dndForm.dnd_active">
                    <div class="form-group">
                        <label>Typ blokady</label>
                        <div class="dnd-type-group">
                            <button
                                :class="['dnd-type-btn', { active: dndForm.dnd_type === 'permanent' }]"
                                type="button"
                                @click="dndForm.dnd_type = 'permanent'"
                            >
                                <span class="material-icons">block</span>
                                Permanentny
                            </button>
                            <button
                                :class="['dnd-type-btn', { active: dndForm.dnd_type === 'temporary' }]"
                                type="button"
                                @click="dndForm.dnd_type = 'temporary'"
                            >
                                <span class="material-icons">schedule</span>
                                Tymczasowy
                            </button>
                        </div>
                    </div>
                    <div v-if="dndForm.dnd_type === 'temporary'" class="form-group">
                        <label>Blokada do dnia</label>
                        <input type="date" v-model="dndForm.dnd_until" />
                    </div>
                    <div class="form-group">
                        <label>Powód (opcjonalnie)</label>
                        <textarea v-model="dndForm.dnd_reason" rows="3" placeholder="Np. prośba klienta, rezygnacja..."></textarea>
                    </div>
                </div>
                <p v-if="dndError" class="err-msg">{{ dndError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showDndModal = false">Anuluj</button>
                    <button class="btn-primary" @click="saveDnd">Zapisz</button>
                </div>
            </div>
        </div>

        <div v-if="showMergeModal" class="modal-overlay" @click.self="showMergeModal = false">
            <div class="modal">
                <h2>Scal kontakty w grupę</h2>
                <p class="merge-info">Scalasz {{ selectedIds.size }} kontaktów.</p>
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
                <div v-if="allTags.length > 0" class="form-group">
                    <label>Tagi</label>
                    <div class="tags-picker">
                        <button
                            v-for="t in allTags"
                            :key="t.id"
                            type="button"
                            :class="['tag-pick-btn', { selected: form.tagIds.includes(t.id) }]"
                            :style="{ '--tag-color': t.color }"
                            @click="toggleFormTag(t.id)"
                        >{{ t.name }}</button>
                    </div>
                </div>
                <div v-if="customFields.length > 0" class="form-group">
                    <label>Pola niestandardowe</label>
                    <div v-for="cf in customFields" :key="cf.id" class="cf-row">
                        <label class="cf-label">{{ cf.name }}</label>
                        <input
                            v-model="form.customValues[cf.id]"
                            :type="cf.field_type === 'number' ? 'number' : cf.field_type === 'date' ? 'date' : cf.field_type === 'url' ? 'url' : cf.field_type === 'email' ? 'email' : 'text'"
                            :placeholder="cf.name"
                            class="cf-input"
                        />
                    </div>
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

.crm-filters { display: flex; gap: 12px; margin-bottom: 12px; flex-wrap: wrap; }
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

.tag-filters { display: flex; flex-wrap: wrap; gap: 8px; margin-bottom: 16px; align-items: center; }
.tag-filter-label { color: #64748b; font-size: 0.82rem; }
.tag-filter-chip {
    padding: 4px 12px; border-radius: 20px; font-size: 0.8rem; font-weight: 600;
    cursor: pointer; border: 1px solid var(--tag-color); color: var(--tag-color);
    background: transparent; transition: 0.15s;
}
.tag-filter-chip.active { background: var(--tag-color); color: #fff; }

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
    border-bottom: 1px solid #334155; white-space: nowrap;
}
.crud-table th.sortable { cursor: pointer; user-select: none; }
.crud-table th.sortable:hover { color: #38bdf8; }
.th-sort-icon { font-size: 0.85rem; vertical-align: middle; margin-left: 2px; }
.crud-table td { padding: 14px 16px; border-bottom: 1px solid #1e293b; color: #e2e8f0; font-size: 0.95rem; }
.crud-table tr:last-child td { border-bottom: none; }
.crud-table tr:hover td { background: #1e293b44; }
.crud-table tr.selected td { background: #1e293b88; }
.crud-table tr.dnd-row td { opacity: 0.45; }
.crud-table tr.dnd-row:hover td { opacity: 0.6; }
.col-check { width: 40px; padding: 12px 8px !important; }
.col-check input[type="checkbox"] { cursor: pointer; width: 16px; height: 16px; accent-color: #38bdf8; }
.name-cell { display: flex; align-items: center; gap: 6px; }
.group-icon { font-size: 1rem; color: #7c3aed; }
.dnd-icon { font-size: 1rem; color: #ef4444; }
.tags-cell { display: flex; flex-wrap: wrap; gap: 4px; }
.tag-chip-small {
    display: inline-block; padding: 2px 8px; border-radius: 12px;
    font-size: 0.72rem; font-weight: 600; border: 1px solid; white-space: nowrap;
}
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
.btn-dnd:hover { background: #3f1414; color: #f87171; }
.btn-dnd-active { color: #ef4444; }
.btn-dnd-active:hover { background: #3f1414; color: #f87171; }

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
    padding: 32px; width: 100%; max-width: 520px; max-height: 90vh; overflow-y: auto;
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

.tags-picker { display: flex; flex-wrap: wrap; gap: 8px; }
.tag-pick-btn {
    padding: 4px 14px; border-radius: 20px; font-size: 0.82rem; font-weight: 600;
    cursor: pointer; border: 1px solid var(--tag-color); color: var(--tag-color);
    background: transparent; transition: 0.15s;
}
.tag-pick-btn.selected { background: var(--tag-color); color: #fff; }

.cf-row { display: flex; align-items: center; gap: 10px; margin-bottom: 10px; }
.cf-label { color: #94a3b8; font-size: 0.85rem; min-width: 120px; flex-shrink: 0; }
.cf-input {
    flex: 1; padding: 8px 10px; background: #0f172a;
    border: 1px solid #334155; border-radius: 8px; color: #f1f5f9; outline: none; font-size: 0.9rem;
}
.cf-input:focus { border-color: #38bdf8; }

.dnd-type-group { display: flex; gap: 10px; }
.dnd-type-btn {
    flex: 1; padding: 10px; background: #0f172a; border: 1px solid #334155;
    border-radius: 8px; color: #94a3b8; cursor: pointer; transition: 0.15s;
    display: flex; align-items: center; justify-content: center; gap: 6px; font-size: 0.9rem;
}
.dnd-type-btn:hover { border-color: #ef4444; color: #ef4444; }
.dnd-type-btn.active { border-color: #ef4444; color: #ef4444; background: #3f141422; }
.dnd-type-btn .material-icons { font-size: 1rem; }

.toggle-row { margin-top: 4px; }
.toggle-label { display: flex; align-items: center; gap: 10px; cursor: pointer; }
.toggle-input { display: none; }
.toggle-track {
    width: 44px; height: 24px; background: #334155; border-radius: 12px;
    position: relative; transition: 0.2s; flex-shrink: 0;
}
.toggle-input:checked + .toggle-track { background: #ef4444; }
.toggle-thumb {
    position: absolute; top: 3px; left: 3px; width: 18px; height: 18px;
    background: white; border-radius: 50%; transition: 0.2s;
}
.toggle-input:checked + .toggle-track .toggle-thumb { left: 23px; }
.toggle-text { color: #e2e8f0; font-size: 0.9rem; }
</style>
