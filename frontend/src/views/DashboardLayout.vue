<script setup>
import { computed } from "vue"
import { useRouter, useRoute } from "vue-router"

const TOKEN_KEY = "jwt_token"
const removeToken = () => localStorage.removeItem(TOKEN_KEY)
const getToken = () => localStorage.getItem(TOKEN_KEY)

const router = useRouter()
const route = useRoute()

const currentView = computed(() => route.name || "dashboard")

const authFetch = (url, options = {}) => {
    const token = getToken()
    return fetch(url, {
        ...options,
        headers: {
            ...(options.headers || {}),
            "Authorization": token ? `Bearer ${token}` : "",
            "Content-Type": "application/json",
        },
    })
}

const logout = () => {
    removeToken()
    router.push({ name: "login" })
}

const navigate = (name) => router.push({ name })

provide("authFetch", authFetch)
</script>

<script>
import { provide } from "vue"
export default {}
</script>

<template>
    <div class="app-container">
        <aside class="sidebar">
            <div class="logo-section">
                <h2>BizmoPol</h2>
                <span class="version">v1.0 - Market Edition</span>
            </div>
            <nav>
                <ul>
                    <li @click="navigate('dashboard')" :class="{ active: currentView === 'dashboard' }">
                        <span class="material-icons nav-icon">store</span> BizmoPol Market
                    </li>
                    <li @click="navigate('funnels')" :class="{ active: currentView === 'funnels' }">
                        <span class="material-icons nav-icon">rocket_launch</span> Lejki &amp; Landing
                    </li>
                    <li @click="navigate('contacts')" :class="{ active: currentView === 'contacts' || currentView === 'contactDetail' }">
                        <span class="material-icons nav-icon">group</span> CRM &amp; Kontakty
                    </li>
                    <li @click="navigate('communication')" :class="{ active: currentView === 'communication' }">
                        <span class="material-icons nav-icon">chat</span> Komunikacja
                    </li>
                    <li @click="navigate('courses')" :class="{ active: currentView === 'courses' }">
                        <span class="material-icons nav-icon">school</span> Kursy &amp; Portal
                    </li>
                    <li @click="navigate('documents')" :class="{ active: currentView === 'documents' }">
                        <span class="material-icons nav-icon">edit_document</span> Dokumenty &amp; E-Sign
                    </li>
                    <li @click="navigate('calendar')" :class="{ active: currentView === 'calendar' }">
                        <span class="material-icons nav-icon">calendar_month</span> Kalendarz
                    </li>
                </ul>
            </nav>
            <div class="sidebar-footer">
                <button @click="logout" class="logout-link">
                    <span class="material-icons nav-icon">logout</span> Wyloguj się
                </button>
            </div>
        </aside>

        <main class="main-content">
            <router-view />
        </main>
    </div>
</template>

<style scoped>
.app-container {
    display: flex;
    width: 100vw;
    height: 100vh;
    background: #0f172a;
    color: #f1f5f9;
    font-family: "Inter", sans-serif;
    overflow: hidden;
}

.sidebar {
    width: 280px;
    min-width: 280px;
    background: #1e293b;
    padding: 20px;
    border-right: 1px solid #334155;
    display: flex;
    flex-direction: column;
}

.logo-section { text-align: center; margin-bottom: 2rem; }
.sidebar h2 { color: #38bdf8; margin-bottom: 5px; }
.version { font-size: 0.7rem; color: #94a3b8; }

.nav-icon {
    font-size: 1.1rem;
    vertical-align: middle;
    margin-right: 8px;
}

.sidebar li {
    padding: 12px;
    margin: 8px 0;
    border-radius: 8px;
    cursor: pointer;
    transition: 0.2s;
    list-style: none;
    display: flex;
    align-items: center;
}
.sidebar li:hover { background: #334155; }
.sidebar li.active { background: #38bdf8; color: #0f172a; font-weight: bold; }

.sidebar-footer { margin-top: auto; padding-top: 20px; }
.logout-link {
    width: 100%;
    background: transparent;
    border: 1px solid #ef4444;
    color: #ef4444;
    padding: 10px;
    border-radius: 8px;
    cursor: pointer;
    transition: 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
}
.logout-link:hover { background: #ef4444; color: white; }

.main-content {
    flex: 1;
    padding: 40px;
    overflow-y: auto;
    background: #0f172a;
}
</style>
