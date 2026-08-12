import { defineStore } from "pinia";
import {
  passwordLogin,
  passwordLogout,
  checkAuthStatus,
  getUserProfile,
} from "@/services/authService";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    isAuthenticated: false,
    provider: window.APP_CONFIG.AUTH_PROVIDER,
    user: {
      name: "",
      email: "",
      profileImage: "",
    }
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
      this.user = await getUserProfile()
      return status
    },
  },
});
