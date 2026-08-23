<template>
  <div>
    <!-- Header Bar -->
    <v-card rounded="lg" variant="elevated" class="pa-4 mb-4">
      <div class="d-flex align-center flex-wrap gap-3">
        <v-btn icon variant="tonal" color="primary" size="small" to="/emails">
          <v-icon icon="mdi-arrow-left" />
          <v-tooltip activator="parent" location="top">Back to Logs</v-tooltip>
        </v-btn>

        <div class="flex-grow-1 min-w-0">
          <div class="text-h6 font-weight-bold text-truncate">
            {{ emailDetail?.subject || "Email Details" }}
          </div>
          <div class="d-flex align-center gap-2 mt-1">
            <v-chip
              size="x-small"
              color="primary"
              variant="outlined"
              class="font-mono"
            >
              {{ emailDetail?.id }}
            </v-chip>
            <span class="text-caption text-medium-emphasis">
              Received
              {{
                formatDate(emailDetail?.receivedAt)
              }}
            </span>
          </div>
        </div>
      </div>
    </v-card>

    <!-- Loading State -->
    <v-card v-if="loading" rounded="lg" class="pa-8 text-center">
      <v-progress-circular indeterminate color="primary" class="mb-4" />
      <div class="text-body-1">Loading email detail & metadata...</div>
    </v-card>

    <!-- Split Layout -->
    <v-row v-else-if="emailDetail">
      <!-- Left Column: Metadata Card (4 cols) -->
      <v-col cols="12" md="4">
        <v-card rounded="lg" variant="elevated" class="pa-4">
          <div class="text-h6 font-weight-bold mb-3">Email Metadata</div>

          <v-list density="compact" class="pa-0 bg-transparent">
            <v-list-item class="px-0">
              <v-list-item-title class="text-caption text-medium-emphasis"
                >Sender (From)</v-list-item-title
              >
              <div class="text-body-2 font-weight-medium text-break">
                {{ emailDetail.fromAddress || "—" }}
              </div>
            </v-list-item>

            <v-divider class="my-2" />

            <v-list-item class="px-0">
              <v-list-item-title class="text-caption text-medium-emphasis"
                >Recipient (To)</v-list-item-title
              >
              <div class="text-body-2 font-weight-medium text-break">
                {{ emailDetail.localPart + '@' + ingestDomain }}
              </div>
            </v-list-item>

            <v-divider class="my-2" />

            <v-list-item class="px-0">
              <v-list-item-title class="text-caption text-medium-emphasis"
                >Reference Token</v-list-item-title
              >
              <v-chip
                size="small"
                color="info"
                variant="tonal"
                class="font-mono mt-1"
              >
                {{ emailDetail.referenceToken || "N/A" }}
              </v-chip>
            </v-list-item>

            <v-divider class="my-2" />

            <v-list-item class="px-0">
              <v-list-item-title class="text-caption text-medium-emphasis mb-1">
                S3 Storage Key Prefix
              </v-list-item-title>
              <div
                class="text-caption font-mono bg-surface-variant pa-2 rounded text-break"
              >
                {{ s3KeyPrefix }}
              </div>
            </v-list-item>

            <v-divider class="my-2" />

            <v-list-item class="px-0">
              <v-list-item-title class="text-caption text-medium-emphasis"
                >Ingestion UUID</v-list-item-title
              >
              <div
                class="text-caption font-mono text-medium-emphasis text-break"
              >
                {{ emailDetail?.id }}
              </div>
            </v-list-item>
          </v-list>
        </v-card>
      </v-col>

      <!-- Right Column: Content & Attachments Card (8 cols) -->
      <v-col cols="12" md="8">
        <v-card rounded="lg" variant="elevated" class="pa-4 mb-4">
          <v-tabs v-model="activeTab" color="primary" class="mb-4">
            <v-tab value="html" prepend-icon="mdi-code-tags">HTML Body</v-tab>
            <v-tab value="text" prepend-icon="mdi-text-short">Text Body</v-tab>
            <v-tab value="json" prepend-icon="mdi-email-text-outline"
              >Email Headers</v-tab
            >
          </v-tabs>

          <v-window v-model="activeTab">
            <v-window-item value="html">
              <div
                v-if="emailDetail.html"
                class="iframe-wrapper border rounded"
              >
                <iframe
                  :srcdoc="emailDetail.html"
                  sandbox="allow-same-origin"
                  class="w-100 border-0"
                  style="height: 380px"
                />
              </div>
              <div v-else class="text-center py-8 text-medium-emphasis">
                <v-icon icon="mdi-code-tags-check" size="40" class="mb-2" />
                <div>No HTML body content present for this MIME message.</div>
              </div>
            </v-window-item>

            <v-window-item value="text">
              <CodePreview
                :code="emailDetail.text || '(No plain text body)'"
                language="text"
                title="Plain Text Content"
                max-height="380px"
              />
            </v-window-item>

            <v-window-item value="json">
              <EmailHeadersList
                :headers="emailDetail?.headers"
                max-height="380px"
              />
            </v-window-item>
          </v-window>
        </v-card>

        <!-- Attachments Section -->
        <v-card rounded="lg" variant="elevated" class="pa-4">
          <div class="d-flex align-center justify-space-between mb-3">
            <div class="text-h6 font-weight-bold">
              Attachments ({{ attachments.length }})
            </div>
            <v-chip size="small" color="primary" variant="tonal">
              S3 Presigned Direct Download
            </v-chip>
          </div>

          <v-divider class="mb-4" />

          <div v-if="attachments.length > 0" class="d-flex flex-column gap-3">
            <v-card
              v-for="(att, idx) in attachments"
              :key="att.id || att.attachment_id || idx"
              variant="outlined"
              rounded="lg"
              class="pa-3"
            >
              <div
                class="d-flex align-center justify-space-between flex-wrap gap-2"
              >
                <div class="d-flex align-center gap-3">
                  <v-avatar
                    color="primary"
                    variant="tonal"
                    size="40"
                    rounded="md"
                  >
                    <v-icon icon="mdi-file-document-outline" size="22" />
                  </v-avatar>
                  <div>
                    <div class="text-subtitle-2 font-weight-bold">
                      {{
                        att.filename || att.fileName || `attachment_${idx + 1}`
                      }}
                    </div>
                    <div class="text-caption text-medium-emphasis font-mono">
                      {{
                        att.content_type ||
                        att.contentType ||
                        "application/octet-stream"
                      }}
                      • {{ formatBytes(att.size || att.byte_size || 0) }}
                    </div>
                  </div>
                </div>

                <v-btn
                  color="primary"
                  variant="flat"
                  size="small"
                  prepend-icon="mdi-download"
                  :loading="
                    downloadingId ===
                    (att.id || att.attachment_id || String(idx))
                  "
                  @click="
                    downloadAttachment(
                      att.id || att.attachment_id || String(idx),
                    )
                  "
                >
                  Download Attachment
                </v-btn>
              </div>
            </v-card>
          </div>

          <div v-else class="text-center py-6 text-medium-emphasis">
            <v-icon icon="mdi-paperclip-off" size="36" class="mb-2" />
            <div>No attachments associated with this email.</div>
          </div>
        </v-card>
      </v-col>
    </v-row>

    <!-- Not Found State -->
    <v-card v-else rounded="lg" class="pa-8 text-center">
      <v-icon
        icon="mdi-email-search-outline"
        size="48"
        color="warning"
        class="mb-2"
      />
      <div class="text-h6">Email Record Not Found</div>
      <div class="text-caption mb-4">
        Could not find email ID {{ route.params.id }} for the current tenant
        scope.
      </div>
      <v-btn color="primary" to="/emails">Return to Ingestion Logs</v-btn>
    </v-card>
  </div>
</template>

<route lang="yaml">
meta:
  requiresAuth: true
  title: Email Details
</route>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import CodePreview from "@/components/shared/CodePreview.vue";
import EmailHeadersList from "@/components/emails/EmailHeadersList.vue";
import { useAppStore } from "@/stores/application";
import { useEmailStore } from "@/stores/emails";
import { useNotificationStore } from "@/stores/notification";

const route = useRoute();
const appStore = useAppStore();
const emailStore = useEmailStore();
const notificationStore = useNotificationStore();

const activeTab = ref("html");
const loading = ref(true);
const downloadingId = ref("");

const ingestDomain = window.APP_CONFIG.INGEST_DOMAIN;
const emailDetail = ref({});

const s3KeyPrefix = computed(() => {
  const appId = appStore.activeAppId || "app-uuid";
  const emailId = route.params.id;
  return `s3://email-ingestion-spool/apps/${appId}/emails/${emailId}/`;
});

const attachments = computed(() => {
  if (!emailDetail.value) return [];
  if (Array.isArray(emailDetail.value.attachments)) {
    return emailDetail.value.attachments;
  }
  return [];
});

const rawJson = computed(() => {
  if (!emailDetail.value) return {};
  return emailDetail.value.headers;
});

function formatDate(val) {
  if (!val) return "—";
  try {
    return new Date(val).toLocaleString();
  } catch {
    return String(val);
  }
}

function formatBytes(bytes) {
  if (!bytes || bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${Number.parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
}

async function downloadAttachment(attachmentId) {
  const appId = appStore.activeAppId;
  const emailId = route.params.id;

  if (!appId || !emailId) {
    notificationStore.error("Missing application ID or email ID scope.");
    return;
  }

  downloadingId.value = attachmentId;
  try {
    const res = await emailStore.fetchAttachmentUrl(
      appId,
      emailId,
      attachmentId,
    );
    const downloadUrl = res.download_url || res.downloadUrl || res.DownloadURL;

    if (downloadUrl) {
      notificationStore.success(
        "Generated S3 presigned URL. Opening download link...",
      );
      window.open(downloadUrl, "_blank");
    } else {
      notificationStore.error(
        "No download URL returned by storage service.",
      );
    }
  } catch (err) {
    notificationStore.error(
      err.response?.data?.message || "Failed to acquire presigned download URL",
    );
  } finally {
    downloadingId.value = "";
  }
}

onMounted(async () => {
  loading.value = true;
  try {
    if (appStore.activeAppId) {
      emailDetail.value = await emailStore.fetchEmailById(
        appStore.activeAppId,
        route.params.id,
      );
    }
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
.gap-3 {
  gap: 12px;
}
.font-mono {
  font-family: "Roboto Mono", monospace;
}
.iframe-wrapper {
  background-color: #ffffff;
  overflow: hidden;
}
</style>
