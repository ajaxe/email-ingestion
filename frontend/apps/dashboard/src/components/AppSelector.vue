<template>
  <div class="d-flex align-center gap-3">
    <v-select
      v-model="appStore.activeAppId"
      :items="appStore.applications"
      item-title="name"
      item-value="id"
      label="Active Tenant Application"
      variant="outlined"
      density="compact"
      hide-details
      max-width="360"
      class="flex-grow-0"
      @update:model-value="onAppChange"
    >
      <template #prepend-inner>
        <v-icon icon="mdi-domain" color="primary" class="me-1" />
      </template>

      <template #append-item>
        <v-divider class="my-1" />
        <v-list-item
          title="Create New Application"
          prepend-icon="mdi-plus"
          class="text-primary font-weight-medium"
          @click="emit('open-create-modal')"
        />
      </template>
    </v-select>

    <StatusChip
      :status="appStore.activeAppStatus"
      size="small"
    />
  </div>
</template>

<script setup>
import { onMounted } from 'vue';
import { useAppStore } from '@/stores/application';
import StatusChip from '@/components/StatusChip.vue';

const emit = defineEmits(['open-create-modal']);
const appStore = useAppStore();

async function onAppChange(appId) {
  if (appId) {
    await appStore.selectApp(appId);
  }
}

onMounted(() => {
  if (appStore.activeAppId && !appStore.application) {
    appStore.fetchAppDetails(appStore.activeAppId);
  }
});
</script>

<style scoped>
.gap-3 {
  gap: 12px;
}
</style>
