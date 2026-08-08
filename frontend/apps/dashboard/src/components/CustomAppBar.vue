<template>
  <!-- Header Bar with Tenant App Selector -->
  <div class="d-flex align-center justify-space-between mb-6 pb-4 border-b">
    <div class="d-flex align-center gap-3">
      <AppSelector />
    </div>
    <div class="d-flex align-center gap-2">
      <v-chip color="info" variant="outlined" size="small">
        <v-icon start icon="mdi-shield-check" />
        {{ idp }} Authenticated
      </v-chip>
    </div>
  </div>
</template>
<script setup>
import { ref } from "vue";
import AppSelector from "@/components/AppSelector.vue";
const { AUTH_PROVIDER } = window.APP_CONFIG;
const idp = ref(AUTH_PROVIDER === "oidc" ? "OIDC" : "Admin Password");
</script>
