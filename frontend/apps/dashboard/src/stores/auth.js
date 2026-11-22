import { defineStore } from "pinia";
import { getUser, getAuthSession, emptyUser } from "@/services/authService";

import { useAppStore } from "@/stores/application";

import router from "@/router";

export const useAuthStore = defineStore("auth", {
  state: () => ({
    provider: window.APP_CONFIG.AUTH_PROVIDER,
    user: emptyUser,
    csrfToken: "",
    csrfFormFieldName: "",
    logoutPath: "/app/logout",
  }),
  getters: {
    isAuthenticated: (state) => {
      return state.user.name !== "";
    },
  },
  actions: {
    async loadUser() {
      const [u, userAccess] = await Promise.allSettled([
        getUser(),
        getAuthSession(),
      ]);
      if (u.status === "fulfilled" && u.value) {
        this.user = u.value;
        this.csrfToken = u.value.csrfToken;
        this.csrfFormFieldName = u.value.csrfFormFieldName;
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
