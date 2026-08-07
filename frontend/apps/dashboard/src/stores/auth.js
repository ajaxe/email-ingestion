import { defineStore } from "pinia";
import {
  passwordLogin,
  passwordLogout,
  checkAuthStatus,
} from "@/services/authService";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    isAuthenticated: false,
  }),
  actions: {
    async login(username, password) {
      if (!username || !password)
        throw new Error("Username and password are required");
      const success = await passwordLogin(username, password);
      this.isAuthenticated = success;
      return success;
    },
    logout() {
      passwordLogout();
      this.isAuthenticated = false;
    },
    async loadUser() {
      const status = await checkAuthStatus();
      this.isAuthenticated = status
      return status
    },
  },
});
