<template>
  <v-dialog
    v-model="dialog"
    max-width="500px"
    persistent
  >
    <v-card rounded="lg">
      <v-card-item class="bg-surface-variant py-3 px-4">
        <template #prepend>
          <v-icon :color="color" icon="mdi-alert-circle" class="me-2" />
        </template>
        <v-card-title class="text-h6 font-weight-bold">
          {{ title }}
        </v-card-title>
      </v-card-item>

      <v-card-text class="pt-4 pb-3">
        <p class="text-body-1 text-medium-emphasis">
          {{ message }}
        </p>
        <slot />
      </v-card-text>

      <v-divider />

      <v-card-actions class="px-4 py-3">
        <v-spacer />
        <v-btn
          variant="outlined"
          color="grey"
          :disabled="loading"
          @click="onCancel"
        >
          {{ cancelText }}
        </v-btn>
        <v-btn
          :color="color"
          variant="flat"
          :loading="loading"
          @click="onConfirm"
        >
          {{ confirmText }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: 'Confirm Action',
  },
  message: {
    type: String,
    default: 'Are you sure you want to perform this action? This action cannot be undone.',
  },
  confirmText: {
    type: String,
    default: 'Confirm',
  },
  cancelText: {
    type: String,
    default: 'Cancel',
  },
  color: {
    type: String,
    default: 'error',
  },
  loading: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['update:modelValue', 'confirm', 'cancel']);

const dialog = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
});

function onCancel() {
  emit('cancel');
  dialog.value = false;
}

function onConfirm() {
  emit('confirm');
}
</script>
