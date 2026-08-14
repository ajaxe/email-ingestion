<template>
  <v-card :elevation="elevation" rounded="lg" class="pa-4 bg-surface border">
    <div class="d-flex align-center justify-space-between flex-wrap gap-3">
      <!-- User Avatar & Identity Details -->
      <div class="d-flex align-center gap-3">
        <v-avatar size="52" :color="avatarColor" variant="tonal" class="font-weight-bold text-subtitle-1">
          <v-img v-if="profileImage" :src="profileImage" :alt="username" />
          <span v-else>{{ userInitials }}</span>
        </v-avatar>

        <div>
          <div class="d-flex align-center gap-2">
            <h3 class="text-subtitle-1 font-weight-bold line-height-tight mb-0">
              {{ username }}
            </h3>
            <v-chip size="x-small" color="primary" variant="tonal" class="font-weight-medium">
              <v-icon start icon="mdi-shield-account-outline" size="12" />
              {{ providerLabel }}
            </v-chip>
          </div>

          <div class="text-caption text-medium-emphasis d-flex align-center gap-1 mt-1" v-if="email">
            <v-icon icon="mdi-email-outline" size="14" />
            <span>{{ email }}</span>
          </div>
        </div>
      </div>

      <!-- Action Controls (Sign Out, Theme Toggle) -->
      <div v-if="showActions" class="d-flex align-center gap-2">

        <v-btn
          variant="outlined"
          color="error"
          size="small"
          prepend-icon="mdi-logout"
          class="font-weight-medium"
          @click="handleLogout"
        >
          Sign Out
        </v-btn>
      </div>

      <!-- Optional Custom Slot for Extra Actions -->
      <slot name="actions" />
    </div>
  </v-card>
</template>

<script setup>
import { computed } from 'vue';
import { useTheme } from 'vuetify';
import { useAuthStore } from '@/stores/auth';

const props = defineProps({
  user: {
    type: Object,
    default: null,
  },
  elevation: {
    type: [Number, String],
    default: 1,
  },
  showActions: {
    type: Boolean,
    default: true,
  },
});

const theme = useTheme();
const authStore = useAuthStore();

const activeUser = computed(() => props.user || authStore.user || {});
const username = computed(() => activeUser.value.name || 'Authenticated User');
const email = computed(() => activeUser.value.email || '');
const profileImage = computed(() => activeUser.value.profileImage || '');

const providerLabel = computed(() => {
  return authStore.provider === 'oidc' ? 'OIDC SSO' : 'Password Auth';
});

const avatarColor = computed(() => {
  if (profileImage.value) return 'transparent';
  return 'primary';
});

const userInitials = computed(() => {
  if (username.value && username.value !== 'Authenticated User') {
    const parts = username.value.trim().split(' ');
    if (parts.length >= 2) {
      return (parts[0][0] + parts[1][0]).toUpperCase();
    }
    return username.value.slice(0, 2).toUpperCase();
  }
  if (email.value) {
    return email.value.slice(0, 2).toUpperCase();
  }
  return 'US';
});

function toggleTheme() {
  theme.global.name.value = theme.global.current.value.dark ? 'light' : 'dark';
}

function handleLogout() {
  authStore.logout();
}
</script>

<style scoped>
.gap-3 {
  gap: 12px;
}
.gap-2 {
  gap: 8px;
}
.gap-1 {
  gap: 4px;
}
.line-height-tight {
  line-height: 1.2;
}
</style>
