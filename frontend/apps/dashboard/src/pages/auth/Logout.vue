<template>
  <div>Loading...</div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const store = useAuthStore()

// When the component mounts, run the callback handler
onMounted(() => {
  store.handleLogoutCallback()
})
</script>
