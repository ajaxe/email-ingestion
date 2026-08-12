<template>
  <div class="pa-2">
    <div class="d-flex align-center justify-space-between mb-2 px-2">
      <span class="text-caption text-medium-emphasis font-weight-medium"
        >Theme</span
      >
      <v-btn
        variant="text"
        icon
        :color="theme.global.current.value.dark ? 'amber' : 'primary'"
        @click="toggleTheme"
      >
        <v-icon icon="mdi-theme-light-dark" />
        <v-tooltip activator="parent" location="top">
          Switch to
          {{ theme.global.current.value.dark ? "Light" : "Dark" }} Mode
        </v-tooltip>
      </v-btn>
      <v-btn variant="text" icon>
        <v-icon icon="mdi-pin-outline" />
      </v-btn>
    </div>

    <v-divider class="mb-2" />

    <v-menu location="top start" offset="8">
      <template #activator="{ props }">
        <v-list-item
          v-bind="props"
          :prepend-avatar="profileImage"
          :title="username"
          :subtitle="email"
          class="rounded-lg cursor-pointer"
        >
          <template #append>
            <v-icon icon="mdi-dots-vertical" size="small" />
          </template>
        </v-list-item>
      </template>

      <v-card min-width="240" rounded="lg" elevation="6">
        <v-card-item class="py-3">
          <v-card-title class="text-subtitle-2 font-weight-bold">
            {{ username }}
          </v-card-title>
          <v-card-subtitle class="text-caption text-truncate" v-if="email">
            {{ email }}
          </v-card-subtitle>
        </v-card-item>

        <v-divider />

        <v-list density="compact" nav class="pa-2">
          <v-list-item
            prepend-icon="mdi-theme-light-dark"
            title="Toggle Theme"
            @click="toggleTheme"
          />
          <v-list-item
            prepend-icon="mdi-logout"
            title="Sign Out"
            color="error"
            @click="handleLogout"
          />
        </v-list>
      </v-card>
    </v-menu>
  </div>
</template>

<script setup>
import { computed } from "vue";
import { useTheme } from "vuetify";
import { useAuthStore } from "@/stores/auth";
import { useRouter } from "vue-router";

const theme = useTheme();
const authStore = useAuthStore();
const router = useRouter();

const profileImage = computed(() => authStore.user.profileImage);
const username = computed(() => authStore.user.name);
const email = computed(() => authStore.user.email);

function toggleTheme() {
  theme.global.name.value = theme.global.current.value.dark ? "light" : "dark";
}

function handleLogout() {
  authStore.logout();
  router.push("/login");
}
</script>

<style scoped>
.cursor-pointer {
  cursor: pointer;
}
</style>
