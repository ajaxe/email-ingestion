<template>
  <div>
    <!-- Table Controls Toolbar -->
    <div class="d-flex align-center justify-space-between flex-wrap gap-3 mb-3 px-2">
      <div class="d-flex align-center gap-3">
        <v-checkbox
          v-model="hideDeleted"
          label="Hide Deleted Emails"
          density="compact"
          hide-details
          color="primary"
        />
      </div>

      <div class="d-flex align-center gap-2">
        <v-btn
          v-if="selectedActiveIds.length > 0"
          color="error"
          variant="flat"
          size="small"
          prepend-icon="mdi-delete-sweep"
          @click="showBulkDeleteModal = true"
        >
          Delete Selected ({{ selectedActiveIds.length }})
        </v-btn>
      </div>
    </div>

    <!-- Data Table -->
    <v-data-table-server
      v-model="selectedItems"
      show-select
      item-value="id"
      :headers="headers"
      :items="emails"
      :items-length="totalCount"
      :loading="emailStore.loading"
      v-model:items-per-page="itemsPerPage"
      @update:options="loadItems"
      density="comfortable"
      hover
    >
      <template #item.created_at="{ item }">
        <span class="font-mono text-caption">
          {{ formatDate(item.receivedAt || item.received_at) }}
        </span>
      </template>

      <template #item.local_part="{ item }">
        <v-chip
          size="small"
          color="primary"
          variant="tonal"
          class="font-mono font-weight-medium"
        >
          {{ item.localPart || item.local_part }}
        </v-chip>
      </template>

      <template #item.from_address="{ item }">
        <span class="text-body-2 font-weight-medium text-truncate max-w-180">
          {{ item.fromAddress || item.from_address }}
        </span>
      </template>

      <template #item.subject="{ item }">
        <div class="d-flex align-center gap-2 max-w-250">
          <span class="text-body-2 text-truncate font-weight-medium">
            {{ item.subject || "(No Subject)" }}
          </span>
          <v-chip
            v-if="item.deletedAt || item.deleted_at"
            size="x-small"
            color="error"
            variant="flat"
            class="font-weight-bold"
          >
            Deleted
          </v-chip>
        </div>
      </template>

      <template #item.reference_token="{ item }">
        <v-chip size="x-small" color="info" variant="outlined" class="font-mono">
          {{ item.referenceToken || item.reference_token || "N/A" }}
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
        <div class="d-flex align-center justify-end gap-1">
          <v-btn icon variant="text" color="primary" size="small" :to="`/emails/${item.id}`">
            <v-icon icon="mdi-eye" size="16" />
            <v-tooltip activator="parent" location="top">Inspect Details</v-tooltip>
          </v-btn>

          <v-btn
            icon
            variant="text"
            color="error"
            size="small"
            :disabled="Boolean(item.deletedAt || item.deleted_at)"
            @click="openSingleDeleteModal(item)"
          >
            <v-icon icon="mdi-delete-outline" size="16" />
            <v-tooltip activator="parent" location="top">
              {{ (item.deletedAt || item.deleted_at) ? 'Email already deleted' : 'Soft Delete Email' }}
            </v-tooltip>
          </v-btn>
        </div>
      </template>

      <template #no-data>
        <div class="text-center py-8 text-medium-emphasis">
          <v-icon icon="mdi-email-search-outline" size="48" class="mb-2" />
          <div class="text-h6">No email records found</div>
          <div class="text-caption">
            Inbound emails sent to provisioned addresses will appear here in real-time.
          </div>
        </div>
      </template>
    </v-data-table-server>

    <!-- Modals -->
    <DeleteEmailModal
      v-model="showSingleDeleteModal"
      :loading="deleting"
      @confirm="confirmSingleDelete"
    />

    <BulkDeleteEmailsModal
      v-model="showBulkDeleteModal"
      :count="selectedActiveIds.length"
      :loading="deleting"
      @confirm="confirmBulkDelete"
    />
  </div>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { useEmailStore } from "@/stores/emails";
import { useAppStore } from "@/stores/application";
import { useNotificationStore } from "@/stores/notification";
import { storeToRefs } from "pinia";
import DeleteEmailModal from "./DeleteEmailModal.vue";
import BulkDeleteEmailsModal from "./BulkDeleteEmailsModal.vue";

const emailStore = useEmailStore();
const appStore = useAppStore();
const notificationStore = useNotificationStore();

const { selectedLocalPart, searchQuery, hideDeleted, totalCount, emails } = storeToRefs(emailStore);
const itemsPerPage = ref(10);
const selectedItems = ref([]);
let searchDebounceTimer = null;

const showSingleDeleteModal = ref(false);
const showBulkDeleteModal = ref(false);
const targetEmail = ref(null);
const deleting = ref(false);

const headers = [
  { title: "Received At", key: "created_at", sortable: false },
  { title: "Local Part", key: "local_part", sortable: false },
  { title: "From Address", key: "from_address", sortable: false },
  { title: "Subject", key: "subject", sortable: false },
  { title: "Ref Token", key: "reference_token", sortable: false },
  {
    title: "Attachments",
    key: "attachments_count",
    align: "center",
    sortable: false,
  },
  { title: "Actions", key: "actions", align: "end", sortable: false },
];

const selectedActiveIds = computed(() => {
  if (!selectedItems.value || selectedItems.value.length === 0) return [];
  const activeIds = [];
  for (const id of selectedItems.value) {
    const item = emails.value.find((e) => e.id === id || e.id === String(id));
    if (item && !item.deletedAt && !item.deleted_at) {
      activeIds.push(item.id);
    }
  }
  return activeIds;
});

function loadItems({ page, itemsPerPage }) {
  if (!appStore.activeAppId) return;
  emailStore.fetchEmails(appStore.activeAppId, {
    limit: itemsPerPage,
    page,
  });
}

function formatDate(val) {
  if (!val) return "—";
  try {
    return new Date(val).toLocaleString();
  } catch {
    return String(val);
  }
}

function getAttachmentCount(item) {
  if (typeof item.attachments_count === "number") return item.attachments_count;
  if (Array.isArray(item.attachments)) return item.attachments.length;
  return 0;
}

function openSingleDeleteModal(item) {
  targetEmail.value = item;
  showSingleDeleteModal.value = true;
}

async function confirmSingleDelete() {
  if (!appStore.activeAppId || !targetEmail.value) return;
  deleting.value = true;
  try {
    await emailStore.deleteEmail(appStore.activeAppId, targetEmail.value.id);
    notificationStore.success("Email soft deleted successfully");
    showSingleDeleteModal.value = false;
    targetEmail.value = null;
    await emailStore.fetchEmails(appStore.activeAppId, { limit: itemsPerPage.value, page: 1 });
  } catch (err) {
    notificationStore.error(err.response?.data?.message || "Failed to soft delete email");
  } finally {
    deleting.value = false;
  }
}

async function confirmBulkDelete() {
  if (!appStore.activeAppId || selectedActiveIds.value.length === 0) return;
  deleting.value = true;
  try {
    const res = await emailStore.bulkDeleteEmails(appStore.activeAppId, selectedActiveIds.value);
    notificationStore.success(`Successfully soft deleted ${res.deletedCount || selectedActiveIds.value.length} emails.`);
    showBulkDeleteModal.value = false;
    selectedItems.value = [];
    await emailStore.fetchEmails(appStore.activeAppId, { limit: itemsPerPage.value, page: 1 });
  } catch (err) {
    notificationStore.error(err.response?.data?.message || "Failed to bulk delete emails");
  } finally {
    deleting.value = false;
  }
}

watch(hideDeleted, () => {
  if (!appStore.activeAppId) return;
  emailStore.fetchEmails(appStore.activeAppId, {
    page: 1,
    limit: itemsPerPage.value,
  });
});

watch(searchQuery, () => {
  if (searchDebounceTimer) clearTimeout(searchDebounceTimer);
  searchDebounceTimer = setTimeout(() => {
    if (!appStore.activeAppId) return;
    emailStore.fetchEmails(appStore.activeAppId, {
      page: 1,
      limit: itemsPerPage.value,
    });
  }, 300);
});

watch(selectedLocalPart, () => {
  if (!appStore.activeAppId) return;
  emailStore.fetchEmails(appStore.activeAppId, {
    page: 1,
    limit: itemsPerPage.value,
  });
});
</script>

<style scoped>
.gap-1 {
  gap: 4px;
}
.gap-2 {
  gap: 8px;
}
.gap-3 {
  gap: 12px;
}
.font-mono {
  font-family: "Roboto Mono", monospace;
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
