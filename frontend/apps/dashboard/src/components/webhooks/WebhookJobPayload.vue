<template>
  <!-- Payload & Response Modal Dialog -->
  <v-dialog v-model="dialogVisible" max-width="800px">
    <v-card rounded="lg" v-if="job">
      <v-card-title
        class="d-flex align-center justify-space-between pa-4 bg-surface-variant"
      >
        <div>
          <span class="text-h6 font-weight-bold"
            >Webhook Payload & Telemetry</span
          >
          <div class="text-caption text-medium-emphasis font-mono">
            Job ID: {{ job.id || job.job_id }}
          </div>
        </div>
        <v-btn
          icon="mdi-close"
          variant="text"
          size="small"
          @click="dialogVisible = false"
        />
      </v-card-title>

      <v-card-text class="pa-4">
        <v-tabs v-model="modalTab" color="primary" class="mb-4">
          <v-tab value="request" prepend-icon="mdi-tray-arrow-up"
            >Request Payload</v-tab
          >
          <v-tab value="response" prepend-icon="mdi-tray-arrow-down"
            >Client Response</v-tab
          >
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
        <v-btn color="primary" variant="flat" @click="dialogVisible = false">
          Close
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { computed, ref } from "vue";
import CodePreview from "@/components/CodePreview.vue";
import { useAppStore } from "@/stores/application";

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  job: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(["update:modelValue"]);

const appStore = useAppStore();
const modalTab = ref("request");

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit("update:modelValue", val),
});

const sampleRequestPayload = computed(() => {
  if (!props.job) return {};
  return {
    event: "email.ingested",
    job_id: props.job.id || props.job.job_id,
    application_id: appStore.activeAppId,
    timestamp: new Date().toISOString(),
    signature:
      "sha256=a8f9b2c3d4e5f67890123456789abcdef0123456789abcdef0123456789abc",
    data: {
      email_id:
        props.job.ingested_email_id ||
        "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
      from: "sender@example.com",
      to: "inbound@domain.com",
      subject: "Sample Webhook Payload Payload",
    },
  };
});

const sampleResponseBody = computed(() => {
  if (!props.job) return {};
  return {
    http_status: props.job.http_status_code || 200,
    body: { status: "ok", message: "Event received successfully" },
    duration_ms: props.job.duration_ms || 42,
  };
});
</script>

<style scoped>
.font-mono {
  font-family: "Roboto Mono", monospace;
}
</style>