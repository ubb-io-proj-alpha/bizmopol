<script setup>
import { onMounted, ref } from "vue"
import { useRouter } from "vue-router"

import PipelineBoard from "@/components/PipelineBoard.vue"
import { clearToken, pipelineAPI } from "@/services/api"

const router = useRouter()

const pipelines = ref([])
const selectedPipelineId = ref(null)
const isLoadingPipelines = ref(false)
const isCreatingPipeline = ref(false)
const showCreatePipelineModal = ref(false)
const newPipelineName = ref("")
const loadError = ref("")

const redirectToLogin = () => {
    clearToken()
    router.push({ name: "login" })
}

const syncSelectedPipeline = (items) => {
    if (!items.length) {
        selectedPipelineId.value = null
        return
    }

    const selectedStillExists = items.some((pipeline) => pipeline.id === selectedPipelineId.value)
    if (!selectedStillExists) {
        selectedPipelineId.value = items[0].id
    }
}

const loadPipelines = async () => {
    isLoadingPipelines.value = true
    loadError.value = ""

    try {
        const data = await pipelineAPI.getPipelines()
        pipelines.value = data
        syncSelectedPipeline(data)
    } catch (error) {
        if (error.status === 401) {
            redirectToLogin()
            return
        }

        loadError.value = error.data?.error || "Nie udało się pobrać lejków."
    } finally {
        isLoadingPipelines.value = false
    }
}

const openCreatePipelineModal = () => {
    newPipelineName.value = ""
    showCreatePipelineModal.value = true
}

const createPipeline = async () => {
    const pipelineName = newPipelineName.value.trim()
    if (!pipelineName) {
        alert("Wpisz nazwę lejka.")
        return
    }

    isCreatingPipeline.value = true

    try {
        const pipeline = await pipelineAPI.createPipeline(pipelineName)
        await loadPipelines()
        selectedPipelineId.value = pipeline.id
        showCreatePipelineModal.value = false
        newPipelineName.value = ""
    } catch (error) {
        if (error.status === 401) {
            redirectToLogin()
            return
        }

        alert(error.data?.error || "Nie udało się utworzyć lejka.")
    } finally {
        isCreatingPipeline.value = false
    }
}

onMounted(async () => {
    await loadPipelines()
})
</script>

<template>
    <section class="view-section pipeline-view">
        <div class="funnels-toolbar">
            <div>
                <h1>Lejki &amp; Landing</h1>
                <p class="subtitle">Pracuj na realnych pipeline'ach powiązanych z Twoim kontem.</p>
            </div>
            <div class="funnels-actions">
                <select
                    v-model="selectedPipelineId"
                    class="pipeline-select"
                    :disabled="isLoadingPipelines || !pipelines.length"
                >
                    <option :value="null" disabled>Wybierz lejek</option>
                    <option v-for="pipeline in pipelines" :key="pipeline.id" :value="pipeline.id">
                        {{ pipeline.name }}
                    </option>
                </select>
                <button class="btn" @click="openCreatePipelineModal">+ Nowy Lejek</button>
            </div>
        </div>

        <div v-if="isLoadingPipelines" class="empty-state">
            <h3>Ładowanie lejków...</h3>
        </div>
        <div v-else-if="loadError" class="empty-state">
            <h3>Nie udało się załadować lejków</h3>
            <p>{{ loadError }}</p>
            <button class="btn" @click="loadPipelines">Spróbuj ponownie</button>
        </div>
        <div v-else-if="!pipelines.length" class="empty-state">
            <h3>Nie masz jeszcze żadnego lejka</h3>
            <p>Utwórz pierwszy pipeline, aby zacząć zarządzać stage'ami i leadami.</p>
            <button class="btn" @click="openCreatePipelineModal">Utwórz lejek</button>
        </div>
        <PipelineBoard v-else-if="selectedPipelineId !== null" :pipelineId="selectedPipelineId" />
    </section>

    <div v-if="showCreatePipelineModal" class="modal">
        <form class="modal-content" @submit.prevent="createPipeline">
            <h2>Nowy lejek</h2>
            <input v-model="newPipelineName" type="text" placeholder="Nazwa pipeline'u" />
            <div class="modal-buttons">
                <button type="submit" class="btn" :disabled="isCreatingPipeline">
                    {{ isCreatingPipeline ? "Dodawanie..." : "Dodaj" }}
                </button>
                <button
                    type="button"
                    class="btn-secondary"
                    @click="showCreatePipelineModal = false"
                >
                    Anuluj
                </button>
            </div>
        </form>
    </div>
</template>

<style scoped>
.view-section {
    width: 100%;
}

.pipeline-view {
    display: flex;
    flex-direction: column;
    gap: 24px;
}

.pipeline-view :deep(.pipeline-container) {
    padding: 0;
    background: transparent;
}

.funnels-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 20px;
}

.funnels-actions {
    display: flex;
    gap: 12px;
    align-items: center;
}

h1 {
    font-size: 2.8rem;
    margin-bottom: 10px;
    color: #f8fafc;
}

.subtitle {
    color: #94a3b8;
    font-size: 1.05rem;
}

.pipeline-select {
    min-width: 240px;
    padding: 12px;
    border-radius: 8px;
    border: 1px solid #4a5f7f;
    background: #1e293b;
    color: white;
}

.empty-state {
    padding: 32px;
    border-radius: 12px;
    border: 1px solid #334155;
    background: #1e293b;
}

.empty-state p {
    margin-top: 10px;
    color: #cbd5e1;
}

.btn {
    background: #38bdf8;
    color: #0f172a;
    border: none;
    border-radius: 10px;
    padding: 14px 18px;
    font-weight: 700;
    cursor: pointer;
}

.btn:disabled {
    opacity: 0.65;
    cursor: not-allowed;
}

.modal {
    position: fixed;
    inset: 0;
    background: rgba(15, 23, 42, 0.72);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 20;
}

.modal-content {
    width: min(420px, calc(100vw - 32px));
    padding: 24px;
    border-radius: 16px;
    background: #1e293b;
    border: 1px solid #334155;
}

.modal-content h2 {
    color: #f8fafc;
}

.modal-content input {
    width: 100%;
    margin-top: 16px;
    padding: 12px;
    border-radius: 8px;
    border: 1px solid #4a5f7f;
    background: #2c3e50;
    color: white;
}

.modal-buttons {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 20px;
}

.btn-secondary {
    background: transparent;
    border: 1px solid #475569;
    color: #e2e8f0;
    border-radius: 10px;
    padding: 14px 18px;
    font-weight: 700;
    cursor: pointer;
}

@media (max-width: 900px) {
    .funnels-toolbar {
        flex-direction: column;
        align-items: stretch;
    }

    .funnels-actions {
        flex-direction: column;
        align-items: stretch;
    }

    .pipeline-select {
        min-width: 0;
        width: 100%;
    }
}
</style>
