import { defineStore } from 'pinia';
import { getAddresses, createAddress, toggleAddressStatus } from '@/services/apiService';

export const useAddressStore = defineStore('addresses', {
  state: () => ({
    addresses: [],
    loading: false,
    error: null,
  }),
  getters: {
    activeAddresses: state => state.addresses.filter(a => a.status === 'ACTIVE'),
  },
  actions: {
    async fetchAddresses(appId, queryParams) {
      this.loading = true;
      try {
        const res = await getAddresses(appId, queryParams);
        this.addresses = res.data || res;
        return this.addresses;
      } catch (err) {
        this.error = err;
        throw err;
      } finally {
        this.loading = false;
      }
    },
    async provisionAddress(appId, description) {
      this.loading = true;
      try {
        return await createAddress(appId, description);
      } finally {
        this.loading = false;
      }
    },
    async toggleAddress(appId, addressId, newStatus) {
      this.loading = true;
      try {
        return await toggleAddressStatus(appId, addressId, newStatus);
      } finally {
        this.loading = false;
      }
    },
  },
});