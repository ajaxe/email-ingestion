<script setup>
import { useRoute } from 'vue-router'
import { computed } from 'vue'
import NoLayout from '@/layouts/NoLayout.vue';

const route = useRoute()
const layouts = { NoLayout }

// Fallback to DefaultLayout if no layout is specified in meta
const layout = computed(() => layouts[route.meta.layout] || NoLayout)
</script>

<template>
  <component :is="layout">
    <RouterView />
  </component>
</template>