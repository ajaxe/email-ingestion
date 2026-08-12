import { defineStore } from "pinia";
import {
  passwordLogin,
  passwordLogout,
  getUser,
  signinRedirect,
  signoutRedirectCallback,
  signinRedirectCallback,
} from "@/services/authService";

import router from "@/router";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    isAuthenticated: false,
    provider: window.APP_CONFIG.AUTH_PROVIDER,
    user: {
      name: "",
      email: "",
      profileImage: "",
    },
  }),
  getters: {
    isPasswordProvider: (state) => state.provider === "password",
  },
  actions: {
    async login(username, password) {
      if (this.isPasswordProvider) {
        if (!username || !password)
          throw new Error("Username and password are required");
        const success = await passwordLogin(username, password);
        this.isAuthenticated = success;
        return success;
      } else {
        await signinRedirect();
      }
    },
    logout() {
      this.isAuthenticated = false;
      if (this.isPasswordProvider) {
        passwordLogout();
      } else {
        this.handleLogoutCallback();
      }
    },

    async loadUser() {
      this.user = await getUser();
      this.isAuthenticated = !!this.user?.name;
      return this.isAuthenticated;
    },

    async handleLoginCallback() {
      if (this.isPasswordProvider) return;

      try {
        // Complete the login, exchange code for tokens
        const u = await signinRedirectCallback();
        this.user = u;
        // Redirect to the home page
        router.push("/");
      } catch (error) {
        console.error("Error handling login callback:", error);
        router.push("/login-failed"); // Or some error page
      }
    },

    async handleLogoutCallback() {
      if (this.isPasswordProvider) return;

      try {
        // Complete the login, exchange code for tokens
        await signoutRedirectCallback();
        this.user = null;
        // Redirect to the home page
        router.push("/");
      } catch (error) {
        console.error("Error handling login callback:", error);
        router.push("/login-failed"); // Or some error page
      }
    },
  },
});

