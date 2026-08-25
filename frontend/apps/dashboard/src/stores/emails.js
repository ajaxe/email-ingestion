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
    totalCount: 0,
    attachments: {},
    loading: false,
    error: null,
    selectedLocalPart: "ALL",
    searchQuery: "",
  }),
  getters: {
    localPartOptions: () => {
      const addressStore = useAddressStore();
      const options = [{ title: "All Local Parts", value: "ALL" }];
      if (addressStore.addresses) {
        for (const addr of addressStore.addresses) {
          const val = addr.localPart;
          if (val) {
            options.push({ title: val, value: val });
          }
        }
      }
      return options;
    },
  },
  actions: {
    async fetchEmails(appId, { limit, page, localPart, search } = {}) {
      this.loading = true;
      limit = limit || 10;
      page = page || 1;
      localPart = localPart?.trim() ? localPart : this.selectedLocalPart;
      search = search?.trim() ? search : this.searchQuery;

      const params = { limit, page };
      if (localPart && localPart !== "ALL") {
        params.localPart = localPart;
      }
      if (search && search.trim()) {
        params.search = search.trim();
      }

      try {
        const res = await getEmailList(appId, params);
        const { emails, pagination } = res.data;
        this.emails = emails;
        this.totalCount = pagination?.totalCount || 0;
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
