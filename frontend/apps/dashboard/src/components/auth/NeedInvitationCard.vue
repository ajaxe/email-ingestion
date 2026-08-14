<template>
  <v-card rounded="lg" elevation="3" class="pa-8 text-center bg-surface border">
    <!-- Top Alert Avatar Icon -->
    <div class="d-flex justify-center mb-4">
      <v-avatar color="amber-lighten-4" size="72" class="elevation-1">
        <v-icon icon="mdi-email-lock-outline" size="40" color="amber-darken-3" />
      </v-avatar>
    </div>

    <!-- Status Chip -->
    <div class="mb-3">
      <v-chip color="warning" variant="tonal" size="small" class="font-weight-bold">
        <v-icon start icon="mdi-clock-outline" size="14" />
        Access Invitation Pending
      </v-chip>
    </div>

    <!-- Main Title & Explanation -->
    <h2 class="text-h5 font-weight-bold mb-3">Invitation Required</h2>

    <p class="text-body-1 text-medium-emphasis mx-auto mb-6 max-w-600">
      Your account <strong class="text-high-emphasis" v-if="userEmail">{{ userEmail }}</strong> is authenticated, but you do not currently have access to any application tenants. An invitation from your organization administrator is required to access the gateway console.
    </p>

    <!-- Primary Action Buttons -->
    <div class="d-flex justify-center flex-wrap gap-3 mb-8">
      <v-btn
        color="primary"
        size="large"
        variant="flat"
        prepend-icon="mdi-refresh"
        :loading="checkingAccess"
        class="px-6 font-weight-bold"
        @click="checkAccess"
      >
        Check Access Status
      </v-btn>

      <v-btn
        color="secondary"
        size="large"
        variant="tonal"
        prepend-icon="mdi-content-copy"
        class="px-5 font-weight-medium"
        @click="copyAccountInfo"
      >
        Copy Email Info
      </v-btn>

      <v-btn
        color="error"
        size="large"
        variant="outlined"
        prepend-icon="mdi-logout"
        class="px-5 font-weight-medium"
        @click="handleLogout"
      >
        Sign Out
      </v-btn>
    </div>

    <v-divider class="my-6" />

    <!-- Next Steps Instructions -->
    <div class="text-start">
      <h3 class="text-subtitle-2 font-weight-bold text-uppercase text-medium-emphasis mb-4 tracking-wider">
        Next Steps to Resolve Access
      </h3>

      <v-row class="justify-center">
        <v-col cols="12" sm="4">
          <div class="d-flex align-start gap-3 pa-3 rounded-lg bg-background border fill-height">
            <v-avatar color="info" variant="tonal" size="32" class="mt-1">
              <span class="font-weight-bold text-caption">1</span>
            </v-avatar>
            <div>
              <div class="text-subtitle-2 font-weight-bold">Request Invitation</div>
              <div class="text-caption text-medium-emphasis">
                Contact your organization administrator and request access to the Gateway.
              </div>
            </div>
          </div>
        </v-col>

        <v-col cols="12" sm="4">
          <div class="d-flex align-start gap-3 pa-3 rounded-lg bg-background border fill-height">
            <v-avatar color="warning" variant="tonal" size="32" class="mt-1">
              <span class="font-weight-bold text-caption">2</span>
            </v-avatar>
            <div>
              <div class="text-subtitle-2 font-weight-bold">Tenant Assignment</div>
              <div class="text-caption text-medium-emphasis">
                Admin provisions your account with application tenant permissions.
              </div>
            </div>
          </div>
        </v-col>

        <v-col cols="12" sm="4">
          <div class="d-flex align-start gap-3 pa-3 rounded-lg bg-background border fill-height">
            <v-avatar color="success" variant="tonal" size="32" class="mt-1">
              <span class="font-weight-bold text-caption">3</span>
            </v-avatar>
            <div>
              <div class="text-subtitle-2 font-weight-bold">Re-check Permission</div>
              <div class="text-caption text-medium-emphasis">
                Click "Check Access Status" above to enter your assigned dashboard.
              </div>
            </div>
          </div>
        </v-col>
      </v-row>
    </div>
  </v-card>
</template>

<script setup>
import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/stores/auth';
import { useNotificationStore } from '@/stores/notification';

const authStore = useAuthStore();
const notificationStore = useNotificationStore();
const router = useRouter();

const checkingAccess = ref(false);

const userEmail = computed(() => authStore.user?.email || authStore.user?.name || '');

async function checkAccess() {
  checkingAccess.value = true;
  try {
    const result = await authStore.loadUser();
    if (result.isAuthenticated && !result.forbidden) {
      notificationStore.success('Access granted! Redirecting to gateway dashboard...');
      router.push('/');
    } else {
      notificationStore.warning('Account still pending invitation. Please contact your administrator.');
    }
  } catch (error) {
    console.error('Error checking user access status:', error);
    notificationStore.error('Failed to verify session status. Please try again.');
  } finally {
    checkingAccess.value = false;
  }
}

function copyAccountInfo() {
  const info = userEmail.value || 'No email available';
  navigator.clipboard.writeText(info);
  notificationStore.info(`Copied "${info}" to clipboard.`);
}

function handleLogout() {
  authStore.logout();
}
</script>

<style scoped>
.max-w-600 {
  max-width: 600px;
}
.gap-3 {
  gap: 12px;
}
.tracking-wider {
  letter-spacing: 0.08em;
}
.fill-height {
  height: 100%;
}
</style>
