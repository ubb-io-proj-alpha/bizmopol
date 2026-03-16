import "./assets/main.css"

import { createApp } from "vue"
import { createRouter, createWebHistory } from "vue-router"
import App from "./App.vue"

const TOKEN_KEY = "jwt_token"
const getToken = () => localStorage.getItem(TOKEN_KEY)

const router = createRouter({
    history: createWebHistory(),
    routes: [
        { path: "/", redirect: "/login" },
        { path: "/login", name: "login", meta: { public: true } },
        { path: "/register", name: "register", meta: { public: true } },
        { path: "/dashboard", name: "dashboard" },
        { path: "/dashboard/funnels", name: "funnels" },
        { path: "/dashboard/contacts", name: "contacts" },
        { path: "/dashboard/contacts/:id", name: "contactDetail" },
        { path: "/dashboard/communication", name: "communication" },
        { path: "/dashboard/courses", name: "courses" },
        { path: "/dashboard/documents", name: "documents" },
        { path: "/dashboard/calendar", name: "calendar" },
    ],
})

router.beforeEach((to, from, next) => {
    if (to.meta.public) {
        next()
    } else if (!getToken()) {
        next({ name: "login" })
    } else {
        next()
    }
})

createApp(App).use(router).mount("#app")
