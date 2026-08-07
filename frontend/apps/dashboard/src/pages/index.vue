<template>
  <DashboardStats />
</template>

<route lang="yaml">
meta:
  requiresAuth: true
  layout: DashboardLayout
</route>

<script setup>
  import DashboardStats from '@/components/dashboard/DashboardStats.vue'
</script>
