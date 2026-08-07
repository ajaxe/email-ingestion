<template>
  <div>
    <!-- Top Metrics Grid -->
    <v-row class="mb-4">
      <v-col cols="12" sm="6" md="3">
        <StatsWidget
          title="Total Ingested Emails"
          :value="totalEmails"
          icon="mdi-email-outline"
          color="primary"
          trend="+12.5%"
          subtitle="All time inbound stream"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <StatsWidget
          title="Active Routing Addresses"
          :value="activeAddressesRatio"
          icon="mdi-at"
          color="success"
          trend="+2"
          subtitle="Provisioned local-parts"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <StatsWidget
          title="Webhook Success Rate"
          :value="webhookSuccessRate"
          icon="mdi-webhook"
          color="info"
          trend="+0.8%"
          subtitle="Handshake & dispatch"
        />
      </v-col>
      <v-col cols="12" sm="6" md="3">
        <StatsWidget
          title="Failed Outbox Jobs"
          :value="failedJobsCount"
          icon="mdi-alert-circle-outline"
          color="warning"
          trend="0"
          subtitle="Requires attention"
        />
      </v-col>
    </v-row>

    <!-- Main Content Layout -->
    <v-row>
      <!-- Left Column: Recent Ingestion Stream -->
      <v-col cols="12" md="8">
        <RecentIgestionStream />
      </v-col>

      <!-- Right Column: Quick Actions & Gateway Status -->
      <v-col cols="12" md="4">
        <QuickActions />
        <GatewayStatus />
      </v-col>
    </v-row>
  </div>
</template>

<route lang="yaml">
meta:
  requiresAuth: true
</route>

<script setup>
import { computed, onMounted, watch } from 'vue';
import StatsWidget from '@/components/dashboard/StatsWidget.vue';
import RecentIgestionStream from '@/components/dashboard/RecentIgestionStream.vue';
import QuickActions from '@/components/dashboard/QuickActions.vue';
import GatewayStatus from '@/components/dashboard/GatewayStatus.vue';
import { useAppStore } from '@/stores/application';
import { useEmailStore } from '@/stores/emails';
import { useAddressStore } from '@/stores/addresses';
import { useWebhookStore } from '@/stores/webhooks';

const appStore = useAppStore();
const emailStore = useEmailStore();
const addressStore = useAddressStore();
const webhookStore = useWebhookStore();

const totalEmails = computed(() => {
  return emailStore.emails ? emailStore.emails.length : 0;
});

const activeAddressesRatio = computed(() => {
  const total = addressStore.addresses ? addressStore.addresses.length : 0;
  const active = addressStore.activeAddresses ? addressStore.activeAddresses.length : 0;
  return `${active} / ${total}`;
});

const webhookSuccessRate = computed(() => {
  const jobs = webhookStore.jobs || [];
  if (jobs.length === 0) return '100%';
  const success = jobs.filter(j => (j.status || j.Status) === 'SUCCESS').length;
  return `${((success / jobs.length) * 100).toFixed(1)}%`;
});

const failedJobsCount = computed(() => {
  const jobs = webhookStore.jobs || [];
  return jobs.filter(j => ['FAILED', 'DEAD'].includes(j.status || j.Status)).length;
});

async function loadDashboardData() {
  if (!appStore.activeAppId) return;
  await Promise.allSettled([
    emailStore.fetchEmails(appStore.activeAppId, { limit: 10 }),
    addressStore.fetchAddresses(appStore.activeAppId),
    webhookStore.fetchJobs(appStore.activeAppId, { limit: 20 }),
  ]);
}

onMounted(() => {
  loadDashboardData();
});

watch(() => appStore.activeAppId, () => {
  loadDashboardData();
});
</script>

<style scoped>
</style>
