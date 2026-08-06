<template>
  <HelloWorld />
</template>

<route lang="yaml">
meta:
  requiresAuth: true
  layout: DashboardLayout
</route>

<script setup>
  import HelloWorld from '@/components/HelloWorld.vue'
</script>
