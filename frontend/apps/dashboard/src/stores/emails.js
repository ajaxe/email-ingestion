import { defineStore } from 'pinia';
import { getEmailList, getAttachmentUrl } from '@/services/apiService';

export const useEmailStore = defineStore('emails', {
  state: () => ({
    emails: [],
    attachments: {},
    loading: false,
    error: null,
  }),
  actions: {
    async fetchEmails(appId, queryParams) {
      this.loading = true;
      try {
        const res = await getEmailList(appId, queryParams);
        this.emails = res.data || res;
        return this.emails;
      } catch (err) {
        this.error = err;
        throw err;
      } finally {
        this.loading = false;
      }
    },
    async fetchAttachmentUrl(appId, emailId, attachmentId) {
      this.loading = true;
      try {
        const res = await getAttachmentUrl(appId, emailId, attachmentId);
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