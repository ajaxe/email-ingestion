<script setup>
import { computed, onMounted, onUnmounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { apiEvents, API_EVENT_TYPES } from '@/services/apiEvents';
import { useAuthStore } from '@/stores/auth';
import NoLayout from '@/layouts/NoLayout.vue';
import DashboardLayout from '@/layouts/DashboardLayout.vue';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();

const layouts = { NoLayout, DashboardLayout };

// Default to DashboardLayout if layout is not explicitly set in route meta
const layout = computed(() => {
  if (route?.meta?.layout && layouts[route.meta.layout]) {
    return layouts[route.meta.layout];
  }
  return DashboardLayout;
});

let isLoggingOut = false;

async function handleTokenInvalid() {
  if (isLoggingOut) return;
  isLoggingOut = true;
  try {
    await authStore.logout();
  } catch (err) {
    console.error("Error during forced logout:", err);
    router.push("/login");
  } finally {
    isLoggingOut = false;
  }
}

function handleForbiddenUser() {
  if (route.name !== 'need-invitation') {
    router.push({ name: 'need-invitation' });
  }
}

onMounted(() => {
  apiEvents.addEventListener(API_EVENT_TYPES.FORBIDDEN_TOKEN_INVALID, handleTokenInvalid);
  apiEvents.addEventListener(API_EVENT_TYPES.FORBIDDEN_USER, handleForbiddenUser);
});

onUnmounted(() => {
  apiEvents.removeEventListener(API_EVENT_TYPES.FORBIDDEN_TOKEN_INVALID, handleTokenInvalid);
  apiEvents.removeEventListener(API_EVENT_TYPES.FORBIDDEN_USER, handleForbiddenUser);
});
</script>

<template>
  <component :is="layout">
    <RouterView />
  </component>
</template>