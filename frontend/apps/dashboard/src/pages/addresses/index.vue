<template>
  <div>
    <!-- Header Toolbar -->
    <v-card rounded="lg" variant="elevated" class="pa-4 mb-4">
      <v-row align="center" justify="space-between" class="ma-0">
        <v-col cols="12" md="4" class="pa-0">
          <div class="text-h6 font-weight-bold">Assigned Email Addresses</div>
          <div class="text-caption text-medium-emphasis">
            Manage 10-character email routing addresses
          </div>
        </v-col>

        <v-col
          cols="12"
          md="8"
          class="pa-0 mt-3 mt-md-0 d-flex flex-wrap align-center justify-md-end gap-3"
        >
          <v-text-field
            v-model="searchQuery"
            prepend-inner-icon="mdi-magnify"
            placeholder="Search addresses..."
            variant="outlined"
            density="compact"
            hide-details
            style="max-width: 240px"
          />

          <v-btn-toggle
            v-model="statusFilter"
            mandatory
            variant="outlined"
            density="compact"
            color="primary"
          >
            <v-btn value="ALL">All</v-btn>
            <v-btn value="ACTIVE">Active</v-btn>
            <v-btn value="INACTIVE">Inactive</v-btn>
          </v-btn-toggle>

          <v-btn
            color="primary"
            prepend-icon="mdi-plus"
            @click="showProvisionModal = true"
          >
            Provision New Address
          </v-btn>
        </v-col>
      </v-row>
    </v-card>

    <!-- Data Table Card -->
    <v-card rounded="lg" variant="elevated">
      <v-data-table
        :headers="headers"
        :items="filteredAddresses"
        :loading="addressStore.loading"
        density="comfortable"
        hover
      >
        <template #item.local_part="{ item }">
          <span class="font-mono font-weight-bold primary--text">
            {{ item.local_part || item.localPart }}
          </span>
        </template>

        <template #item.email_address="{ item }">
          <div class="d-flex align-center gap-1">
            <span class="font-mono text-body-2">
              {{ getFullAddress(item) }}
            </span>
            <v-btn
              icon="mdi-content-copy"
              size="x-small"
              variant="text"
              color="grey"
              @click="copyToClipboard(getFullAddress(item))"
            >
              <v-tooltip activator="parent" location="top"
                >Copy Email</v-tooltip
              >
            </v-btn>
          </div>
        </template>

        <template #item.description="{ item }">
          <span class="text-body-2">
            {{ item.description || "No description provided" }}
          </span>
        </template>

        <template #item.status="{ item }">
          <StatusChip
            :status="item.isActive ? 'ACTIVE' : 'INACTIVE'"
          />
        </template>

        <template #item.created_at="{ item }">
          <span class="text-caption font-mono">
            {{ formatDate(item.createdAt) }}
          </span>
        </template>

        <template #item.actions="{ item }">
          <div class="d-flex justify-end">
            <v-switch
              :model-value="isAddressActive(item)"
              color="success"
              hide-details
              density="compact"
              @update:model-value="handleToggleStatus(item)"
            >
              <v-tooltip activator="parent" location="top">
                {{ isAddressActive(item) ? "Deactivate" : "Activate" }} Address
              </v-tooltip>
            </v-switch>
          </div>
        </template>

        <template #no-data>
          <div class="text-center py-8 text-medium-emphasis">
            <v-icon icon="mdi-email-outline" size="48" class="mb-2" />
            <div class="text-h6">No addresses found</div>
            <div class="text-caption mb-4">
              Provision a 10-character address to start receiving inbound
              emails.
            </div>
            <v-btn
              color="primary"
              size="small"
              prepend-icon="mdi-plus"
              @click="showProvisionModal = true"
            >
              Provision Address Now
            </v-btn>
          </div>
        </template>
      </v-data-table>
    </v-card>

    <!-- Provision Modal Dialog -->
    <v-dialog v-model="showProvisionModal" max-width="520px">
      <v-card rounded="lg">
        <v-card-title
          class="d-flex align-center justify-space-between pa-4 bg-surface-variant"
        >
          <span class="text-h6 font-weight-bold">Provision New Address</span>
          <v-btn
            icon="mdi-close"
            variant="text"
            size="small"
            @click="showProvisionModal = false"
          />
        </v-card-title>

        <v-card-text class="pa-4">
          <v-alert
            type="info"
            variant="tonal"
            icon="mdi-information-outline"
            class="mb-4"
            density="compact"
          >
            Addresses use a system-generated 10-character random routing path
            (e.g. <code>a1b2c3d4e5@domain.com</code>) for maximum security and
            uniqueness.
          </v-alert>

          <v-form ref="formRef" @submit.prevent="handleProvision">
            <v-text-field
              v-model="newDescription"
              label="Address Description"
              placeholder="e.g. Customer Support Webhook Ingestion"
              variant="outlined"
              density="comfortable"
              hint="Optional purpose description for internal reference"
              persistent-hint="false"
              class="mb-2"
            />
          </v-form>
        </v-card-text>

        <v-divider />

        <v-card-actions class="pa-4">
          <v-spacer />
          <v-btn
            variant="outlined"
            color="grey"
            @click="showProvisionModal = false"
          >
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            variant="flat"
            :loading="provisioning"
            @click="handleProvision"
          >
            Generate Address
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<route lang="yaml">
meta:
  requiresAuth: true
  title: Addresses
</route>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import StatusChip from "@/components/shared/StatusChip.vue";
import { useAppStore } from "@/stores/application";
import { useAddressStore } from "@/stores/addresses";
import { useNotificationStore } from "@/stores/notification";

const appStore = useAppStore();
const addressStore = useAddressStore();
const notificationStore = useNotificationStore();

const searchQuery = ref("");
const statusFilter = ref("ALL");
const showProvisionModal = ref(false);
const newDescription = ref("");
const provisioning = ref(false);

const headers = [
  { title: "Local Part", key: "local_part", sortable: true },
  { title: "Full Address", key: "email_address", sortable: false },
  { title: "Description", key: "description", sortable: true },
  { title: "Status", key: "status", sortable: true },
  { title: "Created At", key: "created_at", sortable: true },
  { title: "Toggle Active", key: "actions", align: "end", sortable: false },
];

const filteredAddresses = computed(() => {
  let list = addressStore.addresses || [];

  if (statusFilter.value !== "ALL") {
    list = list.filter((item) => {
      const isAct = isAddressActive(item);
      return statusFilter.value === "ACTIVE" ? isAct : !isAct;
    });
  }

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase().trim();
    list = list.filter((item) => {
      const local = (item.local_part || item.localPart || "").toLowerCase();
      const desc = (item.description || "").toLowerCase();
      return local.includes(q) || desc.includes(q);
    });
  }

  return list;
});

function isAddressActive(item) {
  if (item.status) {
    return item.status.toUpperCase() === "ACTIVE";
  }
  return Boolean(item.is_active || item.isActive);
}

function getFullAddress(item) {
  const local = item.localPart || "";
  const domain = window.APP_CONFIG.INGEST_DOMAIN;
  return `${local}@${domain}`;
}

function formatDate(val) {
  if (!val) return "—";
  try {
    return new Date(val).toLocaleString();
  } catch {
    return String(val);
  }
}

async function copyToClipboard(text) {
  try {
    await navigator.clipboard.writeText(text);
    notificationStore.success(`Copied ${text} to clipboard!`);
  } catch {
    notificationStore.error("Failed to copy text");
  }
}

async function handleProvision() {
  if (!appStore.activeAppId) {
    notificationStore.error("Please select an active application scope first.");
    return;
  }
  provisioning.value = true;
  try {
    await addressStore.provisionAddress(
      appStore.activeAppId,
      newDescription.value,
    );
    notificationStore.success(
      "Successfully provisioned new 10-character email address!",
    );
    showProvisionModal.value = false;
    newDescription.value = "";
    await addressStore.fetchAddresses(appStore.activeAppId);
  } catch (err) {
    notificationStore.error(
      err.response?.data?.message || "Failed to provision address",
    );
  } finally {
    provisioning.value = false;
  }
}

async function handleToggleStatus(item) {
  const currentStatus = isAddressActive(item);
  const nextStatus = currentStatus ? "INACTIVE" : "ACTIVE";
  const addressId = item.id || item.address_id;
  try {
    await addressStore.toggleAddress(
      appStore.activeAppId,
      addressId,
      nextStatus,
    );
    notificationStore.success(`Address updated to ${nextStatus}`);
    await addressStore.fetchAddresses(appStore.activeAppId);
  } catch (err) {
    notificationStore.error(
      err.response?.data?.message || "Failed to toggle address status",
    );
  }
}

function loadAddresses() {
  if (appStore.activeAppId) {
    addressStore.fetchAddresses(appStore.activeAppId);
  }
}

onMounted(() => {
  loadAddresses();
});

watch(
  () => appStore.activeAppId,
  () => {
    loadAddresses();
  },
);
</script>

<style scoped>
.gap-1 {
  gap: 4px;
}
.gap-3 {
  gap: 12px;
}
.font-mono {
  font-family: "Roboto Mono", monospace;
}
</style>
