<script setup>
import { ref, onMounted, inject, computed } from "vue"
import { useWebSocket } from "../composables/useWebSocket.js"

const authFetch = inject("authFetch")
const toast = inject("toast")
const confirm = inject("confirm")
const { onMessage } = useWebSocket()
const getToken = () => localStorage.getItem("jwt_token")
const authedUrl = (path) => `${path}${path.includes('?') ? '&' : '?'}token=${encodeURIComponent(getToken())}`

const documents = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = 20
const search = ref("")
const filterStatus = ref("")
const loading = ref(false)

const selectedDoc = ref(null)
const showUpload = ref(false)
const showSign = ref(false)
const showAuditLog = ref(false)
const showViewer = ref(false)
const auditLogs = ref([])
const viewerMode = ref("signed")

const uploadTitle = ref("")
const uploadContactId = ref("")
const uploadFile = ref(null)
const uploading = ref(false)

const signerName = ref("")
const signerEmail = ref("")
const signing = ref(false)
const signatureCanvas = ref(null)
let signCtx = null
let isDrawing = false

const contacts = ref([])
const contactSearch = ref("")
const filteredContacts = computed(() => {
    if (!contactSearch.value) return contacts.value.slice(0, 20)
    const q = contactSearch.value.toLowerCase()
    return contacts.value.filter(c => c.name.toLowerCase().includes(q) || (c.email && c.email.toLowerCase().includes(q))).slice(0, 20)
})

const stats = computed(() => {
    const all = documents.value.length
    const pending = documents.value.filter(d => d.status === "pending").length
    const signed = documents.value.filter(d => d.status === "signed").length
    return { total: total.value, pending, signed }
})

async function loadDocuments() {
    loading.value = true
    try {
        const params = new URLSearchParams()
        if (search.value) params.set("search", search.value)
        if (filterStatus.value) params.set("status", filterStatus.value)
        params.set("page", page.value)
        params.set("page_size", pageSize)
        const res = await authFetch(`/api/v1/documents/?${params}`)
        if (res.ok) {
            const data = await res.json()
            documents.value = data.documents || []
            total.value = data.total || 0
        }
    } catch (_) {} finally {
        loading.value = false
    }
}

async function loadContacts() {
    try {
        const res = await authFetch("/api/v1/contacts/?page_size=200")
        if (res.ok) {
            const data = await res.json()
            contacts.value = data.data || []
        }
    } catch (_) {}
}

function openUpload() {
    uploadTitle.value = ""
    uploadContactId.value = ""
    uploadFile.value = null
    showUpload.value = true
}

function onFileChange(e) {
    const file = e.target.files[0]
    if (file && file.type !== "application/pdf") {
        toast.show("Dozwolone są tylko pliki PDF", "error")
        e.target.value = ""
        return
    }
    uploadFile.value = file
    if (file && !uploadTitle.value) {
        uploadTitle.value = file.name.replace(/\.pdf$/i, "")
    }
}

async function submitUpload() {
    if (!uploadFile.value) {
        toast.show("Wybierz plik PDF", "error")
        return
    }
    uploading.value = true
    try {
        const fd = new FormData()
        fd.append("file", uploadFile.value)
        fd.append("title", uploadTitle.value || uploadFile.value.name)
        if (uploadContactId.value) fd.append("contact_id", uploadContactId.value)
        const res = await authFetch("/api/v1/documents/", { method: "POST", body: fd })
        if (res.ok) {
            toast.show("Dokument przesłany", "success")
            showUpload.value = false
            loadDocuments()
        } else {
            const err = await res.json()
            toast.show(err.error || "Błąd przesyłania", "error")
        }
    } catch (_) {
        toast.show("Błąd przesyłania", "error")
    } finally {
        uploading.value = false
    }
}

async function selectDocument(doc) {
    const res = await authFetch(`/api/v1/documents/${doc.id}`)
    if (res.ok) {
        selectedDoc.value = await res.json()
    }
}

function viewPdf(doc, mode = "signed") {
    viewerMode.value = mode
    showViewer.value = true
    selectedDoc.value = doc
}

async function deleteDocument(doc) {
    const ok = await confirm.ask(`Usunąć dokument "${doc.title}"?`)
    if (!ok) return
    const res = await authFetch(`/api/v1/documents/${doc.id}`, { method: "DELETE" })
    if (res.ok) {
        toast.show("Dokument usunięty", "success")
        if (selectedDoc.value?.id === doc.id) selectedDoc.value = null
        loadDocuments()
    }
}

function openSignDialog(doc) {
    selectedDoc.value = doc
    signerName.value = ""
    signerEmail.value = ""
    showSign.value = true
    setTimeout(initCanvas, 50)
}

function initCanvas() {
    const canvas = signatureCanvas.value
    if (!canvas) return
    signCtx = canvas.getContext("2d")
    canvas.width = canvas.offsetWidth
    canvas.height = canvas.offsetHeight
    signCtx.fillStyle = "#ffffff"
    signCtx.fillRect(0, 0, canvas.width, canvas.height)
    signCtx.strokeStyle = "#0f172a"
    signCtx.lineWidth = 2.5
    signCtx.lineCap = "round"
    signCtx.lineJoin = "round"

    const getPos = (e) => {
        const rect = canvas.getBoundingClientRect()
        const clientX = e.touches ? e.touches[0].clientX : e.clientX
        const clientY = e.touches ? e.touches[0].clientY : e.clientY
        return { x: clientX - rect.left, y: clientY - rect.top }
    }

    const startDraw = (e) => {
        e.preventDefault()
        isDrawing = true
        const pos = getPos(e)
        signCtx.beginPath()
        signCtx.moveTo(pos.x, pos.y)
    }
    const draw = (e) => {
        if (!isDrawing) return
        e.preventDefault()
        const pos = getPos(e)
        signCtx.lineTo(pos.x, pos.y)
        signCtx.stroke()
    }
    const stopDraw = () => { isDrawing = false }

    canvas.addEventListener("mousedown", startDraw)
    canvas.addEventListener("mousemove", draw)
    canvas.addEventListener("mouseup", stopDraw)
    canvas.addEventListener("mouseleave", stopDraw)
    canvas.addEventListener("touchstart", startDraw, { passive: false })
    canvas.addEventListener("touchmove", draw, { passive: false })
    canvas.addEventListener("touchend", stopDraw)
}

function clearCanvas() {
    if (!signCtx) return
    const canvas = signatureCanvas.value
    signCtx.fillStyle = "#ffffff"
    signCtx.fillRect(0, 0, canvas.width, canvas.height)
}

function isCanvasBlank() {
    const canvas = signatureCanvas.value
    if (!canvas) return true
    const ctx = canvas.getContext("2d")
    const data = ctx.getImageData(0, 0, canvas.width, canvas.height).data
    for (let i = 0; i < data.length; i += 4) {
        if (data[i] !== 255 || data[i + 1] !== 255 || data[i + 2] !== 255) return false
    }
    return true
}

async function submitSignature() {
    if (!signerName.value.trim()) {
        toast.show("Podaj imię i nazwisko", "error")
        return
    }
    if (isCanvasBlank()) {
        toast.show("Narysuj podpis na kanwie", "error")
        return
    }
    signing.value = true
    try {
        const canvas = signatureCanvas.value
        const dataUrl = canvas.toDataURL("image/png")
        const base64 = dataUrl.split(",")[1]
        const res = await authFetch(`/api/v1/documents/${selectedDoc.value.id}/sign`, {
            method: "POST",
            body: JSON.stringify({
                signer_name: signerName.value.trim(),
                signer_email: signerEmail.value.trim(),
                image_data: base64,
            }),
        })
        if (res.ok) {
            toast.show("Dokument podpisany", "success")
            showSign.value = false
            loadDocuments()
            if (selectedDoc.value) selectDocument(selectedDoc.value)
        } else {
            const err = await res.json()
            toast.show(err.error || "Błąd podpisywania", "error")
        }
    } catch (_) {
        toast.show("Błąd podpisywania", "error")
    } finally {
        signing.value = false
    }
}

async function openAuditLog(doc) {
    selectedDoc.value = doc
    showAuditLog.value = true
    const res = await authFetch(`/api/v1/documents/${doc.id}/audit-log`)
    if (res.ok) {
        auditLogs.value = await res.json()
    }
}

function formatDate(d) {
    if (!d) return ""
    return new Date(d).toLocaleString("pl-PL")
}

function formatSize(bytes) {
    if (bytes < 1024) return bytes + " B"
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB"
    return (bytes / (1024 * 1024)).toFixed(1) + " MB"
}

function statusLabel(s) {
    return { pending: "Oczekuje", signed: "Podpisany" }[s] || s
}

function statusClass(s) {
    return { pending: "status-pending", signed: "status-signed" }[s] || ""
}

function actionIcon(a) {
    return { uploaded: "upload_file", signed: "draw", viewed: "visibility", downloaded: "download" }[a] || "info"
}

onMounted(async () => {
    await Promise.all([loadDocuments(), loadContacts()])
    onMessage((msg) => {
        if (msg.type === "document_signed") {
            toast.show(msg.message, "info")
            loadDocuments()
            if (selectedDoc.value?.id === msg.data?.document_id) {
                selectDocument(selectedDoc.value)
            }
        }
    })
})
</script>

<template>
    <div class="doc-wrapper">
        <div class="doc-header">
            <div>
                <h1>Dokumenty & E-Signing</h1>
                <p class="subtitle">Przesyłaj PDF-y, zbieraj podpisy elektroniczne</p>
            </div>
            <div class="header-actions">
                <button class="btn-primary" @click="openUpload">
                    <span class="material-icons">upload_file</span> Prześlij PDF
                </button>
            </div>
        </div>

        <div class="stats-row">
            <div class="stat-card">
                <span class="material-icons stat-icon" style="color:#38bdf8">description</span>
                <div>
                    <div class="stat-value">{{ stats.total }}</div>
                    <div class="stat-label">Wszystkie</div>
                </div>
            </div>
            <div class="stat-card">
                <span class="material-icons stat-icon" style="color:#f59e0b">pending</span>
                <div>
                    <div class="stat-value">{{ stats.pending }}</div>
                    <div class="stat-label">Oczekujące</div>
                </div>
            </div>
            <div class="stat-card">
                <span class="material-icons stat-icon" style="color:#22c55e">verified</span>
                <div>
                    <div class="stat-value">{{ stats.signed }}</div>
                    <div class="stat-label">Podpisane</div>
                </div>
            </div>
        </div>

        <div class="toolbar">
            <div class="search-box">
                <span class="material-icons">search</span>
                <input v-model="search" @input="loadDocuments" placeholder="Szukaj dokumentów..." />
            </div>
            <select v-model="filterStatus" @change="loadDocuments" class="filter-select">
                <option value="">Wszystkie statusy</option>
                <option value="pending">Oczekujące</option>
                <option value="signed">Podpisane</option>
            </select>
        </div>

        <div class="doc-layout">
            <div class="doc-list">
                <div v-if="loading" class="empty-state">
                    <span class="material-icons spinning">sync</span> Ładowanie...
                </div>
                <div v-else-if="documents.length === 0" class="empty-state">
                    <span class="material-icons" style="font-size:3rem;color:#475569">folder_open</span>
                    <p>Brak dokumentów</p>
                    <button class="btn-primary" @click="openUpload">
                        <span class="material-icons">upload_file</span> Prześlij pierwszy PDF
                    </button>
                </div>
                <div
                    v-for="doc in documents" :key="doc.id"
                    class="doc-item" :class="{ active: selectedDoc?.id === doc.id }"
                    @click="selectDocument(doc)"
                >
                    <div class="doc-item-icon">
                        <span class="material-icons" :style="{ color: doc.status === 'signed' ? '#22c55e' : '#f59e0b' }">
                            {{ doc.status === 'signed' ? 'task' : 'description' }}
                        </span>
                    </div>
                    <div class="doc-item-body">
                        <div class="doc-item-title">{{ doc.title }}</div>
                        <div class="doc-item-meta">
                            <span>{{ doc.file_name }}</span>
                            <span>{{ formatSize(doc.file_size) }}</span>
                            <span v-if="doc.contact_name" class="doc-contact-badge">
                                <span class="material-icons" style="font-size:0.7rem">person</span>
                                {{ doc.contact_name }}
                            </span>
                        </div>
                    </div>
                    <div class="doc-item-right">
                        <span :class="['status-badge', statusClass(doc.status)]">{{ statusLabel(doc.status) }}</span>
                        <span class="doc-item-date">{{ formatDate(doc.created_at) }}</span>
                    </div>
                </div>

                <div v-if="total > pageSize" class="pagination">
                    <button :disabled="page <= 1" @click="page--; loadDocuments()">
                        <span class="material-icons">chevron_left</span>
                    </button>
                    <span>{{ page }} / {{ Math.ceil(total / pageSize) }}</span>
                    <button :disabled="page >= Math.ceil(total / pageSize)" @click="page++; loadDocuments()">
                        <span class="material-icons">chevron_right</span>
                    </button>
                </div>
            </div>

            <div class="doc-detail" v-if="selectedDoc">
                <div class="detail-header">
                    <h2>{{ selectedDoc.title }}</h2>
                    <span :class="['status-badge', statusClass(selectedDoc.status)]">{{ statusLabel(selectedDoc.status) }}</span>
                </div>

                <div class="detail-meta">
                    <div class="meta-row">
                        <span class="material-icons meta-icon">insert_drive_file</span>
                        <span>{{ selectedDoc.file_name }} ({{ formatSize(selectedDoc.file_size) }})</span>
                    </div>
                    <div class="meta-row" v-if="selectedDoc.contact_name">
                        <span class="material-icons meta-icon">person</span>
                        <span>{{ selectedDoc.contact_name }}</span>
                    </div>
                    <div class="meta-row">
                        <span class="material-icons meta-icon">calendar_today</span>
                        <span>Przesłano: {{ formatDate(selectedDoc.created_at) }}</span>
                    </div>
                    <div class="meta-row" v-if="selectedDoc.signed_at">
                        <span class="material-icons meta-icon">verified</span>
                        <span>Podpisano: {{ formatDate(selectedDoc.signed_at) }}</span>
                    </div>
                </div>

                <div class="detail-actions">
                    <button class="btn-action" @click="viewPdf(selectedDoc, 'signed')">
                        <span class="material-icons">visibility</span> Podgląd
                    </button>
                    <button v-if="selectedDoc.status === 'signed'" class="btn-action" @click="viewPdf(selectedDoc, 'original')">
                        <span class="material-icons">file_copy</span> Oryginał
                    </button>
                    <a :href="authedUrl(`/api/v1/documents/${selectedDoc.id}/download`)" class="btn-action" target="_blank">
                        <span class="material-icons">download</span> Pobierz
                    </a>
                    <a v-if="selectedDoc.status === 'signed'" :href="authedUrl(`/api/v1/documents/${selectedDoc.id}/download?mode=original`)" class="btn-action" target="_blank">
                        <span class="material-icons">download</span> Oryginał PDF
                    </a>
                    <button class="btn-action btn-sign" @click="openSignDialog(selectedDoc)">
                        <span class="material-icons">draw</span> Podpisz
                    </button>
                    <button class="btn-action" @click="openAuditLog(selectedDoc)">
                        <span class="material-icons">history</span> Historia
                    </button>
                    <button class="btn-action btn-danger" @click="deleteDocument(selectedDoc)">
                        <span class="material-icons">delete</span>
                    </button>
                </div>

                <div v-if="selectedDoc.signatures && selectedDoc.signatures.length > 0" class="signatures-section">
                    <h3><span class="material-icons" style="font-size:1.1rem;vertical-align:middle">draw</span> Podpisy ({{ selectedDoc.signatures.length }})</h3>
                    <div v-for="sig in selectedDoc.signatures" :key="sig.id" class="sig-card">
                        <img :src="authedUrl(`/api/v1/documents/${selectedDoc.id}/signatures/${sig.id}/image`)" class="sig-image" alt="Podpis" />
                        <div class="sig-info">
                            <div class="sig-name">{{ sig.signer_name }}</div>
                            <div class="sig-meta">{{ sig.signer_email }} &middot; {{ formatDate(sig.created_at) }}</div>
                            <div class="sig-meta" v-if="sig.ip_address">IP: {{ sig.ip_address }}</div>
                        </div>
                    </div>
                </div>
            </div>
            <div class="doc-detail doc-empty" v-else>
                <span class="material-icons" style="font-size:4rem;color:#334155">edit_document</span>
                <p>Wybierz dokument z listy</p>
            </div>
        </div>

        <!-- Upload Modal -->
        <div v-if="showUpload" class="modal-overlay" @click.self="showUpload = false">
            <div class="modal-box">
                <div class="modal-header">
                    <h3>Prześlij dokument PDF</h3>
                    <button class="modal-close" @click="showUpload = false">
                        <span class="material-icons">close</span>
                    </button>
                </div>
                <div class="modal-body">
                    <label class="form-label">Plik PDF *</label>
                    <div class="file-drop" @click="$refs.fileInput.click()" @dragover.prevent @drop.prevent="onDrop">
                        <span class="material-icons" style="font-size:2.5rem;color:#475569">cloud_upload</span>
                        <p v-if="!uploadFile">Kliknij lub przeciągnij plik PDF</p>
                        <p v-else class="file-selected">{{ uploadFile.name }} ({{ formatSize(uploadFile.size) }})</p>
                    </div>
                    <input ref="fileInput" type="file" accept=".pdf" @change="onFileChange" style="display:none" />

                    <label class="form-label">Tytuł dokumentu</label>
                    <input v-model="uploadTitle" class="form-input" placeholder="Np. Umowa o współpracę" />

                    <label class="form-label">Powiąż z kontaktem (opcjonalnie)</label>
                    <input v-model="contactSearch" class="form-input" placeholder="Szukaj kontaktu..." />
                    <div v-if="contactSearch || uploadContactId" class="contact-dropdown">
                        <div v-if="uploadContactId" class="contact-selected" @click="uploadContactId = ''; contactSearch = ''">
                            <span class="material-icons" style="font-size:0.9rem">person</span>
                            {{ contacts.find(c => c.id === uploadContactId)?.name }}
                            <span class="material-icons" style="font-size:0.8rem;margin-left:auto">close</span>
                        </div>
                        <template v-else>
                            <div v-for="c in filteredContacts" :key="c.id" class="contact-option" @click="uploadContactId = c.id; contactSearch = c.name">
                                {{ c.name }} <span style="color:#64748b">{{ c.email }}</span>
                            </div>
                            <div v-if="filteredContacts.length === 0" class="contact-option" style="color:#64748b">Brak wyników</div>
                        </template>
                    </div>
                </div>
                <div class="modal-footer">
                    <button class="btn-secondary" @click="showUpload = false">Anuluj</button>
                    <button class="btn-primary" @click="submitUpload" :disabled="uploading || !uploadFile">
                        <span v-if="uploading" class="material-icons spinning">sync</span>
                        {{ uploading ? 'Przesyłanie...' : 'Prześlij' }}
                    </button>
                </div>
            </div>
        </div>

        <!-- Sign Modal -->
        <div v-if="showSign" class="modal-overlay" @click.self="showSign = false">
            <div class="modal-box modal-sign">
                <div class="modal-header">
                    <h3>Podpisz dokument</h3>
                    <button class="modal-close" @click="showSign = false">
                        <span class="material-icons">close</span>
                    </button>
                </div>
                <div class="modal-body">
                    <p class="sign-doc-title">{{ selectedDoc?.title }}</p>

                    <label class="form-label">Imię i nazwisko *</label>
                    <input v-model="signerName" class="form-input" placeholder="Jan Kowalski" />

                    <label class="form-label">Email</label>
                    <input v-model="signerEmail" class="form-input" placeholder="jan@example.com" />

                    <label class="form-label">Podpis (narysuj poniżej) *</label>
                    <div class="canvas-wrapper">
                        <canvas ref="signatureCanvas" class="sign-canvas"></canvas>
                        <button class="canvas-clear" @click="clearCanvas" title="Wyczyść">
                            <span class="material-icons">refresh</span>
                        </button>
                    </div>
                </div>
                <div class="modal-footer">
                    <button class="btn-secondary" @click="showSign = false">Anuluj</button>
                    <button class="btn-primary btn-sign-submit" @click="submitSignature" :disabled="signing">
                        <span class="material-icons">draw</span>
                        {{ signing ? 'Podpisywanie...' : 'Podpisz dokument' }}
                    </button>
                </div>
            </div>
        </div>

        <!-- Audit Log Modal -->
        <div v-if="showAuditLog" class="modal-overlay" @click.self="showAuditLog = false">
            <div class="modal-box">
                <div class="modal-header">
                    <h3>Historia dokumentu</h3>
                    <button class="modal-close" @click="showAuditLog = false">
                        <span class="material-icons">close</span>
                    </button>
                </div>
                <div class="modal-body">
                    <p class="sign-doc-title">{{ selectedDoc?.title }}</p>
                    <div v-if="auditLogs.length === 0" class="empty-state" style="padding:20px">Brak wpisów</div>
                    <div v-for="log in auditLogs" :key="log.id" class="audit-entry">
                        <span class="material-icons audit-icon" :style="{ color: log.action === 'signed' ? '#22c55e' : '#38bdf8' }">
                            {{ actionIcon(log.action) }}
                        </span>
                        <div class="audit-body">
                            <div class="audit-detail">{{ log.details }}</div>
                            <div class="audit-meta">
                                {{ formatDate(log.created_at) }}
                                <span v-if="log.ip_address"> &middot; IP: {{ log.ip_address }}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- PDF Viewer Modal -->
        <div v-if="showViewer" class="modal-overlay viewer-overlay" @click.self="showViewer = false">
            <div class="viewer-box">
                <div class="viewer-header">
                    <h3>{{ selectedDoc?.title }}</h3>
                    <button class="modal-close" @click="showViewer = false">
                        <span class="material-icons">close</span>
                    </button>
                </div>
                <div class="viewer-toolbar" v-if="selectedDoc?.status === 'signed'">
                    <button :class="['vt-btn', { active: viewerMode === 'signed' }]" @click="viewerMode = 'signed'">
                        <span class="material-icons">verified</span> Podpisany
                    </button>
                    <button :class="['vt-btn', { active: viewerMode === 'original' }]" @click="viewerMode = 'original'">
                        <span class="material-icons">file_copy</span> Oryginał
                    </button>
                </div>
                <iframe :src="authedUrl(`/api/v1/documents/${selectedDoc?.id}/view?mode=${viewerMode}`)" class="pdf-iframe"></iframe>
            </div>
        </div>
    </div>
</template>

<style scoped>
.doc-wrapper { width: 100%; max-width: 1400px; margin: 0 auto; }

.doc-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px; flex-wrap: wrap; gap: 16px; }
.doc-header h1 { font-size: 2rem; margin: 0; color: #f8fafc; }
.subtitle { color: #94a3b8; font-size: 0.9rem; margin-top: 4px; }
.header-actions { display: flex; gap: 10px; align-items: center; }

.btn-primary {
    padding: 8px 16px; background: #38bdf8; border: none; border-radius: 8px;
    color: #0f172a; font-weight: bold; cursor: pointer; transition: 0.2s;
    display: flex; align-items: center; gap: 6px; font-size: 0.85rem; height: 38px;
}
.btn-primary:hover { background: #7dd3fc; }
.btn-secondary {
    padding: 8px 16px; background: transparent; border: 1px solid #475569;
    border-radius: 8px; color: #94a3b8; cursor: pointer; transition: 0.2s;
    font-size: 0.85rem; height: 38px;
}
.btn-secondary:hover { border-color: #f1f5f9; color: #f1f5f9; }

.stats-row { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 12px; margin-bottom: 20px; }
.stat-card {
    background: #1e293b; border: 1px solid #334155; border-radius: 10px;
    padding: 16px; display: flex; align-items: center; gap: 12px;
}
.stat-icon { font-size: 1.8rem; }
.stat-value { font-size: 1.4rem; font-weight: bold; color: #f1f5f9; }
.stat-label { font-size: 0.75rem; color: #94a3b8; }

.toolbar { display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.search-box {
    display: flex; align-items: center; gap: 8px;
    background: #1e293b; border: 1px solid #334155; border-radius: 8px;
    padding: 0 12px; flex: 1; min-width: 200px;
}
.search-box input {
    background: transparent; border: none; color: #f1f5f9; padding: 10px 0;
    flex: 1; outline: none; font-size: 0.9rem;
}
.search-box .material-icons { color: #64748b; font-size: 1.1rem; }
.filter-select {
    background: #1e293b; border: 1px solid #334155; border-radius: 8px;
    color: #f1f5f9; padding: 10px 12px; font-size: 0.85rem; cursor: pointer;
}

.doc-layout { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; min-height: 500px; }

.doc-list { display: flex; flex-direction: column; gap: 6px; }
.doc-item {
    display: flex; align-items: center; gap: 12px; padding: 14px 16px;
    background: #1e293b; border: 1px solid #334155; border-radius: 10px;
    cursor: pointer; transition: 0.2s;
}
.doc-item:hover { border-color: #475569; }
.doc-item.active { border-color: #38bdf8; background: #1e293b; }
.doc-item-icon .material-icons { font-size: 1.6rem; }
.doc-item-body { flex: 1; min-width: 0; }
.doc-item-title { font-weight: 600; color: #f1f5f9; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.doc-item-meta { display: flex; gap: 8px; font-size: 0.75rem; color: #64748b; margin-top: 2px; flex-wrap: wrap; }
.doc-contact-badge { display: flex; align-items: center; gap: 2px; color: #94a3b8; }
.doc-item-right { display: flex; flex-direction: column; align-items: flex-end; gap: 4px; }
.doc-item-date { font-size: 0.7rem; color: #64748b; white-space: nowrap; }

.status-badge {
    font-size: 0.7rem; padding: 2px 8px; border-radius: 20px; font-weight: 600; white-space: nowrap;
}
.status-pending { background: #422006; color: #f59e0b; }
.status-signed { background: #052e16; color: #22c55e; }

.doc-detail {
    background: #1e293b; border: 1px solid #334155; border-radius: 10px;
    padding: 24px; overflow-y: auto; max-height: 70vh;
}
.doc-empty { display: flex; flex-direction: column; align-items: center; justify-content: center; color: #64748b; gap: 12px; }
.detail-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; gap: 12px; }
.detail-header h2 { font-size: 1.2rem; margin: 0; color: #f1f5f9; }

.detail-meta { margin-bottom: 20px; }
.meta-row { display: flex; align-items: center; gap: 8px; padding: 6px 0; font-size: 0.85rem; color: #cbd5e1; }
.meta-icon { font-size: 1rem; color: #64748b; }

.detail-actions { display: flex; gap: 8px; margin-bottom: 20px; flex-wrap: wrap; }
.btn-action {
    padding: 6px 12px; background: #0f172a; border: 1px solid #334155;
    border-radius: 8px; color: #94a3b8; cursor: pointer; transition: 0.2s;
    display: flex; align-items: center; gap: 4px; font-size: 0.8rem;
    text-decoration: none;
}
.btn-action:hover { border-color: #38bdf8; color: #38bdf8; }
.btn-sign { border-color: #22c55e; color: #22c55e; }
.btn-sign:hover { background: #052e16; }
.btn-danger:hover { border-color: #ef4444; color: #ef4444; }

.signatures-section { border-top: 1px solid #334155; padding-top: 16px; }
.signatures-section h3 { font-size: 0.95rem; color: #f1f5f9; margin-bottom: 12px; }
.sig-card {
    display: flex; gap: 12px; padding: 12px; background: #0f172a;
    border: 1px solid #334155; border-radius: 8px; margin-bottom: 8px; align-items: center;
}
.sig-image { width: 120px; height: 50px; object-fit: contain; border-radius: 4px; background: #fff; }
.sig-info { flex: 1; }
.sig-name { font-weight: 600; color: #f1f5f9; font-size: 0.9rem; }
.sig-meta { font-size: 0.75rem; color: #64748b; }

.empty-state { text-align: center; padding: 40px; color: #64748b; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.spinning { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.pagination {
    display: flex; justify-content: center; align-items: center; gap: 12px; padding: 12px;
}
.pagination button {
    background: #1e293b; border: 1px solid #334155; border-radius: 6px;
    color: #94a3b8; cursor: pointer; padding: 6px; display: flex; align-items: center;
}
.pagination button:disabled { opacity: 0.3; cursor: default; }
.pagination span { font-size: 0.85rem; color: #94a3b8; }

/* Modals */
.modal-overlay {
    position: fixed; inset: 0; background: rgba(0,0,0,0.7); z-index: 1000;
    display: flex; align-items: center; justify-content: center;
}
.modal-box {
    background: #1e293b; border: 1px solid #334155; border-radius: 12px;
    width: 90%; max-width: 520px; max-height: 90vh; overflow-y: auto;
}
.modal-sign { max-width: 600px; }
.modal-header {
    display: flex; justify-content: space-between; align-items: center;
    padding: 16px 20px; border-bottom: 1px solid #334155;
}
.modal-header h3 { margin: 0; font-size: 1.1rem; color: #f1f5f9; }
.modal-close {
    background: none; border: none; color: #94a3b8; cursor: pointer;
    padding: 4px; display: flex; border-radius: 6px;
}
.modal-close:hover { color: #f1f5f9; background: #334155; }
.modal-body { padding: 20px; }
.modal-footer {
    display: flex; justify-content: flex-end; gap: 10px;
    padding: 16px 20px; border-top: 1px solid #334155;
}

.form-label { display: block; font-size: 0.8rem; color: #94a3b8; margin: 12px 0 4px; }
.form-label:first-child { margin-top: 0; }
.form-input {
    width: 100%; padding: 10px 12px; background: #0f172a; border: 1px solid #334155;
    border-radius: 8px; color: #f1f5f9; font-size: 0.9rem; box-sizing: border-box;
}
.form-input:focus { border-color: #38bdf8; outline: none; }

.file-drop {
    border: 2px dashed #334155; border-radius: 10px; padding: 30px;
    text-align: center; cursor: pointer; transition: 0.2s; color: #94a3b8;
}
.file-drop:hover { border-color: #38bdf8; }
.file-selected { color: #38bdf8; font-weight: 600; }

.contact-dropdown {
    background: #0f172a; border: 1px solid #334155; border-radius: 8px;
    max-height: 150px; overflow-y: auto; margin-top: 4px;
}
.contact-option {
    padding: 8px 12px; cursor: pointer; font-size: 0.85rem; color: #f1f5f9;
    display: flex; gap: 8px;
}
.contact-option:hover { background: #1e293b; }
.contact-selected {
    padding: 8px 12px; display: flex; align-items: center; gap: 6px;
    font-size: 0.85rem; color: #38bdf8; cursor: pointer;
}

.sign-doc-title { color: #94a3b8; font-size: 0.85rem; margin-bottom: 8px; }
.canvas-wrapper { position: relative; margin-top: 4px; }
.sign-canvas {
    width: 100%; height: 180px; border: 2px solid #334155; border-radius: 8px;
    cursor: crosshair; touch-action: none; background: #ffffff;
}
.canvas-clear {
    position: absolute; top: 8px; right: 8px; background: #1e293b;
    border: 1px solid #334155; border-radius: 6px; color: #94a3b8;
    cursor: pointer; padding: 4px; display: flex;
}
.canvas-clear:hover { color: #f1f5f9; }
.btn-sign-submit { background: #22c55e; color: #fff; }
.btn-sign-submit:hover { background: #16a34a; }

.audit-entry {
    display: flex; gap: 12px; padding: 12px 0; border-bottom: 1px solid #1e293b;
    align-items: flex-start;
}
.audit-icon { font-size: 1.2rem; margin-top: 2px; }
.audit-body { flex: 1; }
.audit-detail { color: #f1f5f9; font-size: 0.85rem; }
.audit-meta { color: #64748b; font-size: 0.75rem; margin-top: 2px; }

.viewer-overlay { align-items: stretch; padding: 20px; }
.viewer-box {
    background: #1e293b; border: 1px solid #334155; border-radius: 12px;
    flex: 1; display: flex; flex-direction: column; overflow: hidden;
}
.viewer-header {
    display: flex; justify-content: space-between; align-items: center;
    padding: 12px 20px; border-bottom: 1px solid #334155;
}
.viewer-header h3 { margin: 0; font-size: 1rem; color: #f1f5f9; }
.viewer-toolbar {
    display: flex; gap: 6px; padding: 8px 20px; border-bottom: 1px solid #334155; background: #0f172a;
}
.vt-btn {
    padding: 4px 12px; border-radius: 6px; border: 1px solid #334155; background: transparent;
    color: #94a3b8; cursor: pointer; font-size: 0.8rem; display: flex; align-items: center; gap: 4px; transition: 0.2s;
}
.vt-btn .material-icons { font-size: 0.9rem; }
.vt-btn.active { background: #38bdf8; color: #0f172a; border-color: #38bdf8; font-weight: 600; }
.vt-btn:hover:not(.active) { border-color: #475569; color: #f1f5f9; }
.pdf-iframe { flex: 1; border: none; background: #fff; }
</style>
