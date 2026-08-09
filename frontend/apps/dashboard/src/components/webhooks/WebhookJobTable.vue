<template>
  <!-- Bottom Card: Webhook Delivery Outbox Sandbox -->
  <v-card rounded="lg" variant="elevated" class="pa-4">
    <v-row align="center" justify="space-between" class="ma-0 mb-4">
      <v-col cols="12" md="4" class="pa-0">
        <div class="text-h6 font-weight-bold">
          Outbox Delivery Attempt History
        </div>
        <div class="text-caption text-medium-emphasis">
          Transactional outbox retry jobs & delivery execution telemetry
        </div>
      </v-col>

      <v-col
        cols="12"
        md="8"
        class="pa-0 mt-3 mt-md-0 d-flex flex-wrap align-center justify-md-end gap-3"
      >
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
          icon
          variant="outlined"
          density="compact"
          color="primary"
          :loading="webhookStore.loading"
          @click="loadWebhookJobs"
        >
          <v-icon icon="mdi-refresh" />
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
          {{ formatShortId(item.id) }}
        </span>
      </template>

      <template #item.retry_count="{ item }">
        <v-chip
          size="x-small"
          color="primary"
          variant="outlined"
          class="font-mono"
        >
          #{{ item.attemptNumber || item.retryCount || 1 }}
        </v-chip>
      </template>

      <template #item.status="{ item }">
        <StatusChip :status="item.status || item.Status" />
      </template>

      <template #item.duration_ms="{ item }">
        <span class="font-mono text-caption">
          {{ itemitem.durationMs ? `${item.durationMs} ms` : "—" }}
        </span>
      </template>

      <template #item.created_at="{ item }">
        <span class="font-mono text-caption">
          {{
            formatDate(item.createdAt)
          }}
        </span>
      </template>

      <template #item.is_retry="{ item }">
        <v-chip
          size="x-small"
          :color="
            (item.retryCount || item.attemptNumber || 0) > 1
              ? 'warning'
              : 'grey'
          "
          variant="tonal"
        >
          {{
            (item.retryCount || item.attemptNumber  || 0) > 1 ? "Yes" : "No"
          }}
        </v-chip>
      </template>

      <template #item.actions="{ item }">
        <div class="d-flex align-center justify-end gap-1">
          <v-btn
            icon
            size="small"
            variant="text"
            color="primary"
            @click="openPayloadModal(item)"
          >
            <v-icon icon="mdi-code-json" />
            <v-tooltip activator="parent" location="top"
              >View Payload & Response</v-tooltip
            >
          </v-btn>

          <v-btn
            icon
            size="small"
            variant="text"
            color="secondary"
            :loading="redeliveringId === item.id"
            @click="handleRedeliver(item.id)"
          >
            <v-icon icon="mdi-refresh" />
            <v-tooltip activator="parent" location="top"
              >Re-deliver Webhook</v-tooltip
            >
          </v-btn>
        </div>
      </template>

      <template #no-data>
        <div class="text-center py-8 text-medium-emphasis">
          <v-icon icon="mdi-webhook-off" size="48" class="mb-2" />
          <div class="text-h6">No outbox jobs recorded</div>
          <div class="text-caption">
            Webhook attempt dispatches will be tracked here when inbound emails
            arrive.
          </div>
        </div>
      </template>
    </v-data-table>

    <WebhookJobPayload v-model="showPayloadModal" :job="selectedJob" />
  </v-card>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import StatusChip from "@/components/StatusChip.vue";
import WebhookJobPayload from "./WebhookJobPayload.vue";
import { useAppStore } from "@/stores/application";
import { useWebhookStore } from "@/stores/webhooks";
import { useNotificationStore } from "@/stores/notification";

const appStore = useAppStore();
const webhookStore = useWebhookStore();
const notificationStore = useNotificationStore();

const statusFilter = ref("ALL");
const autoRefresh = ref(false);
let refreshInterval = null;

const redeliveringId = ref("");
const showPayloadModal = ref(false);
const selectedJob = ref(null);

const headers = [
  { title: "Job ID", key: "id", sortable: true },
  { title: "Attempt #", key: "retry_count", align: "center", sortable: true },
  { title: "Status", key: "status", sortable: true },
  { title: "Duration", key: "duration_ms", align: "end", sortable: true },
  { title: "Executed At", key: "created_at", sortable: true },
  { title: "Is Retry", key: "is_retry", align: "center", sortable: false },
  { title: "Actions", key: "actions", align: "end", sortable: false },
];

const filteredJobs = computed(() => {
  let list = webhookStore.jobs || [];
  if (statusFilter.value !== "ALL") {
    list = list.filter((j) => (j.status || j.Status) === statusFilter.value);
  }
  return list;
});

function formatShortId(id) {
  if (!id) return "—";
  return String(id).slice(0, 8);
}

function formatDate(val) {
  if (!val) return "—";
  try {
    return new Date(val).toLocaleString();
  } catch {
    return String(val);
  }
}

async function handleRedeliver(jobId) {
  if (!appStore.activeAppId || !jobId) return;

  redeliveringId.value = jobId;
  try {
    await webhookStore.redeliverJob(appStore.activeAppId, jobId);
    notificationStore.success(
      `Queued outbox job ${formatShortId(jobId)} for re-delivery!`,
    );
    await loadWebhookJobs();
  } catch (err) {
    notificationStore.error(
      err.response?.data?.message || "Failed to queue re-delivery job",
    );
  } finally {
    redeliveringId.value = "";
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
  loadWebhookJobs();
});

watch(
  () => appStore.activeApp,
  () => {
    loadWebhookJobs();
  },
);

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
  font-family: "Roboto Mono", monospace;
}
</style>
