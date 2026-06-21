<script setup>
import { ref, reactive, onMounted, inject, computed } from "vue"
import FunnelBuilder from "./FunnelBuilder.vue"

const authFetch = inject("authFetch")

const funnels = ref([])
const loading = ref(false)
const error = ref("")

const selectedFunnel = ref(null)
const funnelLoading = ref(false)

const showFunnelModal = ref(false)
const editingFunnelId = ref(null)
const funnelForm = reactive({ name: "", subdomain: "", custom_domain: "" })
const funnelFormError = ref("")

const currentPage = ref(1)
const totalCount = ref(0)
const itemsPerPage = ref(9)

const totalPages = computed(() => Math.ceil(totalCount.value / itemsPerPage.value))

const showPageModal = ref(false)
const editingPageId = ref(null)
const pageForm = reactive({ name: "", path: "" })
const pageFormError = ref("")

const activeTab = ref("list")
const activeEditorPage = ref(null)

const baseDomain = import.meta.env.VITE_BASE_DOMAIN;

async function loadFunnels() {
    loading.value = true
    error.value = ""
    try {
    const res = await authFetch(`/api/v1/funnels/?page=${currentPage.value}&limit=${itemsPerPage.value}`)
    if (!res.ok) throw new Error("Błąd pobierania lejków")
    const responseData = await res.json()
    funnels.value = responseData.data
    totalCount.value = responseData.total_count
    } catch (e) {
        error.value = e.message
    } finally {
        loading.value = false
    }
}

async function changePage(page) {
    if (page < 1 || page > totalPages.value) return
    currentPage.value = page
    await loadFunnels()
}

async function loadFunnel(id) {
    funnelLoading.value = true
    try {
        const res = await authFetch("/api/v1/funnels/" + id)
        if (!res.ok) throw new Error("Błąd pobierania lejka")
        selectedFunnel.value = await res.json()
    } catch (e) {
        error.value = e.message
    } finally {
        funnelLoading.value = false
    }
}

onMounted(loadFunnels)

function openCreateFunnel() {
    editingFunnelId.value = null
    Object.assign(funnelForm, { name: "", subdomain: "", custom_domain: "" })
    funnelFormError.value = ""
    showFunnelModal.value = true
}

function openEditFunnel(f) {
    editingFunnelId.value = f.id
    Object.assign(funnelForm, { name: f.name, subdomain: f.subdomain || "", custom_domain: f.custom_domain || "" })
    funnelFormError.value = ""
    showFunnelModal.value = true
}

async function saveFunnel() {
    funnelFormError.value = ""
    if (!funnelForm.name.trim()) {
        funnelFormError.value = "Nazwa jest wymagana."
        return
    }
    try {
        const payload = { name: funnelForm.name, subdomain: funnelForm.subdomain.trim() ? funnelForm.subdomain.trim() + baseDomain : "", custom_domain: funnelForm.custom_domain }
        let res
        if (editingFunnelId.value) {
            res = await authFetch("/api/v1/funnels/" + editingFunnelId.value, { method: "PUT", body: JSON.stringify(payload) })
        } else {
            res = await authFetch("/api/v1/funnels/", { method: "POST", body: JSON.stringify(payload) })
        }
        if (!res.ok) {
            const d = await res.json()
            throw new Error(d.error || "Błąd zapisu")
        }
        showFunnelModal.value = false
        await loadFunnels()
        if (selectedFunnel.value && editingFunnelId.value === selectedFunnel.value.id) {
            await loadFunnel(selectedFunnel.value.id)
        }
    } catch (e) {
        funnelFormError.value = e.message
    }
}

async function deleteFunnel(id) {
    if (!confirm("Usunąć lejek? Wszystkie strony zostaną usunięte.")) return
    try {
        await authFetch("/api/v1/funnels/" + id, { method: "DELETE" })
        if (selectedFunnel.value && selectedFunnel.value.id === id) {
            selectedFunnel.value = null
            activeTab.value = "list"
        }
        await loadFunnels()
    } catch (e) {
        error.value = e.message
    }
}

async function selectFunnel(f) {
    activeTab.value = "detail"
    await loadFunnel(f.id)
}

function funnelUrl(f) {
    if (f.custom_domain) return "https://" + f.custom_domain
    if (f.subdomain) return "https://" + f.subdomain
    return null
}

const pagesCount = computed(() => selectedFunnel.value?.pages?.length || 0)

function openCreatePage() {
    editingPageId.value = null
    Object.assign(pageForm, { name: "", path: "/" })
    pageFormError.value = ""
    showPageModal.value = true
}

function openEditPage(p) {
    editingPageId.value = p.id
    Object.assign(pageForm, { name: p.name, path: p.path })
    pageFormError.value = ""
    showPageModal.value = true
}

async function savePage() {
    pageFormError.value = ""
    if (!pageForm.name.trim()) {
        pageFormError.value = "Nazwa strony jest wymagana."
        return
    }
    if (!pageForm.path.trim()) {
        pageFormError.value = "Ścieżka jest wymagana."
        return
    }
    try {
        const payload = { name: pageForm.name, path: pageForm.path }
        let res
        if (editingPageId.value) {
            res = await authFetch("/api/v1/funnels/pages/" + editingPageId.value, { method: "PUT", body: JSON.stringify(payload) })
        } else {
            res = await authFetch("/api/v1/funnels/" + selectedFunnel.value.id + "/pages", { method: "POST", body: JSON.stringify(payload) })
        }
        if (!res.ok) {
            const d = await res.json()
            throw new Error(d.error || "Błąd zapisu")
        }
        showPageModal.value = false
        await loadFunnel(selectedFunnel.value.id)
    } catch (e) {
        pageFormError.value = e.message
    }
}

async function deletePage(pageId) {
    if (!confirm("Usunąć stronę?")) return
    try {
        await authFetch("/api/v1/funnels/pages/" + pageId, { method: "DELETE" })
        await loadFunnel(selectedFunnel.value.id)
    } catch (e) {
        error.value = e.message
    }
}

function openBuilder(page) {
    activeEditorPage.value = page
    activeTab.value = 'editor'
}

async function onBuilderSaved() {
    if (selectedFunnel.value) {
        await loadFunnel(selectedFunnel.value.id)
    }
}
</script>

<template>
    <section class="funnels-wrapper">

        <div v-if="activeTab !== 'editor'">
            <div class="funnels-header">
                <div>
                    <h1>Lejki Sprzedaży</h1>
                    <p class="subtitle">Zarządzaj lejkami sprzedażowymi i landing pages</p>
                </div>
                <button class="btn-primary" @click="openCreateFunnel">
                    <span class="material-icons">add</span> Nowy lejek
                </button>
            </div>

            <div class="tabs">
                <button :class="['tab-btn', { active: activeTab === 'list' }]" @click="activeTab = 'list'">
                    <span class="material-icons">list</span> Lista lejków
                </button>
                <button v-if="selectedFunnel" :class="['tab-btn', { active: activeTab === 'detail' }]" @click="activeTab = 'detail'">
                    <span class="material-icons">rocket_launch</span> {{ selectedFunnel.name }}
                </button>
            </div>
        </div>

        <div v-if="activeTab === 'list'">
            <p v-if="error" class="err-msg">{{ error }}</p>
            <div v-if="loading" class="loading">Ładowanie...</div>
            <div v-else-if="funnels.length === 0" class="empty">
                <span class="material-icons empty-icon">rocket_launch</span>
                <p>Brak lejków. Utwórz pierwszy lejek sprzedażowy.</p>
            </div>
            <div v-else class="funnels-grid">
                <div v-for="f in funnels" :key="f.id" class="funnel-card">
                    <div class="funnel-card-header">
                        <span class="material-icons funnel-icon">rocket_launch</span>
                        <div class="funnel-info">
                            <h3>{{ f.name }}</h3>
                            <a v-if="funnelUrl(f)" :href="funnelUrl(f)" target="_blank" class="funnel-url">
                                {{ funnelUrl(f) }}
                            </a>
                            <span v-else class="funnel-url-none">Brak domeny</span>
                        </div>
                    </div>
                    <div class="funnel-meta">
                        <span class="meta-badge">
                            <span class="material-icons">web</span>
                            {{ f.pages?.length || 0 }} stron
                        </span>
                        <span v-if="f.subdomain" class="meta-badge subdomain-badge">
                            <span class="material-icons">link</span>
                            {{ f.subdomain }}
                        </span>
                    </div>
                    <div class="funnel-actions">
                        <button class="btn-detail" @click="selectFunnel(f)">
                            <span class="material-icons">open_in_new</span> Zarządzaj
                        </button>
                        <button class="btn-action btn-edit" @click.stop="openEditFunnel(f)" title="Edytuj">
                            <span class="material-icons">edit</span>
                        </button>
                        <button class="btn-action btn-delete" @click.stop="deleteFunnel(f.id)" title="Usuń">
                            <span class="material-icons">delete</span>
                        </button>
                    </div>
                </div>
            </div>
            <div v-if="totalPages > 1" class="pagination-container">
                <button
                    class="btn-secondary btn-pagination"
                    :disabled="currentPage === 1"
                    @click="changePage(currentPage - 1)"
                >
                    <span class="material-icons">chevron_left</span> Poprzednia
                </button>

                <div class="pagination-info">
                    Strona <span>{{ currentPage }}</span> z {{ totalPages }}
                </div>

                <button
                    class="btn-secondary btn-pagination"
                    :disabled="currentPage === totalPages"
                    @click="changePage(currentPage + 1)"
                >
                    Następna <span class="material-icons">chevron_right</span>
                </button>
            </div>
        </div>

        <div v-if="activeTab === 'detail' && selectedFunnel">
            <div v-if="funnelLoading" class="loading">Ładowanie...</div>
            <div v-else>
                <div class="detail-header">
                    <div class="detail-title">
                        <h2>{{ selectedFunnel.name }}</h2>
                        <div class="detail-meta">
                            <span v-if="selectedFunnel.subdomain" class="meta-badge">
                                Subdomena: {{ selectedFunnel.subdomain }}
                            </span>
                            <span v-if="selectedFunnel.custom_domain" class="meta-badge">
                                Domena: {{ selectedFunnel.custom_domain }}
                            </span>
                            <span class="meta-badge">{{ pagesCount }} stron</span>
                        </div>
                    </div>
                    <div class="detail-actions">
                        <button class="btn-secondary" @click="openEditFunnel(selectedFunnel)">
                            <span class="material-icons">edit</span> Edytuj lejek
                        </button>
                        <button class="btn-primary" @click="openCreatePage">
                            <span class="material-icons">add</span> Nowa strona
                        </button>
                    </div>
                </div>

                <div v-if="!selectedFunnel.pages || selectedFunnel.pages.length === 0" class="empty">
                    <span class="material-icons empty-icon">web</span>
                    <p>Ten lejek nie ma jeszcze żadnych stron. Dodaj pierwszą stronę.</p>
                </div>
                <div v-else class="pages-list">
                    <div v-for="page in selectedFunnel.pages" :key="page.id" class="page-card">
                        <div class="page-card-left">
                            <span class="material-icons page-icon">article</span>
                            <div class="page-info">
                                <h4>{{ page.name }}</h4>
                                <span class="page-path">{{ page.path }}</span>
                            </div>
                        </div>
                        <div class="page-card-right">
                            <span v-if="page.structure" class="struct-badge">
                                <span class="material-icons">check_circle</span> Zaprojektowana
                            </span>

                            <button class="btn-primary btn-builder" @click="openBuilder(page)">
                                <span class="material-icons">brush</span> Kreator
                            </button>

                            <button class="btn-action btn-edit" @click="openEditPage(page)" title="Ustawienia">
                                <span class="material-icons">settings</span>
                            </button>
                            <button class="btn-action btn-delete" @click="deletePage(page.id)" title="Usuń">
                                <span class="material-icons">delete</span>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div v-if="activeTab === 'editor' && activeEditorPage" class="editor-tab">
            <FunnelBuilder
                :page="activeEditorPage"
                @close="activeTab = 'detail'"
                @saved="onBuilderSaved"
            />
        </div>

        <div v-if="showFunnelModal" class="modal-overlay" @click.self="showFunnelModal = false">
            <div class="modal">
                <h2>{{ editingFunnelId ? "Edytuj lejek" : "Nowy lejek" }}</h2>
                <div class="form-group">
                    <label>Nazwa *</label>
                    <input v-model="funnelForm.name" placeholder="np. Lejek Premium" />
                </div>
                <div class="form-group">
                    <label>Subdomena</label>
                    <div class="subdomain-row">
                        <input v-model="funnelForm.subdomain" placeholder="np. premium.bizmopol.localhost" />
                    </div>
                </div>
                <div class="form-group">
                    <label>Własna domena</label>
                    <input v-model="funnelForm.custom_domain" placeholder="np. oferta.moja-firma.pl" />
                </div>
                <p v-if="funnelFormError" class="err-msg">{{ funnelFormError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showFunnelModal = false">Anuluj</button>
                    <button class="btn-primary" @click="saveFunnel">{{ editingFunnelId ? "Zapisz" : "Utwórz" }}</button>
                </div>
            </div>
        </div>

        <div v-if="showPageModal" class="modal-overlay" @click.self="showPageModal = false">
            <div class="modal">
                <h2>{{ editingPageId ? "Ustawienia strony" : "Nowa strona" }}</h2>
                <div class="form-group">
                    <label>Nazwa strony *</label>
                    <input v-model="pageForm.name" placeholder="np. Strona główna" />
                </div>
                <div class="form-group">
                    <label>Ścieżka URL *</label>
                    <input v-model="pageForm.path" placeholder="np. /oferta" />
                </div>
                <p v-if="pageFormError" class="err-msg">{{ pageFormError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showPageModal = false">Anuluj</button>
                    <button class="btn-primary" @click="savePage">{{ editingPageId ? "Zapisz" : "Dodaj" }}</button>
                </div>
            </div>
        </div>
    </section>
</template>

<style scoped>
/* Paste ALL of your original styles here */
.funnels-wrapper { width: 100%; max-width: 1200px; margin: 0 auto; }
.funnels-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; }
.funnels-header h1 { font-size: 2rem; color: #f8fafc; margin-bottom: 4px; }
.subtitle { color: #94a3b8; font-size: 0.95rem; }
.tabs { display: flex; gap: 8px; margin-bottom: 24px; border-bottom: 1px solid #334155; padding-bottom: 0; }
.tab-btn { padding: 10px 20px; background: transparent; border: none; border-bottom: 2px solid transparent; color: #94a3b8; cursor: pointer; font-size: 0.95rem; transition: 0.2s; display: flex; align-items: center; gap: 6px; margin-bottom: -1px; }
.tab-btn:hover { color: #f1f5f9; }
.tab-btn.active { color: #38bdf8; border-bottom-color: #38bdf8; }
.tab-btn .material-icons { font-size: 1.1rem; }
.funnels-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap: 20px; }
.funnel-card { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 20px; display: flex; flex-direction: column; gap: 14px; transition: border-color 0.2s; }
.funnel-card:hover { border-color: #38bdf8; }
.funnel-card-header { display: flex; align-items: flex-start; gap: 14px; }
.funnel-icon { font-size: 2rem; color: #38bdf8; margin-top: 2px; }
.funnel-info h3 { color: #f1f5f9; font-size: 1.1rem; margin-bottom: 4px; }
.funnel-url { color: #38bdf8; font-size: 0.82rem; text-decoration: none; word-break: break-all; }
.funnel-url:hover { text-decoration: underline; }
.funnel-url-none { color: #475569; font-size: 0.82rem; }
.funnel-meta { display: flex; gap: 8px; flex-wrap: wrap; }
.meta-badge { display: inline-flex; align-items: center; gap: 4px; background: #0f172a; color: #94a3b8; border: 1px solid #334155; border-radius: 20px; padding: 3px 10px; font-size: 0.78rem; }
.meta-badge .material-icons { font-size: 0.85rem; }
.subdomain-badge { color: #a78bfa; border-color: #4c1d95; background: #2e1065; }
.funnel-actions { display: flex; gap: 8px; align-items: center; margin-top: 4px; }
.btn-detail { flex: 1; padding: 8px 16px; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #38bdf8; font-size: 0.9rem; cursor: pointer; transition: 0.2s; display: flex; align-items: center; justify-content: center; gap: 6px; }
.btn-detail:hover { background: #1e3a5f; border-color: #38bdf8; }
.btn-detail .material-icons { font-size: 1rem; }
.detail-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; gap: 20px; flex-wrap: wrap; }
.detail-title h2 { font-size: 1.6rem; color: #f8fafc; margin-bottom: 8px; }
.detail-meta { display: flex; gap: 8px; flex-wrap: wrap; }
.detail-actions { display: flex; gap: 10px; }
.pages-list { display: flex; flex-direction: column; gap: 12px; }
.page-card { background: #1e293b; border: 1px solid #334155; border-radius: 10px; padding: 16px 20px; display: flex; justify-content: space-between; align-items: center; gap: 20px; transition: border-color 0.2s; }
.page-card:hover { border-color: #475569; }
.page-card-left { display: flex; align-items: center; gap: 14px; }
.page-icon { font-size: 1.5rem; color: #64748b; }
.page-info h4 { color: #f1f5f9; font-size: 1rem; margin-bottom: 4px; }
.page-path { color: #64748b; font-size: 0.85rem; font-family: monospace; }
.page-card-right { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.struct-badge { display: inline-flex; align-items: center; gap: 4px; background: #064e3b; color: #34d399; border: 1px solid #065f46; border-radius: 6px; padding: 2px 10px; font-size: 0.78rem; }
.struct-badge .material-icons { font-size: 0.85rem; }

.btn-builder { padding: 6px 12px; font-size: 0.85rem; background: #38bdf8; color: #0f172a; }
.btn-builder:hover { background: #7dd3fc; }

.btn-primary { padding: 10px 20px; background: #38bdf8; border: none; border-radius: 8px; color: #0f172a; font-weight: bold; cursor: pointer; transition: 0.2s; display: flex; align-items: center; gap: 6px; }
.btn-primary:hover { background: #7dd3fc; }
.btn-primary .material-icons { font-size: 1.1rem; }
.btn-secondary { padding: 10px 20px; background: transparent; border: 1px solid #475569; border-radius: 8px; color: #94a3b8; cursor: pointer; transition: 0.2s; display: flex; align-items: center; gap: 6px; }
.btn-secondary:hover { border-color: #f1f5f9; color: #f1f5f9; }
.btn-secondary .material-icons { font-size: 1.1rem; }
.btn-action { background: none; border: none; cursor: pointer; padding: 6px 8px; border-radius: 6px; transition: 0.15s; display: flex; align-items: center; color: #94a3b8; }
.btn-action .material-icons { font-size: 1.1rem; }
.btn-edit:hover { background: #1e3a5f; color: #38bdf8; }
.btn-delete:hover { background: #3f1414; color: #f87171; }
.loading, .empty { color: #94a3b8; padding: 60px; text-align: center; }
.empty-icon { font-size: 3rem; display: block; margin-bottom: 12px; color: #334155; }
.err-msg { color: #f87171; font-size: 0.9rem; margin-bottom: 12px; }
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.6); display: flex; align-items: center; justify-content: center; z-index: 100; }
.modal { background: #1e293b; border: 1px solid #334155; border-radius: 16px; padding: 32px; width: 100%; max-width: 460px; max-height: 90vh; overflow-y: auto; }
.modal h2 { color: #38bdf8; margin-bottom: 20px; font-size: 1.4rem; }
.form-group { margin-bottom: 16px; }
.form-group label { display: block; margin-bottom: 6px; color: #94a3b8; font-size: 0.88rem; }
.form-group input, .form-group select { width: 100%; padding: 10px 12px; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #f1f5f9; outline: none; font-size: 0.95rem; }
.form-group input:focus, .form-group select:focus { border-color: #38bdf8; }
.subdomain-row { display: flex; align-items: center; gap: 0; }
.subdomain-input { flex: 1; border-radius: 8px 0 0 8px !important; border-right: none !important; }
.subdomain-suffix { background: #0f172a; border: 1px solid #334155; border-radius: 0 8px 8px 0; padding: 10px 12px; color: #64748b; font-size: 0.9rem; white-space: nowrap; }
.modal-actions { display: flex; gap: 12px; justify-content: flex-end; margin-top: 20px; }
.pagination-container {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 16px;
    margin-top: 32px;
    padding-top: 16px;
    border-top: 1px solid #334155;
}
.pagination-info {
    color: #94a3b8;
    font-size: 0.95rem;
}
.pagination-info span {
    color: #38bdf8;
    font-weight: bold;
}
.btn-pagination {
    padding: 6px 14px !important;
    font-size: 0.88rem;
    display: inline-flex;
    align-items: center;
    gap: 4px;
}
.btn-pagination:disabled {
    opacity: 0.4;
    cursor: not-allowed;
    border-color: #334155;
    color: #475569;
}
/* Editor specific styles */
.editor-tab {
    margin-top: -24px; /* Pull it up since we hide the header */
}
</style>
