<template>
  <div>
    <!-- Page Header -->
    <v-card rounded="lg" variant="elevated" class="pa-4 mb-4">
      <div class="text-h6 font-weight-bold">API Keys & Security Settings</div>
      <div class="text-caption text-medium-emphasis">
        Manage M2M API credentials and external service authentication
      </div>
    </v-card>

    <v-row>
      <v-col cols="12">
        <ApiKeyCard />
      </v-col>
    </v-row>
  </div>
</template>

<route lang="yaml">
meta:
  requiresAuth: true
  title: Settings
</route>

<script setup>
import ApiKeyCard from '@/components/settings/ApiKeyCard.vue';
</script>
