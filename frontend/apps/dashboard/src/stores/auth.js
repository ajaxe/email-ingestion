import { defineStore } from "pinia";
import {
  passwordLogin,
  passwordLogout,
  getUser,
  getAuthSession,
  signinRedirect,
  signoutRedirect,
  signoutRedirectCallback,
  signinRedirectCallback,
  emptyUser,
} from "@/services/authService";

import { useAppStore } from "@/stores/application";

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
      const [u, userAccess] = await Promise.allSettled([
        getUser(),
        getAuthSession(),
      ]);
      if (u.status === "fulfilled" && u.value) {
        this.user = u.value;
      }
      
      let isForbiddenUser = false;
      let isTokenInvalid = false;
      let apps = [];

      if (userAccess.status === "rejected") {
        const err = userAccess.reason;
        const msg = err?.message;

        if (msg === "forbidden_token_invalid" || err?.status === 401) {
          isTokenInvalid = true;
          this.user = emptyUser;
          passwordLogout();
        } else if (msg === "forbidden_user" || err?.status === 403) {
          isForbiddenUser = true;
        }
      } else if (userAccess.status === "fulfilled") {
        apps = userAccess.value.applications || [];
        if (apps.length > 0) {
          const appStore = useAppStore();
          appStore.applications = apps;
          appStore.application = apps[0];
          appStore.activeAppId = apps[0].id;
        }
      }
      return {
        isAuthenticated: this.isAuthenticated && !isTokenInvalid,
        forbidden: isForbiddenUser,
        applications: apps,
      };
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

    async checkSession() {
      try {
        await getAuthSession();
      } catch (error) {
        console.error("Error checking session:", error);
        this.user = emptyUser;
        router.push("/login");
      }
    },
  },
});
