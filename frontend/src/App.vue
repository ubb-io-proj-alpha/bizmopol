<script setup>
import { onMounted, ref } from 'vue'
import PipelineBoard from '@/components/PipelineBoard.vue'
import { authAPI, clearToken, getToken, pipelineAPI, setToken } from '@/services/api'

const isLoggedIn = ref(Boolean(getToken()))
const isRegisterMode = ref(false)
const email = ref('')
const password = ref('')
const name = ref('')
const authError = ref('')
const currentView = ref('market')
const pipelines = ref([])
const selectedPipelineId = ref(null)
const isLoadingPipelines = ref(false)
const isCreatingPipeline = ref(false)
const showCreatePipelineModal = ref(false)
const newPipelineName = ref('')

const resetAuthForm = () => {
  email.value = ''
  password.value = ''
  name.value = ''
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

  try {
    const data = await pipelineAPI.getPipelines()
    pipelines.value = data
    syncSelectedPipeline(data)
  } catch (error) {
    if (error.status === 401) {
      logout('Sesja wygasła. Zaloguj się ponownie.')
      return
    }
    throw error
  } finally {
    isLoadingPipelines.value = false
  }
}

const handleAuth = async () => {
  authError.value = ''

  if (!email.value || !password.value || (isRegisterMode.value && !name.value)) {
    authError.value = 'Proszę wypełnić wymagane pola.'
    return
  }

  try {
    const payload = isRegisterMode.value
      ? { email: email.value, password: password.value, name: name.value }
      : { email: email.value, password: password.value }

    const response = isRegisterMode.value
      ? await authAPI.register(payload)
      : await authAPI.login(payload)

    setToken(response.token)
    isLoggedIn.value = true
    resetAuthForm()
    await loadPipelines()
  } catch (error) {
    authError.value = error.data?.error || 'Błąd autoryzacji.'
  }
}

const logout = (message = '') => {
  clearToken()
  isLoggedIn.value = false
  currentView.value = 'market'
  pipelines.value = []
  selectedPipelineId.value = null
  password.value = ''
  authError.value = message
}

const openCreatePipelineModal = () => {
  newPipelineName.value = ''
  showCreatePipelineModal.value = true
}

const createPipeline = async () => {
  const pipelineName = newPipelineName.value.trim()
  if (!pipelineName) {
    alert('Wpisz nazwę lejka.')
    return
  }

  try {
    isCreatingPipeline.value = true
    const pipeline = await pipelineAPI.createPipeline(pipelineName)
    await loadPipelines()
    selectedPipelineId.value = pipeline.id
    currentView.value = 'funnels'
    newPipelineName.value = ''
    showCreatePipelineModal.value = false
  } catch (error) {
    alert(error.data?.error || 'Nie udało się utworzyć lejka.')
  } finally {
    isCreatingPipeline.value = false
  }
}

onMounted(async () => {
  if (!isLoggedIn.value) {
    return
  }

  try {
    await loadPipelines()
  } catch (error) {
    authError.value = error.data?.error || 'Nie udało się pobrać lejków.'
  }
})
</script>

<template>
  <div v-if="!isLoggedIn" class="login-page">
    <div class="login-card">
      <div class="logo-section">
        <h2>BizmoPol</h2>
        <p>{{ isRegisterMode ? 'Załóż nowe konto biznesowe' : 'System zarządzania biznesem online' }}</p>
      </div>

      <form @submit.prevent="handleAuth" class="login-form">
        <div v-if="isRegisterMode" class="input-group">
          <label>Imię / Nazwa firmy</label>
          <input v-model="name" type="text" placeholder="Np. Jan Kowalski" required />
        </div>

        <div class="input-group">
          <label>E-mail</label>
          <input v-model="email" type="email" placeholder="twoj@email.pl" required />
        </div>

        <div class="input-group">
          <label>Hasło</label>
          <input v-model="password" type="password" placeholder="••••••••" required />
        </div>

        <button type="submit" class="login-btn">
          {{ isRegisterMode ? 'Zarejestruj się' : 'Zaloguj się' }}
        </button>
      </form>

      <p v-if="authError" class="auth-error">{{ authError }}</p>

      <div class="auth-toggle">
        <p v-if="!isRegisterMode">
          Nie masz konta? <span @click="isRegisterMode = true">Zarejestruj się</span>
        </p>
        <p v-else>
          Masz już konto? <span @click="isRegisterMode = false">Zaloguj się</span>
        </p>
      </div>
    </div>
  </div>

  <div v-else class="app-container">
    <aside class="sidebar">
      <div class="logo-section">
        <h2>BizmoPol</h2>
        <span class="version">v1.0 - Market Edition</span>
      </div>
      <nav>
        <ul>
          <li @click="currentView = 'market'" :class="{ active: currentView === 'market' }">🏠 BizmoPol Market</li>
          <li @click="currentView = 'funnels'" :class="{ active: currentView === 'funnels' }">🚀 Lejki & Landing</li>
          <li @click="currentView = 'crm'" :class="{ active: currentView === 'crm' }">👥 CRM & Kontakty</li>
          <li @click="currentView = 'comm'" :class="{ active: currentView === 'comm' }">💬 Komunikacja</li>
          <li @click="currentView = 'courses'" :class="{ active: currentView === 'courses' }">🎓 Kursy & Portal</li>
          <li @click="currentView = 'docs'" :class="{ active: currentView === 'docs' }">✍️ Dokumenty & E-Sign</li>
          <li @click="currentView = 'calendar'" :class="{ active: currentView === 'calendar' }">📅 Kalendarz</li>
        </ul>
      </nav>
      <div class="sidebar-footer">
        <button @click="logout()" class="logout-link">🚪 Wyloguj się</button>
      </div>
    </aside>

    <main class="main-content">
      <section v-if="currentView === 'market'" class="view-section">
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
            <button class="btn" @click="currentView = 'funnels'">+ Nowy Lejek</button>
            <button class="btn" @click="currentView = 'crm'">+ Dodaj Leada</button>
            <button class="btn" @click="currentView = 'courses'">+ Dodaj Kurs</button>
          </div>
        </div>
      </section>

      <section v-if="currentView === 'funnels'" class="view-section pipeline-view">
        <div class="funnels-toolbar">
          <div>
            <h1>Lejki & Landing</h1>
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
        <div v-else-if="!pipelines.length" class="empty-state">
          <h3>Nie masz jeszcze żadnego lejka</h3>
          <p>Utwórz pierwszy pipeline, aby zacząć zarządzać stage'ami i leadami.</p>
          <button class="btn" @click="openCreatePipelineModal">Utwórz lejek</button>
        </div>
        <PipelineBoard v-else-if="selectedPipelineId !== null" :pipelineId="selectedPipelineId" />
      </section>

      <section v-if="currentView === 'crm'" class="view-section">
        <h1>CRM — Zarządzanie kontaktami</h1>
        <div class="grid-fill">
          <div class="feature-card"><h3>Baza</h3><p>Centralna baza kontaktów i smart lists </p></div>
          <div class="feature-card"><h3>Pipeline</h3><p>Statusy leadów i etapy sprzedaży </p></div>
          <div class="feature-card"><h3>Historia</h3><p>Historia kontaktów, notatki i zadania </p></div>
        </div>
      </section>

      <section v-if="currentView === 'comm'" class="view-section">
        <h1>Komunikacja wielokanałowa</h1>
        <div class="list-fill">
          <div class="item">📧 Dwukierunkowa komunikacja e-mail (Gmail/Outlook Sync)</div>
          <div class="item">📱 Automatyczne sekwencje e-mail i SMS </div>
          <div class="item">💬 Unified Inbox: Messenger, Instagram, Google Business </div>
        </div>
      </section>

      <section v-if="currentView === 'courses'" class="view-section">
        <h1>Kursy & Portal Członkowski</h1>
        <div class="grid-fill">
          <div class="feature-card"><h3>Hosting</h3><p>Struktura kursu: moduły, lekcje, media </p></div>
          <div class="feature-card"><h3>Dostęp</h3><p>Dostęp po zakupie lub podpisaniu dokumentu </p></div>
          <div class="feature-card"><h3>Egzaminy</h3><p>Quizy, certyfikaty i materiały dodatkowe </p></div>
        </div>
      </section>

      <section v-if="currentView === 'docs'" class="view-section">
        <h1>Dokumenty i E-Signing</h1>
        <div class="list-fill">
          <div class="item">📄 Upload plików PDF i HTML </div>
          <div class="item">✒️ System E-podpisu i powiadomienia o podpisaniu </div>
          <div class="item">🗄️ Archiwum podpisanych plików i historia podpisów </div>
        </div>
      </section>

      <section v-if="currentView === 'calendar'" class="view-section">
        <h1>Kalendarz i Spotkania</h1>
        <div class="grid-fill">
          <div class="feature-card"><h3>Booking</h3><p>Rezerwacja spotkań z poziomu lejka/landing page </p></div>
          <div class="feature-card"><h3>Sync</h3><p>Synchronizacja z Google, Outlook, Zoom i Teams </p></div>
          <div class="feature-card"><h3>Strefy</h3><p>Inteligentny wybór strefy czasowej i blokowanie godzin </p></div>
        </div>
      </section>

      <div v-if="showCreatePipelineModal" class="modal">
        <form class="modal-content" @submit.prevent="createPipeline">
          <h2>Nowy lejek</h2>
          <input v-model="newPipelineName" type="text" placeholder="Nazwa pipeline'u" />
          <div class="modal-buttons">
            <button type="submit" class="btn" :disabled="isCreatingPipeline">
              {{ isCreatingPipeline ? 'Dodawanie...' : 'Dodaj' }}
            </button>
            <button
              type="button"
              @click="showCreatePipelineModal = false"
              class="logout-link modal-cancel"
            >
              Anuluj
            </button>
          </div>
        </form>
      </div>
    </main>
  </div>
</template>

<style scoped>
/* --- STYLE AUTORYZACJI (LOGOWANIE & REJESTRACJA) --- */
.login-page {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  width: 100vw;
  background: #2c3e50;
}
.login-card {
  background: #1a2a3a;
  padding: 40px;
  border-radius: 16px;
  width: 100%;
  max-width: 400px;
  box-shadow: none;
  border: none;
  text-align: center;
}
.login-form { margin-top: 25px; }
.input-group { margin-bottom: 20px; text-align: left; }
.input-group label { display: block; margin-bottom: 8px; font-size: 0.9rem; color: #94a3b8; }
.input-group input {
  width: 100%;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid #4a5f7f;
  background: #2c3e50;
  color: white;
  outline: none;
}
.input-group input:focus { border-color: #38bdf8; }
.login-btn {
  width: 100%;
  padding: 14px;
  background: #38bdf8;
  color: #0f172a;
  border: none;
  border-radius: 8px;
  font-weight: bold;
  font-size: 1rem;
  cursor: pointer;
}
.login-btn:hover { background: #7dd3fc; }
.auth-error { margin-top: 16px; color: #fca5a5; font-size: 0.95rem; }
.auth-toggle { margin-top: 20px; font-size: 0.9rem; color: #94a3b8; }
.auth-toggle span { color: #38bdf8; cursor: pointer; text-decoration: underline; font-weight: bold; }

/* --- STYLE PANELU GŁÓWNEGO --- */
.app-container { 
  display: flex; 
  width: 100vw;
  height: 100vh; 
  background: #2c3e50; 
  color: #f1f5f9; 
  font-family: 'Inter', sans-serif;
  overflow: hidden;
}

.sidebar { 
  width: 280px; 
  min-width: 280px;
  background: #2c3e50; 
  padding: 20px; 
  display: flex;
  flex-direction: column;
}

.logo-section { text-align: center; margin-bottom: 2rem; }
.sidebar h2 { color: #38bdf8; margin-bottom: 5px; }
.version { font-size: 0.7rem; color: #94a3b8; }

.sidebar li { 
  padding: 12px; 
  margin: 8px 0; 
  border-radius: 8px; 
  cursor: pointer; 
  list-style: none; 
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
}
.logout-link:hover { background: #ef4444; color: white; }

.main-content { 
  flex: 1; 
  padding: 40px; 
  overflow-y: auto; 
  background: #2c3e50;
}

.view-section { width: 100%; max-width: 1400px; margin: 0 auto; }
.market-overview, .grid-fill { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin-top: 30px; }
.stat-card, .feature-card { background: #1e293b; padding: 25px; border-radius: 12px; border: 1px solid #334155; }
.stat-card.highlight { border-color: #38bdf8; }
.stat-card .value { font-size: 2.2rem; font-weight: bold; color: #38bdf8; display: block; margin: 10px 0; }
.item { background: #1e293b; padding: 20px; border-radius: 8px; margin-bottom: 15px; border-left: 4px solid #38bdf8; }
.btn-group { display: flex; gap: 15px; margin-top: 20px; }
.btn { background: #38bdf8; border: none; padding: 14px 28px; border-radius: 8px; cursor: pointer; font-weight: bold; color: #0f172a; }
.btn:hover { background: #7dd3fc; }

h1 { font-size: 2.8rem; margin-bottom: 10px; color: #f8fafc; }
.subtitle { color: #94a3b8; font-size: 1.1rem; }
h3 { color: #38bdf8; margin-bottom: 10px; }

.pipeline-view {
  padding: 0 !important;
  overflow: hidden;
}

.funnels-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 20px;
  padding: 40px 40px 0;
}

.funnels-actions {
  display: flex;
  gap: 12px;
  align-items: center;
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
  margin: 40px;
  padding: 32px;
  border-radius: 12px;
  border: 1px solid #334155;
  background: #1e293b;
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

.modal-cancel {
  width: auto;
}
</style>
