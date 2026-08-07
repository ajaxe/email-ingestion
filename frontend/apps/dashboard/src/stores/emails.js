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
        this.emails = await getEmailList(appId, queryParams);
      } finally {
        this.loading = false;
      }
    },
    async fetchAttachmentUrl(appId, emailId, attachmentId) {
      this.loading = true;
      try {
        return await getAttachmentUrl(appId, emailId, attachmentId);
      } finally {
        this.loading = false;
      }
    },
  },
});