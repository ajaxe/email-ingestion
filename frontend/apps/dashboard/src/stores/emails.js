import { defineStore } from "pinia";
import {
  getEmailList,
  getAttachmentUrl,
  getEmailById,
} from "@/services/apiService";
import { useAddressStore } from "./addresses";

export const useEmailStore = defineStore("emails", {
  state: () => ({
    emails: [],
    attachments: {},
    loading: false,
    error: null,
    selectedLocalPart: 'ALL',
    searchQuery: '',
  }),
  getters: {
    localPartOptions: () => {
      const addressStore = useAddressStore();
      const options = [{ title: "All Local Parts", value: "ALL" }];
      if (addressStore.addresses) {
        for (const addr of addressStore.addresses) {
          const val = addr.local_part || addr.localPart;
          if (val) {
            options.push({ title: val, value: val });
          }
        }
      }
      return options;
    },
  },
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

    async fetchEmailById(appId, emailId) {
      this.loading = true;
      try {
        const res = await getEmailById(appId, emailId);
        return res.data || res;
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
