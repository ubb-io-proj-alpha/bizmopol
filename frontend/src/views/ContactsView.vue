<script setup>
import { inject } from "vue"
import CrmView from "../components/CrmView.vue"

const authFetch = inject("authFetch")
</script>

<template>
    <section class="view-section">
        <CrmView :auth-fetch="authFetch" />
    </section>
</template>

<style scoped>
.view-section { width: 100%; max-width: 1400px; margin: 0 auto; }
</style>
