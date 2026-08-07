<script setup>
import { useRoute } from 'vue-router';
import { computed } from 'vue';
import NoLayout from '@/layouts/NoLayout.vue';
import DashboardLayout from '@/layouts/DashboardLayout.vue';

const route = useRoute();
const layouts = { NoLayout, DashboardLayout };

// Default to DashboardLayout if layout is not explicitly set in route meta
const layout = computed(() => {
  if (route?.meta?.layout && layouts[route.meta.layout]) {
    return layouts[route.meta.layout];
  }
  return DashboardLayout;
});
</script>

<template>
  <component :is="layout">
    <RouterView />
  </component>
</template>