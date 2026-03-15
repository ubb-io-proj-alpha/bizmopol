<script setup lang="ts">
import { ref, onMounted } from 'vue'

const message = ref('Loading...')
onMounted(async () => {
  const response = await fetch('/api/v1/tests/');
  if (response.ok) {
    message.value = "communication is working";
  } else {
    message.value = "communication is not working";
  }
})
</script>

<template>
  <main>
    <h1> {{ message }} </h1>
  </main>
</template>
