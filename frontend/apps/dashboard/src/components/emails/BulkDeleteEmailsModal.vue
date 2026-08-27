<template>
  <v-dialog v-model="internalModel" max-width="520">
    <v-card rounded="lg" class="pa-4">
      <div class="d-flex align-center gap-3 mb-3">
        <v-avatar color="error" variant="tonal" size="40">
          <v-icon icon="mdi-delete-sweep-outline" size="24" />
        </v-avatar>
        <div>
          <div class="text-h6 font-weight-bold">
            Bulk Soft Delete {{ count }} {{ count === 1 ? 'Email' : 'Emails' }}
          </div>
          <div class="text-caption text-medium-emphasis">
            Asynchronously purge object storage content for selected records.
          </div>
        </div>
      </div>

      <v-divider class="mb-4" />

      <v-alert color="warning" variant="tonal" class="mb-4" density="compact">
        <div class="text-body-2 font-weight-medium">
          Confirm bulk soft-delete action:
        </div>
        <ul class="text-caption mt-1 ps-4">
          <li><strong>{{ count }}</strong> selected emails will have their raw content and attachments purged from S3 object storage.</li>
          <li>All metadata and <strong>webhook delivery job histories will be retained</strong> intact.</li>
          <li>Storage cleanup jobs will be processed asynchronously by background worker streams.</li>
        </ul>
      </v-alert>

      <div class="d-flex justify-end gap-2 mt-4">
        <v-btn variant="text" :disabled="loading" @click="internalModel = false">
          Cancel
        </v-btn>
        <v-btn
          color="error"
          variant="flat"
          :loading="loading"
          prepend-icon="mdi-delete-sweep"
          @click="confirmBulkDelete"
        >
          Bulk Soft Delete ({{ count }})
        </v-btn>
      </div>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  count: {
    type: Number,
    default: 0,
  },
  loading: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["update:modelValue", "confirm"]);

const internalModel = computed({
  get: () => props.modelValue,
  set: (val) => emit("update:modelValue", val),
});

function confirmBulkDelete() {
  emit("confirm");
}
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
.gap-3 {
  gap: 12px;
}
</style>
