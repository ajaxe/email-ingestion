import { defineStore } from 'pinia';
import { getApplication } from '@/services/apiService';

export const useAppStore = defineStore('app', {
  state: () => ({
    application: null,
    loading: false,
    error: null,
  }),
  actions: {
    async fetchAppDetails(appId) {
      this.loading = true;
      try {
        const res = await getApplication(appId);
        this.application = res.data || res;
        return this.application;
      } catch (err) {
        this.error = err;
        throw err;
      } finally {
        this.loading = false;
      }
    },
  },
});