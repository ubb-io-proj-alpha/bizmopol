<script setup>
import { inject } from "vue"
import ContactDetailViewComponent from "../components/ContactDetailView.vue"

const authFetch = inject("authFetch")
</script>

<template>
    <section class="view-section">
        <ContactDetailViewComponent :auth-fetch="authFetch" />
    </section>
</template>

<style scoped>
.view-section { width: 100%; max-width: 1400px; margin: 0 auto; }
</style>
