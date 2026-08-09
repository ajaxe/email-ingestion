<template>
  <v-card rounded="lg" variant="elevated" class="pa-4">
    <div class="d-flex align-center justify-space-between mb-2">
      <div>
        <div class="text-h6 font-weight-bold">M2M Ingestion API Keys</div>
        <div class="text-caption text-medium-emphasis">
          Manage bearer credentials and scope environments for external REST invocations
        </div>
      </div>
      <v-btn
        color="primary"
        prepend-icon="mdi-key-plus"
        size="small"
        class="font-weight-bold"
        @click="showCreateModal = true"
      >
        Create API Key
      </v-btn>
    </div>

    <v-divider class="my-3" />

    <!-- Table of API Keys -->
    <v-table v-if="appStore.apiKeys && appStore.apiKeys.length > 0" density="comfortable" class="bg-transparent">
      <thead>
        <tr>
          <th class="text-left font-weight-bold">Name</th>
          <th class="text-left font-weight-bold">Prefix / Scope</th>
          <th class="text-left font-weight-bold">Created</th>
          <th class="text-left font-weight-bold">Expires</th>
          <th class="text-right font-weight-bold">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="key in appStore.apiKeys" :key="key.id">
          <td>
            <div class="font-weight-medium text-body-2">{{ key.name }}</div>
          </td>
          <td>
            <div class="d-flex align-center gap-2">
              <v-chip
                :color="key.keyPrefix?.includes('live') ? 'success' : 'warning'"
                size="x-small"
                variant="tonal"
                class="font-weight-bold"
              >
                {{ key.keyPrefix?.includes('live') ? 'LIVE' : 'TEST' }}
              </v-chip>
              <span class="font-mono text-caption text-medium-emphasis">
                {{ key.keyPrefix || 'eg_live_' }}••••••••
              </span>
            </div>
          </td>
          <td class="text-caption text-medium-emphasis">
            {{ formatDate(key.createdAt) }}
          </td>
          <td class="text-caption text-medium-emphasis">
            {{ formatDate(key.expiresAt) }}
          </td>
          <td class="text-right">
            <v-btn
              icon
              size="small"
              variant="text"
              color="error"
              @click="confirmRevokeKey(key)"
            >
              <v-icon icon="mdi-delete-outline" />
              <v-tooltip activator="parent" location="top">Revoke API Key</v-tooltip>
            </v-btn>
          </td>
        </tr>
      </tbody>
    </v-table>

    <!-- Empty State -->
    <div v-else-if="!appStore.loading" class="text-center py-6">
      <v-icon icon="mdi-key-outline" size="40" color="grey" class="mb-2" />
      <div class="text-body-2 font-weight-bold text-medium-emphasis">No API Keys Generated</div>
      <div class="text-caption text-medium-emphasis mb-3">
        Create an API key to allow external services to authenticate with the ingestion gateway.
      </div>
      <v-btn
        color="primary"
        variant="tonal"
        size="small"
        prepend-icon="mdi-key-plus"
        @click="showCreateModal = true"
      >
        Create API Key
      </v-btn>
    </div>

    <!-- Loading Skeleton -->
    <v-skeleton-loader v-else type="table-row-divider@3" />

    <!-- Create API Key Modal -->
    <CreateApiKeyModal
      v-model="showCreateModal"
      :submitting="creating"
      @submit="handleCreateKey"
    />

    <!-- Revoke Confirmation Dialog -->
    <ConfirmDialog
      v-model="showRevokeDialog"
      title="Revoke API Key"
      :message="`Are you sure you want to revoke '${selectedKey?.name}'? Any service using this key will immediately lose access.`"
      confirm-text="Revoke Key"
      color="error"
      :loading="revoking"
      @confirm="handleRevokeKey"
    />

    <!-- One-Time Viewing Modal for Newly Generated Key -->
    <ApiKeyConfirmationModal
      v-model="showKeyModal"
      :api-key="generatedKey"
    />
  </v-card>
</template>

<script setup>
import { onMounted, ref, watch } from 'vue';
import ConfirmDialog from '@/components/ConfirmDialog.vue';
import ApiKeyConfirmationModal from '@/components/settings/ApiKeyConfirmationModal.vue';
import CreateApiKeyModal from '@/components/settings/CreateApiKeyModal.vue';
import { useAppStore } from '@/stores/application';
import { useNotificationStore } from '@/stores/notification';

const appStore = useAppStore();
const notificationStore = useNotificationStore();

const showCreateModal = ref(false);
const showRevokeDialog = ref(false);
const showKeyModal = ref(false);
const creating = ref(false);
const revoking = ref(false);
const selectedKey = ref(null);
const generatedKey = ref('');

onMounted(async () => {
  if (appStore.activeAppId) {
    await appStore.fetchApiKeys(appStore.activeAppId);
  }
});

watch(() => appStore.activeAppId, async (newAppId) => {
  if (newAppId) {
    await appStore.fetchApiKeys(newAppId);
  }
});

function formatDate(dateStr) {
  if (!dateStr) return 'N/A';
  try {
    return new Date(dateStr).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  } catch {
    return dateStr;
  }
}

async function handleCreateKey(payload) {
  if (!appStore.activeAppId) {
    notificationStore.error('No active application scope selected.');
    showCreateModal.value = false;
    return;
  }

  creating.value = true;
  try {
    const keyData = await appStore.generateApiKey(appStore.activeAppId, payload);
    const newKey = keyData?.apiKey || keyData?.api_key || keyData?.APIKey || keyData?.key || appStore.latestApiKey || '';
    generatedKey.value = newKey;
    showCreateModal.value = false;
    showKeyModal.value = true;
    notificationStore.success('Successfully created new API Key!');
  } catch (err) {
    notificationStore.error(err.response?.data?.message || 'Failed to create API key');
  } finally {
    creating.value = false;
  }
}

function confirmRevokeKey(key) {
  selectedKey.value = key;
  showRevokeDialog.value = true;
}

async function handleRevokeKey() {
  if (!selectedKey.value || !appStore.activeAppId) return;

  revoking.value = true;
  try {
    await appStore.revokeApiKey(appStore.activeAppId, selectedKey.value.id);
    notificationStore.success(`API key '${selectedKey.value.name}' revoked successfully.`);
    showRevokeDialog.value = false;
    selectedKey.value = null;
  } catch (err) {
    notificationStore.error(err.response?.data?.message || 'Failed to revoke API key');
  } finally {
    revoking.value = false;
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
