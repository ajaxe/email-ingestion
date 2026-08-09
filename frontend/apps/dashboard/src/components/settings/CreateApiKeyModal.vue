<template>
  <v-dialog v-model="dialogVisible" max-width="520px" persistent>
    <v-card rounded="lg" class="pa-4">
      <v-card-title class="d-flex align-center gap-2 pa-0 mb-2">
        <v-icon color="primary" icon="mdi-key-plus" />
        <span class="text-h6 font-weight-bold">Create New API Key</span>
      </v-card-title>

      <v-card-text class="pa-0 my-3">
        <v-form ref="formRef" v-model="formValid" @submit.prevent="handleSubmit">
          <!-- Key Name -->
          <div class="text-caption font-weight-bold text-medium-emphasis mb-1">
            Key Name / Description
          </div>
          <v-text-field
            v-model="form.name"
            placeholder="e.g. Staging Integration, Worker Service"
            variant="outlined"
            density="comfortable"
            :rules="[v => !!v || 'Key name is required']"
            class="mb-3"
          />

          <!-- Environment / Key Prefix -->
          <div class="text-caption font-weight-bold text-medium-emphasis mb-1">
            Environment Scope
          </div>
          <v-radio-group v-model="form.keyPrefix" inline density="compact" class="mb-3">
            <v-radio value="eg_live_" color="success">
              <template #label>
                <div class="d-flex align-center gap-2">
                  <v-chip color="success" size="x-small" variant="tonal" class="font-weight-bold">
                    LIVE
                  </v-chip>
                  <span class="font-mono text-caption">eg_live_</span>
                </div>
              </template>
            </v-radio>
            <v-radio value="eg_test_" color="warning">
              <template #label>
                <div class="d-flex align-center gap-2">
                  <v-chip color="warning" size="x-small" variant="tonal" class="font-weight-bold">
                    TEST
                  </v-chip>
                  <span class="font-mono text-caption">eg_test_</span>
                </div>
              </template>
            </v-radio>
          </v-radio-group>

          <!-- Expiration Period -->
          <div class="text-caption font-weight-bold text-medium-emphasis mb-1">
            Expiration Duration
          </div>
          <v-select
            v-model="form.expireDays"
            :items="expirationOptions"
            item-title="title"
            item-value="value"
            variant="outlined"
            density="comfortable"
            class="mb-2"
          />
        </v-form>
      </v-card-text>

      <v-card-actions class="pa-0 mt-3 justify-end gap-2">
        <v-btn variant="text" :disabled="submitting" @click="dialogVisible = false">
          Cancel
        </v-btn>
        <v-btn
          color="primary"
          variant="flat"
          :loading="submitting"
          :disabled="!formValid"
          @click="handleSubmit"
        >
          Generate API Key
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { computed, reactive, ref, watch } from 'vue';

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  submitting: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['update:modelValue', 'submit']);

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
});

const formRef = ref(null);
const formValid = ref(false);

const form = reactive({
  name: '',
  keyPrefix: 'eg_live_',
  expireDays: 365,
});

const expirationOptions = [
  { title: '30 Days', value: 30 },
  { title: '90 Days', value: 90 },
  { title: '1 Year (Default)', value: 365 },
  { title: '2 Years', value: 730 },
];

watch(dialogVisible, (visible) => {
  if (visible) {
    form.name = '';
    form.keyPrefix = 'eg_live_';
    form.expireDays = 365;
  }
});

function handleSubmit() {
  if (!form.name) return;
  emit('submit', {
    name: form.name,
    keyPrefix: form.keyPrefix,
    expireDays: form.expireDays,
  });
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
