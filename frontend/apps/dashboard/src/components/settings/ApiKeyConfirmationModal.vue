<template>
  <!-- One-Time M2M API Key Confirmation Dialog -->
  <v-dialog v-model="dialogVisible" max-width="560px" persistent>
    <v-card rounded="lg" class="pa-4">
      <v-card-title class="d-flex align-center gap-2 pa-0 mb-2">
        <v-icon color="warning" icon="mdi-key-star" />
        <span class="text-h6 font-weight-bold">M2M Ingestion API Key Generated</span>
      </v-card-title>

      <v-alert
        type="warning"
        variant="tonal"
        density="comfortable"
        class="mb-4"
      >
        <div class="font-weight-bold">Save your API key now</div>
        <div class="text-caption">
          This key will <strong>only be displayed once</strong>. Store it in a secure location (e.g. environment variables or secrets manager) as you will not be able to retrieve it again.
        </div>
      </v-alert>

      <div class="text-caption font-weight-bold text-medium-emphasis mb-1">
        Ingestion API Key
      </div>
      <v-text-field
        :model-value="apiKey"
        variant="outlined"
        density="comfortable"
        readonly
        class="font-mono mb-2"
      >
        <template #append-inner>
          <v-btn
            icon
            size="small"
            variant="text"
            color="primary"
            @click="copyApiKey"
          >
            <v-icon icon="mdi-content-copy" />
            <v-tooltip activator="parent" location="top">Copy API Key</v-tooltip>
          </v-btn>
        </template>
      </v-text-field>

      <v-card-actions class="pa-0 mt-4 justify-end">
        <v-btn color="primary" variant="flat" @click="dialogVisible = false">
          I Have Saved My Key
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { computed } from 'vue';
import { useNotificationStore } from '@/stores/notification';

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  apiKey: {
    type: String,
    default: '',
  },
});

const emit = defineEmits(['update:modelValue']);

const notificationStore = useNotificationStore();

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
});

async function copyApiKey() {
  try {
    await navigator.clipboard.writeText(props.apiKey);
    notificationStore.success('Copied API key to clipboard!');
  } catch {
    notificationStore.error('Failed to copy API key');
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
