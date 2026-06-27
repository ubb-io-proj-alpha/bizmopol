<script setup>
import { ref, reactive, onMounted, onUnmounted, inject, computed, watch } from "vue"
import { useWebSocket } from "../composables/useWebSocket.js"

const authFetch = inject("authFetch")
const toast = inject("toast")
const confirm = inject("confirm")

const activeTab = ref("calendar")
const loading = ref(false)
const formError = ref("")

const stats = ref(null)
const events = ref([])

const currentDate = ref(new Date())
const selectedDate = ref(null)
const selectedEvent = ref(null)

const showEventModal = ref(false)
const editingEventId = ref(null)
const eventForm = reactive({
    title: "", description: "", eventType: "meeting",
    startDate: "", startTime: "09:00", endDate: "", endTime: "09:30",
    allDay: false, location: "",
    contactId: "", contactName: "", contactEmail: "",
    createZoom: false, color: "", reminder: 15,
    hasZoom: false, removeZoom: false, addZoom: false,
})

const zoomAccount = ref(null)
const zoomLoading = ref(false)
const zoomSaving = ref(false)
const zoomTesting = ref(false)
const zoomTestResult = ref(null)
const zoomForm = reactive({ accountId: "", clientId: "", clientSecret: "" })

const bookingSettings = ref(null)
const bookingLoading = ref(false)
const bookingSaving = ref(false)
const bookingError = ref("")
const WEEKDAYS = [
    { v: 1, l: "Pn" }, { v: 2, l: "Wt" }, { v: 3, l: "Śr" },
    { v: 4, l: "Cz" }, { v: 5, l: "Pt" }, { v: 6, l: "Sb" }, { v: 0, l: "Nd" },
]
const bookingForm = reactive({
    enabled: true,
    workingDays: [1, 2, 3, 4, 5],
    startTime: "09:00",
    endTime: "17:00",
    slotMinutes: 30,
    meetingTitle: "Spotkanie: {name}",
})

const contacts = ref([])
const contactSearch = ref("")

const currentYear = computed(() => currentDate.value.getFullYear())
const currentMonth = computed(() => currentDate.value.getMonth())
const monthName = computed(() => {
    const names = ["Styczeń","Luty","Marzec","Kwiecień","Maj","Czerwiec","Lipiec","Sierpień","Wrzesień","Październik","Listopad","Grudzień"]
    return names[currentMonth.value] + " " + currentYear.value
})

const calendarDays = computed(() => {
    const year = currentYear.value
    const month = currentMonth.value
    const firstDay = new Date(year, month, 1)
    const lastDay = new Date(year, month + 1, 0)
    let startDow = firstDay.getDay()
    if (startDow === 0) startDow = 7
    const days = []
    const prevMonthLast = new Date(year, month, 0).getDate()
    for (let i = startDow - 1; i > 0; i--) {
        days.push({ day: prevMonthLast - i + 1, currentMonth: false, date: new Date(year, month - 1, prevMonthLast - i + 1) })
    }
    for (let d = 1; d <= lastDay.getDate(); d++) {
        days.push({ day: d, currentMonth: true, date: new Date(year, month, d) })
    }
    const remaining = 42 - days.length
    for (let d = 1; d <= remaining; d++) {
        days.push({ day: d, currentMonth: false, date: new Date(year, month + 1, d) })
    }
    return days
})

function getEventsForDay(date) {
    const ds = formatDateKey(date)
    return events.value.filter(e => {
        if (!e.start_time) return false
        return formatDateKey(new Date(e.start_time)) === ds
    })
}

function formatDateKey(d) {
    const y = d.getFullYear()
    const m = String(d.getMonth() + 1).padStart(2, "0")
    const day = String(d.getDate()).padStart(2, "0")
    return `${y}-${m}-${day}`
}

function isToday(date) {
    const t = new Date()
    return date.getFullYear() === t.getFullYear() && date.getMonth() === t.getMonth() && date.getDate() === t.getDate()
}

function prevMonth() { currentDate.value = new Date(currentYear.value, currentMonth.value - 1, 1); loadMonthEvents() }
function nextMonth() { currentDate.value = new Date(currentYear.value, currentMonth.value + 1, 1); loadMonthEvents() }
function goToday() { currentDate.value = new Date(); loadMonthEvents() }

function selectDay(dayObj) {
    selectedDate.value = dayObj.date
    selectedEvent.value = null
}

async function loadStats() {
    try {
        const res = await authFetch("/api/v1/calendar/stats")
        if (res.ok) stats.value = await res.json()
    } catch (_) {}
}

async function loadMonthEvents() {
    loading.value = true
    try {
        const from = formatDateKey(new Date(currentYear.value, currentMonth.value, 1))
        const to = formatDateKey(new Date(currentYear.value, currentMonth.value + 1, 0))
        const res = await authFetch(`/api/v1/calendar/events/?from=${from}&to=${to}&page_size=200`)
        if (res.ok) {
            const data = await res.json()
            events.value = data.data || []
        }
    } catch (_) {} finally { loading.value = false }
}

async function loadContacts() {
    try {
        const params = new URLSearchParams({ page_size: "100" })
        if (contactSearch.value) params.set("search", contactSearch.value)
        const res = await authFetch("/api/v1/contacts/?" + params.toString())
        if (res.ok) {
            const data = await res.json()
            contacts.value = data.data || []
        }
    } catch (_) {}
}

function openNewEvent(date) {
    editingEventId.value = null
    const d = date || selectedDate.value || new Date()
    const ds = formatDateKey(d)
    Object.assign(eventForm, {
        title: "", description: "", eventType: "meeting",
        startDate: ds, startTime: "09:00", endDate: ds, endTime: "09:30",
        allDay: false, location: "",
        contactId: "", contactName: "", contactEmail: "",
        createZoom: false, color: "", reminder: 15,
        hasZoom: false, removeZoom: false, addZoom: false,
    })
    formError.value = ""
    showEventModal.value = true
    loadContacts()
}

function openEditEvent(event) {
    editingEventId.value = event.id
    const st = new Date(event.start_time)
    const et = new Date(event.end_time)
    const eventHasZoom = event.zoom_meeting_id > 0
    Object.assign(eventForm, {
        title: event.title, description: event.description || "",
        eventType: event.event_type || "meeting",
        startDate: formatDateKey(st),
        startTime: `${String(st.getHours()).padStart(2,"0")}:${String(st.getMinutes()).padStart(2,"0")}`,
        endDate: formatDateKey(et),
        endTime: `${String(et.getHours()).padStart(2,"0")}:${String(et.getMinutes()).padStart(2,"0")}`,
        allDay: event.all_day,
        location: event.location || "",
        contactId: event.contact_id || "",
        contactName: event.contact_name || "",
        contactEmail: event.contact_email || "",
        createZoom: false, color: event.color || "",
        reminder: event.reminder || 15,
        hasZoom: eventHasZoom, removeZoom: false, addZoom: false,
    })
    formError.value = ""
    showEventModal.value = true
    loadContacts()
}

function selectContact(c) {
    eventForm.contactId = c.id
    eventForm.contactName = c.name
    eventForm.contactEmail = c.email || ""
}

async function saveEvent() {
    formError.value = ""
    if (!eventForm.title || !eventForm.startDate || !eventForm.startTime) {
        formError.value = "Tytuł, data i godzina rozpoczęcia są wymagane."
        return
    }
    try {
        const startTime = new Date(`${eventForm.startDate}T${eventForm.startTime}:00`).toISOString()
        const endTime = new Date(`${eventForm.endDate || eventForm.startDate}T${eventForm.endTime || eventForm.startTime}:00`).toISOString()

        if (editingEventId.value) {
            const payload = {
                title: eventForm.title, description: eventForm.description,
                event_type: eventForm.eventType, start_time: startTime, end_time: endTime,
                all_day: eventForm.allDay, location: eventForm.location,
                contact_id: eventForm.contactId, contact_name: eventForm.contactName,
                contact_email: eventForm.contactEmail, color: eventForm.color,
                reminder: eventForm.reminder,
            }
            if (eventForm.removeZoom) payload.remove_zoom = true
            if (eventForm.addZoom) payload.add_zoom = true
            const res = await authFetch("/api/v1/calendar/events/" + editingEventId.value, {
                method: "PUT", body: JSON.stringify(payload),
            })
            if (!res.ok) throw new Error("Błąd aktualizacji")
            if (eventForm.removeZoom) toast.show("Spotkanie Zoom zostanie usunięte", "info")
            else if (eventForm.addZoom) toast.show("Spotkanie Zoom jest tworzone...", "info")
            else toast.show("Wydarzenie zaktualizowane", "success")
        } else {
            const payload = {
                title: eventForm.title, description: eventForm.description,
                event_type: eventForm.eventType, start_time: startTime, end_time: endTime,
                all_day: eventForm.allDay, location: eventForm.location,
                contact_id: eventForm.contactId, contact_name: eventForm.contactName,
                contact_email: eventForm.contactEmail, create_zoom: eventForm.createZoom,
                color: eventForm.color, reminder: eventForm.reminder,
            }
            const res = await authFetch("/api/v1/calendar/events/", {
                method: "POST", body: JSON.stringify(payload),
            })
            if (!res.ok) throw new Error("Błąd tworzenia")
            toast.show(eventForm.createZoom ? "Wydarzenie utworzone ze spotkaniem Zoom" : "Wydarzenie utworzone", "success")
        }
        showEventModal.value = false
        selectedEvent.value = null
        await Promise.all([loadMonthEvents(), loadStats()])
    } catch (e) {
        formError.value = e.message
        toast.show(e.message, "error")
    }
}

async function deleteEvent(id) {
    const ev = selectedEvent.value
    const hasZoom = ev && ev.zoom_meeting_id > 0
    const msg = hasZoom ? "Wydarzenie oraz powiązane spotkanie Zoom zostaną trwale usunięte." : "Wydarzenie zostanie trwale usunięte."
    if (!await confirm({ title: "Usuń wydarzenie", message: msg, confirmLabel: "Usuń", variant: "danger" })) return
    try {
        const res = await authFetch("/api/v1/calendar/events/" + id, { method: "DELETE" })
        if (!res.ok) throw new Error("Błąd usuwania")
        selectedEvent.value = null
        toast.show(hasZoom ? "Wydarzenie usunięte, spotkanie Zoom anulowane" : "Wydarzenie usunięte", "success")
        await Promise.all([loadMonthEvents(), loadStats()])
    } catch (e) { toast.show(e.message, "error") }
}

async function updateEventStatus(id, status) {
    try {
        const res = await authFetch("/api/v1/calendar/events/" + id, {
            method: "PUT", body: JSON.stringify({ status }),
        })
        if (!res.ok) throw new Error("Błąd zmiany statusu")
        const labels = { completed: "Wydarzenie zakończone", cancelled: "Wydarzenie anulowane", scheduled: "Wydarzenie przywrócone" }
        toast.show(labels[status] || "Status zmieniony", status === "cancelled" ? "warning" : "success")
        selectedEvent.value = null
        await Promise.all([loadMonthEvents(), loadStats()])
    } catch (e) { toast.show(e.message, "error") }
}

async function removeZoomFromEvent(id) {
    if (!await confirm({ title: "Usuń spotkanie Zoom", message: "Spotkanie Zoom zostanie anulowane. Wydarzenie pozostanie jako offline.", confirmLabel: "Usuń Zoom", variant: "warning" })) return
    try {
        const res = await authFetch("/api/v1/calendar/events/" + id, {
            method: "PUT", body: JSON.stringify({ remove_zoom: true }),
        })
        if (!res.ok) throw new Error("Błąd usuwania Zoom")
        toast.show("Spotkanie Zoom zostanie usunięte", "info")
        selectedEvent.value = null
        await Promise.all([loadMonthEvents(), loadStats()])
    } catch (e) { toast.show(e.message, "error") }
}

async function loadZoomAccount() {
    zoomLoading.value = true
    try {
        const res = await authFetch("/api/v1/calendar/zoom/")
        if (res.ok) {
            const data = await res.json()
            if (data.configured === false) { zoomAccount.value = null }
            else {
                zoomAccount.value = data
                zoomForm.accountId = data.account_id || ""
                zoomForm.clientId = data.client_id || ""
                zoomForm.clientSecret = ""
            }
        }
    } catch (_) {} finally { zoomLoading.value = false }
}

async function saveZoomAccount() {
    formError.value = ""
    if (!zoomForm.accountId || !zoomForm.clientId || !zoomForm.clientSecret) {
        formError.value = "Account ID, Client ID i Client Secret są wymagane."
        return
    }
    zoomSaving.value = true
    try {
        const payload = { account_id: zoomForm.accountId, client_id: zoomForm.clientId, client_secret: zoomForm.clientSecret }
        let res
        if (zoomAccount.value) {
            res = await authFetch("/api/v1/calendar/zoom/", { method: "PUT", body: JSON.stringify(payload) })
        } else {
            res = await authFetch("/api/v1/calendar/zoom/", { method: "POST", body: JSON.stringify(payload) })
        }
        if (!res.ok) { const d = await res.json(); throw new Error(d.error || "Błąd zapisu") }
        zoomAccount.value = await res.json()
        zoomTestResult.value = null
        toast.show("Konto Zoom zapisane", "success")
    } catch (e) { formError.value = e.message; toast.show(e.message, "error") } finally { zoomSaving.value = false }
}

async function testZoom() {
    zoomTesting.value = true; zoomTestResult.value = null
    try {
        const res = await authFetch("/api/v1/calendar/zoom/test", { method: "POST" })
        if (res.ok) zoomTestResult.value = await res.json()
    } catch (_) { zoomTestResult.value = { status: "error", error: "Test failed." } }
    finally { zoomTesting.value = false }
}

async function disconnectZoom() {
    if (!await confirm({ title: "Rozłącz Zoom", message: "Konto Zoom zostanie rozłączone. Istniejące spotkania nie zostaną usunięte.", confirmLabel: "Rozłącz", variant: "danger" })) return
    try {
        await authFetch("/api/v1/calendar/zoom/", { method: "DELETE" })
        zoomAccount.value = null
        Object.assign(zoomForm, { accountId: "", clientId: "", clientSecret: "" })
        zoomTestResult.value = null
        toast.show("Konto Zoom rozłączone", "success")
    } catch (e) { toast.show("Błąd rozłączania", "error") }
}

function minutesToTime(min) {
    const h = Math.floor(min / 60), m = min % 60
    return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}`
}
function timeToMinutes(t) {
    const [h, m] = (t || "0:0").split(":").map(Number)
    return (h || 0) * 60 + (m || 0)
}
function toggleWorkingDay(v) {
    const i = bookingForm.workingDays.indexOf(v)
    if (i === -1) bookingForm.workingDays.push(v)
    else bookingForm.workingDays.splice(i, 1)
}

async function loadBookingSettings() {
    bookingLoading.value = true
    bookingError.value = ""
    try {
        const res = await authFetch("/api/v1/booking/settings")
        if (res.ok) {
            const d = await res.json()
            bookingSettings.value = d
            bookingForm.enabled = d.enabled
            bookingForm.workingDays = (d.working_days && d.working_days.length) ? [...d.working_days] : []
            bookingForm.startTime = minutesToTime(d.start_minutes)
            bookingForm.endTime = minutesToTime(d.end_minutes)
            bookingForm.slotMinutes = d.slot_minutes
            bookingForm.meetingTitle = d.meeting_title || "Spotkanie: {name}"
        }
    } catch (_) {} finally { bookingLoading.value = false }
}

async function saveBookingSettings() {
    bookingError.value = ""
    if (bookingForm.workingDays.length === 0) { bookingError.value = "Wybierz przynajmniej jeden dzień."; return }
    const startM = timeToMinutes(bookingForm.startTime)
    const endM = timeToMinutes(bookingForm.endTime)
    if (startM >= endM) { bookingError.value = "Godzina zakończenia musi być po godzinie rozpoczęcia."; return }
    bookingSaving.value = true
    try {
        const payload = {
            enabled: bookingForm.enabled,
            working_days: bookingForm.workingDays,
            start_minutes: startM,
            end_minutes: endM,
            slot_minutes: Number(bookingForm.slotMinutes),
            meeting_title: bookingForm.meetingTitle,
        }
        const res = await authFetch("/api/v1/booking/settings", { method: "PUT", body: JSON.stringify(payload) })
        if (!res.ok) { const d = await res.json(); throw new Error(d.error || "Błąd zapisu") }
        bookingSettings.value = await res.json()
        toast.show("Ustawienia rezerwacji zapisane", "success")
    } catch (e) { bookingError.value = e.message; toast.show(e.message, "error") } finally { bookingSaving.value = false }
}

function formatTime(d) {
    if (!d) return ""
    const dt = new Date(d)
    return `${String(dt.getHours()).padStart(2,"0")}:${String(dt.getMinutes()).padStart(2,"0")}`
}

function formatDateFull(d) {
    if (!d) return ""
    return new Date(d).toLocaleString("pl-PL")
}

function formatDateShort(d) {
    if (!d) return ""
    const dt = new Date(d)
    return dt.toLocaleDateString("pl-PL", { day: "2-digit", month: "2-digit" })
}

function eventTypeLabel(t) {
    const map = { meeting: "Spotkanie", call: "Rozmowa", follow_up: "Follow-up", other: "Inne" }
    return map[t] || t
}

function eventTypeIcon(t) {
    const map = { meeting: "videocam", call: "phone", follow_up: "replay", other: "event" }
    return map[t] || "event"
}

function statusLabel(s) {
    const map = { scheduled: "Zaplanowane", completed: "Zakończone", cancelled: "Anulowane" }
    return map[s] || s
}

function initials(name) {
    if (!name) return "?"
    return name.split(" ").map(w => w[0]).join("").toUpperCase().slice(0, 2)
}

function avatarColor(name) {
    const colors = ["#38bdf8","#818cf8","#34d399","#f59e0b","#f472b6","#a78bfa","#fb923c"]
    let hash = 0
    for (const c of (name || "")) hash = c.charCodeAt(0) + ((hash << 5) - hash)
    return colors[Math.abs(hash) % colors.length]
}

const selectedDayEvents = computed(() => {
    if (!selectedDate.value) return []
    return getEventsForDay(selectedDate.value)
})

watch(activeTab, (tab) => {
    if (tab === "zoom") loadZoomAccount()
    else if (tab === "booking") loadBookingSettings()
})

const { onMessage } = useWebSocket()

let removeWsListener = null

onMounted(async () => {
    await Promise.all([loadStats(), loadMonthEvents(), loadZoomAccount()])
    removeWsListener = onMessage((msg) => {
        if (msg.type === "zoom_job_done") {
            toast.show(msg.message, "success")
            Promise.all([loadMonthEvents(), loadStats()])
        } else if (msg.type === "zoom_job_failed") {
            toast.show(msg.message, "error")
        }
    })
})

onUnmounted(() => {
    if (removeWsListener) removeWsListener()
})
</script>

<template>
    <div class="cal-wrapper">
        <div class="cal-header">
            <div>
                <h1>Kalendarz i Spotkania</h1>
                <p class="subtitle">Zarządzaj spotkaniami i integruj z Zoom</p>
            </div>
            <div class="header-actions">
                <span v-if="zoomAccount" class="zoom-badge" title="Zoom connected">
                    <span class="material-icons" style="font-size:0.85rem">videocam</span>
                    Zoom połączony
                </span>
                <button class="btn-primary" @click="openNewEvent()">
                    <span class="material-icons">add</span> Nowe wydarzenie
                </button>
            </div>
        </div>

        <div v-if="stats" class="stats-row">
            <div class="stat-card"><span class="material-icons stat-icon" style="color:#38bdf8">event</span><div><div class="stat-value">{{ stats.total_events }}</div><div class="stat-label">Wszystkie</div></div></div>
            <div class="stat-card"><span class="material-icons stat-icon" style="color:#22c55e">upcoming</span><div><div class="stat-value">{{ stats.upcoming_events }}</div><div class="stat-label">Nadchodzące</div></div></div>
            <div class="stat-card"><span class="material-icons stat-icon" style="color:#f59e0b">today</span><div><div class="stat-value">{{ stats.today_events }}</div><div class="stat-label">Dzisiaj</div></div></div>
            <div class="stat-card"><span class="material-icons stat-icon" style="color:#818cf8">videocam</span><div><div class="stat-value">{{ stats.zoom_meetings }}</div><div class="stat-label">Zoom</div></div></div>
        </div>

        <div class="cal-tabs">
            <button :class="['tab-btn', { active: activeTab === 'calendar' }]" @click="activeTab = 'calendar'"><span class="material-icons">calendar_month</span> Kalendarz</button>
            <button :class="['tab-btn', { active: activeTab === 'zoom' }]" @click="activeTab = 'zoom'">
                <span class="material-icons">videocam</span> Zoom
                <span v-if="!zoomAccount" class="setup-dot"></span>
            </button>
            <button :class="['tab-btn', { active: activeTab === 'booking' }]" @click="activeTab = 'booking'">
                <span class="material-icons">event_available</span> Rezerwacje
            </button>
        </div>

        <div v-if="activeTab === 'calendar'" class="calendar-layout">
            <div class="calendar-main">
                <div class="month-nav">
                    <button class="nav-btn" @click="prevMonth"><span class="material-icons">chevron_left</span></button>
                    <h2 class="month-title">{{ monthName }}</h2>
                    <button class="nav-btn" @click="nextMonth"><span class="material-icons">chevron_right</span></button>
                    <button class="btn-today" @click="goToday">Dziś</button>
                </div>
                <div class="cal-grid">
                    <div class="dow-header" v-for="d in ['Pn','Wt','Śr','Cz','Pt','Sb','Nd']" :key="d">{{ d }}</div>
                    <div
                        v-for="(day, i) in calendarDays" :key="i"
                        :class="['cal-day', { other: !day.currentMonth, today: isToday(day.date), selected: selectedDate && formatDateKey(selectedDate) === formatDateKey(day.date) }]"
                        @click="selectDay(day)"
                        @dblclick="openNewEvent(day.date)"
                    >
                        <span class="day-num">{{ day.day }}</span>
                        <div class="day-events">
                            <div
                                v-for="ev in getEventsForDay(day.date).slice(0, 3)" :key="ev.id"
                                class="day-event-dot"
                                :style="{ background: ev.color || '#38bdf8' }"
                                :title="ev.title"
                                @click.stop="selectedEvent = ev"
                            >
                                <span class="dot-text">{{ ev.title.slice(0,12) }}</span>
                            </div>
                            <div v-if="getEventsForDay(day.date).length > 3" class="day-more">+{{ getEventsForDay(day.date).length - 3 }}</div>
                        </div>
                    </div>
                </div>
            </div>

            <div class="calendar-sidebar">
                <div v-if="selectedDate" class="sidebar-date">
                    <h3>{{ selectedDate.toLocaleDateString("pl-PL", { weekday: "long", day: "numeric", month: "long" }) }}</h3>
                    <button class="btn-add-sm" @click="openNewEvent(selectedDate)" title="Dodaj wydarzenie">
                        <span class="material-icons">add</span>
                    </button>
                </div>
                <div v-if="selectedDate && selectedDayEvents.length === 0" class="sidebar-empty">Brak wydarzeń w tym dniu.</div>
                <div v-else-if="selectedDate" class="sidebar-events">
                    <div v-for="ev in selectedDayEvents" :key="ev.id" class="sidebar-event" :style="{ borderLeftColor: ev.color || '#38bdf8' }" @click="selectedEvent = ev">
                        <div class="se-time">{{ formatTime(ev.start_time) }} - {{ formatTime(ev.end_time) }}</div>
                        <div class="se-title">{{ ev.title }}</div>
                        <div class="se-meta">
                            <span class="material-icons se-icon">{{ eventTypeIcon(ev.event_type) }}</span>
                            <span>{{ eventTypeLabel(ev.event_type) }}</span>
                            <span v-if="ev.zoom_join_url" class="zoom-tag"><span class="material-icons" style="font-size:0.7rem">videocam</span> Zoom</span>
                        </div>
                        <div v-if="ev.contact_name" class="se-contact">
                            <span class="material-icons" style="font-size:0.8rem">person</span> {{ ev.contact_name }}
                        </div>
                    </div>
                </div>

                <div v-if="selectedEvent" class="event-detail">
                    <div class="ed-header">
                        <h3>{{ selectedEvent.title }}</h3>
                        <div class="ed-actions">
                            <button class="btn-action-sm" @click="openEditEvent(selectedEvent)" title="Edytuj"><span class="material-icons">edit</span></button>
                            <button class="btn-action-sm btn-danger" @click="deleteEvent(selectedEvent.id)" title="Usuń"><span class="material-icons">delete</span></button>
                        </div>
                    </div>
                    <div class="ed-row"><span class="material-icons ed-ico">schedule</span> {{ formatDateFull(selectedEvent.start_time) }} - {{ formatTime(selectedEvent.end_time) }}</div>
                    <div class="ed-row"><span class="material-icons ed-ico">{{ eventTypeIcon(selectedEvent.event_type) }}</span> {{ eventTypeLabel(selectedEvent.event_type) }} &mdash; <span :class="'status-' + selectedEvent.status">{{ statusLabel(selectedEvent.status) }}</span></div>
                    <div v-if="selectedEvent.contact_name" class="ed-row"><span class="material-icons ed-ico">person</span> {{ selectedEvent.contact_name }} <span v-if="selectedEvent.contact_email" class="ed-email">&lt;{{ selectedEvent.contact_email }}&gt;</span></div>
                    <div v-if="selectedEvent.description" class="ed-desc">{{ selectedEvent.description }}</div>
                    <div v-if="selectedEvent.zoom_join_url" class="ed-zoom">
                        <span class="material-icons" style="color:#2d8cff">videocam</span>
                        <div class="ed-zoom-info">
                            <a :href="selectedEvent.zoom_join_url" target="_blank" class="zoom-link">Dołącz do spotkania Zoom</a>
                            <div v-if="selectedEvent.zoom_passcode" class="zoom-pass">Hasło: {{ selectedEvent.zoom_passcode }}</div>
                        </div>
                        <button class="btn-remove-zoom" @click="removeZoomFromEvent(selectedEvent.id)" title="Usuń spotkanie Zoom">
                            <span class="material-icons">link_off</span>
                        </button>
                    </div>
                    <div class="ed-status-actions">
                        <button v-if="selectedEvent.status === 'scheduled'" class="btn-xs-success" @click="updateEventStatus(selectedEvent.id, 'completed')"><span class="material-icons">check</span> Zakończ</button>
                        <button v-if="selectedEvent.status === 'scheduled'" class="btn-xs-cancel" @click="updateEventStatus(selectedEvent.id, 'cancelled')"><span class="material-icons">close</span> Anuluj</button>
                        <button v-if="selectedEvent.status !== 'scheduled'" class="btn-xs-reopen" @click="updateEventStatus(selectedEvent.id, 'scheduled')"><span class="material-icons">replay</span> Przywróć</button>
                    </div>
                </div>

                <div v-if="!selectedDate && !selectedEvent" class="sidebar-hint">
                    <span class="material-icons" style="font-size:2rem;color:#475569">touch_app</span>
                    <p>Kliknij dzień w kalendarzu, aby zobaczyć wydarzenia. Kliknij dwukrotnie, aby dodać nowe.</p>
                </div>
            </div>
        </div>

        <div v-if="activeTab === 'zoom'" class="zoom-section">
            <h2>Integracja Zoom</h2>
            <p class="settings-desc">Połącz konto Zoom, aby automatycznie tworzyć spotkania wideo przy planowaniu wydarzeń.</p>

            <div v-if="zoomLoading" class="loading">Ładowanie...</div>
            <template v-else>
                <div v-if="zoomAccount" class="account-status-card">
                    <div class="account-status-left">
                        <span class="material-icons account-status-icon connected">videocam</span>
                        <div>
                            <div class="account-provider">ZOOM</div>
                            <div class="account-email">Account ID: {{ zoomAccount.account_id }}</div>
                        </div>
                    </div>
                </div>

                <div class="settings-form-card">
                    <h3>{{ zoomAccount ? 'Zaktualizuj dane Zoom' : 'Połącz konto Zoom' }}</h3>
                    <p class="zoom-instructions">Użyj <strong>Server-to-Server OAuth</strong> z <a href="https://marketplace.zoom.us/develop/create" target="_blank" class="help-link">Zoom Marketplace</a>. Potrzebujesz: Account ID, Client ID i Client Secret.</p>

                    <div class="form-group">
                        <label>Account ID *</label>
                        <input v-model="zoomForm.accountId" placeholder="Account ID z Zoom Marketplace" />
                    </div>
                    <div class="form-group">
                        <label>Client ID *</label>
                        <input v-model="zoomForm.clientId" placeholder="Client ID aplikacji OAuth" />
                    </div>
                    <div class="form-group">
                        <label>Client Secret *</label>
                        <input v-model="zoomForm.clientSecret" type="password" placeholder="Client Secret aplikacji OAuth" />
                    </div>

                    <p v-if="formError" class="err-msg">{{ formError }}</p>

                    <div class="settings-actions">
                        <button class="btn-primary" @click="saveZoomAccount" :disabled="zoomSaving">
                            <span class="material-icons">save</span> {{ zoomSaving ? 'Zapisywanie...' : (zoomAccount ? 'Zaktualizuj' : 'Połącz') }}
                        </button>
                        <button v-if="zoomAccount" class="btn-secondary" @click="testZoom" :disabled="zoomTesting">
                            <span class="material-icons">wifi_tethering</span> {{ zoomTesting ? 'Testowanie...' : 'Testuj połączenie' }}
                        </button>
                        <button v-if="zoomAccount" class="btn-danger-outline" @click="disconnectZoom">
                            <span class="material-icons">link_off</span> Rozłącz
                        </button>
                    </div>

                    <div v-if="zoomTestResult" class="test-result">
                        <div :class="['test-item', zoomTestResult.status === 'ok' ? 'test-ok' : 'test-fail']">
                            <span class="material-icons">{{ zoomTestResult.status === 'ok' ? 'check_circle' : 'error' }}</span>
                            {{ zoomTestResult.status === 'ok' ? 'Połączono z Zoom' : 'Błąd połączenia' }}
                        </div>
                        <div v-if="zoomTestResult.email" class="test-email">Konto: {{ zoomTestResult.email }}</div>
                        <div v-if="zoomTestResult.error" class="test-error">{{ zoomTestResult.error }}</div>
                    </div>
                </div>

                <div class="settings-info-card">
                    <span class="material-icons info-icon">info</span>
                    <div>
                        <h4>Jak skonfigurować Zoom?</h4>
                        <ul>
                            <li>Zaloguj się do <a href="https://marketplace.zoom.us" target="_blank">Zoom Marketplace</a></li>
                            <li>Kliknij <strong>Develop</strong> &rarr; <strong>Build App</strong></li>
                            <li>Wybierz typ <strong>Server-to-Server OAuth</strong></li>
                            <li>Skopiuj <strong>Account ID</strong>, <strong>Client ID</strong> i <strong>Client Secret</strong></li>
                            <li>Dodaj uprawnienia (scopes): <code>meeting:write:admin</code>, <code>meeting:read:admin</code>, <code>user:read:admin</code></li>
                            <li>Aktywuj aplikację i wklej dane powyżej</li>
                        </ul>
                    </div>
                </div>
            </template>
        </div>

        <div v-if="activeTab === 'booking'" class="zoom-section">
            <h2>Rezerwacje online</h2>
            <p class="settings-desc">Udostępnij klientom stronę, na której sami umówią spotkanie. Każda rezerwacja tworzy wydarzenie w tym kalendarzu i blokuje zajętą godzinę.</p>

            <div v-if="bookingLoading" class="loading">Ładowanie...</div>
            <template v-else>
                <div class="settings-form-card">
                    <label class="toggle-label" style="margin-bottom:18px">
                        <input type="checkbox" v-model="bookingForm.enabled" class="toggle-input" />
                        <span class="toggle-track"><span class="toggle-thumb"></span></span>
                        <span class="toggle-text">Rezerwacje włączone</span>
                    </label>

                    <div class="form-group">
                        <label>Dni dostępne na spotkania</label>
                        <div class="weekday-row">
                            <button v-for="d in WEEKDAYS" :key="d.v" type="button"
                                :class="['weekday-chip', { active: bookingForm.workingDays.includes(d.v) }]"
                                @click="toggleWorkingDay(d.v)">{{ d.l }}</button>
                        </div>
                    </div>

                    <div class="form-row">
                        <div class="form-group flex-1">
                            <label>Godzina od</label>
                            <input v-model="bookingForm.startTime" type="time" />
                        </div>
                        <div class="form-group flex-1">
                            <label>Godzina do</label>
                            <input v-model="bookingForm.endTime" type="time" />
                        </div>
                        <div class="form-group flex-1">
                            <label>Długość spotkania</label>
                            <select v-model="bookingForm.slotMinutes">
                                <option :value="15">15 min</option>
                                <option :value="30">30 min</option>
                                <option :value="60">60 min</option>
                            </select>
                        </div>
                    </div>

                    <div class="form-group">
                        <label>Tytuł tworzonego wydarzenia (&#123;name&#125; = imię klienta)</label>
                        <input v-model="bookingForm.meetingTitle" placeholder="Spotkanie: {name}" />
                    </div>

                    <p v-if="bookingError" class="err-msg">{{ bookingError }}</p>

                    <div class="settings-actions">
                        <button class="btn-primary" @click="saveBookingSettings" :disabled="bookingSaving">
                            <span class="material-icons">save</span> {{ bookingSaving ? 'Zapisywanie...' : 'Zapisz' }}
                        </button>
                    </div>
                </div>

                <div class="settings-info-card">
                    <span class="material-icons info-icon">link</span>
                    <div>
                        <h4>Link do strony rezerwacji</h4>
                        <ul>
                            <li>Dodaj na stronie lejka przycisk z blokiem <strong>„Umów spotkanie”</strong> w kreatorze (albo zwykły link).</li>
                            <li>Przycisk prowadzi do ścieżki <code>/umow-spotkanie</code> w domenie lejka, np. <code>https://twoj-lejek.bizmopol.localhost/umow-spotkanie</code></li>
                            <li>Klient wybiera wolny termin i klika „Umów się” — wydarzenie pojawi się w tym kalendarzu.</li>
                        </ul>
                    </div>
                </div>
            </template>
        </div>

        <!-- Event Modal -->
        <div v-if="showEventModal" class="modal-overlay" @click.self="showEventModal = false">
            <div class="modal modal-lg">
                <div class="modal-header">
                    <h2>{{ editingEventId ? 'Edytuj wydarzenie' : 'Nowe wydarzenie' }}</h2>
                    <button class="modal-close" @click="showEventModal = false"><span class="material-icons">close</span></button>
                </div>

                <div class="event-type-row">
                    <button v-for="t in [{v:'meeting',l:'Spotkanie',i:'videocam'},{v:'call',l:'Rozmowa',i:'phone'},{v:'follow_up',l:'Follow-up',i:'replay'},{v:'other',l:'Inne',i:'event'}]" :key="t.v"
                        :class="['etype-btn', { active: eventForm.eventType === t.v }]"
                        @click="eventForm.eventType = t.v"
                    >
                        <span class="material-icons">{{ t.i }}</span> {{ t.l }}
                    </button>
                </div>

                <div class="form-group">
                    <label>Tytuł *</label>
                    <input v-model="eventForm.title" placeholder="np. Spotkanie z klientem" />
                </div>

                <div class="form-row">
                    <div class="form-group flex-1">
                        <label>Data rozpoczęcia *</label>
                        <input v-model="eventForm.startDate" type="date" />
                    </div>
                    <div class="form-group w-120">
                        <label>Godzina *</label>
                        <input v-model="eventForm.startTime" type="time" />
                    </div>
                    <div class="form-group flex-1">
                        <label>Data zakończenia</label>
                        <input v-model="eventForm.endDate" type="date" />
                    </div>
                    <div class="form-group w-120">
                        <label>Godzina</label>
                        <input v-model="eventForm.endTime" type="time" />
                    </div>
                </div>

                <div class="form-group">
                    <label>Opis</label>
                    <textarea v-model="eventForm.description" rows="3" placeholder="Dodatkowe informacje..."></textarea>
                </div>

                <div class="form-group">
                    <label>Kontakt (opcjonalnie)</label>
                    <div class="contact-picker">
                        <div v-if="eventForm.contactName" class="selected-contact">
                            <div class="sc-avatar" :style="{ background: avatarColor(eventForm.contactName) }">{{ initials(eventForm.contactName) }}</div>
                            <div>
                                <div class="sc-name">{{ eventForm.contactName }}</div>
                                <div class="sc-email">{{ eventForm.contactEmail }}</div>
                            </div>
                            <button class="btn-clear-contact" @click="eventForm.contactId = ''; eventForm.contactName = ''; eventForm.contactEmail = ''"><span class="material-icons">close</span></button>
                        </div>
                        <template v-else>
                            <input v-model="contactSearch" @input="loadContacts" placeholder="Szukaj kontaktu..." class="contact-search" />
                            <div v-if="contacts.length > 0" class="contact-dropdown">
                                <div v-for="c in contacts.slice(0, 8)" :key="c.id" class="cd-item" @click="selectContact(c)">
                                    <div class="cd-avatar" :style="{ background: avatarColor(c.name) }">{{ initials(c.name) }}</div>
                                    <div><div class="cd-name">{{ c.name }}</div><div class="cd-email">{{ c.email || 'Brak emaila' }}</div></div>
                                </div>
                            </div>
                        </template>
                    </div>
                </div>

                <div v-if="zoomAccount" class="zoom-toggle">
                    <template v-if="!editingEventId">
                        <label class="toggle-label">
                            <input type="checkbox" v-model="eventForm.createZoom" class="toggle-input" />
                            <span class="toggle-track"><span class="toggle-thumb"></span></span>
                            <span class="toggle-text"><span class="material-icons" style="font-size:1rem;vertical-align:middle;color:#2d8cff">videocam</span> Utwórz spotkanie Zoom automatycznie</span>
                        </label>
                    </template>
                    <template v-else-if="eventForm.hasZoom">
                        <div class="zoom-edit-info">
                            <span class="material-icons" style="color:#2d8cff">videocam</span>
                            <span class="zoom-edit-text">Spotkanie Zoom jest podłączone</span>
                        </div>
                        <label class="toggle-label toggle-danger">
                            <input type="checkbox" v-model="eventForm.removeZoom" class="toggle-input" />
                            <span class="toggle-track"><span class="toggle-thumb"></span></span>
                            <span class="toggle-text"><span class="material-icons" style="font-size:1rem;vertical-align:middle;color:#f87171">link_off</span> Usuń spotkanie Zoom</span>
                        </label>
                    </template>
                    <template v-else>
                        <label class="toggle-label">
                            <input type="checkbox" v-model="eventForm.addZoom" class="toggle-input" />
                            <span class="toggle-track"><span class="toggle-thumb"></span></span>
                            <span class="toggle-text"><span class="material-icons" style="font-size:1rem;vertical-align:middle;color:#2d8cff">videocam</span> Dodaj spotkanie Zoom</span>
                        </label>
                    </template>
                </div>

                <p v-if="formError" class="err-msg">{{ formError }}</p>
                <div class="modal-actions">
                    <button class="btn-secondary" @click="showEventModal = false">Anuluj</button>
                    <button class="btn-primary" @click="saveEvent">
                        <span class="material-icons">{{ editingEventId ? 'save' : 'add' }}</span>
                        {{ editingEventId ? 'Zapisz' : 'Utwórz' }}
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<style scoped>
.cal-wrapper { width: 100%; max-width: 1400px; margin: 0 auto; }
.cal-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 20px; }
.cal-header h1 { font-size: 2rem; color: #f8fafc; margin-bottom: 4px; }
.subtitle { color: #94a3b8; font-size: 0.95rem; }
.header-actions { display: flex; gap: 10px; align-items: center; }

.btn-primary { padding: 10px 18px; background: #38bdf8; border: none; border-radius: 8px; color: #0f172a; font-weight: bold; cursor: pointer; transition: 0.2s; display: flex; align-items: center; gap: 6px; font-size: 0.9rem; }
.btn-primary:hover { background: #7dd3fc; }
.btn-primary .material-icons { font-size: 1.5rem; }
.btn-secondary { padding: 10px 18px; background: transparent; border: 1px solid #475569; border-radius: 8px; color: #94a3b8; cursor: pointer; transition: 0.2s; font-size: 0.9rem; }
.btn-secondary:hover { border-color: #f1f5f9; color: #f1f5f9; }

.zoom-badge { display: flex; align-items: center; gap: 4px; color: #2d8cff; font-size: 0.82rem; padding: 6px 12px; background: #1e3a5f22; border: 1px solid #1e3a5f; border-radius: 8px; }
.setup-dot { width: 8px; height: 8px; border-radius: 50%; background: #f59e0b; display: inline-block; margin-left: 4px; }

.stats-row { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 14px; margin-bottom: 24px; }
.stat-card { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 16px; display: flex; align-items: center; gap: 14px; }
.stat-icon { font-size: 2rem; }
.stat-value { font-size: 1.8rem; font-weight: 800; color: #f1f5f9; line-height: 1; }
.stat-label { font-size: 0.78rem; color: #64748b; margin-top: 2px; }

.cal-tabs { display: flex; gap: 2px; margin-bottom: 20px; border-bottom: 1px solid #334155; }
.tab-btn { padding: 10px 18px; background: transparent; border: none; border-bottom: 2px solid transparent; color: #94a3b8; cursor: pointer; font-size: 0.9rem; transition: 0.2s; display: flex; align-items: center; gap: 6px; margin-bottom: -1px; }
.tab-btn .material-icons { font-size: 1rem; }
.tab-btn:hover { color: #f1f5f9; }
.tab-btn.active { color: #38bdf8; border-bottom-color: #38bdf8; }

.calendar-layout { display: flex; gap: 0; border: 1px solid #334155; border-radius: 12px; overflow: hidden; min-height: 600px; }
.calendar-main { flex: 1; }
.calendar-sidebar { width: 340px; min-width: 340px; border-left: 1px solid #334155; background: #0f172a; padding: 16px; overflow-y: auto; max-height: 700px; }

.month-nav { display: flex; align-items: center; gap: 8px; padding: 14px 16px; background: #1e293b; border-bottom: 1px solid #334155; }
.month-title { color: #f1f5f9; font-size: 1.1rem; flex: 1; text-align: center; }
.nav-btn { background: none; border: 1px solid #334155; color: #94a3b8; border-radius: 6px; padding: 4px 8px; cursor: pointer; display: flex; align-items: center; }
.nav-btn:hover { color: #38bdf8; border-color: #38bdf8; }
.btn-today { background: #1e293b; border: 1px solid #334155; color: #94a3b8; border-radius: 6px; padding: 4px 12px; cursor: pointer; font-size: 0.82rem; }
.btn-today:hover { color: #38bdf8; border-color: #38bdf8; }

.cal-grid { display: grid; grid-template-columns: repeat(7, 1fr); }
.dow-header { padding: 8px; text-align: center; font-size: 0.78rem; color: #64748b; font-weight: 600; background: #1e293b; border-bottom: 1px solid #334155; }
.cal-day { min-height: 90px; padding: 4px 6px; border-bottom: 1px solid #1e293b; border-right: 1px solid #1e293b; cursor: pointer; transition: 0.1s; }
.cal-day:hover { background: #1e293b; }
.cal-day.other { opacity: 0.35; }
.cal-day.today { background: #0f2a4a; }
.cal-day.today .day-num { color: #38bdf8; font-weight: 700; }
.cal-day.selected { background: #1e293b; outline: 2px solid #38bdf8; outline-offset: -2px; }
.day-num { font-size: 0.82rem; color: #94a3b8; display: block; margin-bottom: 2px; }
.day-events { display: flex; flex-direction: column; gap: 2px; }
.day-event-dot { padding: 1px 4px; border-radius: 3px; cursor: pointer; }
.dot-text { font-size: 0.68rem; color: #0f172a; font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; display: block; max-width: 100%; }
.day-more { font-size: 0.68rem; color: #64748b; padding: 1px 4px; }

.sidebar-date { display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px; }
.sidebar-date h3 { color: #f1f5f9; font-size: 0.95rem; text-transform: capitalize; }
.btn-add-sm { background: #38bdf8; border: none; color: #0f172a; border-radius: 6px; padding: 4px; cursor: pointer; display: flex; align-items: center; }
.btn-add-sm .material-icons { font-size: 1rem; }
.sidebar-empty { color: #64748b; font-size: 0.88rem; padding: 20px 0; text-align: center; }
.sidebar-hint { text-align: center; color: #475569; padding: 30px 10px; display: flex; flex-direction: column; align-items: center; gap: 12px; }
.sidebar-hint p { font-size: 0.88rem; line-height: 1.5; }
.sidebar-events { display: flex; flex-direction: column; gap: 8px; margin-bottom: 16px; }
.sidebar-event { background: #1e293b; border: 1px solid #334155; border-left: 3px solid; border-radius: 8px; padding: 10px 12px; cursor: pointer; transition: 0.15s; }
.sidebar-event:hover { background: #334155; }
.se-time { font-size: 0.78rem; color: #38bdf8; font-weight: 600; margin-bottom: 2px; }
.se-title { font-size: 0.9rem; color: #f1f5f9; font-weight: 500; margin-bottom: 4px; }
.se-meta { display: flex; align-items: center; gap: 4px; font-size: 0.78rem; color: #64748b; }
.se-icon { font-size: 0.85rem; }
.zoom-tag { display: inline-flex; align-items: center; gap: 2px; color: #2d8cff; background: #1e3a5f44; padding: 1px 6px; border-radius: 4px; margin-left: 6px; }
.se-contact { font-size: 0.78rem; color: #94a3b8; margin-top: 4px; display: flex; align-items: center; gap: 4px; }

.event-detail { background: #1e293b; border: 1px solid #334155; border-radius: 10px; padding: 16px; margin-top: 16px; }
.ed-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 12px; }
.ed-header h3 { color: #f1f5f9; font-size: 1rem; }
.ed-actions { display: flex; gap: 4px; }
.btn-action-sm { background: none; border: none; cursor: pointer; padding: 4px 6px; border-radius: 6px; transition: 0.15s; display: flex; align-items: center; color: #94a3b8; }
.btn-action-sm .material-icons { font-size: 1rem; }
.btn-action-sm:hover { background: #334155; color: #f1f5f9; }
.btn-action-sm.btn-danger:hover { background: #3f1414; color: #f87171; }
.ed-row { display: flex; align-items: center; gap: 8px; font-size: 0.85rem; color: #94a3b8; margin-bottom: 6px; }
.ed-ico { font-size: 1rem; color: #64748b; }
.ed-email { color: #64748b; font-size: 0.82rem; }
.ed-desc { color: #94a3b8; font-size: 0.85rem; margin-top: 10px; padding-top: 10px; border-top: 1px solid #334155; line-height: 1.5; }
.zoom-link { color: #2d8cff; font-size: 0.88rem; text-decoration: none; font-weight: 500; }
.zoom-link:hover { text-decoration: underline; }
.zoom-pass { color: #64748b; font-size: 0.78rem; margin-top: 2px; }
.ed-status-actions { display: flex; gap: 8px; margin-top: 12px; }
.btn-xs-success { padding: 4px 10px; background: #064e3b; border: 1px solid #065f46; color: #34d399; border-radius: 6px; cursor: pointer; font-size: 0.78rem; display: flex; align-items: center; gap: 4px; }
.btn-xs-success .material-icons { font-size: 0.85rem; }
.btn-xs-cancel { padding: 4px 10px; background: #3f1414; border: 1px solid #7f1d1d; color: #f87171; border-radius: 6px; cursor: pointer; font-size: 0.78rem; display: flex; align-items: center; gap: 4px; }
.btn-xs-cancel .material-icons { font-size: 0.85rem; }
.btn-xs-reopen { padding: 4px 10px; background: #1e293b; border: 1px solid #334155; color: #94a3b8; border-radius: 6px; cursor: pointer; font-size: 0.78rem; display: flex; align-items: center; gap: 4px; }
.btn-xs-reopen .material-icons { font-size: 0.85rem; }

.status-scheduled { color: #38bdf8; }
.status-completed { color: #34d399; }
.status-cancelled { color: #f87171; }

.zoom-section { padding: 4px 0; max-width: 800px; }
.zoom-section h2 { color: #f8fafc; font-size: 1.3rem; margin-bottom: 6px; }
.settings-desc { color: #94a3b8; font-size: 0.92rem; margin-bottom: 24px; }
.zoom-instructions { color: #94a3b8; font-size: 0.88rem; margin-bottom: 16px; line-height: 1.5; }
.zoom-instructions strong { color: #e2e8f0; }

.account-status-card { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 20px; margin-bottom: 20px; display: flex; justify-content: space-between; align-items: center; }
.account-status-left { display: flex; align-items: center; gap: 14px; }
.account-status-icon { font-size: 2rem; }
.account-status-icon.connected { color: #2d8cff; }
.account-provider { color: #f1f5f9; font-weight: 700; font-size: 1.1rem; }
.account-email { color: #94a3b8; font-size: 0.9rem; }

.settings-form-card { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 24px; margin-bottom: 20px; }
.settings-form-card h3 { color: #38bdf8; font-size: 1.1rem; margin-bottom: 16px; }
.settings-actions { display: flex; gap: 10px; margin-top: 20px; flex-wrap: wrap; }
.btn-danger-outline { padding: 10px 18px; background: transparent; border: 1px solid #7f1d1d; border-radius: 8px; color: #f87171; cursor: pointer; transition: 0.2s; font-size: 0.9rem; display: flex; align-items: center; gap: 6px; }
.btn-danger-outline:hover { background: #3f1414; border-color: #f87171; }
.btn-danger-outline .material-icons { font-size: 1rem; }
.help-link { color: #38bdf8; font-size: inherit; text-decoration: none; }
.help-link:hover { text-decoration: underline; }

.test-result { margin-top: 16px; padding: 16px; background: #0f172a; border: 1px solid #334155; border-radius: 10px; display: flex; flex-direction: column; gap: 6px; }
.test-item { display: flex; align-items: center; gap: 8px; font-size: 0.9rem; font-weight: 500; }
.test-item .material-icons { font-size: 1.1rem; }
.test-ok { color: #34d399; }
.test-fail { color: #f87171; }
.test-email { color: #94a3b8; font-size: 0.85rem; }
.test-error { color: #f87171; font-size: 0.82rem; margin-top: 4px; padding-top: 6px; border-top: 1px solid #334155; }

.settings-info-card { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 20px; display: flex; gap: 16px; align-items: flex-start; }
.info-icon { color: #38bdf8; font-size: 1.5rem; flex-shrink: 0; margin-top: 2px; }
.settings-info-card h4 { color: #f1f5f9; margin-bottom: 10px; font-size: 0.95rem; }
.settings-info-card ul { color: #94a3b8; font-size: 0.85rem; padding-left: 18px; line-height: 1.8; }
.settings-info-card strong { color: #e2e8f0; }
.settings-info-card code { background: #0f172a; color: #38bdf8; padding: 2px 6px; border-radius: 4px; font-size: 0.8rem; }
.settings-info-card a { color: #38bdf8; text-decoration: none; }
.settings-info-card a:hover { text-decoration: underline; }

.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.65); display: flex; align-items: center; justify-content: center; z-index: 100; }
.modal { background: #1e293b; border: 1px solid #334155; border-radius: 16px; padding: 28px; width: 100%; max-width: 520px; max-height: 90vh; overflow-y: auto; }
.modal-lg { max-width: 680px; }
.modal h2 { color: #38bdf8; font-size: 1.3rem; }
.modal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.modal-header h2 { margin-bottom: 0; }
.modal-close { background: none; border: none; color: #94a3b8; cursor: pointer; padding: 4px; }
.modal-close .material-icons { font-size: 1.3rem; }
.modal-actions { display: flex; gap: 10px; justify-content: flex-end; margin-top: 20px; }

.event-type-row { display: flex; gap: 6px; margin-bottom: 18px; flex-wrap: wrap; }
.etype-btn { padding: 8px 14px; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #94a3b8; cursor: pointer; transition: 0.2s; font-size: 0.85rem; display: flex; align-items: center; gap: 6px; }
.etype-btn:hover { border-color: #475569; }
.etype-btn.active { border-color: #38bdf8; background: #0f2a4a; color: #38bdf8; }
.etype-btn .material-icons { font-size: 1rem; }

.form-group { margin-bottom: 14px; }
.form-group label { display: block; margin-bottom: 6px; color: #94a3b8; font-size: 0.85rem; }
.form-group input, .form-group select, .form-group textarea { width: 100%; padding: 10px 12px; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #f1f5f9; outline: none; font-size: 0.9rem; resize: vertical; }
.form-group input:focus, .form-group select:focus, .form-group textarea:focus { border-color: #38bdf8; }
.form-row { display: flex; gap: 12px; }
.flex-1 { flex: 1; }
.w-120 { width: 120px; }
.err-msg { color: #f87171; font-size: 0.88rem; margin-bottom: 10px; }
.loading { color: #94a3b8; padding: 40px; text-align: center; }

.contact-picker { position: relative; }
.contact-search { width: 100%; padding: 10px 12px; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #f1f5f9; outline: none; font-size: 0.9rem; }
.contact-search:focus { border-color: #38bdf8; }
.contact-dropdown { position: absolute; top: 100%; left: 0; right: 0; background: #1e293b; border: 1px solid #334155; border-radius: 8px; max-height: 200px; overflow-y: auto; z-index: 10; margin-top: 4px; }
.cd-item { padding: 8px 12px; display: flex; align-items: center; gap: 10px; cursor: pointer; transition: 0.1s; }
.cd-item:hover { background: #334155; }
.cd-avatar { width: 28px; height: 28px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 0.68rem; font-weight: 700; color: #0f172a; flex-shrink: 0; }
.cd-name { color: #e2e8f0; font-size: 0.85rem; }
.cd-email { color: #64748b; font-size: 0.75rem; }
.selected-contact { display: flex; align-items: center; gap: 10px; padding: 8px 12px; background: #0f172a; border: 1px solid #334155; border-radius: 8px; }
.sc-avatar { width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 0.75rem; font-weight: 700; color: #0f172a; flex-shrink: 0; }
.sc-name { color: #f1f5f9; font-size: 0.9rem; font-weight: 500; }
.sc-email { color: #64748b; font-size: 0.78rem; }
.btn-clear-contact { background: none; border: none; color: #64748b; cursor: pointer; margin-left: auto; padding: 2px; }
.btn-clear-contact:hover { color: #f87171; }
.btn-clear-contact .material-icons { font-size: 0.9rem; }

.zoom-toggle { margin-top: 4px; margin-bottom: 10px; }
.toggle-label { display: flex; align-items: center; gap: 10px; cursor: pointer; }
.toggle-input { display: none; }
.toggle-track { width: 44px; height: 24px; background: #334155; border-radius: 12px; position: relative; transition: 0.2s; flex-shrink: 0; }
.toggle-input:checked + .toggle-track { background: #2d8cff; }
.toggle-thumb { position: absolute; top: 3px; left: 3px; width: 18px; height: 18px; background: white; border-radius: 50%; transition: 0.2s; }
.toggle-input:checked + .toggle-track .toggle-thumb { left: 23px; }
.toggle-text { color: #e2e8f0; font-size: 0.9rem; display: flex; align-items: center; gap: 4px; }

.toggle-danger .toggle-input:checked + .toggle-track { background: #ef4444; }

.btn-remove-zoom { background: none; border: none; color: #64748b; cursor: pointer; padding: 4px; border-radius: 6px; margin-left: auto; flex-shrink: 0; }
.btn-remove-zoom:hover { color: #f87171; background: #3f1414; }
.btn-remove-zoom .material-icons { font-size: 1rem; }

.ed-zoom { display: flex; align-items: center; gap: 10px; margin-top: 12px; padding: 10px; background: #0f172a; border-radius: 8px; border: 1px solid #1e3a5f; }
.ed-zoom-info { flex: 1; }

.zoom-edit-info { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; color: #94a3b8; font-size: 0.88rem; }
.zoom-edit-text { color: #e2e8f0; }

.weekday-row { display: flex; gap: 8px; flex-wrap: wrap; }
.weekday-chip { width: 44px; height: 40px; border-radius: 8px; border: 1px solid #334155; background: #0f172a; color: #94a3b8; cursor: pointer; font-size: 0.85rem; transition: 0.15s; }
.weekday-chip:hover { border-color: #475569; }
.weekday-chip.active { background: #0f2a4a; border-color: #38bdf8; color: #38bdf8; font-weight: 700; }
.form-group select { width: 100%; padding: 10px 12px; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #f1f5f9; outline: none; font-size: 0.9rem; }
.form-group select:focus { border-color: #38bdf8; }
</style>
