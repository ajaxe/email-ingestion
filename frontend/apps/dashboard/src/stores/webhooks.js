import { defineStore } from "pinia";
import {
  registerWebhook,
  updateWebhook,
  getWebhookJobs,
  redeliverWebhook,
} from "@/services/apiService";

export const useWebhookStore = defineStore("webhooks", {
  state: () => ({
    config: null,
    jobs: [],
    loading: false,
    error: null,
    totalCount: 0,
    statusFilter:  "ALL",
  }),
  actions: {
    async registerWebhook(appId, config) {
      this.loading = true;
      try {
        const res = await registerWebhook(appId, config);
        this.config = res.data || res;
        return this.config;
      } catch (err) {
        this.error = err;
        throw err;
      } finally {
        this.loading = false;
      }
    },
    async updateWebhook(appId, config) {
      this.loading = true;
      try {
        const res = await updateWebhook(appId, config);
        this.config = res.data || res;
        return this.config;
      } catch (err) {
        this.error = err;
        throw err;
      } finally {
        this.loading = false;
      }
    },
    
    async verifyWebhook(appId, config) {
      this.loading = true;
      try {
        const res = await updateWebhook(appId, config);
        this.config = res.data || res;
        return this.config;
      } catch (err) {
        this.error = err;
        throw err;
      } finally {
        this.loading = false;
      }
    },

    async setupWebhook(appId, config) {
      return this.updateWebhook(appId, config);
    },

    async fetchJobs(appId, queryParams) {
      this.loading = true;
      queryParams = queryParams || {};
      queryParams.limit = queryParams.limit || 10;
      queryParams.page = queryParams.page || 1;
      if(this.statusFilter !== "ALL") {
        queryParams.status = this.statusFilter;
      } else if(queryParams.status === "ALL") {
        delete queryParams.status;
      }
      try {
        const res = await getWebhookJobs(appId, queryParams);
        const { jobs, pagination } = res.data;
        this.jobs = jobs;
        this.totalCount = pagination?.totalCount || 0;
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
