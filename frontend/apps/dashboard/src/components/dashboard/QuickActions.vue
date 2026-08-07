<template>
  <v-card rounded="lg" variant="elevated" class="pa-4 mb-4">
    <div class="text-h6 font-weight-bold mb-1">Quick Actions</div>
    <div class="text-caption text-medium-emphasis mb-4">
      Common tenant management shortcuts
    </div>

    <div class="d-flex flex-column gap-2">
      <v-btn
        color="primary"
        variant="tonal"
        block
        prepend-icon="mdi-plus"
        to="/addresses"
        class="justify-start text-none"
      >
        Provision Address
      </v-btn>

      <v-btn
        color="secondary"
        variant="tonal"
        block
        prepend-icon="mdi-lightning-bolt-outline"
        to="/webhooks"
        class="justify-start text-none"
      >
        Test Webhook
      </v-btn>
    </div>
  </v-card>
</template>

<script setup>
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
</style>
