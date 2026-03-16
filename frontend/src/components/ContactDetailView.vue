<script setup>
import { ref, onMounted } from "vue"
import { useRoute, useRouter } from "vue-router"

const props = defineProps({
    authFetch: { type: Function, required: true }
})

const route = useRoute()
const router = useRouter()

const contact = ref(null)
const loading = ref(false)
const error = ref("")

onMounted(async () => {
    loading.value = true
    try {
        const res = await props.authFetch("/api/v1/contacts/" + route.params.id)
        if (!res.ok) throw new Error("Błąd pobierania kontaktu")
        contact.value = await res.json()
    } catch (e) {
        error.value = e.message
    } finally {
        loading.value = false
    }
})

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
</style>
