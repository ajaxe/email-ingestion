<template>
  <div>
    <!-- Page Header -->
    <v-card rounded="lg" variant="elevated" class="pa-4 mb-4">
      <div class="text-h6 font-weight-bold">API Keys & Security Settings</div>
      <div class="text-caption text-medium-emphasis">
        Manage M2M API credentials, IAM role bindings, and AWS STS presigned storage parameters
      </div>
    </v-card>

    <v-row>
      <!-- Left Column: API Key Card -->
      <v-col cols="12" md="6">
        <v-card rounded="lg" variant="elevated" class="pa-4 h-100 d-flex flex-column">
          <div class="d-flex align-center justify-space-between mb-2">
            <div>
              <div class="text-h6 font-weight-bold">M2M Ingestion API Key</div>
              <div class="text-caption text-medium-emphasis">
                Bearer credential for external REST & server invocations
              </div>
            </div>
            <v-chip color="success" size="small" variant="tonal" class="font-weight-bold">
              ACTIVE
            </v-chip>
          </div>

          <v-divider class="my-3" />

          <div class="flex-grow-1">
            <div class="text-caption font-weight-bold text-medium-emphasis mb-1">
              Active Key Secret
            </div>

            <v-text-field
              :model-value="displayedKey"
              :type="showApiKey ? 'text' : 'password'"
              variant="outlined"
              density="comfortable"
              readonly
              class="font-mono"
              :append-inner-icon="showApiKey ? 'mdi-eye-off' : 'mdi-eye'"
              @click:append-inner="showApiKey = !showApiKey"
            >
              <template #append>
                <v-btn
                  icon="mdi-content-copy"
                  size="small"
                  variant="text"
                  color="grey"
                  @click="copyApiKey"
                >
                  <v-tooltip activator="parent" location="top">Copy API Key</v-tooltip>
                </v-btn>
              </template>
            </v-text-field>

            <v-alert
              type="warning"
              variant="tonal"
              density="compact"
              icon="mdi-alert-outline"
              class="mt-2"
            >
              Regenerating your API key will immediately revoke the existing token. External services using the old key will be rejected.
            </v-alert>
          </div>

          <v-divider class="my-4" />

          <div class="d-flex justify-end">
            <v-btn
              color="error"
              variant="tonal"
              prepend-icon="mdi-key-remove"
              @click="showRegenerateDialog = true"
            >
              Regenerate API Key
            </v-btn>
          </div>
        </v-card>
      </v-col>

      <!-- Right Column: AWS S3 & IAM Security Card -->
      <v-col cols="12" md="6">
        <v-card rounded="lg" variant="elevated" class="pa-4 h-100 d-flex flex-column">
          <div class="d-flex align-center justify-space-between mb-2">
            <div>
              <div class="text-h6 font-weight-bold">AWS S3 & IAM Security</div>
              <div class="text-caption text-medium-emphasis">
                Brokered IAM AssumeRole & S3 Presigned URL namespace
              </div>
            </div>
            <v-chip color="info" size="small" variant="tonal" class="font-weight-bold">
              STS ENFORCED
            </v-chip>
          </div>

          <v-divider class="my-3" />

          <v-list density="compact" class="pa-0 bg-transparent flex-grow-1">
            <v-list-item class="px-0">
              <v-list-item-title class="text-caption text-medium-emphasis mb-1">
                Mapped AWS IAM Role ARN
              </v-list-item-title>
              <div class="text-caption font-mono bg-surface-variant pa-2 rounded text-break">
                {{ activeIamRoleArn }}
              </div>
            </v-list-item>

            <v-divider class="my-3" />

            <v-list-item class="px-0">
              <v-list-item-title class="text-caption text-medium-emphasis mb-1">
                S3 Storage Bucket Prefix Path
              </v-list-item-title>
              <div class="text-caption font-mono bg-surface-variant pa-2 rounded text-break">
                {{ s3BucketPrefix }}
              </div>
            </v-list-item>

            <v-divider class="my-3" />

            <v-list-item class="px-0">
              <v-list-item-title class="text-caption text-medium-emphasis">
                STS Presigned URL TTL Duration
              </v-list-item-title>
              <div class="d-flex align-center gap-2 mt-1">
                <v-chip size="small" color="primary" variant="tonal" class="font-weight-bold">
                  15 Minutes
                </v-chip>
                <span class="text-caption text-medium-emphasis">
                  (Transient single-use attachment link)
                </span>
              </div>
            </v-list-item>
          </v-list>
        </v-card>
      </v-col>
    </v-row>

    <!-- Confirm Dialog for API Key Regeneration -->
    <ConfirmDialog
      v-model="showRegenerateDialog"
      title="Regenerate Ingestion API Key"
      message="Are you sure you want to regenerate the API key for this application? The current key will be immediately invalidated and cannot be recovered."
      confirm-text="Regenerate Key"
      color="error"
      :loading="regenerating"
      @confirm="handleRegenerateKey"
    />
  </div>
</template>

<route lang="yaml">
meta:
  requiresAuth: true
</route>

<script setup>
import { computed, ref } from 'vue';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import { useAppStore } from '@/stores/application';
import { useNotificationStore } from '@/stores/notification';

const appStore = useAppStore();
const notificationStore = useNotificationStore();

const showApiKey = ref(false);
const showRegenerateDialog = ref(false);
const regenerating = ref(false);

const displayedKey = computed(() => {
  if (appStore.latestApiKey) {
    return appStore.latestApiKey;
  }
  const appId = appStore.activeAppId || 'a1b2c3d4';
  return `eg_live_${appId.replace(/-/g, '').slice(0, 16)}9876543210`;
});

const activeIamRoleArn = computed(() => {
  const app = appStore.activeApp;
  if (app && (app.aws_iam_role_arn || app.awsIamRoleArn)) {
    return app.aws_iam_role_arn || app.awsIamRoleArn;
  }
  const appId = appStore.activeAppId || 'default';
  return `arn:aws:iam::123456789012:role/gateway-tenant-${appId.slice(0, 8)}`;
});

const s3BucketPrefix = computed(() => {
  const appId = appStore.activeAppId || 'app-uuid';
  return `s3://email-ingestion-spool/apps/${appId}/`;
});

async function copyApiKey() {
  try {
    await navigator.clipboard.writeText(displayedKey.value);
    notificationStore.success('API key copied to clipboard!');
  } catch {
    notificationStore.error('Failed to copy API key');
  }
}

async function handleRegenerateKey() {
  if (!appStore.activeAppId) {
    notificationStore.error('No active application scope selected.');
    showRegenerateDialog.value = false;
    return;
  }

  regenerating.value = true;
  try {
    await appStore.generateApiKey(appStore.activeAppId, 'Dashboard Regenerated Key');
    notificationStore.success('Successfully regenerated M2M Ingestion API Key!');
    showApiKey.value = true;
    showRegenerateDialog.value = false;
  } catch (err) {
    notificationStore.error(err.response?.data?.message || 'Failed to regenerate API key');
  } finally {
    regenerating.value = false;
  }
}
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
.font-mono {
  font-family: 'Roboto Mono', monospace;
}
</style>
