<template>
  <div class="next-steps-container">
    <!-- Header Banner -->
    <v-alert
      color="success"
      variant="tonal"
      rounded="lg"
      icon="mdi-check-circle-outline"
      class="mb-4"
    >
      <template #title>
        <div class="text-subtitle-1 font-weight-bold">
          Tenant Initialized Successfully!
        </div>
      </template>
      <div class="text-body-2">
        Application <strong>{{ appName }}</strong> is active. Complete the configuration steps below to start processing inbound email traffic.
      </div>
    </v-alert>

    <!-- Provisioned Address Banner (if available) -->
    <v-card
      v-if="fullAddress"
      rounded="lg"
      variant="outlined"
      color="primary"
      class="mb-4 pa-3 bg-blue-lighten-5"
    >
      <div class="d-flex align-center justify-space-between flex-wrap gap-2">
        <div>
          <div class="text-caption font-weight-bold text-primary text-uppercase">
            Auto-Provisioned Routing Mailbox
          </div>
          <div class="font-mono text-subtitle-1 font-weight-bold text-high-emphasis">
            {{ fullAddress }}
          </div>
        </div>
        <v-btn
          size="small"
          color="primary"
          variant="flat"
          prepend-icon="mdi-content-copy"
          @click="copyAddress"
        >
          Copy Address
        </v-btn>
      </div>
    </v-card>

    <!-- Stepper List of Required Configurations -->
    <v-card rounded="lg" variant="flat" border class="pa-4">
      <div class="text-subtitle-2 font-weight-bold text-uppercase text-medium-emphasis mb-3">
        Recommended Configuration Workflow
      </div>

      <v-timeline density="compact" align="start" side="end">
        <!-- Step 1: Inbound Address -->
        <v-timeline-item
          :dot-color="hasAddress ? 'success' : 'primary'"
          :icon="hasAddress ? 'mdi-check' : 'mdi-at'"
          size="small"
        >
          <div class="d-flex align-center justify-space-between flex-wrap gap-2">
            <div>
              <div class="text-subtitle-2 font-weight-bold">
                1. Inbound Routing Address
                <v-chip
                  v-if="hasAddress"
                  color="success"
                  size="x-small"
                  class="ms-2 font-weight-bold"
                >
                  Ready
                </v-chip>
              </div>
              <div class="text-caption text-medium-emphasis">
                Receives inbound SMTP traffic and parses MIME payloads into tenant partitions.
              </div>
            </div>
            <v-btn
              v-if="!hasAddress"
              size="small"
              variant="tonal"
              color="primary"
              prepend-icon="mdi-plus"
              @click="goTo('/addresses')"
            >
              Provision Address
            </v-btn>
          </div>
        </v-timeline-item>

        <!-- Step 2: Webhook Endpoint -->
        <v-timeline-item
          :dot-color="hasWebhook ? 'success' : 'warning'"
          :icon="hasWebhook ? 'mdi-check' : 'mdi-webhook'"
          size="small"
        >
          <div class="d-flex align-center justify-space-between flex-wrap gap-2">
            <div>
              <div class="text-subtitle-2 font-weight-bold">
                2. Real-Time Webhook Endpoint
                <v-chip
                  :color="hasWebhook ? 'success' : 'warning'"
                  size="x-small"
                  class="ms-2 font-weight-bold"
                >
                  {{ hasWebhook ? 'Configured' : 'Recommended' }}
                </v-chip>
              </div>
              <div class="text-caption text-medium-emphasis">
                Delivers signed HMAC-SHA256 HTTP webhooks to your application server on email arrival.
              </div>
            </div>
            <v-btn
              size="small"
              :variant="hasWebhook ? 'outlined' : 'flat'"
              color="primary"
              prepend-icon="mdi-lightning-bolt-outline"
              @click="goTo('/webhooks')"
            >
              {{ hasWebhook ? 'View Webhook' : 'Setup Webhook' }}
            </v-btn>
          </div>
        </v-timeline-item>

        <!-- Step 3: API Key & Security -->
        <v-timeline-item
          :dot-color="hasApiKey ? 'success' : 'info'"
          :icon="hasApiKey ? 'mdi-check' : 'mdi-key-chain'"
          size="small"
        >
          <div class="d-flex align-center justify-space-between flex-wrap gap-2">
            <div>
              <div class="text-subtitle-2 font-weight-bold">
                3. REST API Key Authentication
                <v-chip
                  :color="hasApiKey ? 'success' : 'info'"
                  size="x-small"
                  class="ms-2 font-weight-bold"
                >
                  {{ hasApiKey ? 'Active Key' : 'Optional' }}
                </v-chip>
              </div>
              <div class="text-caption text-medium-emphasis">
                Generate M2M API Keys to fetch emails and pre-signed S3 attachment URLs via backend.
              </div>
            </div>
            <v-btn
              size="small"
              variant="tonal"
              color="primary"
              prepend-icon="mdi-key-plus"
              @click="goTo('/settings')"
            >
              {{ hasApiKey ? 'Manage Keys' : 'Create API Key' }}
            </v-btn>
          </div>
        </v-timeline-item>
      </v-timeline>
    </v-card>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useAppStore } from '@/stores/application';
import { useNotificationStore } from '@/stores/notification';

const props = defineProps({
  app: {
    type: Object,
    default: null,
  },
  provisionedAddress: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(['navigate', 'close']);

const router = useRouter();
const appStore = useAppStore();
const notificationStore = useNotificationStore();

const currentApp = computed(() => props.app || appStore.activeApp || {});
const appName = computed(() => currentApp.value.name || 'New Application');

const fullAddress = computed(() => {
  if (props.provisionedAddress) {
    const local = props.provisionedAddress.localPart;
    const domain = window.APP_CONFIG?.INGEST_DOMAIN || 'localhost';
    if (local) return `${local}@${domain}`;
  }
  return '';
});

const hasAddress = computed(() => {
  return Boolean(props.provisionedAddress || appStore.stats.totalAddresses > 0);
});

const hasWebhook = computed(() => {
  return Boolean(currentApp.value.webhookUrl || currentApp.value.webhook_url);
});

const hasApiKey = computed(() => {
  return Boolean(appStore.apiKeys && appStore.apiKeys.length > 0);
});

async function copyAddress() {
  if (!fullAddress.value) return;
  try {
    await navigator.clipboard.writeText(fullAddress.value);
    notificationStore.success(`Copied ${fullAddress.value} to clipboard!`);
  } catch {
    notificationStore.error('Failed to copy address');
  }
}

function goTo(path) {
  emit('close');
  emit('navigate', path);
  router.push(path);
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
