import { defineStore } from 'pinia'
// import { userManager, authSettings } from '@/services/auth.service'
// import router from '@/router'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    isAuthenticated: false,
  })
})