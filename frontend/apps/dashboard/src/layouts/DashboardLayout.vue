<template>
  <v-layout class="fill-height" v-if="!isNavigating">
    <!-- Left Navigation Drawer -->
    <v-navigation-drawer
      v-model="drawer"
      permanent
      elevation="1"
      width="280"
      :rail="rail"
    >
      <!-- Top Branding Section -->
      <template #prepend>
        <v-list class="py-3">
          <v-list-item
            title="Email Ingestion"
            subtitle="Gateway Console"
            class="brand-title px-2"
          >
            <template #prepend>
              <v-hover v-slot="{ isHovering, props }">
                <div
                  v-bind="props"
                  class="d-flex align-center justify-center me-2"
                  style="width: 36px; height: 36px"
                >
                  <v-btn
                    v-if="isHovering && rail"
                    variant="text"
                    icon
                    size="small"
                    @click.stop="collapseToggle(false)"
                  >
                    <v-icon icon="mdi-arrow-collapse-right" />
                  </v-btn>
                  <v-avatar
                    v-else
                    image="/favicon.svg"
                    variant="flat"
                    size="36"
                  />
                </div>
              </v-hover>
            </template>
            <template #append>
              <v-btn
                variant="text"
                icon
                class="ms-3"
                @click.stop="collapseToggle(true)"
              >
                <v-icon icon="mdi-arrow-collapse-left" />
              </v-btn>
            </template>
          </v-list-item>
        </v-list>
        <v-divider />
      </template>

      <!-- Navigation Links -->
      <NavigationLinks :disable="!hasApps" />

      <!-- Bottom Profile & Theme AddOns -->
      <template #append>
        <v-divider />
        <NavigationAddOns :is-collapsed="rail" />
      </template>
    </v-navigation-drawer>

    <!-- Main Content Workspace Area -->
    <v-main class="bg-background min-vh-100">
      <v-container fluid class="pa-6">
        <CustomAppBar />

        <!-- Router View Slot inside Content Sheet -->
        <v-sheet rounded="lg" class="pa-6 elevation-1 bg-surface">
          <slot />
        </v-sheet>
      </v-container>
    </v-main>

    <!-- Global Toast Notification Bar -->
    <v-snackbar
      v-model="notificationStore.show"
      :color="notificationStore.color"
      :timeout="notificationStore.timeout"
      location="top right"
      rounded="lg"
      elevation="6"
    >
      <div class="d-flex align-center">
        <v-icon
          start
          :icon="getSnackbarIcon(notificationStore.color)"
          class="me-2"
        />
        <span>{{ notificationStore.message }}</span>
      </div>
      <template #actions>
        <v-btn
          variant="text"
          icon="mdi-close"
          size="small"
          @click="notificationStore.close()"
        />
      </template>
    </v-snackbar>
  </v-layout>
</template>

<script setup>
import { ref, computed } from "vue";
import NavigationLinks from "@/components/navigation/NavigationLinks.vue";
import NavigationAddOns from "@/components/navigation/NavigationAddOns.vue";
import CustomAppBar from "@/components/CustomAppBar.vue";
import { useNotificationStore } from "@/stores/notification";
import { useAppStore } from "@/stores/application";
import { isNavigating } from "@/router";

const drawer = ref(true);
const notificationStore = useNotificationStore();
const appStore = useAppStore();

const rail = ref(true);
const hasApps = computed(() => appStore.applications.length > 0);

function getSnackbarIcon(color) {
  switch (color) {
    case "success":
      return "mdi-check-circle";
    case "error":
      return "mdi-alert-circle";
    case "warning":
      return "mdi-alert-amber";
    default:
      return "mdi-information";
  }
}

function collapseToggle(open) {
  rail.value = open;
}
</script>

<style scoped>
.brand-title :deep(.v-list-item-title) {
  font-weight: 700;
  font-size: 1.05rem;
  letter-spacing: -0.01em;
}
.brand-title :deep(.v-list-item-subtitle) {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.gap-3 {
  gap: 12px;
}
.gap-2 {
  gap: 8px;
}
.border-b {
  border-bottom: 1px solid rgba(var(--v-border-color), var(--v-border-opacity));
}
.min-vh-100 {
  min-height: 100vh;
}
</style>
