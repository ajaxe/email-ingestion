/**
 * router/index.js
 *
 * Automatic routes for ./src/pages/*.vue
 */

// Composables
import { ref } from "vue";
import { createRouter, createWebHistory } from "vue-router";
import { routes } from "vue-router/auto-routes";
import { useAuthStore } from "@/stores/auth";
import { useAppStore } from "@/stores/app";

export const isNavigating = ref(false);

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

router.beforeEach(async (to) => {
  isNavigating.value = true;
  // Check if route requires authentication
  if (to.meta.requiresAuth) {
    const authStore = useAuthStore();
    await authStore.loadUser();
    if (!authStore.isAuthenticated) {
      // Redirect to login page
      return { name: "/login/" };
    }
    console.log("to", to)
    const appStore = useAppStore()
    if (to.path !== "/" && !appStore.hasApplication) {
      return { path: "/" }
    }

    // Load user from storage

    // If not logged in, redirect to login
    // Check if user is authorized
    return true;
  }
});

// Always reset loading status regardless of outcome
router.afterEach(() => {
  isNavigating.value = false;
});

router.onError(() => {
  isNavigating.value = false;
});

export default router;
