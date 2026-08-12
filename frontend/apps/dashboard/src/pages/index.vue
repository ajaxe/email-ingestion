<template>
  <div>
    <!-- Zero Applications Banner (Shown when user has zero applications) -->
    <NoApplicationsBanner v-if="!hasApps" />

    <!-- Active Dashboard Workspace (Shown when applications exist) -->
    <template v-else>
      <!-- Configuration Next Steps Banner (Shown when tenant requires setup) -->
      <div v-if="needsConfiguration" class="mb-6">
        <NextStepsGuide />
      </div>

      <!-- Top Metrics Grid -->
      <v-row class="mb-4">
        <v-col cols="12" sm="6" md="3">
          <StatsWidget
            title="Total Ingested Emails"
            :value="totalEmails"
            icon="mdi-email-outline"
            color="primary"
            trend="0"
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
    </template>
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
import NextStepsGuide from '@/components/NextStepsGuide.vue';
import NoApplicationsBanner from '@/components/dashboard/NoApplicationsBanner.vue';
import { useAppStore } from '@/stores/application';
import { useEmailStore } from '@/stores/emails';

const appStore = useAppStore();
const emailStore = useEmailStore();

const hasApps = computed(() => appStore.applications.length > 0);

const needsConfiguration = computed(() => {
  if (!appStore.activeAppId) return false;
  const noAddresses = appStore.stats.totalAddresses === 0;
  const noWebhook = !appStore.activeApp?.webhookUrl;
  return (noAddresses || noWebhook) && !!appStore.activeApp;
});

const totalEmails = computed(() => {
  return appStore.stats.totalEmails;
});

const activeAddressesRatio = computed(() => {
  const total = appStore.stats.totalAddresses
  const active = appStore.stats.activeAddresses
  if (total === 0) return '0%';
  return `${active} / ${total}`;
});

const webhookSuccessRate = computed(() => {
  return `${((appStore.stats.webhookSuccessRate) * 100).toFixed(1)}%`;
});

const failedJobsCount = computed(() => {
  return appStore.stats.failWebhookJobCount;
});

async function loadDashboardData() {
  if (!appStore.activeAppId) return;
  await Promise.allSettled([
    emailStore.fetchEmails(appStore.activeAppId, { limit: 10 }),
    appStore.fetchStatistics(appStore.activeAppId),
  ]);
}

onMounted(() => {
  void loadDashboardData();
});

watch(() => appStore.activeAppId, () => {
  void loadDashboardData();
});
</script>

<style scoped>
</style>
