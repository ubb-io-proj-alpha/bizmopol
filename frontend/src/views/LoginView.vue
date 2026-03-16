<script setup>
import { ref, computed } from "vue"
import { useRouter, useRoute } from "vue-router"

const TOKEN_KEY = "jwt_token"
const setToken = (t) => localStorage.setItem(TOKEN_KEY, t)

const router = useRouter()
const route = useRoute()

const isRegisterMode = computed(() => route.name === "register")

const email = ref("")
const password = ref("")
const name = ref("")
const authError = ref("")

const handleAuth = async () => {
    authError.value = ""
    if (!email.value || !password.value) {
        authError.value = "Proszę wypełnić wymagane pola."
        return
    }

    const url = isRegisterMode.value ? "/api/v1/auth/register" : "/api/v1/auth/login"
    const body = isRegisterMode.value
        ? { email: email.value, password: password.value, name: name.value }
        : { email: email.value, password: password.value }

    const response = await fetch(url, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
    })

    if (!response.ok) {
        const data = await response.json()
        authError.value = data.error || "Błąd autoryzacji."
        return
    }

    const data = await response.json()
    setToken(data.token)
    email.value = ""
    password.value = ""
    name.value = ""
    router.push({ name: "dashboard" })
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
                    <input v-model="email" type="email" placeholder="twoj@email.pl" required />
                </div>

                <div class="input-group">
                    <label>Hasło</label>
                    <input v-model="password" type="password" placeholder="••••••••" required />
                </div>

                <p v-if="authError" class="auth-error">{{ authError }}</p>

                <button type="submit" class="login-btn">
                    {{ isRegisterMode ? "Zarejestruj się" : "Zaloguj się" }}
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
.auth-toggle { margin-top: 20px; font-size: 0.9rem; color: #94a3b8; }
.auth-toggle span { color: #38bdf8; cursor: pointer; text-decoration: underline; font-weight: bold; }
</style>
