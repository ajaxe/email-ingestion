import { defineStore } from "pinia";
import {
  passwordLogin,
  passwordLogout,
  getUser,
  userManager,
} from "@/services/authService";

import router from '@/router'

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
  getters:{
    isPasswordProvider: state => state.provider === 'password',

  },
  actions: {
    async login(username, password) {
      if (!this.isPasswordProvider) return;
      if (!username || !password)
        throw new Error("Username and password are required");
      const success = await passwordLogin(username, password);
      this.isAuthenticated = success;
      return success;
    },
    logout() {
      if(this.isPasswordProvider) {
        passwordLogout();
      } else {
        this.handleLogoutCallback()
      }
      this.isAuthenticated = false;
    },
    
    async loadUser() {
      this.user = await getUser();
      this.isAuthenticated = !!this.user.name;
      return this.isAuthenticated
    },

    async  handleLoginCallback() {
      if(this.isPasswordProvider) return;

      try {
        handleRedirectWithState(true)

        // Complete the login, exchange code for tokens
        await userManager.signoutRedirectCallback()
        this.user = null
        // Redirect to the home page
        router.push('/')
      } catch (error) {
        console.error('Error handling login callback:', error)
        router.push('/login-failed') // Or some error page
      }
    },

    async handleLogoutCallback() {
      if(this.isPasswordProvider) return;
      
      try {
        handleRedirectWithState(true)

        // Complete the login, exchange code for tokens
        await userManager.signoutRedirectCallback()
        this.user = null
        // Redirect to the home page
        router.push('/')
      } catch (error) {
        console.error('Error handling login callback:', error)
        router.push('/login-failed') // Or some error page
      }
    },
  },
});


function handleRedirectWithState(isLogout = false) {
  // get state for auth tenant context
  const urlParams = new URLSearchParams(window.location.search)
  const code = urlParams.get('code')
  const [fwdState, authState] = urlParams.get('state').split(';')

  if (authState) {
    const tenantContext = JSON.parse(atob(authState))

    if (tenantContext.returnUrl !== authSettings.appUrl) {
      window.location.href = isLogout
        ? `${tenantContext.returnUrl}/auth/logout?state=${fwdState}`
        : `${tenantContext.returnUrl}/auth/callback?code=${code}&state=${fwdState}`
      return
    }
  }
}
