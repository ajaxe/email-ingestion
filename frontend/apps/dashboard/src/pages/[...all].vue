<template>
  <v-container class="fill-height d-flex align-center justify-center py-12">
    <v-card rounded="lg" variant="elevated" max-width="500px" class="pa-8 text-center w-100">
      <v-avatar color="error" variant="tonal" size="72" class="mb-4">
        <v-icon icon="mdi-alert-octagon-outline" size="44" color="error" />
      </v-avatar>

      <div class="text-h3 font-weight-bold mb-2">404</div>
      <div class="text-h6 font-weight-bold mb-2">Page Not Found</div>
      <p class="text-body-2 text-medium-emphasis mb-6">
        The requested page route does not exist or has been moved within the Email Ingestion Gateway Console.
      </p>

      <v-btn
        color="primary"
        variant="flat"
        size="large"
        prepend-icon="mdi-home"
        to="/"
      >
        Back to Dashboard
      </v-btn>
    </v-card>
  </v-container>
</template>

<route lang="yaml">
meta:
  requiresAuth: true
</route>

<script setup>
</script>

<style scoped>
</style>
