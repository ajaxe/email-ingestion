<template>
  <v-list density="compact" nav class="py-2">
    <v-list-item>
      <div class="d-flex align-center justify-space-between mb-2 px-2">
        <span
          v-if="!isCollapsed"
          class="text-caption text-medium-emphasis font-weight-medium"
          >Theme</span
        >
        <v-btn
          variant="text"
          density="compact"
          :class="{ 'ms-3': !isCollapsed }"
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
      </div>
    </v-list-item>
    <v-list-item>
      <v-menu location="top start" offset="8">
        <template #activator="{ props }">
          <v-list-item v-bind="props" class="rounded-lg cursor-pointer">
            <template #prepend>
              <v-avatar :image="profileImage" />
            </template>
            <template #append>
              <v-icon icon="mdi-dots-vertical" size="small" />
            </template>
            <v-list-item-title>{{ username }}</v-list-item-title>
            <v-list-item-subtitle>{{ email }}</v-list-item-subtitle>
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
    </v-list-item>
  </v-list>
</template>

<script setup>
import { computed } from "vue";
import { useTheme } from "vuetify";
import { useAuthStore } from "@/stores/auth";

const { isCollapsed } = defineProps({
  isCollapsed: {
    type: Boolean,
    default: true,
  },
});

const emit = defineEmits(["collapseToggle"]);

const theme = useTheme();
const authStore = useAuthStore();

const profileImage = computed(() => authStore.user.profileImage);
const username = computed(() => authStore.user.name);
const email = computed(() => authStore.user.email);

function toggleTheme() {
  theme.global.name.value = theme.global.current.value.dark ? "light" : "dark";
}

function handleLogout() {
  authStore.logout();
}
</script>

<style scoped>
.cursor-pointer {
  cursor: pointer;
}
</style>
