<template>
    <div class="builder-wrapper">
        <div class="builder-topbar">
            <button class="btn-secondary" @click="$emit('close')">
                <span class="material-icons">arrow_back</span> Wróć
            </button>
            <div class="page-info">
                Edytujesz stronę: <strong>{{ page.name }}</strong> <span class="path-badge">{{ page.path }}</span>
            </div>
            <button class="btn-primary" @click="savePage" :disabled="isSaving">
                <span class="material-icons">save</span> {{ isSaving ? 'Zapisywanie...' : 'Zapisz stronę' }}
            </button>
        </div>

        <div ref="editorContainer" class="editor-container"></div>
    </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, inject } from 'vue';
import grapesjs from 'grapesjs';
import 'grapesjs/dist/css/grapes.min.css';

import pluginWebpage from 'grapesjs-preset-webpage';
import pluginBlocks from 'grapesjs-blocks-basic';
import pluginFlexbox from 'grapesjs-blocks-flexbox';
import pluginNavbar from 'grapesjs-navbar';
import pluginForms from 'grapesjs-plugin-forms';

const props = defineProps({
    page: {
        type: Object,
        required: true
    }
});

const emit = defineEmits(['close', 'saved']);
const authFetch = inject('authFetch');

const editorContainer = ref(null);
let editor = null;
const isSaving = ref(false);

onMounted(() => {
    editor = grapesjs.init({
        container: editorContainer.value,
        fromElement: true,
        height: '100%',
        width: 'auto',
        storageManager: false, // We handle storage manually via API
        plugins: [pluginWebpage, pluginBlocks, pluginFlexbox, pluginNavbar, pluginForms]
    });

    // Custom block: link button to the booking page served at /umow-spotkanie
    editor.BlockManager.add('booking-button', {
        label: 'Umów spotkanie',
        category: 'Lejki',
        media: '<svg viewBox="0 0 24 24" width="22" height="22"><path fill="currentColor" d="M19 4h-1V2h-2v2H8V2H6v2H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2m0 16H5V10h14zm0-12H5V6h14z"/></svg>',
        content: '<a href="/umow-spotkanie" style="display:inline-block;padding:14px 28px;background:#38bdf8;color:#0f172a;font-weight:700;text-decoration:none;border-radius:8px;font-family:sans-serif;">Umów spotkanie</a>',
    });

    if (props.page && props.page.structure) {
        try {
            const projectData = typeof props.page.structure === 'string' 
                ? JSON.parse(props.page.structure) 
                : props.page.structure;
            editor.loadProjectData(projectData);
        } catch (e) {
            console.error("Błąd ładowania struktury GrapesJS:", e);
        }
    }
});

onBeforeUnmount(() => {
    if (editor) {
        editor.destroy();
    }
});

async function savePage() {
    if (!editor) return;
    isSaving.value = true;

    try {
        const projectData = editor.getProjectData();
        const htmlContent = editor.getHtml();
        const cssContent = editor.getCss();
        
        const payload = {
            name: props.page.name,
            path: props.page.path,
            structure: JSON.stringify(projectData),
            html_content: htmlContent,
            css_content: cssContent
        };

        const res = await authFetch("/api/v1/funnels/pages/" + props.page.id, { 
            method: "PUT", 
            body: JSON.stringify(payload) 
        });

        if (!res.ok) {
            const d = await res.json();
            throw new Error(d.error || "Błąd zapisu");
        }

        emit('saved');
    } catch (e) {
        alert("Błąd zapisu: " + e.message);
    } finally {
        isSaving.value = false;
    }
}
</script>

<style scoped>
.builder-wrapper {
    display: flex;
    flex-direction: column;
    height: 85vh; /* Adjust based on your layout needs */
    border: 1px solid #334155;
    border-radius: 12px;
    overflow: hidden;
    background: #1e293b;
}

.builder-topbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 20px;
    background: #0f172a;
    border-bottom: 1px solid #334155;
}

.page-info {
    color: #f8fafc;
    font-size: 1rem;
}

.path-badge {
    background: #1e3a5f;
    color: #38bdf8;
    padding: 2px 8px;
    border-radius: 6px;
    font-family: monospace;
    margin-left: 8px;
    font-size: 0.85rem;
}

.editor-container {
    flex-grow: 1;
    width: 100%;
}

/* Inherit button styles from your main file or copy them here */
.btn-primary { padding: 8px 16px; background: #38bdf8; border: none; border-radius: 8px; color: #0f172a; font-weight: bold; cursor: pointer; display: flex; align-items: center; gap: 6px; }
.btn-primary:disabled { opacity: 0.7; cursor: not-allowed; }
.btn-secondary { padding: 8px 16px; background: transparent; border: 1px solid #475569; border-radius: 8px; color: #94a3b8; cursor: pointer; display: flex; align-items: center; gap: 6px; }
</style>