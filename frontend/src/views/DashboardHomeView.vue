<script setup>
import { useRouter } from "vue-router"
const router = useRouter()
const navigate = (name) => router.push({ name })
</script>

<template>
    <section class="view-section">
        <h1>Witaj w BizmoPol Market!</h1>
        <p class="subtitle">Podsumowanie Twojego biznesu online</p>

        <div class="market-overview">
            <div class="stat-card highlight">
                <h4>Całkowity przychód</h4>
                <span class="value">24 500 PLN</span>
                <p>+12% od ostatniego miesiąca</p>
            </div>
            <div class="stat-card">
                <h4>Aktywni klienci</h4>
                <span class="value">1 240</span>
            </div>
            <div class="stat-card">
                <h4>Konwersja lejków</h4>
                <span class="value">8.5%</span>
            </div>
        </div>

        <div class="quick-actions">
            <h3>Szybkie akcje</h3>
            <div class="btn-group">
                <button class="btn" @click="navigate('funnels')">+ Nowy Lejek</button>
                <button class="btn" @click="navigate('contacts')">+ Dodaj Leada</button>
                <button class="btn" @click="navigate('courses')">+ Dodaj Kurs</button>
            </div>
        </div>
    </section>
</template>

<style scoped>
.view-section { width: 100%; max-width: 1400px; margin: 0 auto; }
.market-overview { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin-top: 30px; }
.stat-card { background: #1e293b; padding: 25px; border-radius: 12px; border: 1px solid #334155; }
.stat-card.highlight { border-color: #38bdf8; }
.stat-card .value { font-size: 2.2rem; font-weight: bold; color: #38bdf8; display: block; margin: 10px 0; }
.btn-group { display: flex; gap: 15px; margin-top: 20px; }
.btn { background: #38bdf8; border: none; padding: 14px 28px; border-radius: 8px; cursor: pointer; font-weight: bold; color: #0f172a; }
.btn:hover { background: #7dd3fc; }
h1 { font-size: 2.8rem; margin-bottom: 10px; color: #f8fafc; }
.subtitle { color: #94a3b8; font-size: 1.1rem; }
h3 { color: #38bdf8; margin-bottom: 10px; }
</style>
