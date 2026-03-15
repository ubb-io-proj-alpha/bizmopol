<script setup lang="ts">
import { ref, onMounted } from 'vue'

const message = ref('Loading...')
onMounted(async () => {
  const response = await fetch('/api/hello')
  const data = await response.json()
  message.value = data.message
})
</script>

<template>
  <main>
    <h1> {{ message }} </h1>
  </main>
</template>
