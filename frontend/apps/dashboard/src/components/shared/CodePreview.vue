<template>
  <v-card variant="outlined" rounded="lg" class="code-preview-card">
    <v-card-title v-if="title || copyable" class="d-flex align-center py-2 px-4 bg-surface-variant text-caption font-weight-bold">
      <span>{{ title || language.toUpperCase() }}</span>
      <v-spacer />
      <v-btn
        v-if="copyable"
        variant="text"
        size="x-small"
        :icon="copied ? 'mdi-check' : 'mdi-content-copy'"
        :color="copied ? 'success' : undefined"
        @click="copyCode"
      >
        <v-tooltip activator="parent" location="top">
          {{ copied ? 'Copied!' : 'Copy Code' }}
        </v-tooltip>
      </v-btn>
    </v-card-title>
    <v-divider v-if="title || copyable" />
    <v-card-text class="pa-0">
      <pre
        class="code-block pa-4 overflow-x-auto overflow-y-auto"
        :style="{ maxHeight: maxHeightStyle }"
      ><code>{{ formattedCode }}</code></pre>
    </v-card-text>
  </v-card>
</template>

<script setup>
import { computed, ref } from 'vue';

const props = defineProps({
  code: {
    type: [String, Object, Array],
    default: '',
  },
  language: {
    type: String,
    default: 'json',
  },
  title: {
    type: String,
    default: '',
  },
  maxHeight: {
    type: [String, Number],
    default: '400px',
  },
  copyable: {
    type: Boolean,
    default: true,
  },
});

const copied = ref(false);

const formattedCode = computed(() => {
  if (props.code === null || props.code === undefined) {
    return '';
  }
  if (typeof props.code === 'object') {
    try {
      return JSON.stringify(props.code, null, 2);
    } catch {
      return String(props.code);
    }
  }
  if (typeof props.code === 'string' && props.language === 'json') {
    try {
      const parsed = JSON.parse(props.code);
      return JSON.stringify(parsed, null, 2);
    } catch {
      return props.code;
    }
  }
  return String(props.code);
});

const maxHeightStyle = computed(() => {
  if (typeof props.maxHeight === 'number') {
    return `${props.maxHeight}px`;
  }
  return props.maxHeight;
});

async function copyCode() {
  try {
    await navigator.clipboard.writeText(formattedCode.value);
    copied.value = true;
    setTimeout(() => {
      copied.value = false;
    }, 2000);
  } catch (err) {
    console.error('Failed to copy code', err);
  }
}
</script>

<style scoped>
.code-preview-card {
  border-color: rgba(var(--v-border-color), var(--v-border-opacity));
}
.code-block {
  font-family: 'Roboto Mono', monospace, SFMono-Regular, Menlo, Monaco, Consolas;
  font-size: 0.85rem;
  line-height: 1.45;
  background-color: rgba(0, 0, 0, 0.04);
  white-space: pre-wrap;
  word-break: break-word;
}
.v-theme--dark .code-block {
  background-color: rgba(255, 255, 255, 0.04);
}
</style>
