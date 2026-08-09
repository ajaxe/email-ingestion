import { defineStore } from 'pinia';
import { getApplicationById, getApplications, getApiKeys, createApiKey, revokeApiKey, getApplicationStats } from '@/services/apiService';

export const useAppStore = defineStore('app', {
  state: () => ({
    applications: [],
    activeAppId: '',
    application: null,
    apiKeys: [],
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
      await this.fetchApiKeys(appId);
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
      this.loading = true;
      try {
        const res = await getApplications();
        this.applications = res.data;
        if(!this.activeAppId && this.applications.length > 0) {
          this.activeAppId = this.applications[0].id;
        }
      } catch (err) {
        this.error = err;
      } finally {
        this.loading = false;
      }
    },

    async fetchApiKeys(appId) {
      const targetId = appId || this.activeAppId;
      if (!targetId) return [];
      this.loading = true;
      try {
        const res = await getApiKeys(targetId);
        this.apiKeys = res.data || [];
        return this.apiKeys;
      } catch (err) {
        this.error = err;
        return [];
      } finally {
        this.loading = false;
      }
    },

    async generateApiKey(appId, payload) {
      const targetId = appId || this.activeAppId;
      const body = payload || { name: 'Dashboard Key' };
      this.loading = true;
      try {
        const res = await createApiKey(targetId, body);
        const apiKeyData = res.data || res;
        this.latestApiKey = apiKeyData.apiKey || apiKeyData.api_key || apiKeyData.APIKey || apiKeyData.key || '';
        await this.fetchApiKeys(targetId);
        return apiKeyData;
      } catch (err) {
        this.error = err;
        throw err;
      } finally {
        this.loading = false;
      }
    },

    async revokeApiKey(appId, keyId) {
      const targetId = appId || this.activeAppId;
      this.loading = true;
      try {
        await revokeApiKey(targetId, keyId);
        await this.fetchApiKeys(targetId);
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
        this.stats = res.data;
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