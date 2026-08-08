<template>
  <v-card rounded="lg" variant="elevated" class="pa-4">
    <div class="d-flex align-center justify-space-between mb-4">
      <div>
        <div class="text-h6 font-weight-bold">Recent Ingestion Stream</div>
        <div class="text-caption text-medium-emphasis">
          Real-time inbound SMTP email logs
        </div>
      </div>
      <v-btn
        color="primary"
        variant="text"
        size="small"
        to="/emails"
        prepend-icon="mdi-format-list-bulleted"
      >
        View All Logs
      </v-btn>
    </div>

    <v-data-table
      :headers="headers"
      :items="recentEmails"
      :loading="emailStore.loading"
      density="compact"
      class="bg-transparent"
      hide-default-footer
    >
      <template #item.received_at="{ item }">
        <span class="text-caption font-mono">
          {{ formatDate(item.receivedAt) }}
        </span>
      </template>

      <template #item.to_address="{ item }">
        <span class="text-body-2 text-truncate max-w-200">
          {{ item.localPart }}
        </span>
      </template>

      <template #item.from_address="{ item }">
        <span class="text-body-2 text-truncate max-w-200">
          {{ item.fromAddress }}
        </span>
      </template>

      <template #item.subject="{ item }">
        <span class="text-body-2 text-truncate font-weight-medium">
          {{ item.subject || "(No Subject)" }}
        </span>
      </template>

      <template #item.actions="{ item }">
        <v-btn
          icon
          density="compact"
          variant="text"
          color="primary"
          :to="`/emails/${item.id}`"
        >
          <v-icon icon="mdi-eye-outline" />
          <v-tooltip activator="parent" location="top">Inspect Email</v-tooltip>
        </v-btn>
      </template>

      <template #no-data>
        <div class="text-center py-6 text-medium-emphasis">
          <v-icon icon="mdi-email-open-outline" size="40" class="mb-2" />
          <div>No ingested emails found for this tenant.</div>
        </div>
      </template>
    </v-data-table>
  </v-card>
</template>

<script setup>
import { computed } from "vue";
import { useEmailStore } from "@/stores/emails";

const emailStore = useEmailStore();

const headers = [
  { title: "Received At", key: "received_at", sortable: false },
  { title: "Recipient (To)", key: "to_address", sortable: false },
  { title: "From", key: "from_address", sortable: false },
  { title: "Subject", key: "subject", sortable: false },
  { title: "Actions", key: "actions", align: "end", sortable: false },
];

const recentEmails = computed(() => {
  const list = emailStore.emails || [];
  return list.slice(0, 5);
});

function formatDate(val) {
  if (!val) return "—";
  try {
    return new Date(val).toLocaleString();
  } catch {
    return String(val);
  }
}
</script>

<style scoped>
.font-mono {
  font-family: "Roboto Mono", monospace;
}
.max-w-200 {
  max-width: 200px;
  display: inline-block;
}
</style>
