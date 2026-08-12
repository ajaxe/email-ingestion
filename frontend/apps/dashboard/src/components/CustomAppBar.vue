<template>
  <!-- Header Bar with Tenant App Selector & Add Application Action -->
  <div class="d-flex align-center justify-space-between mb-6 pb-4 border-b">
    <div class="d-flex align-center gap-3">
      <AppSelector @open-create-modal="showCreateModal = true" />
      <v-btn
        color="primary"
        variant="tonal"
        size="small"
        prepend-icon="mdi-plus-circle"
        class="font-weight-medium ms-2"
        @click="showCreateModal = true"
      >
        Add Application
      </v-btn>
    </div>
    <div class="d-flex align-center gap-2">
      <v-chip color="info" variant="outlined" size="small">
        <v-icon start icon="mdi-shield-check" />
        {{ idp }} Authenticated
      </v-chip>
    </div>
  </div>

  <!-- Modal for Creating New Application Tenant -->
  <CreateAppModal
    v-model="showCreateModal"
    @created="onAppCreated"
  />
</template>
<script setup>
import { ref } from "vue";
import AppSelector from "@/components/AppSelector.vue";
import CreateAppModal from "@/components/CreateAppModal.vue";
import { useNotificationStore } from "@/stores/notification";
import {useAuthStore} from "@/stores/auth";

const authStore = useAuthStore();
const idp = ref(authStore.provider === "oidc" ? "OIDC" : "Admin Password");
const showCreateModal = ref(false);
const notificationStore = useNotificationStore();

function onAppCreated(app) {
  notificationStore.success(`Application "${app?.name || 'New App'}" created successfully!`);
}
</script>
