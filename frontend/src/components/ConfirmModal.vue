<script setup>
import { ref, provide } from "vue"

const visible = ref(false)
const title = ref("")
const message = ref("")
const confirmLabel = ref("Potwierdź")
const cancelLabel = ref("Anuluj")
const variant = ref("danger")
let resolvePromise = null

function ask(opts = {}) {
    title.value = opts.title || "Potwierdzenie"
    message.value = opts.message || "Czy na pewno?"
    confirmLabel.value = opts.confirmLabel || "Potwierdź"
    cancelLabel.value = opts.cancelLabel || "Anuluj"
    variant.value = opts.variant || "danger"
    visible.value = true
    return new Promise((resolve) => { resolvePromise = resolve })
}

function onConfirm() {
    visible.value = false
    if (resolvePromise) resolvePromise(true)
    resolvePromise = null
}

function onCancel() {
    visible.value = false
    if (resolvePromise) resolvePromise(false)
    resolvePromise = null
}

provide("confirm", ask)
</script>

<template>
    <slot></slot>
    <Teleport to="body">
        <Transition name="confirm-fade">
            <div v-if="visible" class="confirm-overlay" @click.self="onCancel">
                <Transition name="confirm-scale">
                    <div v-if="visible" class="confirm-modal">
                        <div class="confirm-icon-wrap" :class="'confirm-icon-' + variant">
                            <span class="material-icons confirm-icon">{{
                                variant === 'danger' ? 'warning' :
                                variant === 'warning' ? 'help_outline' : 'info'
                            }}</span>
                        </div>
                        <h3 class="confirm-title">{{ title }}</h3>
                        <p class="confirm-message">{{ message }}</p>
                        <div class="confirm-actions">
                            <button class="confirm-btn-cancel" @click="onCancel">{{ cancelLabel }}</button>
                            <button :class="['confirm-btn-ok', 'confirm-btn-' + variant]" @click="onConfirm">
                                {{ confirmLabel }}
                            </button>
                        </div>
                    </div>
                </Transition>
            </div>
        </Transition>
    </Teleport>
</template>

<style scoped>
.confirm-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 10000;
}
.confirm-modal {
    background: #1e293b;
    border: 1px solid #334155;
    border-radius: 16px;
    padding: 32px;
    width: 100%;
    max-width: 400px;
    text-align: center;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}
.confirm-icon-wrap {
    width: 56px;
    height: 56px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 16px;
}
.confirm-icon { font-size: 1.8rem; }
.confirm-icon-danger { background: #3f1414; }
.confirm-icon-danger .confirm-icon { color: #f87171; }
.confirm-icon-warning { background: #422006; }
.confirm-icon-warning .confirm-icon { color: #f59e0b; }
.confirm-icon-info { background: #0c2d48; }
.confirm-icon-info .confirm-icon { color: #38bdf8; }

.confirm-title {
    color: #f1f5f9;
    font-size: 1.15rem;
    margin-bottom: 8px;
}
.confirm-message {
    color: #94a3b8;
    font-size: 0.9rem;
    line-height: 1.5;
    margin-bottom: 24px;
}
.confirm-actions {
    display: flex;
    gap: 10px;
    justify-content: center;
}
.confirm-btn-cancel {
    padding: 10px 20px;
    background: transparent;
    border: 1px solid #475569;
    border-radius: 8px;
    color: #94a3b8;
    cursor: pointer;
    font-size: 0.9rem;
    transition: 0.2s;
}
.confirm-btn-cancel:hover { border-color: #f1f5f9; color: #f1f5f9; }
.confirm-btn-ok {
    padding: 10px 20px;
    border: none;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    font-size: 0.9rem;
    transition: 0.2s;
}
.confirm-btn-danger { background: #ef4444; color: white; }
.confirm-btn-danger:hover { background: #dc2626; }
.confirm-btn-warning { background: #f59e0b; color: #0f172a; }
.confirm-btn-warning:hover { background: #d97706; }
.confirm-btn-info { background: #38bdf8; color: #0f172a; }
.confirm-btn-info:hover { background: #7dd3fc; }

.confirm-fade-enter-active, .confirm-fade-leave-active { transition: opacity 0.2s ease; }
.confirm-fade-enter-from, .confirm-fade-leave-to { opacity: 0; }

.confirm-scale-enter-active { transition: all 0.2s ease; }
.confirm-scale-leave-active { transition: all 0.15s ease; }
.confirm-scale-enter-from { opacity: 0; transform: scale(0.9); }
.confirm-scale-leave-to { opacity: 0; transform: scale(0.95); }
</style>
