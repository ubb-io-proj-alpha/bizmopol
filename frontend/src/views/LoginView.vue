<script setup>
import { ref, computed } from "vue"
import { useRouter, useRoute } from "vue-router"
import { authAPI, setToken } from "@/services/api"

const router = useRouter()
const route = useRoute()

const isRegisterMode = computed(() => route.name === "register")

const email = ref("")
const password = ref("")
const name = ref("")
const authError = ref("")
const isSubmitting = ref(false)

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms))

const shouldRetryAuthError = (error) => {
    const status = Number(error?.status || 0)
    return status === 0 || status >= 500
}

const authWithRetry = async (requestFn) => {
    const maxAttempts = 30
    const retryDelayMs = 1000

    for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
        try {
            return await requestFn()
        } catch (error) {
            if (attempt < maxAttempts && shouldRetryAuthError(error)) {
                await sleep(retryDelayMs)
                continue
            }
            throw error
        }
    }

    throw new Error("auth_request_failed")
}

const mapAuthError = (error) => {
    const status = Number(error?.status || 0)

    if (status === 401) {
        return "Nieprawidłowy e-mail lub hasło."
    }
    if (status === 400) {
        return error?.data?.error || "Nieprawidłowe dane logowania."
    }
    if (status === 409) {
        return error?.data?.error || "Użytkownik już istnieje."
    }
    if (status >= 500 || status === 0) {
        return "Backend uruchamia się lub został zrestartowany. Spróbuj ponownie za chwilę."
    }

    return error?.data?.error || error?.message || "Błąd autoryzacji."
}

const handleAuth = async () => {
    authError.value = ""
    const normalizedEmail = email.value.trim().toLowerCase()
    const rawPassword = password.value
    const trimmedName = name.value.trim()

    if (!normalizedEmail || !rawPassword) {
        authError.value = "Proszę wypełnić wymagane pola."
        return
    }

    if (isRegisterMode.value && !trimmedName) {
        authError.value = "Podaj imię lub nazwę firmy."
        return
    }

    isSubmitting.value = true

    const payload = isRegisterMode.value
        ? { email: normalizedEmail, password: rawPassword, name: trimmedName }
        : { email: normalizedEmail, password: rawPassword }

    const requestFn = () => (isRegisterMode.value ? authAPI.register(payload) : authAPI.login(payload))

    try {
        const data = await authWithRetry(requestFn)

        if (!data?.token) {
            authError.value = "Brak tokenu w odpowiedzi serwera."
            return
        }

        setToken(data.token)
        email.value = ""
        password.value = ""
        name.value = ""
        router.push({ name: "dashboard" })
    } catch (error) {
        authError.value = mapAuthError(error)
    } finally {
        isSubmitting.value = false
    }
}

const goToRegister = () => router.push({ name: "register" })
const goToLogin = () => router.push({ name: "login" })
</script>

<template>
    <div class="login-page">
        <div class="login-card">
            <div class="logo-section">
                <h2>BizmoPol</h2>
                <p>{{ isRegisterMode ? "Załóż nowe konto biznesowe" : "System zarządzania biznesem online" }}</p>
            </div>

            <form @submit.prevent="handleAuth" class="login-form">
                <div v-if="isRegisterMode" class="input-group">
                    <label>Imię / Nazwa firmy</label>
                    <input v-model="name" type="text" placeholder="Np. Jan Kowalski" />
                </div>

                <div class="input-group">
                    <label>E-mail</label>
                    <input
                        v-model="email"
                        type="email"
                        placeholder="twoj@email.pl"
                        autocomplete="username"
                        :disabled="isSubmitting"
                        required
                    />
                </div>

                <div class="input-group">
                    <label>Hasło</label>
                    <input
                        v-model="password"
                        type="password"
                        placeholder="••••••••"
                        autocomplete="current-password"
                        :disabled="isSubmitting"
                        required
                    />
                </div>

                <p v-if="authError" class="auth-error">{{ authError }}</p>

                <button type="submit" class="login-btn" :disabled="isSubmitting">
                    {{ isSubmitting ? "Trwa logowanie..." : (isRegisterMode ? "Zarejestruj się" : "Zaloguj się") }}
                </button>
            </form>

            <div class="auth-toggle">
                <p v-if="!isRegisterMode">
                    Nie masz konta? <span @click="goToRegister">Zarejestruj się</span>
                </p>
                <p v-else>
                    Masz już konto? <span @click="goToLogin">Zaloguj się</span>
                </p>
            </div>
        </div>
    </div>
</template>

<style scoped>
.login-page {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100vh;
    width: 100vw;
    background: #0f172a;
}
.login-card {
    background: #1e293b;
    padding: 40px;
    border-radius: 16px;
    width: 100%;
    max-width: 400px;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.5);
    border: 1px solid #334155;
    text-align: center;
}
.login-form { margin-top: 25px; }
.input-group { margin-bottom: 20px; text-align: left; }
.input-group label { display: block; margin-bottom: 8px; font-size: 0.9rem; color: #94a3b8; }
.input-group input {
    width: 100%;
    padding: 12px;
    border-radius: 8px;
    border: 1px solid #334155;
    background: #0f172a;
    color: white;
    outline: none;
}
.input-group input:focus { border-color: #38bdf8; }
.auth-error { color: #f87171; font-size: 0.9rem; margin-bottom: 12px; }
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
    transition: 0.2s;
}
.login-btn:hover { background: #7dd3fc; }
.login-btn:disabled {
    opacity: 0.75;
    cursor: not-allowed;
}
.auth-toggle { margin-top: 20px; font-size: 0.9rem; color: #94a3b8; }
.auth-toggle span { color: #38bdf8; cursor: pointer; text-decoration: underline; font-weight: bold; }
</style>
