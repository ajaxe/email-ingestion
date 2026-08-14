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
              :code="parsedRequestPayload"
              language="json"
              title="Webhook Outbox Dispatch Payload"
              max-height="400px"
            />
          </v-window-item>

          <v-window-item value="response">
            <CodePreview
              :code="parsedResponseBody"
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

const modalTab = ref("request");

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit("update:modelValue", val),
});

const parsedRequestPayload = computed(() => {
  if (!props.job) return {};
  const raw = props.job.requestPayload || props.job.request_payload;
  if (!raw) return { message: "No request payload recorded" };
  if (typeof raw === "object") return raw;
  try {
    return JSON.parse(raw);
  } catch {
    return raw;
  }
});

const parsedResponseBody = computed(() => {
  if (!props.job) return {};
  const raw = props.job.responseBody || props.job.response_body;
  if (!raw) {
    return {
      http_status: props.job.httpStatusCode || props.job.http_status_code || 0,
      message: "No response body recorded",
    };
  }
  if (typeof raw === "object") return raw;
  try {
    return JSON.parse(raw);
  } catch {
    return raw;
  }
});
</script>

<style scoped>
.font-mono {
  font-family: "Roboto Mono", monospace;
}
</style>