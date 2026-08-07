<template>
  <v-chip
    :color="statusColor"
    :size="size"
    :variant="variant"
    class="font-weight-medium text-uppercase"
  >
    <v-icon start :icon="statusIcon" v-if="showIcon" />
    {{ formattedStatus }}
  </v-chip>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  status: {
    type: String,
    default: 'INACTIVE',
  },
  size: {
    type: String,
    default: 'small',
  },
  variant: {
    type: String,
    default: 'tonal',
  },
  showIcon: {
    type: Boolean,
    default: true,
  },
});

const formattedStatus = computed(() => {
  return props.status ? String(props.status).toUpperCase() : 'UNKNOWN';
});

const statusColor = computed(() => {
  const val = formattedStatus.value;
  switch (val) {
    case 'ACTIVE':
    case 'SUCCESS':
    case 'DELIVERED':
    case 'OK':
      return 'success';
    case 'DEAD':
    case 'FAILED':
    case 'ERROR':
      return 'error';
    case 'PENDING':
    case 'PROCESSING':
    case 'RETRYING':
    case 'QUEUED':
      return 'warning';
    case 'INACTIVE':
    case 'DISABLED':
    default:
      return 'grey';
  }
});

const statusIcon = computed(() => {
  const val = formattedStatus.value;
  switch (val) {
    case 'ACTIVE':
    case 'SUCCESS':
    case 'DELIVERED':
    case 'OK':
      return 'mdi-check-circle-outline';
    case 'DEAD':
    case 'FAILED':
    case 'ERROR':
      return 'mdi-alert-circle-outline';
    case 'PENDING':
    case 'PROCESSING':
    case 'RETRYING':
    case 'QUEUED':
      return 'mdi-clock-outline';
    case 'INACTIVE':
    case 'DISABLED':
    default:
      return 'mdi-minus-circle-outline';
  }
});
</script>
