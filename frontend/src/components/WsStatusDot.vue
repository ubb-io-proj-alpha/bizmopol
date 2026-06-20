<script setup>
import { useWebSocket } from "../composables/useWebSocket.js"
const { wsStatus } = useWebSocket()
</script>

<template>
    <span class="ws-dot" :class="wsStatus" :title="wsStatus === 'connected' ? 'WebSocket połączony' : 'WebSocket rozłączony'"></span>
</template>

<style scoped>
.ws-dot {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
    transition: background 0.3s;
}
.ws-dot.connected { background: #22c55e; box-shadow: 0 0 4px #22c55e88; }
.ws-dot.disconnected { background: #64748b; }
</style>
