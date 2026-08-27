<template>
  <v-dialog v-model="internalModel" max-width="500">
    <v-card rounded="lg" class="pa-4">
      <div class="d-flex align-center gap-3 mb-3">
        <v-avatar color="error" variant="tonal" size="40">
          <v-icon icon="mdi-delete-outline" size="24" />
        </v-avatar>
        <div>
          <div class="text-h6 font-weight-bold">Soft Delete Email</div>
          <div class="text-caption text-medium-emphasis">
            Permanently remove raw MIME content & attachments from storage.
          </div>
        </div>
      </div>

      <v-divider class="mb-4" />

      <v-alert color="warning" variant="tonal" class="mb-4" density="compact">
        <div class="text-body-2 font-weight-medium">
          What happens when this email is soft-deleted?
        </div>
        <ul class="text-caption mt-1 ps-4">
          <li>Raw email body (HTML & text) and attachments will be purged from object storage.</li>
          <li>Email metadata and <strong>webhook delivery job logs are permanently preserved</strong>.</li>
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
          prepend-icon="mdi-delete"
          @click="confirmDelete"
        >
          Confirm Soft Delete
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

function confirmDelete() {
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
