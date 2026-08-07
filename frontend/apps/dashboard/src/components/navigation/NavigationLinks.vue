<template>

  <v-divider></v-divider>

  <v-list density="compact" nav class="me-auto">
    <v-list-item
      prepend-icon="mdi-view-dashboard-outline"
      title="Dashboard"
      to="/"
      exact-path
    ></v-list-item>

    <v-list-item
      prepend-icon="mdi-email-multiple-outline"
      title="Routing Addresses"
      to="/addresses"
      exact-path
    ></v-list-item>

    <v-list-item
      prepend-icon="mdi-table-clock"
      title="Ingestion Logs"
      to="/emails"
    ></v-list-item>

    <v-list-item
      prepend-icon="mdi-webhook"
      title="Webhook Console"
      to="/webhooks"
    ></v-list-item>

    <v-list-item
      prepend-icon="mdi-key-chain"
      title="API Keys & Security"
      to="/settings"
    ></v-list-item>
  </v-list>
</template>
