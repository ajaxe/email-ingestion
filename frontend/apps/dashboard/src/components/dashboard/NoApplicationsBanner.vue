<template>
  <v-card rounded="lg" elevation="2" class="pa-8 text-center bg-surface border">
    <div class="d-flex justify-center mb-4">
      <v-avatar color="primary" variant="tonal" size="72">
        <v-icon icon="mdi-domain-plus" size="40" color="primary" />
      </v-avatar>
    </div>

    <h2 class="text-h5 font-weight-bold mb-2">No Applications Found</h2>
    <p class="text-body-1 text-medium-emphasis mx-auto mb-6 max-w-600">
      You haven't created any application tenants yet. Create your first application to begin receiving inbound emails, provisioning routing addresses, and configuring webhook dispatches.
    </p>

    <div class="d-flex justify-center gap-3 mb-8">
      <v-btn
        color="primary"
        size="large"
        variant="flat"
        prepend-icon="mdi-plus-circle"
        class="px-6 font-weight-bold"
        @click="showCreateModal = true"
      >
        Add Application
      </v-btn>
    </div>

    <v-divider class="my-6" />

    <!-- Feature highlights list -->
    <v-row class="text-start justify-center">
      <v-col cols="12" sm="4">
        <div class="d-flex align-start gap-3 pa-2">
          <v-avatar color="info" variant="tonal" size="36" class="mt-1">
            <v-icon icon="mdi-shield-lock-outline" size="20" color="info" />
          </v-avatar>
          <div>
            <div class="text-subtitle-2 font-weight-bold">Multi-Tenant Security</div>
            <div class="text-caption text-medium-emphasis">
              Isolated database partitioning and gateway-brokered email storage per application.
            </div>
          </div>
        </div>
      </v-col>

      <v-col cols="12" sm="4">
        <div class="d-flex align-start gap-3 pa-2">
          <v-avatar color="success" variant="tonal" size="36" class="mt-1">
            <v-icon icon="mdi-routes" size="20" color="success" />
          </v-avatar>
          <div>
            <div class="text-subtitle-2 font-weight-bold">Flexible Inbound Routing</div>
            <div class="text-caption text-medium-emphasis">
              Provision custom local-part routing addresses with active domain validation.
            </div>
          </div>
        </div>
      </v-col>

      <v-col cols="12" sm="4">
        <div class="d-flex align-start gap-3 pa-2">
          <v-avatar color="warning" variant="tonal" size="36" class="mt-1">
            <v-icon icon="mdi-webhook" size="20" color="warning" />
          </v-avatar>
          <div>
            <div class="text-subtitle-2 font-weight-bold">Reliable Webhooks</div>
            <div class="text-caption text-medium-emphasis">
              HMAC-SHA256 signed webhooks with exponential backoff and circuit-breaking delivery.
            </div>
          </div>
        </div>
      </v-col>
    </v-row>

    <!-- Create Application Modal -->
    <CreateAppModal
      v-model="showCreateModal"
      @created="onAppCreated"
    />
  </v-card>
</template>

<script setup>
import { ref } from 'vue';
import CreateAppModal from '@/components/CreateAppModal.vue';
import { useNotificationStore } from '@/stores/notification';

const showCreateModal = ref(false);
const notificationStore = useNotificationStore();

function onAppCreated(app) {
  notificationStore.success(`Application "${app?.name || 'New App'}" created successfully!`);
}
</script>

<style scoped>
.max-w-600 {
  max-width: 600px;
}
.gap-3 {
  gap: 12px;
}
</style>
