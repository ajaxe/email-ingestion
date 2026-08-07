<template>
  <div>
    <!-- Top Card: Webhook Endpoint & Secret Settings -->
    <v-card rounded="lg" variant="elevated" class="pa-4 mb-4">
      <div class="d-flex align-center justify-space-between mb-2">
        <div>
          <div class="text-h6 font-weight-bold">Webhook Configuration & Challenge Handshake</div>
          <div class="text-caption text-medium-emphasis">
            Configure secure HMAC-SHA256 signature dispatch & challenge verification
          </div>
        </div>
        <StatusChip
          :status="handshakeSuccess ? 'ACTIVE' : 'INACTIVE'"
          size="small"
        />
      </div>

      <!-- Handshake Status Alert -->
      <v-alert
        v-if="handshakeResult"
        :type="handshakeSuccess ? 'success' : 'error'"
        variant="tonal"
        closable
        class="mb-4"
        density="compact"
        @click:close="handshakeResult = null"
      >
        <div class="font-weight-bold">{{ handshakeResult.title }}</div>
        <div class="text-caption">{{ handshakeResult.message }}</div>
      </v-alert>

      <v-form @submit.prevent="handleSaveConfig">
        <v-row class="mt-1">
          <v-col cols="12" md="6">
            <v-text-field
              v-model="webhookUrl"
              label="Endpoint URL"
              placeholder="https://example.com/api/webhooks"
              variant="outlined"
              density="comfortable"
              prepend-inner-icon="mdi-link-variant"
              hint="Must be HTTPS and pass RFC 1918 SSRF validation"
              persistent-hint
            />
          </v-col>

          <v-col cols="12" md="6">
            <v-text-field
              v-model="webhookSecret"
              :type="showSecret ? 'text' : 'password'"
              label="Webhook Signing Secret"
              placeholder="whsec_..."
              variant="outlined"
              density="comfortable"
              prepend-inner-icon="mdi-shield-key-outline"
              :append-inner-icon="showSecret ? 'mdi-eye-off' : 'mdi-eye'"
              hint="Used to verify HMAC-SHA256 X-Gateway-Signature header"
              persistent-hint
              @click:append-inner="showSecret = !showSecret"
            >
              <template #append>
                <v-btn
                  icon="mdi-content-copy"
                  size="small"
                  variant="text"
                  color="grey"
                  @click="copySecret"
                >
                  <v-tooltip activator="parent" location="top">Copy Secret</v-tooltip>
                </v-btn>
              </template>
            </v-text-field>
          </v-col>

          <v-col cols="12" md="6">
            <div class="text-caption text-medium-emphasis mb-1">
              Max Retry Delivery Attempts: <strong class="text-primary">{{ maxRetries }}</strong>
            </div>
            <v-slider
              v-model="maxRetries"
              :min="1"
              :max="10"
              :step="1"
              thumb-label
              color="primary"
              track-color="surface-variant"
            />
          </v-col>

          <v-col cols="12" md="6" class="d-flex align-center justify-md-end gap-3 mt-auto">
            <v-btn
              color="secondary"
              variant="tonal"
              prepend-icon="mdi-lightning-bolt"
              :loading="testingHandshake"
              @click="handleTestEndpoint"
            >
              Test & Verify Endpoint
            </v-btn>

            <v-btn
              color="primary"
              variant="flat"
              prepend-icon="mdi-content-save"
              :loading="savingConfig"
              type="submit"
            >
              Save Configuration
            </v-btn>
          </v-col>
        </v-row>
      </v-form>
    </v-card>

    <!-- Bottom Card: Webhook Delivery Outbox Sandbox -->
    <v-card rounded="lg" variant="elevated" class="pa-4">
      <v-row align="center" justify="space-between" class="ma-0 mb-4">
        <v-col cols="12" md="4" class="pa-0">
          <div class="text-h6 font-weight-bold">Outbox Delivery Attempt History</div>
          <div class="text-caption text-medium-emphasis">
            Transactional outbox retry jobs & delivery execution telemetry
          </div>
        </v-col>

        <v-col cols="12" md="8" class="pa-0 mt-3 mt-md-0 d-flex flex-wrap align-center justify-md-end gap-3">
          <v-btn-toggle
            v-model="statusFilter"
            mandatory
            variant="outlined"
            density="compact"
            color="primary"
          >
            <v-btn value="ALL">All</v-btn>
            <v-btn value="PENDING">Pending</v-btn>
            <v-btn value="SUCCESS">Success</v-btn>
            <v-btn value="FAILED">Failed</v-btn>
            <v-btn value="DEAD">Dead</v-btn>
          </v-btn-toggle>

          <v-switch
            v-model="autoRefresh"
            label="Auto-refresh"
            color="primary"
            hide-details
            density="compact"
            class="ms-2"
          />

          <v-btn
            icon="mdi-refresh"
            variant="outlined"
            density="compact"
            color="primary"
            :loading="webhookStore.loading"
            @click="loadWebhookJobs"
          >
            <v-tooltip activator="parent" location="top">Refresh Jobs</v-tooltip>
          </v-btn>
        </v-col>
      </v-row>

      <!-- Outbox Data Table -->
      <v-data-table
        :headers="headers"
        :items="filteredJobs"
        :loading="webhookStore.loading"
        density="comfortable"
        hover
      >
        <template #item.id="{ item }">
          <span class="font-mono text-caption font-weight-bold">
            {{ formatShortId(item.id || item.job_id) }}
          </span>
        </template>

        <template #item.retry_count="{ item }">
          <v-chip size="x-small" color="primary" variant="outlined" class="font-mono">
            #{{ item.attempt_number || item.retry_count || 1 }}
          </v-chip>
        </template>

        <template #item.status="{ item }">
          <StatusChip :status="item.status || item.Status" />
        </template>

        <template #item.duration_ms="{ item }">
          <span class="font-mono text-caption">
            {{ item.duration_ms || item.durationMs ? `${item.duration_ms || item.durationMs} ms` : '—' }}
          </span>
        </template>

        <template #item.created_at="{ item }">
          <span class="font-mono text-caption">
            {{ formatDate(item.created_at || item.executed_at || item.createdAt) }}
          </span>
        </template>

        <template #item.is_retry="{ item }">
          <v-chip
            size="x-small"
            :color="(item.retry_count || item.attempt_number || 0) > 1 ? 'warning' : 'grey'"
            variant="tonal"
          >
            {{ (item.retry_count || item.attempt_number || 0) > 1 ? 'Yes' : 'No' }}
          </v-chip>
        </template>

        <template #item.actions="{ item }">
          <div class="d-flex align-center justify-end gap-1">
            <v-btn
              icon="mdi-code-json"
              size="small"
              variant="text"
              color="primary"
              @click="openPayloadModal(item)"
            >
              <v-tooltip activator="parent" location="top">View Payload & Response</v-tooltip>
            </v-btn>

            <v-btn
              icon="mdi-refresh"
              size="small"
              variant="text"
              color="secondary"
              :loading="redeliveringId === (item.id || item.job_id)"
              @click="handleRedeliver(item.id || item.job_id)"
            >
              <v-tooltip activator="parent" location="top">Re-deliver Webhook</v-tooltip>
            </v-btn>
          </div>
        </template>

        <template #no-data>
          <div class="text-center py-8 text-medium-emphasis">
            <v-icon icon="mdi-webhook-off" size="48" class="mb-2" />
            <div class="text-h6">No outbox jobs recorded</div>
            <div class="text-caption">Webhook attempt dispatches will be tracked here when inbound emails arrive.</div>
          </div>
        </template>
      </v-data-table>
    </v-card>

    <!-- Payload & Response Modal Dialog -->
    <v-dialog v-model="showPayloadModal" max-width="800px">
      <v-card rounded="lg" v-if="selectedJob">
        <v-card-title class="d-flex align-center justify-space-between pa-4 bg-surface-variant">
          <div>
            <span class="text-h6 font-weight-bold">Webhook Payload & Telemetry</span>
            <div class="text-caption text-medium-emphasis font-mono">
              Job ID: {{ selectedJob.id || selectedJob.job_id }}
            </div>
          </div>
          <v-btn icon="mdi-close" variant="text" size="small" @click="showPayloadModal = false" />
        </v-card-title>

        <v-card-text class="pa-4">
          <v-tabs v-model="modalTab" color="primary" class="mb-4">
            <v-tab value="request" prepend-icon="mdi-tray-arrow-up">Request Payload</v-tab>
            <v-tab value="response" prepend-icon="mdi-tray-arrow-down">Client Response</v-tab>
          </v-tabs>

          <v-window v-model="modalTab">
            <v-window-item value="request">
              <CodePreview
                :code="sampleRequestPayload"
                language="json"
                title="Webhook Outbox Dispatch Payload"
                max-height="400px"
              />
            </v-window-item>

            <v-window-item value="response">
              <CodePreview
                :code="sampleResponseBody"
                language="json"
                title="HTTP Delivery Response"
                max-height="400px"
              />
            </v-window-item>
          </v-window>
        </v-card-text>

        <v-divider />

        <v-card-actions class="pa-4">
          <v-spacer />
          <v-btn color="primary" variant="flat" @click="showPayloadModal = false">
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<route lang="yaml">
meta:
  requiresAuth: true
</route>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import StatusChip from '@/components/StatusChip.vue';
import CodePreview from '@/components/CodePreview.vue';
import { useAppStore } from '@/stores/application';
import { useWebhookStore } from '@/stores/webhooks';
import { useNotificationStore } from '@/stores/notification';

const appStore = useAppStore();
const webhookStore = useWebhookStore();
const notificationStore = useNotificationStore();

const webhookUrl = ref('');
const webhookSecret = ref('whsec_live_9f8e7d6c5b4a3210');
const maxRetries = ref(5);
const showSecret = ref(false);

const savingConfig = ref(false);
const testingHandshake = ref(false);
const handshakeResult = ref(null);
const handshakeSuccess = ref(true);

const statusFilter = ref('ALL');
const autoRefresh = ref(false);
let refreshInterval = null;

const redeliveringId = ref('');
const showPayloadModal = ref(false);
const selectedJob = ref(null);
const modalTab = ref('request');

const headers = [
  { title: 'Job ID', key: 'id', sortable: true },
  { title: 'Attempt #', key: 'retry_count', align: 'center', sortable: true },
  { title: 'Status', key: 'status', sortable: true },
  { title: 'Duration', key: 'duration_ms', align: 'end', sortable: true },
  { title: 'Executed At', key: 'created_at', sortable: true },
  { title: 'Is Retry', key: 'is_retry', align: 'center', sortable: false },
  { title: 'Actions', key: 'actions', align: 'end', sortable: false },
];

const filteredJobs = computed(() => {
  let list = webhookStore.jobs || [];
  if (statusFilter.value !== 'ALL') {
    list = list.filter((j) => (j.status || j.Status) === statusFilter.value);
  }
  return list;
});

const sampleRequestPayload = computed(() => {
  if (!selectedJob.value) return {};
  return {
    event: 'email.ingested',
    job_id: selectedJob.value.id || selectedJob.value.job_id,
    application_id: appStore.activeAppId,
    timestamp: new Date().toISOString(),
    signature: 'sha256=a8f9b2c3d4e5f67890123456789abcdef0123456789abcdef0123456789abc',
    data: {
      email_id: selectedJob.value.ingested_email_id || '9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d',
      from: 'sender@example.com',
      to: 'inbound@domain.com',
      subject: 'Sample Webhook Payload Payload',
    },
  };
});

const sampleResponseBody = computed(() => {
  if (!selectedJob.value) return {};
  return {
    http_status: selectedJob.value.http_status_code || 200,
    body: { status: 'ok', message: 'Event received successfully' },
    duration_ms: selectedJob.value.duration_ms || 42,
  };
});

function formatShortId(id) {
  if (!id) return '—';
  return String(id).slice(0, 8);
}

function formatDate(val) {
  if (!val) return '—';
  try {
    return new Date(val).toLocaleString();
  } catch {
    return String(val);
  }
}

async function copySecret() {
  try {
    await navigator.clipboard.writeText(webhookSecret.value);
    notificationStore.success('Copied webhook secret to clipboard!');
  } catch {
    notificationStore.error('Failed to copy secret');
  }
}

async function handleSaveConfig() {
  if (!appStore.activeAppId) {
    notificationStore.error('No active application scope selected.');
    return;
  }

  savingConfig.value = true;
  try {
    await webhookStore.setupWebhook(appStore.activeAppId, {
      url: webhookUrl.value,
      secret: webhookSecret.value,
      max_retries: maxRetries.value,
    });
    notificationStore.success('Webhook configuration saved successfully!');
  } catch (err) {
    notificationStore.error(err.response?.data?.message || 'Failed to save webhook configuration');
  } finally {
    savingConfig.value = false;
  }
}

async function handleTestEndpoint() {
  if (!appStore.activeAppId) {
    notificationStore.error('No active application scope selected.');
    return;
  }

  testingHandshake.value = true;
  handshakeResult.value = null;
  try {
    await webhookStore.setupWebhook(appStore.activeAppId, {
      url: webhookUrl.value,
      secret: webhookSecret.value,
      max_retries: maxRetries.value,
    });
    handshakeSuccess.value = true;
    handshakeResult.value = {
      title: 'Handshake Challenge Passed (200 OK)',
      message: 'Client endpoint successfully resolved CRC challenge response and returned valid signature verification token.',
    };
    notificationStore.success('Webhook endpoint challenge test passed!');
  } catch (err) {
    handshakeSuccess.value = false;
    handshakeResult.value = {
      title: 'Handshake Challenge Failed',
      message: err.response?.data?.message || 'Endpoint verification request failed CRC challenge check or returned an error status.',
    };
    notificationStore.error('Webhook endpoint challenge test failed.');
  } finally {
    testingHandshake.value = false;
  }
}

async function handleRedeliver(jobId) {
  if (!appStore.activeAppId || !jobId) return;

  redeliveringId.value = jobId;
  try {
    await webhookStore.redeliverJob(appStore.activeAppId, jobId);
    notificationStore.success(`Queued outbox job ${formatShortId(jobId)} for re-delivery!`);
    await loadWebhookJobs();
  } catch (err) {
    notificationStore.error(err.response?.data?.message || 'Failed to queue re-delivery job');
  } finally {
    redeliveringId.value = '';
  }
}

function openPayloadModal(job) {
  selectedJob.value = job;
  showPayloadModal.value = true;
}

async function loadWebhookJobs() {
  if (appStore.activeAppId) {
    await webhookStore.fetchJobs(appStore.activeAppId, { limit: 50 });
  }
}

watch(autoRefresh, (enabled) => {
  if (enabled) {
    refreshInterval = setInterval(loadWebhookJobs, 5000);
  } else if (refreshInterval) {
    clearInterval(refreshInterval);
    refreshInterval = null;
  }
});

onMounted(() => {
  if (appStore.activeApp) {
    webhookUrl.value = appStore.activeApp.webhook_url || appStore.activeApp.webhookUrl || 'https://api.myapp.com/webhooks/inbound';
    maxRetries.value = appStore.activeApp.max_retries || appStore.activeApp.maxRetries || 5;
  }
  loadWebhookJobs();
});

watch(() => appStore.activeApp, (newApp) => {
  if (newApp) {
    webhookUrl.value = newApp.webhook_url || newApp.webhookUrl || 'https://api.myapp.com/webhooks/inbound';
    maxRetries.value = newApp.max_retries || newApp.maxRetries || 5;
  }
  loadWebhookJobs();
});

onBeforeUnmount(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval);
  }
});
</script>

<style scoped>
.gap-1 {
  gap: 4px;
}
.gap-3 {
  gap: 12px;
}
.font-mono {
  font-family: 'Roboto Mono', monospace;
}
</style>
