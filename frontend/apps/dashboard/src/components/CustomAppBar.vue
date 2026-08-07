<template>
  <!-- Header Bar with Tenant App Selector -->
  <div class="d-flex align-center justify-space-between mb-6 pb-4 border-b">
    <div class="d-flex align-center gap-3">
      <AppSelector />
    </div>
    <div class="d-flex align-center gap-2">
      <v-chip color="info" variant="outlined" size="small">
        <v-icon start icon="mdi-shield-check" />
        OIDC Authenticated
      </v-chip>
    </div>
  </div>
</template>
<script setup>
import AppSelector from '@/components/AppSelector.vue'

</script>