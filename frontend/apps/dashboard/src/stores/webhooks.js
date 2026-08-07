import { defineStore } from 'pinia';
import { configureWebhook, getWebhookJobs, redeliverWebhook } from '@/services/apiService';

export const useWebhookStore = defineStore('webhooks', {
  state: () => ({
    config: null,
    jobs: [],
    loading: false,
    error: null,
  }),
  actions: {
    async setupWebhook(appId, config) {
      this.loading = true;
      try {
        const res = await configureWebhook(appId, config);
        this.config = res.data || res;
        return this.config;
      } catch (err) {
        this.error = err;
        throw err;
      } finally {
        this.loading = false;
      }
    },
    async fetchJobs(appId, queryParams) {
      this.loading = true;
      try {
        const res = await getWebhookJobs(appId, queryParams);
        this.jobs = res.data || res;
        return this.jobs;
      } catch (err) {
        this.error = err;
        throw err;
      } finally {
        this.loading = false;
      }
    },
    async redeliverJob(appId, jobId) {
      this.loading = true;
      try {
        const res = await redeliverWebhook(appId, jobId);
        return res.data || res;
      } catch (err) {
        this.error = err;
        throw err;
      } finally {
        this.loading = false;
      }
    },
  },
});