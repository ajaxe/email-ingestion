/**
 * router/index.js
 *
 * Automatic routes for ./src/pages/*.vue
 */

// Composables
import { createRouter, createWebHistory } from "vue-router";
import { routes } from "vue-router/auto-routes";
import { useAuthStore } from "@/stores/auth";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

router.beforeEach(async (to) => {
  // Check if route requires authentication
  if (to.meta.requiresAuth) {
    const authStore = useAuthStore();
    await authStore.loadUser()
    if (!authStore.isAuthenticated) {
      // Redirect to login page
      return await Promise.resolve({ name: "/login/" });
    }

    // Load user from storage

    // If not logged in, redirect to login
    // Check if user is authorized
    return true
  }
});

export default router;
