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
        { path: "/login", name: "login", component: () => import("./views/LoginView.vue"), meta: { public: true } },
        { path: "/register", name: "register", component: () => import("./views/LoginView.vue"), meta: { public: true } },
        {
            path: "/dashboard",
            component: () => import("./views/DashboardLayout.vue"),
            children: [
                { path: "", name: "dashboard", component: () => import("./views/DashboardHomeView.vue") },
                { path: "funnels", name: "funnels", component: () => import("./views/FunnelsView.vue") },
                { path: "contacts", name: "contacts", component: () => import("./views/ContactsView.vue") },
                { path: "contacts/:id", name: "contactDetail", component: () => import("./views/ContactDetailView.vue") },
                { path: "communication", name: "communication", component: () => import("./views/CommunicationView.vue") },
                { path: "courses", name: "courses", component: () => import("./views/CoursesView.vue") },
                { path: "documents", name: "documents", component: () => import("./views/DocumentsView.vue") },
                { path: "calendar", name: "calendar", component: () => import("./views/CalendarView.vue") },
            ],
        },
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
