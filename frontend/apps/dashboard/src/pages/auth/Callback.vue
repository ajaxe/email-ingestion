<template>
  <div>Loading...</div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuth } from '@/composables/useAuth'

// Get the callback handler from our composable
const { handleLoginCallback } = useAuth()

// When the component mounts, run the callback handler
onMounted(() => {
  handleLoginCallback()
})
</script>
