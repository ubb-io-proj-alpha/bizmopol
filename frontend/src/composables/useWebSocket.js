import { ref, onMounted, onUnmounted } from "vue"

const wsStatus = ref("disconnected")
const listeners = []
let socket = null
let reconnectTimer = null
let refCount = 0

function connectWebSocket() {
    const token = localStorage.getItem("jwt_token")
    if (!token) return

    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) return

    const proto = location.protocol === "https:" ? "wss:" : "ws:"
    const wsUrl = `${proto}//${location.host}/api/v1/ws?token=${encodeURIComponent(token)}`

    socket = new WebSocket(wsUrl)

    socket.onopen = () => {
        wsStatus.value = "connected"
    }

    socket.onmessage = (event) => {
        try {
            const msg = JSON.parse(event.data)
            listeners.forEach(fn => fn(msg))
        } catch (_) {}
    }

    socket.onclose = () => {
        wsStatus.value = "disconnected"
        if (refCount > 0) {
            reconnectTimer = setTimeout(connectWebSocket, 5000)
        }
    }

    socket.onerror = () => {
        socket.close()
    }
}

function disconnect() {
    if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
    if (socket) {
        socket.onclose = null
        socket.close()
        socket = null
    }
    wsStatus.value = "disconnected"
}

export function useWebSocket() {
    const onMessage = (fn) => {
        listeners.push(fn)
        return () => {
            const idx = listeners.indexOf(fn)
            if (idx !== -1) listeners.splice(idx, 1)
        }
    }

    onMounted(() => {
        refCount++
        if (refCount === 1) connectWebSocket()
    })

    onUnmounted(() => {
        refCount--
        if (refCount === 0) disconnect()
    })

    return { wsStatus, onMessage }
}
