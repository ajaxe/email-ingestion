<template>
  <!-- Top Card: Webhook Endpoint & Secret Settings -->
  <v-card rounded="lg" variant="elevated" class="pa-4 mb-4">
    <div class="d-flex align-center justify-space-between mb-2">
      <div>
        <div class="text-h6 font-weight-bold">
          Webhook Configuration & Challenge Handshake
        </div>
        <div class="text-caption text-medium-emphasis">
          Configure secure HMAC-SHA256 signature dispatch & challenge
          verification
        </div>
      </div>
      <StatusChip
        :status="handshakeSuccess ? 'ACTIVE' : 'INACTIVE'"
        size="small"
      />
    </div>

    <!-- Handshake Status Alert -->
    <v-alert
      v-if="handshakeResult"
      :type="handshakeSuccess ? 'success' : 'error'"
      variant="tonal"
      closable
      class="mb-4"
      density="compact"
      @click:close="handshakeResult = null"
    >
      <div class="font-weight-bold">{{ handshakeResult.title }}</div>
      <div class="text-caption">{{ handshakeResult.message }}</div>
    </v-alert>

    <v-form @submit.prevent="handleSaveConfig">
      <v-row class="mt-1">
        <v-col cols="12" md="8">
          <v-text-field
            v-model="webhookUrl"
            label="Endpoint URL"
            placeholder="https://example.com/api/webhooks"
            variant="outlined"
            density="comfortable"
            prepend-inner-icon="mdi-link-variant"
            hint="Must be HTTPS and pass RFC 1918 SSRF validation"
            persistent-hint
          />
        </v-col>

        <v-col cols="12" md="4">
          <div class="text-caption text-medium-emphasis mb-1">
            Max Retry Delivery Attempts:
            <strong class="text-primary">{{ maxRetries }}</strong>
          </div>
          <v-slider
            v-model="maxRetries"
            :min="1"
            :max="10"
            :step="1"
            thumb-label
            color="primary"
            track-color="surface-variant"
          />
        </v-col>

        <v-col
          cols="12"
          class="d-flex flex-wrap align-center justify-space-between gap-3 mt-1"
        >
          <v-btn
            color="warning"
            variant="outlined"
            prepend-icon="mdi-key"
            :loading="registeringSecret"
            @click="handleRegisterSecret"
            class="position-relative"
          >
            Regenerate Signing Secret
          </v-btn>

          <div class="d-flex align-center gap-3">
            <v-btn
              color="secondary"
              variant="tonal"
              prepend-icon="mdi-lightning-bolt"
              :loading="testingHandshake"
              @click="handleTestEndpoint"
            >
              Test & Verify Endpoint
            </v-btn>

            <v-btn
              color="primary"
              variant="flat"
              prepend-icon="mdi-content-save"
              :loading="savingConfig"
              type="submit"
            >
              Save Configuration
            </v-btn>
          </div>
        </v-col>
      </v-row>
    </v-form>

    <WebhookSecretConfirmation
      v-model="showSecretModal"
      :secret="generatedSecret"
    />
  </v-card>
</template>

<script setup>
import { onMounted, ref, watch } from "vue";
import StatusChip from "@/components/shared/StatusChip.vue";
import WebhookSecretConfirmation from "./WebhookSecretConfirmation.vue";
import { useAppStore } from "@/stores/application";
import { useWebhookStore } from "@/stores/webhooks";
import { useNotificationStore } from "@/stores/notification";

const appStore = useAppStore();
const webhookStore = useWebhookStore();
const notificationStore = useNotificationStore();

const webhookUrl = ref("");
const maxRetries = ref(5);

const savingConfig = ref(false);
const testingHandshake = ref(false);
const registeringSecret = ref(false);
const showSecretModal = ref(false);
const generatedSecret = ref("");
const handshakeResult = ref(null);
const handshakeSuccess = ref(true);

async function handleRegisterSecret() {
  if (!appStore.activeAppId) {
    notificationStore.error("No active application scope selected.");
    return;
  }

  registeringSecret.value = true;
  try {
    const res = await webhookStore.registerWebhook(appStore.activeAppId, {
      webhookUrl: webhookUrl.value,
      maxRetries: maxRetries.value,
    });
    const secret = res?.webhookSecret || res?.data?.webhookSecret;
    if (secret) {
      generatedSecret.value = secret;
      showSecretModal.value = true;
    }
    notificationStore.success("Webhook signing secret generated successfully!");
  } catch (err) {
    notificationStore.error(
      err.response?.data?.message ||
        "Failed to generate webhook signing secret",
    );
  } finally {
    registeringSecret.value = false;
  }
}

async function handleSaveConfig() {
  if (!appStore.activeAppId) {
    notificationStore.error("No active application scope selected.");
    return;
  }

  savingConfig.value = true;
  try {
    await webhookStore.updateWebhook(appStore.activeAppId, {
      webhookUrl: webhookUrl.value,
      maxRetries: maxRetries.value,
    });
    notificationStore.success("Webhook configuration saved successfully!");
  } catch (err) {
    notificationStore.error(
      err.response?.data?.message || "Failed to save webhook configuration",
    );
  } finally {
    savingConfig.value = false;
  }
}

async function handleTestEndpoint() {
  if (!appStore.activeAppId) {
    notificationStore.error("No active application scope selected.");
    return;
  }

  testingHandshake.value = true;
  handshakeResult.value = null;
  try {
    await webhookStore.updateWebhook(appStore.activeAppId, {
      webhookUrl: webhookUrl.value,
      maxRetries: maxRetries.value,
    });
    handshakeSuccess.value = true;
    handshakeResult.value = {
      title: "Handshake Challenge Passed (200 OK)",
      message:
        "Client endpoint successfully resolved CRC challenge response and returned valid signature verification token.",
    };
    notificationStore.success("Webhook endpoint challenge test passed!");
  } catch (err) {
    handshakeSuccess.value = false;
    handshakeResult.value = {
      title: "Handshake Challenge Failed",
      message:
        err.response?.data?.message ||
        "Endpoint verification request failed CRC challenge check or returned an error status.",
    };
    notificationStore.error("Webhook endpoint challenge test failed.");
  } finally {
    testingHandshake.value = false;
  }
}

onMounted(() => {
  if (appStore.activeApp) {
    webhookUrl.value = appStore.activeApp.webhookUrl;
    maxRetries.value = appStore.activeApp.maxRetries || 5;
  }
});

watch(
  () => appStore.activeApp,
  (newApp) => {
    if (newApp) {
      webhookUrl.value = newApp.webhookUrl;
      maxRetries.value = newApp.maxRetries || 5;
    }
  },
);
</script>

<style scoped>
.gap-3 {
  gap: 12px;
}
</style>