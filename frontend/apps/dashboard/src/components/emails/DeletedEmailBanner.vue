<template>
  <v-card rounded="lg" color="amber-lighten-5" class="pa-4 mb-4 border border-warning">
    <div class="d-flex align-center gap-3">
      <v-avatar color="warning" variant="tonal" size="44" rounded="md">
        <v-icon icon="mdi-delete-clock-outline" size="26" />
      </v-avatar>

      <div class="flex-grow-1 min-w-0">
        <div class="d-flex align-center gap-2 flex-wrap">
          <span class="text-subtitle-1 font-weight-bold text-amber-darken-4">
            Email Content Soft-Deleted
          </span>
          <v-chip color="warning" size="x-small" variant="flat" class="font-weight-bold">
            DELETED
          </v-chip>
        </div>

        <div class="text-caption text-medium-emphasis mt-1">
          Soft deleted on <strong>{{ formatDate(deletedAt) }}</strong>. Raw MIME text/HTML bodies and attachments were permanently purged from object storage.
        </div>
      </div>
    </div>
  </v-card>
</template>

<script setup>
defineProps({
  deletedAt: {
    type: [String, Date],
    default: null,
  },
});

function formatDate(val) {
  if (!val) return "N/A";
  try {
    return new Date(val).toLocaleString();
  } catch {
    return String(val);
  }
}
</script>

<style scoped>
.gap-2 {
  gap: 8px;
}
.gap-3 {
  gap: 12px;
}
.border-warning {
  border-color: rgba(255, 160, 0, 0.4) !important;
}
</style>
