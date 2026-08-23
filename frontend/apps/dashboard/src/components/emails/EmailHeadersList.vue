<template>
  <v-card variant="outlined" rounded="lg" class="email-headers-card">
    <!-- Header Toolbar -->
    <v-card-title class="d-flex align-center flex-wrap gap-2 py-2 px-4 bg-surface-variant">
      <span class="text-caption font-weight-bold text-uppercase">Email Headers</span>
      <v-chip size="x-small" color="primary" variant="tonal" class="ml-1 font-weight-medium">
        {{ filteredHeaders.length }} {{ filteredHeaders.length === 1 ? 'header' : 'headers' }}
      </v-chip>

      <v-spacer />

      <!-- Search Input -->
      <v-text-field
        v-model="searchQuery"
        density="compact"
        variant="solo-filled"
        flat
        hide-details
        placeholder="Filter headers..."
        prepend-inner-icon="mdi-magnify"
        clearable
        class="header-search-field mr-2"
        style="max-width: 220px;"
      />

      <!-- Copy All Button -->
      <v-btn
        variant="text"
        size="small"
        :icon="copiedAll ? 'mdi-check' : 'mdi-content-copy'"
        :color="copiedAll ? 'success' : undefined"
        @click="copyAllHeaders"
      >
        <v-tooltip activator="parent" location="top">
          {{ copiedAll ? 'Copied All!' : 'Copy All Headers' }}
        </v-tooltip>
      </v-btn>

      <!-- Raw JSON / List View Toggle -->
      <v-btn
        variant="text"
        size="small"
        :icon="showRawJson ? 'mdi-format-list-bulleted' : 'mdi-code-json'"
        @click="showRawJson = !showRawJson"
      >
        <v-tooltip activator="parent" location="top">
          {{ showRawJson ? 'Switch to Formatted List' : 'Switch to Raw JSON' }}
        </v-tooltip>
      </v-btn>
    </v-card-title>
    <v-divider />

    <v-card-text class="pa-0">
      <!-- Raw JSON View -->
      <div v-if="showRawJson">
        <pre class="code-block pa-4 overflow-x-auto overflow-y-auto" :style="{ maxHeight: maxHeightStyle }"><code>{{ rawJsonString }}</code></pre>
      </div>

      <!-- Formatted Headers List View -->
      <div v-else class="headers-scroll-container overflow-y-auto" :style="{ maxHeight: maxHeightStyle }">
        <v-table v-if="filteredHeaders.length > 0" density="compact" class="headers-table">
          <thead>
            <tr>
              <th class="text-left font-weight-bold text-caption text-uppercase" style="width: 220px;">Header Name</th>
              <th class="text-left font-weight-bold text-caption text-uppercase">Header Value</th>
              <th style="width: 48px;"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(header, index) in filteredHeaders" :key="index" class="header-row">
              <td class="align-top py-2 font-mono text-caption font-weight-bold">
                <v-chip
                  size="x-small"
                  :color="getHeaderColor(header.key)"
                  variant="tonal"
                  class="font-mono text-caption text-none"
                >
                  {{ header.key }}
                </v-chip>
              </td>
              <td class="align-top py-2 font-mono text-caption text-break header-value-cell">
                {{ header.value }}
              </td>
              <td class="align-top py-1 text-right">
                <v-btn
                  variant="text"
                  size="x-small"
                  :icon="copiedIndex === index ? 'mdi-check' : 'mdi-content-copy'"
                  :color="copiedIndex === index ? 'success' : 'medium-emphasis'"
                  @click="copyHeaderValue(header.value, index)"
                >
                  <v-tooltip activator="parent" location="top">
                    {{ copiedIndex === index ? 'Copied!' : 'Copy Value' }}
                  </v-tooltip>
                </v-btn>
              </td>
            </tr>
          </tbody>
        </v-table>

        <!-- Empty State -->
        <div v-else class="text-center py-8 px-4 text-medium-emphasis">
          <v-icon icon="mdi-text-search" size="36" class="mb-2" />
          <div v-if="searchQuery" class="text-body-2">
            No headers matching "<strong>{{ searchQuery }}</strong>"
          </div>
          <div v-else class="text-body-2">
            No email headers available.
          </div>
        </div>
      </div>
    </v-card-text>
  </v-card>
</template>

<script setup>
import { computed, ref } from 'vue';

const props = defineProps({
  headers: {
    type: [Object, Array, String],
    default: () => ({}),
  },
  maxHeight: {
    type: [String, Number],
    default: '380px',
  },
});

const searchQuery = ref('');
const showRawJson = ref(false);
const copiedIndex = ref(null);
const copiedAll = ref(false);

const maxHeightStyle = computed(() => {
  if (typeof props.maxHeight === 'number') {
    return `${props.maxHeight}px`;
  }
  return props.maxHeight;
});

// Normalize headers prop into a flat list of { key, value }
const parsedHeaders = computed(() => {
  if (!props.headers) return [];

  let rawData = props.headers;

  // If string, try JSON parse or parse MIME headers string
  if (typeof rawData === 'string') {
    try {
      rawData = JSON.parse(rawData);
    } catch {
      return parseMimeHeaderString(rawData);
    }
  }

  // If Array
  if (Array.isArray(rawData)) {
    const list = [];
    for (const item of rawData) {
      if (typeof item === 'object' && item !== null) {
        const key = item.key || item.name || item.header || item.field || '';
        const value = item.value || item.val || item.content || '';
        if (key) list.push({ key: String(key), value: String(value) });
      } else if (typeof item === 'string') {
        const idx = item.indexOf(':');
        if (idx !== -1) {
          list.push({
            key: item.substring(0, idx).trim(),
            value: item.substring(idx + 1).trim(),
          });
        }
      }
    }
    return list;
  }

  // If Object
  if (typeof rawData === 'object' && rawData !== null) {
    const list = [];
    for (const [key, value] of Object.entries(rawData)) {
      if (Array.isArray(value)) {
        for (const v of value) {
          list.push({ key: String(key), value: typeof v === 'object' ? JSON.stringify(v) : String(v) });
        }
      } else if (typeof value === 'object' && value !== null) {
        list.push({ key: String(key), value: JSON.stringify(value) });
      } else {
        list.push({ key: String(key), value: String(value ?? '') });
      }
    }
    return list;
  }

  return [];
});

function parseMimeHeaderString(str) {
  const lines = str.split(/\r?\n/);
  const list = [];
  let currentKey = '';
  let currentValue = '';

  for (const line of lines) {
    if (!line.trim()) continue;
    if (line.startsWith(' ') || line.startsWith('\t')) {
      // Multiline continuation
      currentValue += ' ' + line.trim();
    } else {
      if (currentKey) {
        list.push({ key: currentKey, value: currentValue });
      }
      const idx = line.indexOf(':');
      if (idx !== -1) {
        currentKey = line.substring(0, idx).trim();
        currentValue = line.substring(idx + 1).trim();
      } else {
        currentKey = line.trim();
        currentValue = '';
      }
    }
  }
  if (currentKey) {
    list.push({ key: currentKey, value: currentValue });
  }
  return list;
}

const filteredHeaders = computed(() => {
  if (!searchQuery.value) return parsedHeaders.value;
  const q = searchQuery.value.toLowerCase().trim();
  return parsedHeaders.value.filter(
    (h) => h.key.toLowerCase().includes(q) || h.value.toLowerCase().includes(q)
  );
});

const rawJsonString = computed(() => {
  if (!props.headers) return '{}';
  if (typeof props.headers === 'string') {
    try {
      return JSON.stringify(JSON.parse(props.headers), null, 2);
    } catch {
      return props.headers;
    }
  }
  return JSON.stringify(props.headers, null, 2);
});

function getHeaderColor(key) {
  const k = key.toLowerCase();
  if (['from', 'to', 'cc', 'bcc', 'subject', 'date', 'reply-to'].includes(k)) {
    return 'primary';
  }
  if (['dkim-signature', 'spf', 'authentication-results', 'received-spf', 'arc-seal'].includes(k)) {
    return 'success';
  }
  if (['message-id', 'in-reply-to', 'references'].includes(k)) {
    return 'info';
  }
  if (['content-type', 'content-transfer-encoding', 'mime-version'].includes(k)) {
    return 'secondary';
  }
  if (k.startsWith('x-')) {
    return 'warning';
  }
  return undefined;
}

async function copyHeaderValue(val, index) {
  try {
    await navigator.clipboard.writeText(val);
    copiedIndex.value = index;
    setTimeout(() => {
      if (copiedIndex.value === index) {
        copiedIndex.value = null;
      }
    }, 2000);
  } catch (err) {
    console.error('Failed to copy header value', err);
  }
}

async function copyAllHeaders() {
  try {
    const formatted = parsedHeaders.value
      .map((h) => `${h.key}: ${h.value}`)
      .join('\n');
    await navigator.clipboard.writeText(formatted || rawJsonString.value);
    copiedAll.value = true;
    setTimeout(() => {
      copiedAll.value = false;
    }, 2000);
  } catch (err) {
    console.error('Failed to copy all headers', err);
  }
}
</script>

<style scoped>
.email-headers-card {
  border-color: rgba(var(--v-border-color), var(--v-border-opacity));
}
.font-mono {
  font-family: 'Roboto Mono', monospace, SFMono-Regular, Menlo, Monaco, Consolas;
}
.headers-scroll-container {
  background-color: var(--v-theme-surface);
}
.headers-table {
  width: 100%;
}
.header-row:hover {
  background-color: rgba(var(--v-theme-on-surface), 0.04);
}
.header-value-cell {
  word-break: break-all;
  white-space: pre-wrap;
  line-height: 1.4;
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
.gap-2 {
  gap: 8px;
}
</style>
