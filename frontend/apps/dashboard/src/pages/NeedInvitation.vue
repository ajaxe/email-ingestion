<template>
  <div class="min-vh-100 bg-background d-flex flex-column justify-space-between pa-4 pa-sm-8">
    <!-- Top Branding & Navigation Bar -->
    <v-container max-width="840" class="pa-0 mb-6">
      <div class="d-flex align-center justify-space-between px-2">
        <AppName />

        <div class="d-flex align-center gap-2">
          <v-btn
            variant="text"
            icon
            :color="theme.global.current.value.dark ? 'amber' : 'primary'"
            @click="toggleTheme"
          >
            <v-icon icon="mdi-theme-light-dark" />
            <v-tooltip activator="parent" location="bottom">
              Switch to {{ theme.global.current.value.dark ? 'Light' : 'Dark' }} Mode
            </v-tooltip>
          </v-btn>
        </div>
      </div>
    </v-container>

    <!-- Main Workspace Section -->
    <v-container max-width="840" class="pa-0 flex-grow-1 d-flex flex-column justify-center">
      <!-- 1. Logged In User Profile Header Card -->
      <div class="mb-6">
        <div class="text-caption text-uppercase text-medium-emphasis font-weight-bold tracking-wider mb-2 px-1">
          Authenticated Identity
        </div>
        <UserProfileCard :show-actions="true" elevation="2" />
      </div>

      <!-- 2. Need Invitation Banner Card -->
      <NeedInvitationCard />
    </v-container>

    <!-- Footer Support Area -->
    <v-container max-width="840" class="pa-0 mt-8 text-center">
      <div class="text-caption text-medium-emphasis">
        Email Ingestion Gateway &bull; Multi-Tenant Inbound Microservices Suite
      </div>
    </v-container>
  </div>
</template>

<route lang="yaml">
meta:
  requiresAuth: false
  layout: NoLayout
</route>

<script setup>
import { useTheme } from 'vuetify';
import UserProfileCard from '@/components/auth/UserProfileCard.vue';
import NeedInvitationCard from '@/components/auth/NeedInvitationCard.vue';
import AppName from '@/components/AppName.vue';

definePage({
  path: '/need-invitation',
  name: 'need-invitation',
});

const theme = useTheme();

function toggleTheme() {
  theme.global.name.value = theme.global.current.value.dark ? 'light' : 'dark';
}
</script>

<style scoped>
.min-vh-100 {
  min-height: 100vh;
}
.gap-3 {
  gap: 12px;
}
.gap-2 {
  gap: 8px;
}
.line-height-tight {
  line-height: 1.2;
}
.tracking-wider {
  letter-spacing: 0.08em;
}
</style>
