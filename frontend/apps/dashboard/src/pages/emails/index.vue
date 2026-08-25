<template>
  <div>
    <!-- Filter Toolbar -->
    <v-card rounded="lg" variant="elevated" class="pa-4 mb-4">
      <v-row align="center" justify="space-between" class="ma-0">
        <v-col cols="12" md="3" class="pa-0">
          <div class="text-h6 font-weight-bold">Ingested Email Logs</div>
          <div class="text-caption text-medium-emphasis">
            Inspect inbound MIME stream & attachment records
          </div>
        </v-col>

        <v-col cols="12" md="9" class="pa-0 mt-3 mt-md-0 d-flex flex-wrap align-center justify-md-end gap-3">
          <v-text-field
            v-model="searchQuery"
            prepend-inner-icon="mdi-magnify"
            placeholder="Filter From or Subject..."
            variant="outlined"
            density="compact"
            hide-details
            clearable
            style="max-width: 220px"
          />

          <v-select
            v-model="selectedLocalPart"
            :items="localPartOptions"
            label="Local Part"
            variant="outlined"
            density="compact"
            hide-details
            style="max-width: 180px"
          />

          <v-switch
            v-model="autoRefresh"
            label="Auto-refresh"
            color="primary"
            hide-details
            density="compact"
            class="ms-2"
          />

          <v-btn
            icon
            variant="outlined"
            density="compact"
            color="primary"
            :loading="emailStore.loading"
            @click="refreshLogs"
          >
            <v-icon icon="mdi-refresh" />
            <v-tooltip activator="parent" location="top">Manual Refresh</v-tooltip>
          </v-btn>
        </v-col>
      </v-row>
    </v-card>

    <!-- Data Table Card -->
    <v-card rounded="lg" variant="elevated">
      <EmailList />
    </v-card>
  </div>
</template>

<route lang="yaml">
meta:
  requiresAuth: true
  title: Emails
</route>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useAppStore } from '@/stores/application';
import { useEmailStore } from '@/stores/emails';
import { useAddressStore } from '@/stores/addresses';
import EmailList from '@/components/emails/EmailList.vue';
import { storeToRefs } from 'pinia';

const appStore = useAppStore();
const emailStore = useEmailStore();
const addressStore = useAddressStore();

const { selectedLocalPart, searchQuery } = storeToRefs(emailStore);
const autoRefresh = ref(false);
let refreshInterval = null;



const localPartOptions = computed(() => {
  const options = [{ title: 'All Local Parts', value: 'ALL' }];
  if (addressStore.addresses) {
    for(const addr of addressStore.addresses) {
      const val = addr.local_part || addr.localPart;
      if (val) {
        options.push({ title: val, value: val });
      }
    }
  }
  return options;
});

async function refreshLogs() {
  if (appStore.activeAppId) {
    await emailStore.fetchEmails(appStore.activeAppId, { limit: 50 });
  }
}

watch(autoRefresh, (enabled) => {
  if (enabled) {
    refreshInterval = setInterval(refreshLogs, 5000);
  } else if (refreshInterval) {
    clearInterval(refreshInterval);
    refreshInterval = null;
  }
});

onMounted(async () => {
  if (appStore.activeAppId) {
    await Promise.allSettled([
      refreshLogs(),
      addressStore.fetchAddresses(appStore.activeAppId),
    ]);
  }
});

watch(() => appStore.activeAppId, () => {
  if (appStore.activeAppId) {
    refreshLogs();
    addressStore.fetchAddresses(appStore.activeAppId);
  }
});

onBeforeUnmount(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval);
  }
});
</script>

<style scoped>
.gap-3 {
  gap: 12px;
}
</style>
