<template>
  <v-card rounded="lg" variant="elevated" class="pa-4">
    <div class="text-h6 font-weight-bold mb-1">Gateway Status</div>
    <div class="text-caption text-medium-emphasis mb-4">
      Security & System Diagnostics
    </div>

    <v-list density="compact" class="pa-0 bg-transparent">
      <v-list-item class="px-0">
        <v-list-item-title class="text-caption text-medium-emphasis">
          System Status
        </v-list-item-title>
        <template #append>
          <StatusChip status="ACTIVE" size="x-small" />
        </template>
      </v-list-item>

      <v-divider class="my-2" />

      <v-list-item class="px-0">
        <v-list-item-title class="text-caption text-medium-emphasis mb-1">
          AWS IAM Role ARN
        </v-list-item-title>
        <div class="text-caption font-mono bg-surface-variant pa-2 rounded text-truncate">
          {{ activeRoleArn }}
        </div>
      </v-list-item>

      <v-divider class="my-2" />

      <v-list-item class="px-0">
        <v-list-item-title class="text-caption text-medium-emphasis">
          Active Tenant App ID
        </v-list-item-title>
        <div class="text-caption font-mono text-medium-emphasis">
          {{ appStore.activeAppId || 'None Selected' }}
        </div>
      </v-list-item>
    </v-list>
  </v-card>
</template>

<script setup>
import { computed } from 'vue';
import StatusChip from '@/components/StatusChip.vue';
import { useAppStore } from '@/stores/application';

const appStore = useAppStore();

const activeRoleArn = computed(() => {
  const app = appStore.activeApp;
  if (app && (app.aws_iam_role_arn || app.awsIamRoleArn)) {
    return app.aws_iam_role_arn || app.awsIamRoleArn;
  }
  const appId = appStore.activeAppId || 'default';
  return `arn:aws:iam::123456789012:role/gateway-tenant-${appId.slice(0, 8)}`;
});
</script>

<style scoped>
.font-mono {
  font-family: 'Roboto Mono', monospace;
}
</style>
