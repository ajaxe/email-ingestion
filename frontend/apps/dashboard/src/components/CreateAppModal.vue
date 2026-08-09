<template>
  <v-dialog
    v-model="dialog"
    max-width="580px"
    persistent
  >
    <!-- Step 1: Create Application Form -->
    <v-card v-if="step === 1" rounded="lg">
      <v-card-item class="bg-surface-variant py-3 px-4">
        <template #prepend>
          <v-avatar color="primary" variant="flat" size="32" class="me-2">
            <v-icon icon="mdi-domain-plus" color="white" size="18" />
          </v-avatar>
        </template>
        <v-card-title class="text-h6 font-weight-bold">
          Create New Application Tenant
        </v-card-title>
        <v-card-subtitle class="text-caption text-medium-emphasis">
          Provision an isolated tenant partition for email routing and webhooks.
        </v-card-subtitle>
      </v-card-item>

      <v-divider />

      <v-card-text class="pt-4 pb-2">
        <v-form ref="formRef" @submit.prevent="handleCreate">
          <!-- Application Name Input -->
          <v-text-field
            v-model="appName"
            label="Application Name *"
            placeholder="e.g. Production Service, Staging Ingest"
            variant="outlined"
            density="compact"
            :rules="nameRules"
            counter="255"
            autofocus
            clearable
            class="mb-3"
            :error-messages="errorMessage"
            @input="errorMessage = ''"
          >
            <template #prepend-inner>
              <v-icon icon="mdi-card-text-outline" color="primary" class="me-1" />
            </template>
          </v-text-field>

          <!-- Quick Name Presets / Suggestions for Great UX -->
          <div class="mb-3">
            <div class="text-caption font-weight-medium text-medium-emphasis mb-1">
              Suggested naming presets:
            </div>
            <div class="d-flex flex-wrap gap-2">
              <v-chip
                v-for="preset in namePresets"
                :key="preset"
                size="small"
                variant="outlined"
                color="secondary"
                class="cursor-pointer"
                @click="applyPreset(preset)"
              >
                <v-icon start icon="mdi-label-outline" size="12" />
                {{ preset }}
              </v-chip>
            </div>
          </div>

          <!-- Auto-Provision Initial Address Checkbox Option -->
          <v-checkbox
            v-model="autoProvision"
            color="primary"
            density="compact"
            hide-details
            class="mb-3"
          >
            <template #label>
              <span class="text-body-2 font-weight-medium">
                Auto-provision initial 10-character email routing address
              </span>
            </template>
          </v-checkbox>

          <!-- Feature & Architecture Info Box -->
          <v-sheet rounded="md" class="pa-3 bg-grey-lighten-4 mb-2 border">
            <div class="d-flex align-start gap-2">
              <v-icon icon="mdi-shield-lock-outline" color="success" size="20" class="mt-1" />
              <div>
                <div class="text-subtitle-2 font-weight-bold text-high-emphasis">
                  Multi-Tenant Isolation Enforced
                </div>
                <div class="text-caption text-medium-emphasis">
                  Creating an application initializes a logically partitioned schema namespace, dedicated attachment storage prefix, and HMAC webhook security configuration.
                </div>
              </div>
            </div>
          </v-sheet>
        </v-form>
      </v-card-text>

      <v-divider />

      <v-card-actions class="px-4 py-3">
        <v-spacer />
        <v-btn
          variant="outlined"
          color="grey"
          :disabled="loading"
          @click="onClose"
        >
          Cancel
        </v-btn>
        <v-btn
          color="primary"
          variant="flat"
          :loading="loading"
          :disabled="!appName || appName.trim().length < 3"
          @click="handleCreate"
        >
          <v-icon start icon="mdi-plus" />
          Create & Configure
        </v-btn>
      </v-card-actions>
    </v-card>

    <!-- Step 2: Next Steps Configuration Guidance View -->
    <v-card v-else rounded="lg" class="pa-4">
      <NextStepsGuide
        :app="createdApp"
        :provisioned-address="provisionedAddress"
        @close="onClose"
      />
      <v-divider class="my-3" />
      <div class="d-flex justify-end">
        <v-btn
          color="primary"
          variant="flat"
          @click="onClose"
        >
          Done / Go to Dashboard
        </v-btn>
      </div>
    </v-card>
  </v-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { useAppStore } from '@/stores/application';
import { useAddressStore } from '@/stores/addresses';
import NextStepsGuide from '@/components/NextStepsGuide.vue';

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['update:modelValue', 'created']);

const appStore = useAppStore();
const addressStore = useAddressStore();

const formRef = ref(null);
const appName = ref('');
const autoProvision = ref(true);
const loading = ref(false);
const errorMessage = ref('');
const step = ref(1);
const createdApp = ref(null);
const provisionedAddress = ref(null);

const namePresets = [
  'Production Service',
  'Staging Environment',
  'Payment Webhooks Gateway',
  'Support Desk Ingest',
];

const nameRules = [
  v => !!v || 'Application name is required',
  v => (v && v.trim().length >= 3) || 'Name must be at least 3 characters',
  v => (v && v.length <= 255) || 'Name must be 255 characters or less',
];

const dialog = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
});

watch(dialog, (val) => {
  if (val) {
    step.value = 1;
    appName.value = '';
    errorMessage.value = '';
    createdApp.value = null;
    provisionedAddress.value = null;
  }
});

function applyPreset(preset) {
  appName.value = preset;
  errorMessage.value = '';
}

function onClose() {
  dialog.value = false;
  step.value = 1;
  appName.value = '';
  errorMessage.value = '';
  createdApp.value = null;
  provisionedAddress.value = null;
}

async function handleCreate() {
  if (!appName.value || appName.value.trim().length < 3) return;

  loading.value = true;
  errorMessage.value = '';

  try {
    const resApp = await appStore.createApp(appName.value.trim());
    createdApp.value = resApp || appStore.activeApp;

    // Auto provision initial address if selected
    if (autoProvision.value && appStore.activeAppId) {
      try {
        const addrRes = await addressStore.provisionAddress(
          appStore.activeAppId,
          'Primary Inbound Address'
        );
        provisionedAddress.value = addrRes?.data || addrRes;
      } catch (e) {
        console.warn('Auto-provisioning initial address failed:', e);
      }
    }

    emit('created', createdApp.value);
    // Move to Step 2: Next Steps Configuration Guidance
    step.value = 2;
  } catch (err) {
    const backendMsg = err.response?.data?.message || err.message || 'Failed to create application';
    errorMessage.value = backendMsg;
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
.cursor-pointer {
  cursor: pointer;
}
</style>
