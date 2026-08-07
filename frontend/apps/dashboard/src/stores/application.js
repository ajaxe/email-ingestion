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
        this.application = await getApplication(appId);
      } catch (err) {
        this.error = err;
      } finally {
        this.loading = false;
      }
    },
  },
});