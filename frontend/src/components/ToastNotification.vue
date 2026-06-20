<script setup>
import { ref, provide } from "vue"

const toasts = ref([])
let nextId = 0

function show(message, type = "info", duration = 4000) {
    const id = nextId++
    toasts.value.push({ id, message, type, leaving: false })
    setTimeout(() => dismiss(id), duration)
}

function dismiss(id) {
    const t = toasts.value.find(t => t.id === id)
    if (t) {
        t.leaving = true
        setTimeout(() => {
            toasts.value = toasts.value.filter(t => t.id !== id)
        }, 300)
    }
}

provide("toast", { show })

defineExpose({ show })
</script>

<template>
    <slot></slot>
    <div class="toast-container">
        <div
            v-for="t in toasts" :key="t.id"
            :class="['toast', 'toast-' + t.type, { 'toast-leaving': t.leaving }]"
            @click="dismiss(t.id)"
        >
            <span class="material-icons toast-icon">{{
                t.type === 'success' ? 'check_circle' :
                t.type === 'error' ? 'error' :
                t.type === 'warning' ? 'warning' : 'info'
            }}</span>
            <span class="toast-msg">{{ t.message }}</span>
        </div>
    </div>
</template>

<style scoped>
.toast-container {
    position: fixed;
    top: 20px;
    right: 20px;
    z-index: 9999;
    display: flex;
    flex-direction: column;
    gap: 8px;
    pointer-events: none;
}
.toast {
    pointer-events: auto;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 12px 18px;
    border-radius: 10px;
    color: #f1f5f9;
    font-size: 0.9rem;
    cursor: pointer;
    animation: slideIn 0.3s ease;
    max-width: 420px;
    box-shadow: 0 4px 20px rgba(0,0,0,0.4);
    border: 1px solid #334155;
}
.toast-leaving {
    animation: slideOut 0.3s ease forwards;
}
.toast-icon { font-size: 1.2rem; flex-shrink: 0; }
.toast-msg { line-height: 1.4; }

.toast-success { background: #064e3b; border-color: #065f46; }
.toast-success .toast-icon { color: #34d399; }

.toast-error { background: #3f1414; border-color: #7f1d1d; }
.toast-error .toast-icon { color: #f87171; }

.toast-warning { background: #422006; border-color: #713f12; }
.toast-warning .toast-icon { color: #f59e0b; }

.toast-info { background: #1e293b; border-color: #334155; }
.toast-info .toast-icon { color: #38bdf8; }

@keyframes slideIn {
    from { transform: translateX(100%); opacity: 0; }
    to { transform: translateX(0); opacity: 1; }
}
@keyframes slideOut {
    from { transform: translateX(0); opacity: 1; }
    to { transform: translateX(100%); opacity: 0; }
}
</style>
