<script setup>
import { ref, reactive, onMounted, onUnmounted, inject, computed, watch } from "vue"
import { useWebSocket } from "../composables/useWebSocket.js"

const authFetch = inject("authFetch")
const confirm = inject("confirm")
const toast = inject("toast")
const { onMessage } = useWebSocket()
const emailQueueCount = ref(0)

const activeTab = ref("inbox")
const loading = ref(false)
const error = ref("")

const threads = ref([])
const total = ref(0)
const totalPages = ref(1)
const page = ref(1)
const pageSize = ref(20)
const search = ref("")
const filterStatus = ref("")
const filterDirection = ref("")

const selectedThread = ref(null)
const threadLoading = ref(false)

const templates = ref([])
const signatures = ref([])
const bulkJobs = ref([])
const stats = ref(null)

const showComposeModal = ref(false)
const showBulkModal = ref(false)
const showTemplateModal = ref(false)
const showSignatureModal = ref(false)
const editingTemplateId = ref(null)
const editingSignatureId = ref(null)

const composeForm = reactive({
    contactId: "",
    to: "",
    subject: "",
    body: "",
    bodyHTML: "",
    templateId: "",
    signatureId: "",
    cc: "",
    threadId: "",
})

const replyForm = reactive({ body: "", cc: "", signatureId: "" })
const showReply = ref(false)

const bulkForm = reactive({
    name: "",
    subject: "",
    body: "",
    templateId: "",
    contactIds: [],
    filterStatus: "",
})

const templateForm = reactive({ name: "", subject: "", body: "", category: "" })
const signatureForm = reactive({ name: "", body: "", isDefault: false })

const bulkContacts = ref([])
const bulkLoading = ref(false)

const formError = ref("")

const syncStatus = ref(null)
const syncing = ref(false)
const lastSyncAt = ref(null)
const now = ref(Date.now())
let nowTimer = null

const emailAccount = ref(null)
const accountLoading = ref(false)
const accountSaving = ref(false)
const testingConnection = ref(false)
const testResult = ref(null)
const accountForm = reactive({
    provider: "gmail",
    email: "",
    password: "",
    imapHost: "",
    imapPort: 993,
    smtpHost: "",
    smtpPort: 587,
})
const showCustomServers = ref(false)

const providerOptions = [
    { value: "gmail", label: "Gmail", icon: "mail", imapHost: "imap.gmail.com", imapPort: 993, smtpHost: "smtp.gmail.com", smtpPort: 587 },
    { value: "outlook", label: "Outlook / Hotmail", icon: "mail", imapHost: "outlook.office365.com", imapPort: 993, smtpHost: "smtp-mail.outlook.com", smtpPort: 587 },
    { value: "yahoo", label: "Yahoo Mail", icon: "mail", imapHost: "imap.mail.yahoo.com", imapPort: 993, smtpHost: "smtp.mail.yahoo.com", smtpPort: 587 },
    { value: "mailpit", label: "Mailpit (Test)", icon: "science", imapHost: "mailpit", imapPort: 1143, smtpHost: "mailpit", smtpPort: 1025 },
    { value: "fake-imap", label: "Fake IMAP (Dev)", icon: "bug_report", imapHost: "fake-imap", imapPort: 1993, smtpHost: "mailpit", smtpPort: 1025 },
    { value: "custom", label: "Custom IMAP/SMTP", icon: "dns", imapHost: "", imapPort: 993, smtpHost: "", smtpPort: 587 },
]

async function loadStats() {
    try {
        const res = await authFetch("/api/v1/communication/stats")
        if (res.ok) stats.value = await res.json()
    } catch (_) {}
}

async function loadThreads() {
    loading.value = true
    error.value = ""
    try {
        const params = new URLSearchParams()
        if (search.value) params.set("search", search.value)
        if (filterStatus.value) params.set("status", filterStatus.value)
        if (filterDirection.value) params.set("direction", filterDirection.value)
        params.set("page", page.value)
        params.set("page_size", pageSize.value)
        const res = await authFetch("/api/v1/communication/threads/?" + params.toString())
        if (!res.ok) throw new Error("Błąd pobierania wątków")
        const data = await res.json()
        threads.value = data.data || []
        total.value = data.total || 0
        totalPages.value = data.total_pages || 1
    } catch (e) {
        error.value = e.message
    } finally {
        loading.value = false
    }
}

async function loadTemplates() {
    try {
        const res = await authFetch("/api/v1/communication/templates/")
        if (res.ok) templates.value = await res.json()
    } catch (_) {}
}

async function loadSignatures() {
    try {
        const res = await authFetch("/api/v1/communication/signatures/")
        if (res.ok) signatures.value = await res.json()
    } catch (_) {}
}

async function loadBulkJobs() {
    bulkLoading.value = true
    try {
        const res = await authFetch("/api/v1/communication/bulk/")
        if (res.ok) bulkJobs.value = await res.json()
    } catch (_) {} finally {
        bulkLoading.value = false
    }
}

async function openThread(thread) {
    threadLoading.value = true
    selectedThread.value = thread
    showReply.value = false
    replyForm.body = ""
    replyForm.cc = ""
    try {
        const res = await authFetch("/api/v1/communication/threads/" + thread.id)
        if (res.ok) {
            selectedThread.value = await res.json()
        }
        await authFetch("/api/v1/communication/threads/" + thread.id + "/read", { method: "PUT" })
        thread.unread_count = 0
        await loadStats()
    } catch (_) {} finally {
        threadLoading.value = false
    }
}

async function syncInbox() {
    syncing.value = true
    syncStatus.value = null
    try {
        const res = await authFetch("/api/v1/communication/sync", { method: "POST" })
        if (res.ok) {
            syncStatus.value = await res.json()
            lastSyncAt.value = new Date()
            if (syncStatus.value.error) {
                formError.value = syncStatus.value.error
            } else if (syncStatus.value.new_messages > 0) {
                toast.show(`Zsynchronizowano ${syncStatus.value.new_messages} nowych wiadomości`, "success")
            }
        }
        await loadThreads()
        await loadStats()
    } catch (_) {} finally {
        syncing.value = false
    }
}

async function sendCompose() {
    formError.value = ""
    if (!composeForm.to || !composeForm.subject || !composeForm.body) {
        formError.value = "Pola Do, Temat i Treść są wymagane."
        return
    }
    try {
        const payload = {
            contact_id: composeForm.contactId || "manual",
            to: composeForm.to,
            subject: composeForm.subject,
            body: composeForm.body,
            body_html: composeForm.bodyHTML,
            cc: composeForm.cc ? composeForm.cc.split(",").map(s => s.trim()) : [],
            template_id: composeForm.templateId,
            signature_id: composeForm.signatureId,
        }
        const res = await authFetch("/api/v1/communication/threads/", {
            method: "POST",
            body: JSON.stringify(payload),
        })
        if (!res.ok) {
            const d = await res.json()
            throw new Error(d.error || "Błąd wysyłki")
        }
        showComposeModal.value = false
        toast.show("Wiadomość dodana do kolejki wysyłki", "info")
        Object.assign(composeForm, { contactId: "", to: "", subject: "", body: "", bodyHTML: "", templateId: "", signatureId: "", cc: "", threadId: "" })
        await loadThreads()
        await loadStats()
    } catch (e) {
        formError.value = e.message
    }
}

async function sendReply() {
    formError.value = ""
    if (!replyForm.body) {
        formError.value = "Treść odpowiedzi jest wymagana."
        return
    }
    try {
        const payload = {
            body: replyForm.body,
            cc: replyForm.cc ? replyForm.cc.split(",").map(s => s.trim()) : [],
            signature_id: replyForm.signatureId,
        }
        const res = await authFetch("/api/v1/communication/threads/" + selectedThread.value.id + "/reply", {
            method: "POST",
            body: JSON.stringify(payload),
        })
        if (!res.ok) throw new Error("Błąd wysyłki odpowiedzi")
        showReply.value = false
        replyForm.body = ""
        toast.show("Odpowiedź dodana do kolejki wysyłki", "info")
        await openThread(selectedThread.value)
        await loadStats()
    } catch (e) {
        formError.value = e.message
    }
}

async function archiveThread(threadId) {
    await authFetch("/api/v1/communication/threads/" + threadId + "/archive", { method: "PUT" })
    if (selectedThread.value?.id === threadId) selectedThread.value = null
    await loadThreads()
}

async function deleteThread(threadId) {
    if (!await confirm({ title: "Usuń wątek", message: "Wątek i wszystkie wiadomości zostaną trwale usunięte.", confirmLabel: "Usuń", variant: "danger" })) return
    await authFetch("/api/v1/communication/threads/" + threadId, { method: "DELETE" })
    if (selectedThread.value?.id === threadId) selectedThread.value = null
    await loadThreads()
    await loadStats()
}

async function updateThreadStatus(threadId, status) {
    await authFetch("/api/v1/communication/threads/" + threadId + "/status", {
        method: "PUT",
        body: JSON.stringify({ status }),
    })
    if (selectedThread.value?.id === threadId) selectedThread.value.status = status
    await loadThreads()
}

async function starMessage(msgId, starred) {
    await authFetch("/api/v1/communication/messages/" + msgId + "/star", {
        method: "PUT",
        body: JSON.stringify({ starred }),
    })
    if (selectedThread.value) {
        const msg = selectedThread.value.messages?.find(m => m.id === msgId)
        if (msg) msg.is_starred = starred
    }
}

async function loadBulkContacts() {
    try {
        const params = new URLSearchParams()
        if (bulkForm.filterStatus) params.set("status", bulkForm.filterStatus)
        params.set("page_size", "200")
        const res = await authFetch("/api/v1/contacts/?" + params.toString())
        if (res.ok) {
            const data = await res.json()
            bulkContacts.value = data.data || []
            bulkForm.contactIds = bulkContacts.value.filter(c => c.email).map(c => c.id)
        }
    } catch (_) {}
}

async function sendBulk() {
    formError.value = ""
    if (!bulkForm.name || !bulkForm.subject || !bulkForm.body) {
        formError.value = "Nazwa, Temat i Treść są wymagane."
        return
    }
    if (bulkForm.contactIds.length === 0) {
        formError.value = "Wybierz co najmniej jeden kontakt."
        return
    }
    try {
        const payload = {
            name: bulkForm.name,
            subject: bulkForm.subject,
            body: bulkForm.body,
            contact_ids: bulkForm.contactIds,
            template_id: bulkForm.templateId,
        }
        const res = await authFetch("/api/v1/communication/bulk/", {
            method: "POST",
            body: JSON.stringify(payload),
        })
        if (!res.ok) throw new Error("Błąd uruchomienia kampanii")
        showBulkModal.value = false
        toast.show("Kampania uruchomiona - wysyłanie w toku", "info")
        Object.assign(bulkForm, { name: "", subject: "", body: "", templateId: "", contactIds: [], filterStatus: "" })
        activeTab.value = "bulk"
        await loadBulkJobs()
        await loadStats()
    } catch (e) {
        formError.value = e.message
    }
}

async function saveTemplate() {
    formError.value = ""
    if (!templateForm.name || !templateForm.subject || !templateForm.body) {
        formError.value = "Nazwa, temat i treść są wymagane."
        return
    }
    try {
        const payload = {
            name: templateForm.name,
            subject: templateForm.subject,
            body: templateForm.body,
            category: templateForm.category,
        }
        let res
        if (editingTemplateId.value) {
            res = await authFetch("/api/v1/communication/templates/" + editingTemplateId.value, {
                method: "PUT",
                body: JSON.stringify(payload),
            })
        } else {
            res = await authFetch("/api/v1/communication/templates/", {
                method: "POST",
                body: JSON.stringify(payload),
            })
        }
        if (!res.ok) throw new Error("Błąd zapisu szablonu")
        showTemplateModal.value = false
        Object.assign(templateForm, { name: "", subject: "", body: "", category: "" })
        editingTemplateId.value = null
        await loadTemplates()
    } catch (e) {
        formError.value = e.message
    }
}

async function deleteTemplate(id) {
    if (!await confirm({ title: "Usuń szablon", message: "Szablon zostanie trwale usunięty.", confirmLabel: "Usuń", variant: "danger" })) return
    await authFetch("/api/v1/communication/templates/" + id, { method: "DELETE" })
    await loadTemplates()
}

function openEditTemplate(t) {
    editingTemplateId.value = t.id
    Object.assign(templateForm, { name: t.name, subject: t.subject, body: t.body, category: t.category })
    formError.value = ""
    showTemplateModal.value = true
}

function useTemplate(t) {
    composeForm.subject = t.subject
    composeForm.body = t.body
    composeForm.templateId = t.id
    showTemplateModal.value = false
}

async function saveSignature() {
    formError.value = ""
    if (!signatureForm.name || !signatureForm.body) {
        formError.value = "Nazwa i treść są wymagane."
        return
    }
    try {
        const payload = { name: signatureForm.name, body: signatureForm.body, is_default: signatureForm.isDefault }
        let res
        if (editingSignatureId.value) {
            res = await authFetch("/api/v1/communication/signatures/" + editingSignatureId.value, {
                method: "PUT",
                body: JSON.stringify(payload),
            })
        } else {
            res = await authFetch("/api/v1/communication/signatures/", {
                method: "POST",
                body: JSON.stringify(payload),
            })
        }
        if (!res.ok) throw new Error("Błąd zapisu podpisu")
        showSignatureModal.value = false
        Object.assign(signatureForm, { name: "", body: "", isDefault: false })
        editingSignatureId.value = null
        await loadSignatures()
    } catch (e) {
        formError.value = e.message
    }
}

async function deleteSignature(id) {
    if (!await confirm({ title: "Usuń podpis", message: "Podpis zostanie trwale usunięty.", confirmLabel: "Usuń", variant: "danger" })) return
    await authFetch("/api/v1/communication/signatures/" + id, { method: "DELETE" })
    await loadSignatures()
}

function openEditSignature(s) {
    editingSignatureId.value = s.id
    Object.assign(signatureForm, { name: s.name, body: s.body, isDefault: s.is_default })
    formError.value = ""
    showSignatureModal.value = true
}

function openCompose() {
    Object.assign(composeForm, { contactId: "", to: "", subject: "", body: "", bodyHTML: "", templateId: "", signatureId: "", cc: "", threadId: "" })
    formError.value = ""
    showComposeModal.value = true
}

function openBulk() {
    Object.assign(bulkForm, { name: "", subject: "", body: "", templateId: "", contactIds: [], filterStatus: "" })
    formError.value = ""
    bulkContacts.value = []
    showBulkModal.value = true
    loadBulkContacts()
}

function applyTemplateToCompose(t) {
    composeForm.subject = t.subject
    composeForm.body = t.body
    composeForm.templateId = t.id
}

function applyTemplateToReply(t) {
    replyForm.body = t.body
}

function applyTemplateToBulk(t) {
    bulkForm.subject = t.subject
    bulkForm.body = t.body
    bulkForm.templateId = t.id
}

function applySignatureToCompose(s) {
    composeForm.body = composeForm.body + "\n\n--\n" + s.body
    composeForm.signatureId = s.id
}

function applySignatureToReply(s) {
    replyForm.body = replyForm.body + "\n\n--\n" + s.body
    replyForm.signatureId = s.id
}

function formatDate(d) {
    if (!d) return ""
    const dt = new Date(d)
    const now = new Date()
    const diff = now - dt
    if (diff < 86400000) return dt.toLocaleTimeString("pl-PL", { hour: "2-digit", minute: "2-digit" })
    if (diff < 7 * 86400000) {
        const days = ["Nd", "Pn", "Wt", "Śr", "Cz", "Pt", "Sb"]
        return days[dt.getDay()]
    }
    return dt.toLocaleDateString("pl-PL", { day: "2-digit", month: "2-digit", year: "2-digit" })
}

function formatDateFull(d) {
    if (!d) return ""
    return new Date(d).toLocaleString("pl-PL")
}

function timeAgo(d) {
    if (!d) return ""
    const sec = Math.floor((now.value - new Date(d).getTime()) / 1000)
    if (sec < 5) return "właśnie teraz"
    if (sec < 60) return `${sec}s temu`
    const min = Math.floor(sec / 60)
    if (min < 60) return `${min}min temu`
    return formatDateFull(d)
}

function statusLabel(s) {
    const map = { open: "Otwarty", archived: "Archiwum", closed: "Zamknięty" }
    return map[s] || s
}

function statusClass(s) {
    const map = { open: "status-open", archived: "status-archived", closed: "status-closed" }
    return map[s] || ""
}

function jobStatusLabel(s) {
    const map = { pending: "Oczekuje", running: "W toku", completed: "Zakończony", failed: "Błąd" }
    return map[s] || s
}

function jobStatusClass(s) {
    const map = { pending: "job-pending", running: "job-running", completed: "job-completed", failed: "job-failed" }
    return map[s] || ""
}

function updateBulkJobProgress(data) {
    const job = bulkJobs.value.find(j => j.id === data.bulk_job_id)
    if (!job) {
        loadBulkJobs()
        return
    }
    job.sent_count = data.bulk_sent
    job.failed_count = data.bulk_failed || 0
    if (data.completed) {
        job.status = "completed"
        job.completed_at = new Date().toISOString()
    } else {
        job.status = "running"
    }
}

function directionIcon(d) {
    return d === "inbound" ? "call_received" : "call_made"
}

function initials(name) {
    if (!name) return "?"
    return name.split(" ").map(w => w[0]).join("").toUpperCase().slice(0, 2)
}

function avatarColor(name) {
    const colors = ["#38bdf8", "#818cf8", "#34d399", "#f59e0b", "#f472b6", "#a78bfa", "#fb923c"]
    let hash = 0
    for (const c of (name || "")) hash = c.charCodeAt(0) + ((hash << 5) - hash)
    return colors[Math.abs(hash) % colors.length]
}

const toggleBulkContact = (id) => {
    const idx = bulkForm.contactIds.indexOf(id)
    if (idx === -1) bulkForm.contactIds.push(id)
    else bulkForm.contactIds.splice(idx, 1)
}

const selectAllBulkContacts = () => {
    bulkForm.contactIds = bulkContacts.value.filter(c => c.email).map(c => c.id)
}

const clearBulkContacts = () => {
    bulkForm.contactIds = []
}

async function loadEmailAccount() {
    accountLoading.value = true
    try {
        const res = await authFetch("/api/v1/communication/account/")
        if (res.ok) {
            const data = await res.json()
            if (data.configured === false) {
                emailAccount.value = null
            } else {
                emailAccount.value = data
                accountForm.provider = data.provider || "gmail"
                accountForm.email = data.email || ""
                accountForm.password = ""
                accountForm.imapHost = data.imap_host || ""
                accountForm.imapPort = data.imap_port || 993
                accountForm.smtpHost = data.smtp_host || ""
                accountForm.smtpPort = data.smtp_port || 587
                showCustomServers.value = data.provider === "custom" || data.provider === "mailpit" || data.provider === "fake-imap"
            }
        }
    } catch (_) {} finally {
        accountLoading.value = false
    }
}

function onProviderChange() {
    const provider = providerOptions.find(p => p.value === accountForm.provider)
    if (provider) {
        accountForm.imapHost = provider.imapHost
        accountForm.imapPort = provider.imapPort
        accountForm.smtpHost = provider.smtpHost
        accountForm.smtpPort = provider.smtpPort
        showCustomServers.value = accountForm.provider === "custom" || accountForm.provider === "mailpit"
    }
}

async function saveEmailAccount() {
    formError.value = ""
    if (!accountForm.email || !accountForm.password) {
        formError.value = "Email and app password are required."
        return
    }
    accountSaving.value = true
    try {
        const payload = {
            provider: accountForm.provider,
            email: accountForm.email,
            password: accountForm.password,
            imap_host: accountForm.imapHost,
            imap_port: accountForm.imapPort,
            smtp_host: accountForm.smtpHost,
            smtp_port: accountForm.smtpPort,
        }
        let res
        if (emailAccount.value) {
            res = await authFetch("/api/v1/communication/account/", {
                method: "PUT",
                body: JSON.stringify(payload),
            })
        } else {
            res = await authFetch("/api/v1/communication/account/", {
                method: "POST",
                body: JSON.stringify(payload),
            })
        }
        if (!res.ok) {
            const d = await res.json()
            throw new Error(d.error || "Failed to save account")
        }
        const data = await res.json()
        emailAccount.value = data
        testResult.value = null
        formError.value = ""
    } catch (e) {
        formError.value = e.message
    } finally {
        accountSaving.value = false
    }
}

async function testConnection() {
    testingConnection.value = true
    testResult.value = null
    try {
        const res = await authFetch("/api/v1/communication/account/test", { method: "POST" })
        if (res.ok) testResult.value = await res.json()
    } catch (_) {
        testResult.value = { imap_status: "error", smtp_status: "error", error: "Connection test failed." }
    } finally {
        testingConnection.value = false
    }
}

async function disconnectAccount() {
    if (!await confirm({ title: "Rozłącz konto email", message: "Konto email zostanie rozłączone. Zsynchronizowane wiadomości pozostaną.", confirmLabel: "Rozłącz", variant: "danger" })) return
    try {
        await authFetch("/api/v1/communication/account/", { method: "DELETE" })
        emailAccount.value = null
        Object.assign(accountForm, { provider: "gmail", email: "", password: "", imapHost: "", imapPort: 993, smtpHost: "", smtpPort: 587 })
        testResult.value = null
    } catch (_) {}
}

watch(filterStatus, () => { page.value = 1; loadThreads() })
watch(filterDirection, () => { page.value = 1; loadThreads() })

onMounted(async () => {
    await Promise.all([loadStats(), loadThreads(), loadTemplates(), loadSignatures(), loadEmailAccount()])
    onMessage((msg) => {
        if (msg.type === "email_sent" || msg.type === "email_send_failed") {
            emailQueueCount.value = msg.data?.queue_remaining || 0
            if (!msg.data?.bulk_job_id) {
                toast.show(msg.message, msg.type === "email_sent" ? "success" : "error")
                loadThreads()
                loadStats()
            }
        } else if (msg.type === "bulk_progress") {
            updateBulkJobProgress(msg.data)
            loadThreads()
            loadStats()
        } else if (msg.type === "bulk_completed") {
            updateBulkJobProgress(msg.data)
            toast.show(msg.message, "success")
            loadThreads()
            loadStats()
        } else if (msg.type === "email_sync_done") {
            lastSyncAt.value = new Date()
            const n = msg.data?.new_messages || 0
            if (n > 0) {
                toast.show(`Nowa poczta: ${n} ${n === 1 ? 'wiadomość' : 'wiadomości'}`, "info")
            }
            emailQueueCount.value = 0
            loadThreads()
            loadStats()
        } else if (msg.type === "email_sync_failed") {
            toast.show(msg.message, "error")
        }
    })
    nowTimer = setInterval(() => { now.value = Date.now() }, 5000)
})

onUnmounted(() => {
    if (nowTimer) clearInterval(nowTimer)
})

watch(activeTab, (tab) => {
    if (tab === "bulk") loadBulkJobs()
    if (tab === "inbox") loadThreads()
    if (tab === "settings") loadEmailAccount()
})
</script>

<template>
    <div class="comm-wrapper">
        <div class="comm-header">
            <div>
                <h1>Komunikacja</h1>
                <p class="subtitle">Zarządzaj korespondencją z kontaktami</p>
            </div>
            <div class="header-actions">
                <span v-if="emailAccount" class="connected-badge" title="Connected">
                    <span class="material-icons" style="font-size:0.8rem">circle</span>
                    {{ emailAccount.email }}
                </span>
                <div class="sync-group">
                    <button class="btn-sync" @click="syncInbox" :disabled="syncing || !emailAccount" :title="!emailAccount ? 'Configure email in Settings first' : 'Sync inbox'">
                        <span class="material-icons" :class="{ spinning: syncing }">sync</span>
                        <template v-if="syncing">Synchronizuję...</template>
                        <template v-else-if="syncStatus && syncStatus.new_messages > 0">+{{ syncStatus.new_messages }} nowych</template>
                        <template v-else>Synchronizuj</template>
                    </button>
                    <span v-if="lastSyncAt" class="last-sync-label">
                        <span class="material-icons" style="font-size:0.75rem">schedule</span>
                        {{ timeAgo(lastSyncAt) }}
                    </span>
                </div>
                <button class="btn-bulk" @click="openBulk">
                    <span class="material-icons">send</span> Wyślij masowo
                </button>
                <button class="btn-primary" @click="openCompose">
                    <span class="material-icons">edit</span> Nowa wiadomość
                </button>
            </div>
        </div>

        <div v-if="stats" class="stats-row">
            <div class="stat-card">
                <span class="material-icons stat-icon" style="color:#38bdf8">forum</span>
                <div>
                    <div class="stat-value">{{ stats.total_threads }}</div>
                    <div class="stat-label">Wszystkie wątki</div>
                </div>
            </div>
            <div class="stat-card">
                <span class="material-icons stat-icon" style="color:#22c55e">mark_email_unread</span>
                <div>
                    <div class="stat-value">{{ stats.unread_messages }}</div>
                    <div class="stat-label">Nieprzeczytane</div>
                </div>
            </div>
            <div class="stat-card">
                <span class="material-icons stat-icon" style="color:#f59e0b">send</span>
                <div>
                    <div class="stat-value">{{ stats.sent_today }}</div>
                    <div class="stat-label">Wysłane dziś</div>
                </div>
            </div>
            <div class="stat-card">
                <span class="material-icons stat-icon" style="color:#a78bfa">campaign</span>
                <div>
                    <div class="stat-value">{{ stats.bulk_jobs_run }}</div>
                    <div class="stat-label">Kampanie</div>
                </div>
            </div>
        </div>

        <div v-if="emailQueueCount > 0" class="queue-bar">
            <div class="queue-bar-inner">
                <span class="material-icons queue-icon spinning">sync</span>
                <span>Wysyłanie wiadomości... ({{ emailQueueCount }} w kolejce)</span>
            </div>
            <div class="queue-progress">
                <div class="queue-progress-fill"></div>
            </div>
        </div>

        <div class="comm-tabs">
            <button :class="['tab-btn', { active: activeTab === 'inbox' }]" @click="activeTab = 'inbox'">
                <span class="material-icons">inbox</span> Skrzynka
                <span v-if="stats?.unread_messages > 0" class="unread-badge">{{ stats.unread_messages }}</span>
            </button>
            <button :class="['tab-btn', { active: activeTab === 'bulk' }]" @click="activeTab = 'bulk'">
                <span class="material-icons">campaign</span> Kampanie masowe
            </button>
            <button :class="['tab-btn', { active: activeTab === 'templates' }]" @click="activeTab = 'templates'">
                <span class="material-icons">description</span> Szablony
            </button>
            <button :class="['tab-btn', { active: activeTab === 'signatures' }]" @click="activeTab = 'signatures'">
                <span class="material-icons">draw</span> Podpisy
            </button>
            <button :class="['tab-btn', { active: activeTab === 'recent' }]" @click="activeTab = 'recent'">
                <span class="material-icons">people</span> Ostatnie kontakty
            </button>
            <button :class="['tab-btn', { active: activeTab === 'settings' }]" @click="activeTab = 'settings'">
                <span class="material-icons">settings</span> Ustawienia
                <span v-if="!emailAccount" class="setup-dot"></span>
            </button>
        </div>

        <div v-if="activeTab === 'inbox'" class="inbox-layout">
            <div class="inbox-sidebar">
                <div class="inbox-filters">
                    <input v-model="search" class="search-input" placeholder="Szukaj..." @input="() => { page = 1; loadThreads() }" />
                    <select v-model="filterStatus" class="filter-sel">
                        <option value="">Wszystkie statusy</option>
                        <option value="open">Otwarte</option>
                        <option value="archived">Archiwum</option>
                        <option value="closed">Zamknięte</option>
                    </select>
                    <select v-model="filterDirection" class="filter-sel">
                        <option value="">Oba kierunki</option>
                        <option value="inbound">Przychodzące</option>
                        <option value="outbound">Wychodzące</option>
                    </select>
                </div>

                <div v-if="loading" class="loading-sm">Ładowanie...</div>
                <div v-else-if="threads.length === 0" class="empty-sm">Brak wiadomości.</div>
                <div v-else class="thread-list">
                    <div
                        v-for="t in threads"
                        :key="t.id"
                        :class="['thread-item', { active: selectedThread?.id === t.id, unread: t.unread_count > 0 }]"
                        @click="openThread(t)"
                    >
                        <div class="thread-avatar" :style="{ background: avatarColor(t.contact_name) }">
                            {{ initials(t.contact_name || t.contact_email) }}
                        </div>
                        <div class="thread-info">
                            <div class="thread-row1">
                                <span class="thread-name">{{ t.contact_name || t.contact_email }}</span>
                                <span class="thread-date">{{ formatDate(t.last_message_at) }}</span>
                            </div>
                            <div class="thread-row2">
                                <span class="material-icons dir-icon" :title="t.direction === 'inbound' ? 'Przychodzące' : 'Wychodzące'">
                                    {{ directionIcon(t.direction) }}
                                </span>
                                <span class="thread-subject">{{ t.subject }}</span>
                                <span v-if="t.unread_count > 0" class="unread-dot">{{ t.unread_count }}</span>
                            </div>
                            <div class="thread-row3">
                                <span :class="['thread-status', statusClass(t.status)]">{{ statusLabel(t.status) }}</span>
                                <span class="thread-count">{{ t.message_count }} wiad.</span>
                            </div>
                        </div>
                    </div>
                </div>

                <div v-if="totalPages > 1" class="pagination-sm">
                    <button class="page-btn-sm" :disabled="page <= 1" @click="page--; loadThreads()">‹</button>
                    <span>{{ page }} / {{ totalPages }}</span>
                    <button class="page-btn-sm" :disabled="page >= totalPages" @click="page++; loadThreads()">›</button>
                </div>
            </div>

            <div class="inbox-main">
                <div v-if="!selectedThread" class="empty-thread">
                    <span class="material-icons empty-icon">mail_outline</span>
                    <p>Wybierz wątek, aby zobaczyć wiadomości</p>
                </div>
                <div v-else-if="threadLoading" class="loading">Ładowanie...</div>
                <div v-else class="thread-view">
                    <div class="thread-header">
                        <div class="thread-header-left">
                            <div class="thread-avatar-lg" :style="{ background: avatarColor(selectedThread.contact_name) }">
                                {{ initials(selectedThread.contact_name || selectedThread.contact_email) }}
                            </div>
                            <div>
                                <h2>{{ selectedThread.subject }}</h2>
                                <p class="thread-meta">
                                    {{ selectedThread.contact_name || selectedThread.contact_email }}
                                    <span v-if="selectedThread.contact_name">&lt;{{ selectedThread.contact_email }}&gt;</span>
                                </p>
                            </div>
                        </div>
                        <div class="thread-header-actions">
                            <button class="btn-action-sm" @click="showReply = !showReply" title="Odpowiedz">
                                <span class="material-icons">reply</span>
                            </button>
                            <button class="btn-action-sm" @click="updateThreadStatus(selectedThread.id, selectedThread.status === 'open' ? 'closed' : 'open')" title="Zamknij/Otwórz">
                                <span class="material-icons">{{ selectedThread.status === 'open' ? 'check_circle' : 'radio_button_unchecked' }}</span>
                            </button>
                            <button class="btn-action-sm" @click="archiveThread(selectedThread.id)" title="Archiwizuj">
                                <span class="material-icons">archive</span>
                            </button>
                            <button class="btn-action-sm btn-danger" @click="deleteThread(selectedThread.id)" title="Usuń">
                                <span class="material-icons">delete</span>
                            </button>
                        </div>
                    </div>

                    <div class="messages-list">
                        <div v-for="msg in (selectedThread.messages || [])" :key="msg.id" :class="['message-item', msg.direction]">
                            <div class="msg-avatar" :style="{ background: msg.direction === 'outbound' ? '#38bdf8' : avatarColor(selectedThread.contact_name) }">
                                {{ msg.direction === 'outbound' ? 'Ja' : initials(selectedThread.contact_name) }}
                            </div>
                            <div class="msg-body">
                                <div class="msg-header">
                                    <span class="msg-from">{{ msg.from }}</span>
                                    <div class="msg-actions">
                                        <button class="btn-star" :class="{ starred: msg.is_starred }" @click="starMessage(msg.id, !msg.is_starred)" title="Oznacz gwiazdką">
                                            <span class="material-icons">{{ msg.is_starred ? 'star' : 'star_border' }}</span>
                                        </button>
                                        <span class="msg-date">{{ formatDateFull(msg.sent_at || msg.created_at) }}</span>
                                    </div>
                                </div>
                                <div class="msg-to">Do: {{ msg.to }}<span v-if="msg.cc"> CC: {{ msg.cc }}</span></div>
                                <div class="msg-text">{{ msg.body }}</div>
                            </div>
                        </div>
                    </div>

                    <div v-if="showReply" class="reply-form">
                        <div class="reply-header">
                            <h3>Odpowiedz</h3>
                            <div class="reply-quick-actions">
                                <span class="quick-label">Szablon:</span>
                                <select class="quick-sel" @change="e => { const t = templates.find(t => t.id === e.target.value); if(t) applyTemplateToReply(t); e.target.value = '' }">
                                    <option value="">Wybierz szablon...</option>
                                    <option v-for="t in templates" :key="t.id" :value="t.id">{{ t.name }}</option>
                                </select>
                                <span class="quick-label">Podpis:</span>
                                <select class="quick-sel" @change="e => { const s = signatures.find(s => s.id === e.target.value); if(s) applySignatureToReply(s); e.target.value = '' }">
                                    <option value="">Dodaj podpis...</option>
                                    <option v-for="s in signatures" :key="s.id" :value="s.id">{{ s.name }}</option>
                                </select>
                            </div>
                        </div>
                        <div class="form-group">
                            <label>CC (opcjonalnie)</label>
                            <input v-model="replyForm.cc" placeholder="email1@example.com, email2@example.com" />
                        </div>
                        <div class="form-group">
                            <label>Treść *</label>
                            <textarea v-model="replyForm.body" rows="8" placeholder="Treść odpowiedzi..."></textarea>
                        </div>
                        <p v-if="formError" class="err-msg">{{ formError }}</p>
                        <div class="reply-actions">
                            <button class="btn-secondary" @click="showReply = false">Anuluj</button>
                            <button class="btn-primary" @click="sendReply">
                                <span class="material-icons">send</span> Wyślij odpowiedź
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div v-if="activeTab === 'bulk'" class="bulk-section">
            <div class="section-header">
                <h2>Kampanie masowe</h2>
                <button class="btn-primary" @click="openBulk">
                    <span class="material-icons">add</span> Nowa kampania
                </button>
            </div>
            <div v-if="bulkLoading" class="loading">Ładowanie...</div>
            <div v-else-if="bulkJobs.length === 0" class="empty">Brak kampanii.</div>
            <div v-else class="bulk-grid">
                <div v-for="job in bulkJobs" :key="job.id" class="bulk-card">
                    <div class="bulk-card-header">
                        <div>
                            <div class="bulk-name">{{ job.name }}</div>
                            <div class="bulk-subject">{{ job.subject }}</div>
                        </div>
                        <span :class="['job-badge', jobStatusClass(job.status)]">{{ jobStatusLabel(job.status) }}</span>
                    </div>
                    <div class="bulk-progress">
                        <div class="progress-bar-wrap">
                            <div :class="['progress-bar', { 'progress-bar-animated': job.status === 'running' }]"
                                 :style="{ width: job.total_count > 0 ? (job.sent_count / job.total_count * 100) + '%' : '0%' }"></div>
                        </div>
                        <div class="progress-stats">
                            <span class="prog-sent">✓ {{ job.sent_count }}</span>
                            <span class="prog-fail" v-if="job.failed_count > 0">✗ {{ job.failed_count }}</span>
                            <span class="prog-total">/ {{ job.total_count }}</span>
                            <span v-if="job.status === 'running'" class="prog-pct">{{ Math.round(job.sent_count / job.total_count * 100) }}%</span>
                        </div>
                    </div>
                    <div class="bulk-dates">
                        <span v-if="job.started_at">Start: {{ formatDateFull(job.started_at) }}</span>
                        <span v-if="job.completed_at">Koniec: {{ formatDateFull(job.completed_at) }}</span>
                    </div>
                </div>
            </div>
        </div>

        <div v-if="activeTab === 'templates'" class="templates-section">
            <div class="section-header">
                <h2>Szablony wiadomości</h2>
                <button class="btn-primary" @click="() => { editingTemplateId = null; Object.assign(templateForm, { name: '', subject: '', body: '', category: '' }); formError = ''; showTemplateModal = true }">
                    + Nowy szablon
                </button>
            </div>
            <div v-if="templates.length === 0" class="empty">Brak szablonów.</div>
            <div v-else class="templates-grid">
                <div v-for="t in templates" :key="t.id" class="template-card">
                    <div class="template-header">
                        <div>
                            <div class="template-name">{{ t.name }}</div>
                            <div class="template-cat" v-if="t.category">{{ t.category }}</div>
                        </div>
                        <div class="template-usage">
                            <span class="material-icons usage-icon">bar_chart</span>
                            {{ t.used_count }}x
                        </div>
                    </div>
                    <div class="template-subject">{{ t.subject }}</div>
                    <div class="template-body">{{ t.body?.slice(0, 120) }}{{ t.body?.length > 120 ? '...' : '' }}</div>
                    <div class="template-actions">
                        <button class="btn-use" @click="() => { openCompose(); applyTemplateToCompose(t) }">
                            <span class="material-icons">send</span> Użyj
                        </button>
                        <button class="btn-action-sm" @click="openEditTemplate(t)" title="Edytuj">
                            <span class="material-icons">edit</span>
                        </button>
                        <button class="btn-action-sm btn-danger" @click="deleteTemplate(t.id)" title="Usuń">
                            <span class="material-icons">delete</span>
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <div v-if="activeTab === 'signatures'" class="signatures-section">
            <div class="section-header">
                <h2>Podpisy e-mail</h2>
                <button class="btn-primary" @click="() => { editingSignatureId = null; Object.assign(signatureForm, { name: '', body: '', isDefault: false }); formError = ''; showSignatureModal = true }">
                    + Nowy podpis
                </button>
            </div>
            <div v-if="signatures.length === 0" class="empty">Brak podpisów.</div>
            <div v-else class="signatures-grid">
                <div v-for="s in signatures" :key="s.id" class="signature-card">
                    <div class="sig-header">
                        <div class="sig-name">{{ s.name }}</div>
                        <span v-if="s.is_default" class="default-badge">Domyślny</span>
                    </div>
                    <div class="sig-body">{{ s.body }}</div>
                    <div class="sig-actions">
                        <button class="btn-action-sm" @click="openEditSignature(s)">
                            <span class="material-icons">edit</span>
                        </button>
                        <button class="btn-action-sm btn-danger" @click="deleteSignature(s.id)">
                            <span class="material-icons">delete</span>
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <div v-if="activeTab === 'recent'" class="recent-section">
            <h2>Ostatnio kontaktowane osoby</h2>
            <div v-if="!stats?.recent_contacts?.length" class="empty">Brak danych.</div>
            <div v-else class="recent-grid">
                <div v-for="rc in stats.recent_contacts" :key="rc.contact_id" class="recent-card">
                    <div class="recent-avatar" :style="{ background: avatarColor(rc.contact_name) }">
                        {{ initials(rc.contact_name || rc.contact_email) }}
                    </div>
                    <div class="recent-info">
                        <div class="recent-name">{{ rc.contact_name || rc.contact_email }}</div>
                        <div class="recent-email">{{ rc.contact_email }}</div>
                        <div class="recent-meta">
                            <span>{{ rc.thread_count }} wątki</span>
                            <span>Ostatni kontakt: {{ formatDate(rc.last_contact) }}</span>
                        </div>
                    </div>
                    <button class="btn-compose-quick" @click="() => { openCompose(); composeForm.to = rc.contact_email; composeForm.contactId = rc.contact_id }">
                        <span class="material-icons">reply</span>
                    </button>
                </div>
            </div>
        </div>

        <div v-if="activeTab === 'settings'" class="settings-section">
            <h2>Ustawienia konta e-mail</h2>
            <p class="settings-desc">Połącz konto e-mail, aby synchronizować skrzynkę odbiorczą i wysyłać wiadomości bezpośrednio z CRM.</p>

            <div v-if="accountLoading" class="loading">Ładowanie...</div>
            <template v-else>
                <div v-if="emailAccount" class="account-status-card">
                    <div class="account-status-header">
                        <div class="account-status-left">
                            <span class="material-icons account-status-icon connected">check_circle</span>
                            <div>
                                <div class="account-provider">{{ emailAccount.provider?.toUpperCase() }}</div>
                                <div class="account-email">{{ emailAccount.email }}</div>
                            </div>
                        </div>
                        <div class="account-status-right">
                            <div class="account-meta">
                                <span v-if="emailAccount.last_sync_at">Ostatnia synchronizacja: {{ formatDateFull(emailAccount.last_sync_at) }}</span>
                                <span v-else>Nigdy nie synchronizowano</span>
                            </div>
                            <div class="account-meta">Zsynchronizowano łącznie: {{ emailAccount.synced_count || 0 }} wiadomości</div>
                        </div>
                    </div>
                </div>

                <div class="settings-form-card">
                    <h3>{{ emailAccount ? 'Zaktualizuj połączenie' : 'Połącz konto e-mail' }}</h3>

                    <div class="provider-grid">
                        <div
                            v-for="p in providerOptions"
                            :key="p.value"
                            :class="['provider-card', { selected: accountForm.provider === p.value }]"
                            @click="accountForm.provider = p.value; onProviderChange()"
                        >
                            <span class="material-icons provider-icon">{{ p.icon }}</span>
                            <span class="provider-label">{{ p.label }}</span>
                        </div>
                    </div>

                    <div class="form-group">
                        <label>Adres e-mail *</label>
                        <input v-model="accountForm.email" type="email" placeholder="twoj@gmail.com" />
                    </div>
                    <div class="form-group">
                        <label>
                            Hasło aplikacji *
                            <a v-if="accountForm.provider === 'gmail'" href="https://myaccount.google.com/apppasswords" target="_blank" class="help-link">(Jak wygenerować?)</a>
                            <a v-else-if="accountForm.provider === 'outlook'" href="https://account.live.com/proofs/manage/additional" target="_blank" class="help-link">(Jak wygenerować?)</a>
                        </label>
                        <input v-model="accountForm.password" type="password" placeholder="Wklej hasło aplikacji..." />
                    </div>

                    <div v-if="accountForm.provider === 'custom' || showCustomServers" class="custom-server-fields">
                        <div class="form-row">
                            <div class="form-group flex-1">
                                <label>IMAP Host</label>
                                <input v-model="accountForm.imapHost" placeholder="imap.example.com" />
                            </div>
                            <div class="form-group w-100">
                                <label>IMAP Port</label>
                                <input v-model.number="accountForm.imapPort" type="number" />
                            </div>
                        </div>
                        <div class="form-row">
                            <div class="form-group flex-1">
                                <label>SMTP Host</label>
                                <input v-model="accountForm.smtpHost" placeholder="smtp.example.com" />
                            </div>
                            <div class="form-group w-100">
                                <label>SMTP Port</label>
                                <input v-model.number="accountForm.smtpPort" type="number" />
                            </div>
                        </div>
                    </div>

                    <button v-if="accountForm.provider !== 'custom' && !showCustomServers" class="btn-link" @click="showCustomServers = true">
                        Pokaż ustawienia serwera
                    </button>

                    <p v-if="formError" class="err-msg">{{ formError }}</p>

                    <div class="settings-actions">
                        <button class="btn-primary" @click="saveEmailAccount" :disabled="accountSaving">
                            <span class="material-icons">save</span>
                            {{ accountSaving ? 'Zapisywanie...' : (emailAccount ? 'Zaktualizuj' : 'Połącz') }}
                        </button>
                        <button v-if="emailAccount" class="btn-secondary" @click="testConnection" :disabled="testingConnection">
                            <span class="material-icons">wifi_tethering</span>
                            {{ testingConnection ? 'Testowanie...' : 'Testuj połączenie' }}
                        </button>
                        <button v-if="emailAccount" class="btn-danger-outline" @click="disconnectAccount">
                            <span class="material-icons">link_off</span> Rozłącz
                        </button>
                    </div>

                    <div v-if="testResult" class="test-result">
                        <div :class="['test-item', testResult.imap_status === 'ok' ? 'test-ok' : testResult.imap_status === 'skip' ? 'test-skip' : 'test-fail']">
                            <span class="material-icons">{{ testResult.imap_status === 'ok' ? 'check_circle' : testResult.imap_status === 'skip' ? 'info' : 'error' }}</span>
                            IMAP (odbiór poczty): {{ testResult.imap_status === 'ok' ? 'Połączono' : testResult.imap_status === 'skip' ? 'Niedostępny (Mailpit)' : 'Błąd' }}
                        </div>
                        <div :class="['test-item', testResult.smtp_status === 'ok' ? 'test-ok' : 'test-fail']">
                            <span class="material-icons">{{ testResult.smtp_status === 'ok' ? 'check_circle' : 'error' }}</span>
                            SMTP (wysyłanie poczty): {{ testResult.smtp_status === 'ok' ? 'Połączono' : 'Błąd' }}
                        </div>
                        <div v-if="testResult.error" class="test-error">{{ testResult.error }}</div>
                    </div>
                </div>

                <div class="settings-info-card">
                    <span class="material-icons info-icon">info</span>
                    <div>
                        <h4>Jak to działa?</h4>
                        <ul>
                            <li>BizmoPol łączy się z Twoim kontem e-mail przez protokoły <strong>IMAP</strong> (odbiór) i <strong>SMTP</strong> (wysyłanie).</li>
                            <li>Dla <strong>Gmail</strong>: włącz IMAP w ustawieniach i wygeneruj <em>hasło aplikacji</em> (wymaga weryfikacji dwuetapowej).</li>
                            <li>Dla <strong>Outlook</strong>: wygeneruj hasło aplikacji w ustawieniach bezpieczeństwa konta Microsoft.</li>
                            <li>Kliknij <strong>"Synchronizuj"</strong> w nagłówku, aby pobrać nowe wiadomości ze skrzynki.</li>
                            <li>Wiadomości wysyłane z CRM trafią bezpośrednio na wskazany adres e-mail.</li>
                        </ul>
                    </div>
                </div>
            </template>
        </div>

        <div v-if="showComposeModal" class="modal-overlay" @click.self="showComposeModal = false">
            <div class="modal modal-lg">
                <div class="modal-header">
                    <h2>Nowa wiadomość</h2>
                    <button class="modal-close" @click="showComposeModal = false">
                        <span class="material-icons">close</span>
                    </button>
                </div>

                <div class="compose-quick-actions">
                    <div class="qaction-group">
                        <span class="qaction-label">Szablon:</span>
                        <select class="quick-sel" @change="e => { const t = templates.find(t => t.id === e.target.value); if(t) applyTemplateToCompose(t); e.target.value = '' }">
                            <option value="">Wybierz szablon...</option>
                            <option v-for="t in templates" :key="t.id" :value="t.id">{{ t.name }}</option>
                        </select>
                    </div>
                    <div class="qaction-group">
                        <span class="qaction-label">Podpis:</span>
                        <select class="quick-sel" @change="e => { const s = signatures.find(s => s.id === e.target.value); if(s) applySignatureToCompose(s); e.target.value = '' }">
                            <option value="">Dodaj podpis...</option>
                            <option v-for="s in signatures" :key="s.id" :value="s.id">{{ s.name }}</option>
                        </select>
                    </div>
                </div>

                <div class="form-group">
                    <label>Do *</label>
                    <input v-model="composeForm.to" type="email" placeholder="kontakt@firma.pl" />
                </div>
                <div class="form-group">
                    <label>CC</label>
                    <input v-model="composeForm.cc" placeholder="cc1@email.com, cc2@email.com" />
                </div>
                <div class="form-group">
                    <label>Temat *</label>
                    <input v-model="composeForm.subject" placeholder="Temat wiadomości" />
                </div>
                <div class="form-group">
                    <label>Treść *</label>
                    <textarea v-model="composeForm.body" rows="12" placeholder="Treść wiadomości..."></textarea>
                </div>
                <p v-if="formError" class="err-msg">{{ formError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showComposeModal = false">Anuluj</button>
                    <button class="btn-primary" @click="sendCompose">
                        <span class="material-icons">send</span> Wyślij
                    </button>
                </div>
            </div>
        </div>

        <div v-if="showBulkModal" class="modal-overlay" @click.self="showBulkModal = false">
            <div class="modal modal-xl">
                <div class="modal-header">
                    <h2>Nowa kampania masowa</h2>
                    <button class="modal-close" @click="showBulkModal = false">
                        <span class="material-icons">close</span>
                    </button>
                </div>

                <div class="bulk-modal-layout">
                    <div class="bulk-modal-left">
                        <div class="form-group">
                            <label>Nazwa kampanii *</label>
                            <input v-model="bulkForm.name" placeholder="np. Kampania Majowa 2025" />
                        </div>
                        <div class="form-group">
                            <label>Szablon (opcjonalnie)</label>
                            <select v-model="bulkForm.templateId" @change="() => { const t = templates.find(t => t.id === bulkForm.templateId); if(t) applyTemplateToBulk(t) }">
                                <option value="">Wybierz szablon...</option>
                                <option v-for="t in templates" :key="t.id" :value="t.id">{{ t.name }}</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label>Temat *</label>
                            <input v-model="bulkForm.subject" placeholder="Temat wiadomości" />
                        </div>
                        <div class="form-group">
                            <label>Treść *</label>
                            <textarea v-model="bulkForm.body" rows="10" placeholder="Treść wiadomości...&#10;&#10;Możesz użyć zmiennych: {{name}}, {{company}}"></textarea>
                        </div>
                        <div class="bulk-selected-info">
                            <span class="material-icons">group</span>
                            Wybrano: <strong>{{ bulkForm.contactIds.length }}</strong> kontaktów
                        </div>
                    </div>
                    <div class="bulk-modal-right">
                        <div class="bulk-contact-header">
                            <h3>Wybierz odbiorców</h3>
                            <div class="bulk-filter">
                                <select v-model="bulkForm.filterStatus" @change="loadBulkContacts">
                                    <option value="">Wszystkie statusy</option>
                                    <option value="lead">Lead</option>
                                    <option value="prospect">Prospect</option>
                                    <option value="customer">Klient</option>
                                    <option value="inactive">Nieaktywny</option>
                                </select>
                            </div>
                            <div class="bulk-select-actions">
                                <button class="btn-xs" @click="selectAllBulkContacts">Zaznacz wszystkich</button>
                                <button class="btn-xs" @click="clearBulkContacts">Odznacz</button>
                            </div>
                        </div>
                        <div class="bulk-contact-list">
                            <div v-for="c in bulkContacts" :key="c.id" :class="['bulk-contact-item', { selected: bulkForm.contactIds.includes(c.id), 'no-email': !c.email }]" @click="c.email && toggleBulkContact(c.id)">
                                <input type="checkbox" :checked="bulkForm.contactIds.includes(c.id)" :disabled="!c.email" @change="c.email && toggleBulkContact(c.id)" @click.stop />
                                <div class="bci-avatar" :style="{ background: avatarColor(c.name) }">{{ initials(c.name) }}</div>
                                <div class="bci-info">
                                    <div class="bci-name">{{ c.name }}</div>
                                    <div class="bci-email">{{ c.email || "Brak emaila" }}</div>
                                </div>
                                <span v-if="!c.email" class="no-email-badge">brak e-mail</span>
                            </div>
                        </div>
                    </div>
                </div>

                <p v-if="formError" class="err-msg">{{ formError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showBulkModal = false">Anuluj</button>
                    <button class="btn-primary" @click="sendBulk">
                        <span class="material-icons">campaign</span> Uruchom kampanię ({{ bulkForm.contactIds.length }})
                    </button>
                </div>
            </div>
        </div>

        <div v-if="showTemplateModal" class="modal-overlay" @click.self="showTemplateModal = false">
            <div class="modal">
                <h2>{{ editingTemplateId ? "Edytuj szablon" : "Nowy szablon" }}</h2>
                <div class="form-group">
                    <label>Nazwa *</label>
                    <input v-model="templateForm.name" placeholder="Nazwa szablonu" />
                </div>
                <div class="form-group">
                    <label>Kategoria</label>
                    <input v-model="templateForm.category" placeholder="np. Onboarding, Sales, Marketing" />
                </div>
                <div class="form-group">
                    <label>Temat *</label>
                    <input v-model="templateForm.subject" placeholder="Temat wiadomości" />
                </div>
                <div class="form-group">
                    <label>Treść * (możesz użyć: {{name}}, {{company}})</label>
                    <textarea v-model="templateForm.body" rows="8" placeholder="Treść szablonu..."></textarea>
                </div>
                <p v-if="formError" class="err-msg">{{ formError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showTemplateModal = false">Anuluj</button>
                    <button class="btn-primary" @click="saveTemplate">{{ editingTemplateId ? "Zapisz" : "Dodaj" }}</button>
                </div>
            </div>
        </div>

        <div v-if="showSignatureModal" class="modal-overlay" @click.self="showSignatureModal = false">
            <div class="modal">
                <h2>{{ editingSignatureId ? "Edytuj podpis" : "Nowy podpis" }}</h2>
                <div class="form-group">
                    <label>Nazwa *</label>
                    <input v-model="signatureForm.name" placeholder="np. Podpis główny, Krótki podpis" />
                </div>
                <div class="form-group">
                    <label>Treść podpisu *</label>
                    <textarea v-model="signatureForm.body" rows="5" placeholder="Z poważaniem,&#10;Imię Nazwisko&#10;Firma&#10;Tel: ..."></textarea>
                </div>
                <div class="form-group">
                    <label class="toggle-label">
                        <input type="checkbox" v-model="signatureForm.isDefault" class="toggle-input" />
                        <span class="toggle-track"><span class="toggle-thumb"></span></span>
                        <span class="toggle-text">Podpis domyślny</span>
                    </label>
                </div>
                <p v-if="formError" class="err-msg">{{ formError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showSignatureModal = false">Anuluj</button>
                    <button class="btn-primary" @click="saveSignature">{{ editingSignatureId ? "Zapisz" : "Dodaj" }}</button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.comm-wrapper { width: 100%; max-width: 1400px; margin: 0 auto; }
.comm-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 20px; }
.comm-header h1 { font-size: 2rem; color: #f8fafc; margin-bottom: 4px; }
.subtitle { color: #94a3b8; font-size: 0.95rem; }
.header-actions { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }

.header-actions .material-icons { font-size: 1.1rem; }
.btn-primary, .btn-secondary, .btn-sync, .btn-bulk {
    padding: 8px 16px; border-radius: 8px; cursor: pointer; transition: 0.2s;
    display: flex; align-items: center; gap: 6px; font-size: 0.85rem;
    height: 38px; box-sizing: border-box; white-space: nowrap;
}
.btn-primary {
    background: #38bdf8; border: none; color: #0f172a; font-weight: bold;
}
.btn-primary:hover { background: #7dd3fc; }
.btn-secondary {
    background: transparent; border: 1px solid #475569; color: #94a3b8;
}
.btn-secondary:hover { border-color: #f1f5f9; color: #f1f5f9; }
.btn-sync {
    background: #1e293b; border: 1px solid #334155; color: #94a3b8;
}
.btn-sync:hover { border-color: #38bdf8; color: #38bdf8; }
.sync-group { position: relative; display: flex; flex-direction: column; align-items: flex-end; gap: 2px; }
.last-sync-label {
    position: absolute; top: 100%; left: 0; font-size: 0.7rem; color: #64748b;
    display: flex; align-items: center; gap: 3px; white-space: nowrap;
}
.btn-bulk {
    background: #7c3aed; border: none; color: #ede9fe; font-weight: bold;
}
.btn-bulk:hover { background: #6d28d9; }
.spinning { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.queue-bar { background: #1e293b; border: 1px solid #334155; border-radius: 10px; padding: 12px 16px; margin-bottom: 16px; }
.queue-bar-inner { display: flex; align-items: center; gap: 8px; color: #94a3b8; font-size: 0.88rem; margin-bottom: 8px; }
.queue-icon { font-size: 1rem; color: #38bdf8; }
.queue-progress { height: 4px; background: #334155; border-radius: 2px; overflow: hidden; }
.queue-progress-fill { height: 100%; background: linear-gradient(90deg, #38bdf8, #818cf8); border-radius: 2px; animation: progress-pulse 1.5s ease-in-out infinite; width: 40%; }
@keyframes progress-pulse { 0%,100% { transform: translateX(-50%); } 50% { transform: translateX(200%); } }

.stats-row { display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 14px; margin-bottom: 24px; }
.stat-card {
    background: #1e293b; border: 1px solid #334155; border-radius: 12px;
    padding: 16px; display: flex; align-items: center; gap: 14px;
}
.stat-icon { font-size: 2rem; }
.stat-value { font-size: 1.8rem; font-weight: 800; color: #f1f5f9; line-height: 1; }
.stat-label { font-size: 0.78rem; color: #64748b; margin-top: 2px; }

.comm-tabs { display: flex; gap: 2px; margin-bottom: 20px; border-bottom: 1px solid #334155; }
.tab-btn {
    padding: 10px 18px; background: transparent; border: none; border-bottom: 2px solid transparent;
    color: #94a3b8; cursor: pointer; font-size: 0.9rem; transition: 0.2s;
    display: flex; align-items: center; gap: 6px; margin-bottom: -1px;
    position: relative;
}
.tab-btn .material-icons { font-size: 1rem; }
.tab-btn:hover { color: #f1f5f9; }
.tab-btn.active { color: #38bdf8; border-bottom-color: #38bdf8; }
.unread-badge {
    background: #ef4444; color: white; font-size: 0.65rem; font-weight: 700;
    padding: 2px 6px; border-radius: 10px; min-width: 18px; text-align: center;
}

.inbox-layout { display: flex; gap: 0; height: calc(100vh - 340px); min-height: 500px; border: 1px solid #334155; border-radius: 12px; overflow: hidden; }
.inbox-sidebar { width: 340px; min-width: 340px; border-right: 1px solid #334155; display: flex; flex-direction: column; background: #0f172a; }
.inbox-filters { padding: 12px; display: flex; flex-direction: column; gap: 8px; border-bottom: 1px solid #334155; }
.search-input {
    padding: 8px 12px; background: #1e293b; border: 1px solid #334155; border-radius: 8px;
    color: #f1f5f9; outline: none; font-size: 0.9rem; width: 100%;
}
.search-input:focus { border-color: #38bdf8; }
.filter-sel {
    padding: 7px 10px; background: #1e293b; border: 1px solid #334155; border-radius: 8px;
    color: #f1f5f9; outline: none; cursor: pointer; font-size: 0.85rem; width: 100%;
}
.thread-list { flex: 1; overflow-y: auto; }
.thread-item {
    padding: 12px 14px; border-bottom: 1px solid #1e293b; cursor: pointer;
    transition: 0.15s; display: flex; gap: 10px; align-items: flex-start;
}
.thread-item:hover { background: #1e293b; }
.thread-item.active { background: #1e293b; border-left: 3px solid #38bdf8; }
.thread-item.unread .thread-subject { font-weight: 700; color: #f8fafc; }
.thread-item.unread .thread-name { font-weight: 700; }
.thread-avatar {
    width: 38px; height: 38px; border-radius: 50%; display: flex; align-items: center;
    justify-content: center; font-weight: 700; font-size: 0.85rem; color: #0f172a; flex-shrink: 0;
}
.thread-info { flex: 1; min-width: 0; }
.thread-row1 { display: flex; justify-content: space-between; align-items: center; margin-bottom: 3px; }
.thread-name { color: #e2e8f0; font-size: 0.9rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 170px; }
.thread-date { color: #64748b; font-size: 0.75rem; flex-shrink: 0; }
.thread-row2 { display: flex; align-items: center; gap: 4px; margin-bottom: 4px; }
.dir-icon { font-size: 0.8rem; color: #64748b; }
.thread-subject { color: #94a3b8; font-size: 0.85rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; flex: 1; }
.unread-dot {
    background: #38bdf8; color: #0f172a; font-size: 0.65rem; font-weight: 700;
    padding: 1px 6px; border-radius: 10px; flex-shrink: 0;
}
.thread-row3 { display: flex; justify-content: space-between; align-items: center; }
.thread-status { font-size: 0.72rem; font-weight: 600; padding: 1px 8px; border-radius: 10px; }
.status-open { background: #064e3b22; color: #34d399; border: 1px solid #064e3b; }
.status-archived { background: #37415122; color: #9ca3af; border: 1px solid #374151; }
.status-closed { background: #1e3a5f22; color: #38bdf8; border: 1px solid #1e3a5f; }
.thread-count { font-size: 0.72rem; color: #64748b; }
.pagination-sm { padding: 10px; display: flex; align-items: center; justify-content: center; gap: 10px; border-top: 1px solid #334155; color: #94a3b8; font-size: 0.85rem; }
.page-btn-sm {
    background: #1e293b; border: 1px solid #334155; color: #94a3b8;
    width: 28px; height: 28px; border-radius: 6px; cursor: pointer;
    display: flex; align-items: center; justify-content: center;
}
.page-btn-sm:hover:not(:disabled) { color: #38bdf8; border-color: #38bdf8; }
.page-btn-sm:disabled { opacity: 0.4; cursor: default; }
.loading-sm { padding: 20px; text-align: center; color: #94a3b8; font-size: 0.9rem; }
.empty-sm { padding: 30px; text-align: center; color: #64748b; font-size: 0.9rem; }

.inbox-main { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.empty-thread { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #475569; gap: 12px; }
.empty-icon { font-size: 3rem; }
.loading { color: #94a3b8; padding: 40px; text-align: center; }

.thread-view { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.thread-header {
    padding: 16px 20px; border-bottom: 1px solid #334155;
    display: flex; justify-content: space-between; align-items: center;
    background: #1e293b; flex-shrink: 0;
}
.thread-header-left { display: flex; align-items: center; gap: 14px; }
.thread-avatar-lg {
    width: 46px; height: 46px; border-radius: 50%; display: flex; align-items: center;
    justify-content: center; font-weight: 700; font-size: 1rem; color: #0f172a; flex-shrink: 0;
}
.thread-header h2 { color: #f1f5f9; font-size: 1.05rem; margin-bottom: 3px; }
.thread-meta { color: #94a3b8; font-size: 0.82rem; }
.thread-header-actions { display: flex; gap: 6px; }
.btn-action-sm {
    background: none; border: none; cursor: pointer; padding: 6px 8px; border-radius: 6px;
    transition: 0.15s; display: flex; align-items: center; color: #94a3b8;
}
.btn-action-sm .material-icons { font-size: 1.1rem; }
.btn-action-sm:hover { background: #334155; color: #f1f5f9; }
.btn-action-sm.btn-danger:hover { background: #3f1414; color: #f87171; }

.messages-list { flex: 1; overflow-y: auto; padding: 20px; display: flex; flex-direction: column; gap: 16px; }
.message-item { display: flex; gap: 12px; }
.message-item.outbound { flex-direction: row-reverse; }
.msg-avatar {
    width: 36px; height: 36px; border-radius: 50%; display: flex; align-items: center;
    justify-content: center; font-weight: 700; font-size: 0.75rem; color: #0f172a; flex-shrink: 0;
}
.msg-body {
    max-width: 75%; background: #1e293b; border: 1px solid #334155; border-radius: 12px;
    padding: 14px; flex: 1;
}
.message-item.outbound .msg-body { background: #0f2a4a; border-color: #1e3a5f; }
.msg-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 4px; }
.msg-from { font-size: 0.82rem; font-weight: 600; color: #38bdf8; }
.msg-actions { display: flex; align-items: center; gap: 8px; }
.btn-star { background: none; border: none; cursor: pointer; color: #475569; padding: 0; }
.btn-star:hover, .btn-star.starred { color: #f59e0b; }
.btn-star .material-icons { font-size: 1rem; }
.msg-date { font-size: 0.75rem; color: #64748b; }
.msg-to { font-size: 0.78rem; color: #64748b; margin-bottom: 8px; }
.msg-text { color: #e2e8f0; font-size: 0.9rem; line-height: 1.6; white-space: pre-wrap; }

.reply-form { padding: 16px 20px; border-top: 1px solid #334155; background: #1e293b; flex-shrink: 0; }
.reply-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.reply-header h3 { color: #38bdf8; font-size: 1rem; }
.reply-quick-actions { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.quick-label { color: #64748b; font-size: 0.8rem; }
.quick-sel {
    padding: 5px 8px; background: #0f172a; border: 1px solid #334155; border-radius: 6px;
    color: #f1f5f9; font-size: 0.82rem; outline: none; cursor: pointer;
}
.reply-actions { display: flex; gap: 10px; justify-content: flex-end; margin-top: 12px; }
.form-group { margin-bottom: 14px; }
.form-group label { display: block; margin-bottom: 6px; color: #94a3b8; font-size: 0.85rem; }
.form-group input, .form-group select, .form-group textarea {
    width: 100%; padding: 10px 12px; background: #0f172a; border: 1px solid #334155;
    border-radius: 8px; color: #f1f5f9; outline: none; font-size: 0.9rem; resize: vertical;
}
.form-group input:focus, .form-group select:focus, .form-group textarea:focus { border-color: #38bdf8; }
.err-msg { color: #f87171; font-size: 0.88rem; margin-bottom: 10px; }
.empty { color: #64748b; padding: 40px; text-align: center; }

.bulk-section, .templates-section, .signatures-section, .recent-section { padding: 4px 0; }
.section-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.section-header h2 { color: #f8fafc; font-size: 1.3rem; }

.bulk-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 16px; }
.bulk-card { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 20px; }
.bulk-card-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 14px; }
.bulk-name { color: #f1f5f9; font-weight: 600; margin-bottom: 4px; }
.bulk-subject { color: #94a3b8; font-size: 0.85rem; }
.job-badge { padding: 4px 10px; border-radius: 20px; font-size: 0.75rem; font-weight: 600; white-space: nowrap; }
.job-pending { background: #37415122; color: #9ca3af; border: 1px solid #374151; }
.job-running { background: #1d4ed822; color: #60a5fa; border: 1px solid #1d4ed8; }
.job-completed { background: #064e3b22; color: #34d399; border: 1px solid #064e3b; }
.job-failed { background: #7f1d1d22; color: #f87171; border: 1px solid #7f1d1d; }
.bulk-progress { margin-bottom: 12px; }
.progress-bar-wrap { height: 8px; background: #334155; border-radius: 4px; overflow: hidden; margin-bottom: 6px; }
.progress-bar { height: 100%; background: #38bdf8; border-radius: 4px; transition: width 0.8s ease-in-out; }
.progress-bar-animated {
    background: linear-gradient(90deg, #38bdf8 0%, #818cf8 50%, #38bdf8 100%);
    background-size: 200% 100%;
    animation: shimmer 2s linear infinite;
}
@keyframes shimmer { 0% { background-position: 200% 0; } 100% { background-position: -200% 0; } }
.prog-pct { margin-left: auto; color: #38bdf8; font-weight: 600; }
.progress-stats { display: flex; gap: 10px; font-size: 0.8rem; }
.prog-sent { color: #34d399; }
.prog-fail { color: #f87171; }
.prog-total { color: #64748b; }
.bulk-dates { display: flex; flex-direction: column; gap: 2px; color: #64748b; font-size: 0.78rem; }

.templates-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 14px; }
.template-card { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 18px; }
.template-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 8px; }
.template-name { color: #f1f5f9; font-weight: 600; }
.template-cat { color: #94a3b8; font-size: 0.78rem; margin-top: 2px; }
.template-usage { display: flex; align-items: center; gap: 4px; color: #64748b; font-size: 0.8rem; }
.usage-icon { font-size: 0.9rem; }
.template-subject { color: #38bdf8; font-size: 0.85rem; margin-bottom: 8px; font-weight: 500; }
.template-body { color: #94a3b8; font-size: 0.82rem; line-height: 1.5; margin-bottom: 12px; }
.template-actions { display: flex; gap: 8px; align-items: center; }
.btn-use {
    padding: 6px 14px; background: #0f172a; border: 1px solid #334155; border-radius: 6px;
    color: #38bdf8; cursor: pointer; transition: 0.15s; font-size: 0.82rem;
    display: flex; align-items: center; gap: 4px;
}
.btn-use:hover { background: #1e3a5f; border-color: #38bdf8; }
.btn-use .material-icons { font-size: 0.9rem; }

.signatures-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 14px; }
.signature-card { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 18px; }
.sig-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.sig-name { color: #f1f5f9; font-weight: 600; }
.default-badge { background: #38bdf822; color: #38bdf8; border: 1px solid #38bdf8; padding: 2px 8px; border-radius: 10px; font-size: 0.72rem; font-weight: 600; }
.sig-body { color: #94a3b8; font-size: 0.82rem; white-space: pre-wrap; margin-bottom: 12px; line-height: 1.5; }
.sig-actions { display: flex; gap: 8px; }

.recent-grid { display: flex; flex-direction: column; gap: 10px; }
.recent-card {
    background: #1e293b; border: 1px solid #334155; border-radius: 10px;
    padding: 16px; display: flex; align-items: center; gap: 14px;
}
.recent-avatar {
    width: 44px; height: 44px; border-radius: 50%; display: flex; align-items: center;
    justify-content: center; font-weight: 700; font-size: 0.9rem; color: #0f172a; flex-shrink: 0;
}
.recent-info { flex: 1; }
.recent-name { color: #f1f5f9; font-weight: 600; margin-bottom: 2px; }
.recent-email { color: #94a3b8; font-size: 0.85rem; margin-bottom: 6px; }
.recent-meta { display: flex; gap: 16px; font-size: 0.78rem; color: #64748b; }
.btn-compose-quick {
    background: #1e3a5f; border: 1px solid #1e3a5f; color: #38bdf8; border-radius: 8px;
    padding: 8px 12px; cursor: pointer; transition: 0.15s; display: flex; align-items: center;
}
.btn-compose-quick:hover { background: #38bdf8; color: #0f172a; }
.btn-compose-quick .material-icons { font-size: 1.1rem; }

.modal-overlay {
    position: fixed; inset: 0; background: rgba(0,0,0,0.65);
    display: flex; align-items: center; justify-content: center; z-index: 100;
}
.modal {
    background: #1e293b; border: 1px solid #334155; border-radius: 16px;
    padding: 28px; width: 100%; max-width: 520px; max-height: 90vh; overflow-y: auto;
}
.modal-lg { max-width: 680px; }
.modal-xl { max-width: 1020px; max-height: 92vh; }
.modal h2 { color: #38bdf8; margin-bottom: 20px; font-size: 1.3rem; }
.modal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.modal-header h2 { margin-bottom: 0; }
.modal-close { background: none; border: none; color: #94a3b8; cursor: pointer; padding: 4px; }
.modal-close .material-icons { font-size: 1.3rem; }
.modal-actions { display: flex; gap: 10px; justify-content: flex-end; margin-top: 20px; }

.compose-quick-actions { display: flex; gap: 16px; margin-bottom: 16px; padding: 12px; background: #0f172a; border-radius: 8px; border: 1px solid #334155; flex-wrap: wrap; }
.qaction-group { display: flex; align-items: center; gap: 8px; }
.qaction-label { color: #64748b; font-size: 0.82rem; white-space: nowrap; }

.bulk-modal-layout { display: grid; grid-template-columns: 1fr 380px; gap: 20px; margin-bottom: 16px; }
.bulk-modal-right { display: flex; flex-direction: column; border: 1px solid #334155; border-radius: 10px; overflow: hidden; max-height: 500px; }
.bulk-contact-header { padding: 12px; border-bottom: 1px solid #334155; background: #0f172a; }
.bulk-contact-header h3 { color: #f1f5f9; font-size: 0.95rem; margin-bottom: 8px; }
.bulk-filter select { width: 100%; padding: 7px; background: #1e293b; border: 1px solid #334155; border-radius: 6px; color: #f1f5f9; outline: none; margin-bottom: 8px; font-size: 0.85rem; }
.bulk-select-actions { display: flex; gap: 8px; }
.btn-xs { padding: 4px 10px; font-size: 0.78rem; background: #1e293b; border: 1px solid #334155; border-radius: 6px; color: #94a3b8; cursor: pointer; }
.btn-xs:hover { color: #38bdf8; border-color: #38bdf8; }
.bulk-contact-list { flex: 1; overflow-y: auto; }
.bulk-contact-item {
    padding: 10px 12px; display: flex; align-items: center; gap: 10px;
    cursor: pointer; transition: 0.15s; border-bottom: 1px solid #1e293b;
}
.bulk-contact-item:hover { background: #1e293b; }
.bulk-contact-item.selected { background: #0f2a4a; }
.bulk-contact-item.no-email { opacity: 0.5; cursor: not-allowed; }
.bulk-contact-item input[type="checkbox"] { accent-color: #38bdf8; flex-shrink: 0; }
.bci-avatar { width: 30px; height: 30px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 0.7rem; font-weight: 700; color: #0f172a; flex-shrink: 0; }
.bci-info { flex: 1; min-width: 0; }
.bci-name { color: #e2e8f0; font-size: 0.85rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.bci-email { color: #64748b; font-size: 0.78rem; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.no-email-badge { background: #7f1d1d22; color: #f87171; font-size: 0.7rem; padding: 1px 6px; border-radius: 6px; white-space: nowrap; }
.bulk-selected-info { display: flex; align-items: center; gap: 8px; color: #94a3b8; font-size: 0.9rem; padding: 10px; background: #0f172a; border-radius: 8px; margin-top: 8px; }
.bulk-selected-info .material-icons { font-size: 1.1rem; color: #38bdf8; }
.bulk-selected-info strong { color: #38bdf8; }

.toggle-label { display: flex; align-items: center; gap: 10px; cursor: pointer; }
.toggle-input { display: none; }
.toggle-track { width: 44px; height: 24px; background: #334155; border-radius: 12px; position: relative; transition: 0.2s; flex-shrink: 0; }
.toggle-input:checked + .toggle-track { background: #38bdf8; }
.toggle-thumb { position: absolute; top: 3px; left: 3px; width: 18px; height: 18px; background: white; border-radius: 50%; transition: 0.2s; }
.toggle-input:checked + .toggle-track .toggle-thumb { left: 23px; }
.toggle-text { color: #e2e8f0; font-size: 0.9rem; }

.connected-badge {
    display: flex; align-items: center; gap: 4px;
    color: #34d399; font-size: 0.6rem; padding: 10px 18px;
    background: #064e3b22; border: 1px solid #064e3b; border-radius: 8px;
}
.setup-dot {
    width: 8px; height: 8px; border-radius: 50%; background: #f59e0b;
    display: inline-block; margin-left: 4px;
}

.settings-section { padding: 4px 0; max-width: 800px; }
.settings-section h2 { color: #f8fafc; font-size: 1.3rem; margin-bottom: 6px; }
.settings-desc { color: #94a3b8; font-size: 0.92rem; margin-bottom: 24px; }

.account-status-card {
    background: #1e293b; border: 1px solid #334155; border-radius: 12px;
    padding: 20px; margin-bottom: 20px;
}
.account-status-header { display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 16px; }
.account-status-left { display: flex; align-items: center; gap: 14px; }
.account-status-icon { font-size: 2rem; }
.account-status-icon.connected { color: #34d399; }
.account-provider { color: #f1f5f9; font-weight: 700; font-size: 1.1rem; }
.account-email { color: #94a3b8; font-size: 0.9rem; }
.account-status-right { text-align: right; }
.account-meta { color: #64748b; font-size: 0.82rem; }

.settings-form-card {
    background: #1e293b; border: 1px solid #334155; border-radius: 12px;
    padding: 24px; margin-bottom: 20px;
}
.settings-form-card h3 { color: #38bdf8; font-size: 1.1rem; margin-bottom: 20px; }

.provider-grid {
    display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
    gap: 10px; margin-bottom: 20px;
}
.provider-card {
    background: #0f172a; border: 2px solid #334155; border-radius: 10px;
    padding: 16px; cursor: pointer; transition: 0.2s;
    display: flex; flex-direction: column; align-items: center; gap: 8px; text-align: center;
}
.provider-card:hover { border-color: #475569; }
.provider-card.selected { border-color: #38bdf8; background: #0f2a4a; }
.provider-icon { font-size: 1.8rem; color: #94a3b8; }
.provider-card.selected .provider-icon { color: #38bdf8; }
.provider-label { color: #e2e8f0; font-size: 0.88rem; font-weight: 500; }

.custom-server-fields { margin-top: 16px; padding-top: 16px; border-top: 1px solid #334155; }
.form-row { display: flex; gap: 12px; }
.flex-1 { flex: 1; }
.w-100 { width: 100px; }

.settings-actions { display: flex; gap: 10px; margin-top: 20px; flex-wrap: wrap; }

.btn-danger-outline {
    padding: 10px 18px; background: transparent; border: 1px solid #7f1d1d;
    border-radius: 8px; color: #f87171; cursor: pointer; transition: 0.2s;
    font-size: 0.9rem; display: flex; align-items: center; gap: 6px;
}
.btn-danger-outline:hover { background: #3f1414; border-color: #f87171; }
.btn-danger-outline .material-icons { font-size: 1rem; }

.btn-link {
    background: none; border: none; color: #38bdf8; cursor: pointer;
    font-size: 0.85rem; padding: 4px 0; margin-top: 4px;
}
.btn-link:hover { text-decoration: underline; }

.help-link { color: #38bdf8; font-size: 0.78rem; text-decoration: none; margin-left: 4px; }
.help-link:hover { text-decoration: underline; }

.test-result {
    margin-top: 16px; padding: 16px; background: #0f172a;
    border: 1px solid #334155; border-radius: 10px;
    display: flex; flex-direction: column; gap: 8px;
}
.test-item {
    display: flex; align-items: center; gap: 8px; font-size: 0.9rem; font-weight: 500;
}
.test-item .material-icons { font-size: 1.1rem; }
.test-ok { color: #34d399; }
.test-fail { color: #f87171; }
.test-skip { color: #94a3b8; }
.test-error { color: #f87171; font-size: 0.82rem; margin-top: 6px; padding-top: 8px; border-top: 1px solid #334155; }

.settings-info-card {
    background: #1e293b; border: 1px solid #334155; border-radius: 12px;
    padding: 20px; display: flex; gap: 16px; align-items: flex-start;
}
.info-icon { color: #38bdf8; font-size: 1.5rem; flex-shrink: 0; margin-top: 2px; }
.settings-info-card h4 { color: #f1f5f9; margin-bottom: 10px; font-size: 0.95rem; }
.settings-info-card ul { color: #94a3b8; font-size: 0.85rem; padding-left: 18px; line-height: 1.8; }
.settings-info-card strong { color: #e2e8f0; }
.settings-info-card em { color: #f59e0b; font-style: normal; }
</style>
