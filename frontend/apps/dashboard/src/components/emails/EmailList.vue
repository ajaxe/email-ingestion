<template>
  <v-data-table
    :headers="headers"
    :items="filteredEmails"
    :loading="emailStore.loading"
    density="comfortable"
    hover
  >
    <template #item.created_at="{ item }">
      <span class="font-mono text-caption">
        {{ formatDate(item.receivedAt) }}
      </span>
    </template>

    <template #item.local_part="{ item }">
      <v-chip
        size="small"
        color="primary"
        variant="tonal"
        class="font-mono font-weight-medium"
      >
        {{ item.localPart }}
      </v-chip>
    </template>

    <template #item.from_address="{ item }">
      <span class="text-body-2 font-weight-medium text-truncate max-w-180">
        {{ item.fromAddress }}
      </span>
    </template>

    <template #item.subject="{ item }">
      <span class="text-body-2 text-truncate font-weight-medium max-w-250">
        {{ item.subject || "(No Subject)" }}
        <v-tooltip activator="parent" location="top" v-if="item.subject">
          {{ item.subject }}
        </v-tooltip>
      </span>
    </template>

    <template #item.reference_token="{ item }">
      <v-chip size="x-small" color="info" variant="outlined" class="font-mono">
        {{ item.referenceToken || "N/A" }}
      </v-chip>
    </template>

    <template #item.attachments_count="{ item }">
      <v-chip
        size="x-small"
        :color="getAttachmentCount(item) > 0 ? 'success' : 'grey'"
        variant="tonal"
        class="font-weight-bold"
      >
        <v-icon start icon="mdi-paperclip" size="12" class="me-1" />
        {{ getAttachmentCount(item) }}
      </v-chip>
    </template>

    <template #item.actions="{ item }">
      <v-btn icon variant="text" color="primary" :to="`/emails/${item.id}`">
        <v-icon icon="mdi-eye" size="16" />
        <v-tooltip activator="parent" location="top">Inspect Details</v-tooltip>
      </v-btn>
    </template>

    <template #no-data>
      <div class="text-center py-8 text-medium-emphasis">
        <v-icon icon="mdi-email-search-outline" size="48" class="mb-2" />
        <div class="text-h6">No email records found</div>
        <div class="text-caption">
          Inbound emails sent to provisioned addresses will appear here in
          real-time.
        </div>
      </div>
    </template>
  </v-data-table>
</template>
<script setup>
import { computed } from "vue";
import { useEmailStore } from "@/stores/emails";
import { storeToRefs } from "pinia";

const emailStore = useEmailStore();
const { selectedLocalPart, searchQuery } = storeToRefs(emailStore);

const headers = [
  { title: "Received At", key: "created_at", sortable: true },
  { title: "Local Part", key: "local_part", sortable: true },
  { title: "From Address", key: "from_address", sortable: true },
  { title: "Subject", key: "subject", sortable: true },
  { title: "Ref Token", key: "reference_token", sortable: false },
  {
    title: "Attachments",
    key: "attachments_count",
    align: "center",
    sortable: true,
  },
  { title: "Actions", key: "actions", align: "end", sortable: false },
];

const filteredEmails = computed(() => {
  let list = emailStore.emails || [];

  if (selectedLocalPart.value !== "ALL") {
    list = list.filter((item) => item.localPart === selectedLocalPart.value);
  }

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim();
    list = list.filter((item) => {
      const from = (item.fromAddress || item.sender || "").toLowerCase();
      const subj = (item.subject || "").toLowerCase();
      return from.includes(q) || subj.includes(q);
    });
  }

  return list;
});


function formatDate(val) {
  if (!val) return '—';
  try {
    return new Date(val).toLocaleString();
  } catch {
    return String(val);
  }
}

function getAttachmentCount(item) {
  if (typeof item.attachments_count === 'number') return item.attachments_count;
  if (Array.isArray(item.attachments)) return item.attachments.length;
  return 0;
}
</script>
<style scoped>
.gap-3 {
  gap: 12px;
}
.font-mono {
  font-family: 'Roboto Mono', monospace;
}
.max-w-180 {
  max-width: 180px;
  display: inline-block;
}
.max-w-250 {
  max-width: 250px;
  display: inline-block;
}
</style>