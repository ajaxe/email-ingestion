<template>
  <v-card rounded="lg" variant="elevated" class="h-100 pa-4">
    <div class="d-flex align-center justify-space-between mb-2">
      <span class="text-caption text-medium-emphasis text-uppercase font-weight-bold tracking-wider">
        {{ label || title }}
      </span>
      <v-avatar :color="color" variant="tonal" size="40" rounded="md">
        <v-icon :icon="icon" size="22" />
      </v-avatar>
    </div>

    <div class="d-flex align-baseline mt-2">
      <span class="text-h4 font-weight-bold me-2">
        {{ value }}
      </span>
      <!--
      <div v-if="trend !== undefined && trend !== null" class="d-flex align-center">
        <v-chip
          :color="computedTrendColor"
          size="x-small"
          variant="tonal"
          class="font-weight-medium px-1"
        >
          <v-icon
            start
            :icon="computedTrendIcon"
            size="12"
            class="me-1"
            v-if="computedTrendIcon"
          />
          {{ formattedTrend }}
        </v-chip>
      </div> -->
    </div>

    <div v-if="caption || subtitle" class="text-caption text-medium-emphasis mt-2">
      {{ caption || subtitle }}
    </div>
  </v-card>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  title: {
    type: String,
    default: '',
  },
  label: {
    type: String,
    default: '',
  },
  value: {
    type: [String, Number],
    default: '0',
  },
  icon: {
    type: String,
    default: 'mdi-chart-bar',
  },
  color: {
    type: String,
    default: 'primary',
  },
  trend: {
    type: [String, Number],
    default: null,
  },
  trendColor: {
    type: String,
    default: '',
  },
  subtitle: {
    type: String,
    default: '',
  },
  caption: {
    type: String,
    default: '',
  },
});

const formattedTrend = computed(() => {
  if (props.trend === null || props.trend === undefined) return '';
  const str = String(props.trend);
  if (typeof props.trend === 'number' && props.trend > 0) {
    return `+${str}%`;
  }
  return str.endsWith('%') ? str : `${str}%`;
});

const computedTrendColor = computed(() => {
  if (props.trendColor) return props.trendColor;
  const str = String(props.trend || '');
  if (str.startsWith('+') || (typeof props.trend === 'number' && props.trend > 0)) {
    return 'success';
  }
  if (str.startsWith('-') || (typeof props.trend === 'number' && props.trend < 0)) {
    return 'error';
  }
  return 'info';
});

const computedTrendIcon = computed(() => {
  const color = computedTrendColor.value;
  if (color === 'success') return 'mdi-arrow-up-bold';
  if (color === 'error') return 'mdi-arrow-down-bold';
  return '';
});
</script>

<style scoped>
.tracking-wider {
  letter-spacing: 0.05em;
}
</style>
