import { defineStore } from 'pinia';
import { getApplicationById, getApplications, createApiKey, getApplicationStats } from '@/services/apiService';

export const useAppStore = defineStore('app', {
  state: () => ({
    applications: [],
    activeAppId: '',
    application: null,
    loading: false,
    error: null,
    latestApiKey: '',
    stats: {
      totalEmails: 0,
      totalAddresses: 0,
      activeAddresses: 0,
      webhookSuccessRate: 0,
      failWebhookJobCount: 0,
    }
  }),
  getters: {
    activeApp(state) {
      return state.application || state.applications.find(a => a.id === state.activeAppId) || null;
    },
    activeAppStatus(state) {
      return state.application?.status || state.activeApp?.status || 'ACTIVE';
    },
  },
  actions: {
    async selectApp(appId) {
      this.activeAppId = appId;
      await this.fetchAppDetails(appId);
    },
    async fetchAppDetails(appId) {
      const targetId = appId || this.activeAppId;
      this.loading = true;
      try {
        const res = await getApplicationById(targetId);
        this.application = res.data || res;
        return this.application;
      } catch (err) {
        this.error = err;
        // Fallback to local item if backend request fails (e.g., in dev/mock)
        const local = this.applications.find(a => a.id === targetId);
        if (local) {
          this.application = { ...local };
        }
        return this.application;
      } finally {
        this.loading = false;
      }
    },

    async fetchApplications() {
      this.loading = true
      try {
        const res = await getApplications();
        this.applications = res.data
        if(!this.activeAppId && this.applications.length > 0) {
          this.activeAppId = this.applications[0].id
        }
      } catch (err) {
        this.error = err;
      } finally {
        this.loading = false
      }
    },

    async generateApiKey(appId, name = 'Dashboard Key') {
      const targetId = appId || this.activeAppId;
      this.loading = true;
      try {
        const res = await createApiKey(targetId, name);
        const apiKeyData = res.data || res;
        this.latestApiKey = apiKeyData.api_key || apiKeyData.APIKey || apiKeyData.key || '';
        return apiKeyData;
      } catch (err) {
        this.error = err;
        throw err;
      } finally {
        this.loading = false;
      }
    },

    async fetchStatistics(appId) {
      const targetId = appId || this.activeAppId;
      this.loading = true;
      try {
        const res = await getApplicationStats(targetId);
        this.stats = res.data
        return res.data;
      } catch (err) {
        this.error = err;
        throw err;
      } finally {
        this.loading = false;
      }
    }
  },
});