<script setup>
import { ref, reactive, onMounted, inject } from "vue"

const authFetch = inject("authFetch")
const confirm = inject("confirm")

const activeTab = ref("tags")

const tags = ref([])
const tagsLoading = ref(false)
const tagsError = ref("")
const showTagModal = ref(false)
const editingTagId = ref(null)
const tagForm = reactive({ name: "", color: "#6366f1" })
const tagFormError = ref("")

const fields = ref([])
const fieldsLoading = ref(false)
const fieldsError = ref("")
const showFieldModal = ref(false)
const editingFieldId = ref(null)
const fieldForm = reactive({ name: "", field_type: "text", visible: true, sort_order: 0 })
const fieldFormError = ref("")

const fieldTypes = ["text", "number", "date", "boolean", "url", "email"]

async function loadTags() {
    tagsLoading.value = true
    tagsError.value = ""
    try {
        const res = await authFetch("/api/v1/tags/")
        if (!res.ok) throw new Error("Błąd pobierania tagów")
        tags.value = await res.json()
    } catch (e) {
        tagsError.value = e.message
    } finally {
        tagsLoading.value = false
    }
}

async function loadFields() {
    fieldsLoading.value = true
    fieldsError.value = ""
    try {
        const res = await authFetch("/api/v1/custom-fields/")
        if (!res.ok) throw new Error("Błąd pobierania pól")
        fields.value = await res.json()
    } catch (e) {
        fieldsError.value = e.message
    } finally {
        fieldsLoading.value = false
    }
}

onMounted(() => {
    loadTags()
    loadFields()
})

function openCreateTag() {
    editingTagId.value = null
    Object.assign(tagForm, { name: "", color: "#6366f1" })
    tagFormError.value = ""
    showTagModal.value = true
}

function openEditTag(t) {
    editingTagId.value = t.id
    Object.assign(tagForm, { name: t.name, color: t.color })
    tagFormError.value = ""
    showTagModal.value = true
}

async function saveTag() {
    tagFormError.value = ""
    if (!tagForm.name.trim()) {
        tagFormError.value = "Nazwa jest wymagana."
        return
    }
    try {
        let res
        if (editingTagId.value) {
            res = await authFetch("/api/v1/tags/" + editingTagId.value, {
                method: "PUT",
                body: JSON.stringify({ name: tagForm.name, color: tagForm.color }),
            })
        } else {
            res = await authFetch("/api/v1/tags/", {
                method: "POST",
                body: JSON.stringify({ name: tagForm.name, color: tagForm.color }),
            })
        }
        if (!res.ok) {
            const d = await res.json()
            throw new Error(d.error || "Błąd zapisu")
        }
        showTagModal.value = false
        await loadTags()
    } catch (e) {
        tagFormError.value = e.message
    }
}

async function deleteTag(id) {
    if (!await confirm({ title: "Usuń tag", message: "Tag zostanie trwale usunięty ze wszystkich kontaktów.", confirmLabel: "Usuń", variant: "danger" })) return
    try {
        await authFetch("/api/v1/tags/" + id, { method: "DELETE" })
        await loadTags()
    } catch (e) {
        tagsError.value = e.message
    }
}

function openCreateField() {
    editingFieldId.value = null
    Object.assign(fieldForm, { name: "", field_type: "text", visible: true, sort_order: 0 })
    fieldFormError.value = ""
    showFieldModal.value = true
}

function openEditField(f) {
    editingFieldId.value = f.id
    Object.assign(fieldForm, { name: f.name, field_type: f.field_type, visible: f.visible, sort_order: f.sort_order })
    fieldFormError.value = ""
    showFieldModal.value = true
}

async function saveField() {
    fieldFormError.value = ""
    if (!fieldForm.name.trim()) {
        fieldFormError.value = "Nazwa jest wymagana."
        return
    }
    try {
        let res
        const payload = {
            name: fieldForm.name,
            field_type: fieldForm.field_type,
            visible: fieldForm.visible,
            sort_order: Number(fieldForm.sort_order),
        }
        if (editingFieldId.value) {
            res = await authFetch("/api/v1/custom-fields/" + editingFieldId.value, {
                method: "PUT",
                body: JSON.stringify(payload),
            })
        } else {
            res = await authFetch("/api/v1/custom-fields/", {
                method: "POST",
                body: JSON.stringify(payload),
            })
        }
        if (!res.ok) {
            const d = await res.json()
            throw new Error(d.error || "Błąd zapisu")
        }
        showFieldModal.value = false
        await loadFields()
    } catch (e) {
        fieldFormError.value = e.message
    }
}

async function deleteField(id) {
    if (!await confirm({ title: "Usuń pole niestandardowe", message: "Pole i wszystkie zapisane wartości dla tego pola zostaną trwale usunięte.", confirmLabel: "Usuń", variant: "danger" })) return
    try {
        await authFetch("/api/v1/custom-fields/" + id, { method: "DELETE" })
        await loadFields()
    } catch (e) {
        fieldsError.value = e.message
    }
}

function visibilityLabel(v) {
    return v ? "Widoczne w tabeli (sortowalne)" : "Tylko w szczegółach"
}
</script>

<template>
    <section class="settings-wrapper">
        <div class="settings-header">
            <h1>Ustawienia CRM</h1>
            <p class="subtitle">Zarządzaj tagami i polami niestandardowymi</p>
        </div>

        <div class="tabs">
            <button :class="['tab-btn', { active: activeTab === 'tags' }]" @click="activeTab = 'tags'">
                <span class="material-icons">label</span> Tagi
            </button>
            <button :class="['tab-btn', { active: activeTab === 'fields' }]" @click="activeTab = 'fields'">
                <span class="material-icons">view_column</span> Pola niestandardowe
            </button>
        </div>

        <div v-if="activeTab === 'tags'" class="tab-content">
            <div class="section-header">
                <h2>Tagi kontaktów</h2>
                <button class="btn-primary" @click="openCreateTag">+ Nowy tag</button>
            </div>
            <p class="section-desc">Tagi pozwalają na kategoryzację kontaktów. Możesz filtrować listę kontaktów według tagów.</p>
            <p v-if="tagsError" class="err-msg">{{ tagsError }}</p>
            <div v-if="tagsLoading" class="loading">Ładowanie...</div>
            <div v-else-if="tags.length === 0" class="empty">Brak tagów. Dodaj pierwszy tag.</div>
            <div v-else class="tags-grid">
                <div v-for="t in tags" :key="t.id" class="tag-card">
                    <div class="tag-preview">
                        <span class="tag-chip" :style="{ background: t.color + '33', color: t.color, borderColor: t.color }">
                            {{ t.name }}
                        </span>
                    </div>
                    <div class="tag-actions">
                        <button class="btn-action btn-edit" @click="openEditTag(t)" title="Edytuj">
                            <span class="material-icons">edit</span>
                        </button>
                        <button class="btn-action btn-delete" @click="deleteTag(t.id)" title="Usuń">
                            <span class="material-icons">delete</span>
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <div v-if="activeTab === 'fields'" class="tab-content">
            <div class="section-header">
                <h2>Pola niestandardowe</h2>
                <button class="btn-primary" @click="openCreateField">+ Nowe pole</button>
            </div>
            <p class="section-desc">
                Pola z opcją <strong>„Widoczne w tabeli"</strong> pojawią się jako kolumny w liście kontaktów i będą sortowalne.
                Pozostałe pola są dostępne w widoku szczegółów kontaktu.
            </p>
            <p v-if="fieldsError" class="err-msg">{{ fieldsError }}</p>
            <div v-if="fieldsLoading" class="loading">Ładowanie...</div>
            <div v-else-if="fields.length === 0" class="empty">Brak pól niestandardowych.</div>
            <div v-else class="table-wrap">
                <table class="crud-table">
                    <thead>
                        <tr>
                            <th>Nazwa</th>
                            <th>Typ</th>
                            <th>Kolejność</th>
                            <th>Widoczność</th>
                            <th>Akcje</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="f in fields" :key="f.id">
                            <td>{{ f.name }}</td>
                            <td><span class="type-badge">{{ f.field_type }}</span></td>
                            <td>{{ f.sort_order }}</td>
                            <td>
                                <span :class="['vis-badge', f.visible ? 'vis-yes' : 'vis-no']">
                                    <span class="material-icons vis-icon">{{ f.visible ? 'visibility' : 'visibility_off' }}</span>
                                    {{ visibilityLabel(f.visible) }}
                                </span>
                            </td>
                            <td class="actions-cell">
                                <button class="btn-action btn-edit" @click="openEditField(f)" title="Edytuj">
                                    <span class="material-icons">edit</span>
                                </button>
                                <button class="btn-action btn-delete" @click="deleteField(f.id)" title="Usuń">
                                    <span class="material-icons">delete</span>
                                </button>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <div v-if="showTagModal" class="modal-overlay" @click.self="showTagModal = false">
            <div class="modal">
                <h2>{{ editingTagId ? "Edytuj tag" : "Nowy tag" }}</h2>
                <div class="form-group">
                    <label>Nazwa *</label>
                    <input v-model="tagForm.name" placeholder="np. VIP, Premium..." />
                </div>
                <div class="form-group">
                    <label>Kolor</label>
                    <div class="color-row">
                        <input type="color" v-model="tagForm.color" class="color-picker" />
                        <span class="tag-chip" :style="{ background: tagForm.color + '33', color: tagForm.color, borderColor: tagForm.color }">
                            {{ tagForm.name || "Podgląd" }}
                        </span>
                    </div>
                </div>
                <p v-if="tagFormError" class="err-msg">{{ tagFormError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showTagModal = false">Anuluj</button>
                    <button class="btn-primary" @click="saveTag">{{ editingTagId ? "Zapisz" : "Dodaj" }}</button>
                </div>
            </div>
        </div>

        <div v-if="showFieldModal" class="modal-overlay" @click.self="showFieldModal = false">
            <div class="modal">
                <h2>{{ editingFieldId ? "Edytuj pole" : "Nowe pole" }}</h2>
                <div class="form-group">
                    <label>Nazwa *</label>
                    <input v-model="fieldForm.name" placeholder="np. NIP, Branża..." />
                </div>
                <div class="form-group">
                    <label>Typ pola</label>
                    <select v-model="fieldForm.field_type">
                        <option v-for="ft in fieldTypes" :key="ft" :value="ft">{{ ft }}</option>
                    </select>
                </div>
                <div class="form-group">
                    <label>Kolejność sortowania</label>
                    <input v-model="fieldForm.sort_order" type="number" min="0" />
                </div>
                <div class="form-group">
                    <label>Widoczność w tabeli kontaktów</label>
                    <div class="toggle-row">
                        <label class="toggle-label">
                            <input type="checkbox" v-model="fieldForm.visible" class="toggle-input" />
                            <span class="toggle-track">
                                <span class="toggle-thumb"></span>
                            </span>
                            <span class="toggle-text">{{ fieldForm.visible ? "Widoczne w tabeli (sortowalne)" : "Tylko w szczegółach" }}</span>
                        </label>
                    </div>
                </div>
                <p v-if="fieldFormError" class="err-msg">{{ fieldFormError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showFieldModal = false">Anuluj</button>
                    <button class="btn-primary" @click="saveField">{{ editingFieldId ? "Zapisz" : "Dodaj" }}</button>
                </div>
            </div>
        </div>
    </section>
</template>

<style scoped>
.settings-wrapper { width: 100%; max-width: 1000px; margin: 0 auto; }
.settings-header { margin-bottom: 24px; }
.settings-header h1 { font-size: 2rem; color: #f8fafc; margin-bottom: 4px; }
.subtitle { color: #94a3b8; font-size: 0.95rem; }

.tabs { display: flex; gap: 8px; margin-bottom: 24px; border-bottom: 1px solid #334155; padding-bottom: 0; }
.tab-btn {
    padding: 10px 20px; background: transparent; border: none; border-bottom: 2px solid transparent;
    color: #94a3b8; cursor: pointer; font-size: 0.95rem; transition: 0.2s;
    display: flex; align-items: center; gap: 6px; margin-bottom: -1px;
}
.tab-btn:hover { color: #f1f5f9; }
.tab-btn.active { color: #38bdf8; border-bottom-color: #38bdf8; }
.tab-btn .material-icons { font-size: 1.1rem; }

.tab-content { }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.section-header h2 { color: #f8fafc; font-size: 1.3rem; }
.section-desc { color: #94a3b8; font-size: 0.88rem; margin-bottom: 20px; line-height: 1.5; }

.tags-grid { display: flex; flex-wrap: wrap; gap: 12px; }
.tag-card {
    background: #1e293b; border: 1px solid #334155; border-radius: 10px;
    padding: 14px 16px; display: flex; align-items: center; gap: 16px;
}
.tag-chip {
    display: inline-block; padding: 4px 12px; border-radius: 20px;
    font-size: 0.85rem; font-weight: 600; border: 1px solid; white-space: nowrap;
}
.tag-actions { display: flex; gap: 6px; }

.table-wrap { overflow-x: auto; border-radius: 12px; border: 1px solid #334155; }
.crud-table { width: 100%; border-collapse: collapse; }
.crud-table th {
    background: #1e293b; padding: 12px 16px; text-align: left;
    color: #94a3b8; font-size: 0.85rem; font-weight: 600; text-transform: uppercase;
    border-bottom: 1px solid #334155;
}
.crud-table td { padding: 14px 16px; border-bottom: 1px solid #1e293b; color: #e2e8f0; font-size: 0.9rem; }
.crud-table tr:last-child td { border-bottom: none; }
.crud-table tr:hover td { background: #1e293b44; }
.actions-cell { display: flex; gap: 8px; }

.type-badge {
    background: #1e3a5f; color: #38bdf8; padding: 2px 10px;
    border-radius: 6px; font-size: 0.8rem; font-weight: 600;
}
.vis-badge {
    display: inline-flex; align-items: center; gap: 4px;
    padding: 3px 10px; border-radius: 20px; font-size: 0.78rem; font-weight: 600;
}
.vis-yes { background: #065f4622; color: #34d399; border: 1px solid #065f46; }
.vis-no { background: #37415122; color: #9ca3af; border: 1px solid #374151; }
.vis-icon { font-size: 0.9rem; }

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
    background: none; border: none; cursor: pointer;
    padding: 4px 8px; border-radius: 6px; transition: 0.15s;
    display: flex; align-items: center; color: #94a3b8;
}
.btn-action .material-icons { font-size: 1.1rem; }
.btn-edit:hover { background: #1e3a5f; color: #38bdf8; }
.btn-delete:hover { background: #3f1414; color: #f87171; }

.loading, .empty { color: #94a3b8; padding: 40px; text-align: center; }
.err-msg { color: #f87171; font-size: 0.9rem; margin-bottom: 12px; }

.modal-overlay {
    position: fixed; inset: 0; background: rgba(0,0,0,0.6);
    display: flex; align-items: center; justify-content: center; z-index: 100;
}
.modal {
    background: #1e293b; border: 1px solid #334155; border-radius: 16px;
    padding: 32px; width: 100%; max-width: 440px; max-height: 90vh; overflow-y: auto;
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
.modal-actions { display: flex; gap: 12px; justify-content: flex-end; margin-top: 20px; }

.toggle-row { margin-top: 4px; }
.toggle-label { display: flex; align-items: center; gap: 10px; cursor: pointer; }
.toggle-input { display: none; }
.toggle-track {
    width: 44px; height: 24px; background: #334155; border-radius: 12px;
    position: relative; transition: 0.2s; flex-shrink: 0;
}
.toggle-input:checked + .toggle-track { background: #38bdf8; }
.toggle-thumb {
    position: absolute; top: 3px; left: 3px; width: 18px; height: 18px;
    background: white; border-radius: 50%; transition: 0.2s;
}
.toggle-input:checked + .toggle-track .toggle-thumb { left: 23px; }
.toggle-text { color: #e2e8f0; font-size: 0.9rem; }
</style>
