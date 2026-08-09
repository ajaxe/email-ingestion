<template>
  <div>
    <WebhookConfigHandshake />
    <WebhookJobTable />
  </div>
</template>

<route lang="yaml">
meta:
  requiresAuth: true
</route>

<script setup>
import WebhookConfigHandshake from "@/components/webhooks/WebhookConfigHandshake.vue";
import WebhookJobTable from "@/components/webhooks/WebhookJobTable.vue";
</script>
