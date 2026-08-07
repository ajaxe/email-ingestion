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
        this.config = await configureWebhook(appId, config);
      } finally {
        this.loading = false;
      }
    },
    async fetchJobs(appId, queryParams) {
      this.loading = true;
      try {
        this.jobs = await getWebhookJobs(appId, queryParams);
      } finally {
        this.loading = false;
      }
    },
    async redeliverJob(appId, jobId) {
      this.loading = true;
      try {
        return await redeliverWebhook(appId, jobId);
      } finally {
        this.loading = false;
      }
    },
  },
});