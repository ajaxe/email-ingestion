<template>
  <v-divider />

  <v-list density="compact" nav class="py-2">
    <v-list-item
      prepend-icon="mdi-view-dashboard-outline"
      title="Dashboard Overview"
      to="/"
      exact
      color="primary"
    />

    <v-list-item
      prepend-icon="mdi-email-multiple-outline"
      title="Routing Addresses"
      to="/addresses"
      color="primary"
    />

    <v-list-item
      prepend-icon="mdi-table-clock"
      title="Ingestion Logs"
      to="/emails"
      color="primary"
    />

    <v-list-item
      prepend-icon="mdi-webhook"
      title="Webhook Console"
      to="/webhooks"
      color="primary"
    />

    <v-list-item
      prepend-icon="mdi-key-chain"
      title="API Keys & Security"
      to="/settings"
      color="primary"
    />
  </v-list>
</template>
