<template>
  <v-card rounded="lg" variant="elevated" class="pa-4 mt-4">
    <div class="d-flex align-center justify-space-between mb-3">
      <div class="d-flex align-center gap-2">
        <v-icon icon="mdi-webhook" color="primary" size="22" />
        <span class="text-h6 font-weight-bold">Webhook Delivery Job History</span>
      </div>
      <v-chip size="small" color="primary" variant="tonal" class="font-weight-medium">
        Preserved History ({{ jobs.length }})
      </v-chip>
    </div>

    <v-divider class="mb-4" />

    <!-- Loading state -->
    <div v-if="loading" class="text-center py-6">
      <v-progress-circular indeterminate color="primary" size="28" class="mb-2" />
      <div class="text-caption text-medium-emphasis">Fetching webhook delivery jobs & attempt logs...</div>
    </div>

    <!-- Jobs Table / Expansion List -->
    <div v-else-if="jobs.length > 0">
      <v-expansion-panels variant="accordion">
        <v-expansion-panel
          v-for="job in jobs"
          :key="job.id"
          rounded="lg"
          class="mb-2 border"
        >
          <v-expansion-panel-title class="py-2 px-3">
            <div class="d-flex align-center justify-space-between w-100 me-2 flex-wrap gap-2">
              <div class="d-flex align-center gap-2">
                <StatusChip :status="job.status" />
                <span class="font-mono text-caption text-medium-emphasis">
                  Job: {{ job.id }}
                </span>
              </div>

              <div class="d-flex align-center gap-3 text-caption text-medium-emphasis">
                <span>Retries: {{ job.retryCount || job.retry_count || 0 }}</span>
                <span>Scheduled: {{ formatDate(job.nextDeliveryAt || job.next_delivery_at) }}</span>
                <v-chip
                  v-if="job.lastHttpStatusCode || job.last_http_status_code"
                  size="x-small"
                  :color="isSuccessStatus(job.lastHttpStatusCode || job.last_http_status_code) ? 'success' : 'error'"
                  variant="outlined"
                  class="font-mono"
                >
                  HTTP {{ job.lastHttpStatusCode || job.last_http_status_code }}
                </v-chip>
              </div>
            </div>
          </v-expansion-panel-title>

          <v-expansion-panel-text>
            <div class="pa-2 bg-surface-variant rounded">
              <div class="text-subtitle-2 font-weight-bold mb-2">Latest Attempt Log Details</div>
              
              <v-row density="compact">
                <v-col cols="12" sm="4">
                  <span class="text-caption text-medium-emphasis">Attempt Number:</span>
                  <div class="text-body-2 font-mono">{{ job.lastAttemptNumber || job.last_attempt_number || 1 }}</div>
                </v-col>
                <v-col cols="12" sm="4">
                  <span class="text-caption text-medium-emphasis">HTTP Response Code:</span>
                  <div class="text-body-2 font-mono">{{ job.lastHttpStatusCode || job.last_http_status_code || "N/A" }}</div>
                </v-col>
                <v-col cols="12" sm="4">
                  <span class="text-caption text-medium-emphasis">Execution Duration:</span>
                  <div class="text-body-2 font-mono">{{ job.lastDurationMs || job.last_duration_ms || 0 }} ms</div>
                </v-col>
              </v-row>

              <v-divider class="my-2" />

              <div class="text-caption text-medium-emphasis mb-1">Last Response Body Snippet:</div>
              <CodePreview
                :code="job.lastResponseBody || job.last_response_body || '(No response body log captured)'"
                language="text"
                max-height="120px"
              />
            </div>
          </v-expansion-panel-text>
        </v-expansion-panel>
      </v-expansion-panels>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-6 text-medium-emphasis">
      <v-icon icon="mdi-webhook-off" size="36" class="mb-2" />
      <div>No webhook delivery jobs registered for this email.</div>
    </div>
  </v-card>
</template>

<script setup>
import StatusChip from "@/components/shared/StatusChip.vue";
import CodePreview from "@/components/shared/CodePreview.vue";

defineProps({
  jobs: {
    type: Array,
    default: () => [],
  },
  loading: {
    type: Boolean,
    default: false,
  },
});

function formatDate(val) {
  if (!val) return "—";
  try {
    return new Date(val).toLocaleString();
  } catch {
    return String(val);
  }
}

function isSuccessStatus(code) {
  return typeof code === "number" && code >= 200 && code < 300;
}
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
.gap-3 {
  gap: 12px;
}
.font-mono {
  font-family: "Roboto Mono", monospace;
}
</style>
