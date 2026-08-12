import { defineStore } from "pinia";
import {
  passwordLogin,
  passwordLogout,
  getUser,
  signinRedirect,
  signoutRedirect,
  signoutRedirectCallback,
  signinRedirectCallback,
  emptyUser,
} from "@/services/authService";

import router from "@/router";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    provider: window.APP_CONFIG.AUTH_PROVIDER,
    user: emptyUser,
  }),
  getters: {
    isPasswordProvider: (state) => state.provider === "password",
    isAuthenticated: (state) => {
      return state.user.name !== "";
    },
  },
  actions: {
    async login(username, password) {
      if (this.isPasswordProvider) {
        if (!username || !password)
          throw new Error("Username and password are required");
        const success = await passwordLogin(username, password);
        return success;
      } else {
        await signinRedirect();
      }
    },
    async logout() {
      if (this.isPasswordProvider) {
        passwordLogout();
        router.push("/login");
      } else {
        await signoutRedirect();
      }
    },

    async loadUser() {
      this.user = await getUser();
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
        this.user = emptyUser;
        // Redirect to the home page
        router.push("/login");
      } catch (error) {
        console.error("Error handling login callback:", error);
        router.push("/login-failed"); // Or some error page
      }
    },
  },
});

