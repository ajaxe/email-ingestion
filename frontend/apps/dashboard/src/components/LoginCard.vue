<template>
  <v-card class="mx-auto pa-12 pb-8" elevation="3" width="448" rounded="lg">
    <AppName />
    <v-form ref="form" v-if="usePassword">
      <div class="text-body-large text-medium-emphasis">Account</div>

      <span class="text-red" v-if="loginErr">Invalid username or password</span>

      <v-text-field
        density="compact"
        placeholder="Email address"
        prepend-inner-icon="mdi-email-outline"
        variant="outlined"
        v-model="username"
        :rules="[(v) => !!v || 'Username is required.']"
      ></v-text-field>

      <div
        class="text-body-large text-medium-emphasis d-flex align-center justify-space-between"
      >
        Password
      </div>

      <v-text-field
        :append-inner-icon="visible ? 'mdi-eye-off' : 'mdi-eye'"
        :type="visible ? 'text' : 'password'"
        density="compact"
        placeholder="Enter your password"
        prepend-inner-icon="mdi-lock-outline"
        variant="outlined"
        @click:append-inner="visible = !visible"
        v-model="password"
        :rules="[(v) => !!v || 'Password is required.']"
      ></v-text-field>

      <v-btn
        class="mb-8 mt-4"
        color="blue"
        size="large"
        variant="tonal"
        block
        :loading="loading"
        :disabled="disabled"
        @click="login"
      >
        Log In
      </v-btn>
    </v-form>

    <v-btn
      v-if="!usePassword"
      class="mb-8 mt-4"
      color="blue"
      size="large"
      variant="tonal"
      block
      :loading="loading"
      @click="login"
    >
      <template #append>
        <v-icon icon="mdi-open-in-new" />
      </template>
      Log In using Idp
    </v-btn>
  </v-card>
</template>

<script setup>
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import AppName from "./AppName.vue";

const form = ref();

const visible = ref(false);
const loginErr = ref(false);
const loading = ref(false);
const username = ref("");
const password = ref("");
const disabled = computed(() => !username.value || !password.value);
const store = useAuthStore();
const router = useRouter();
const usePassword = ref(store.isPasswordProvider);

async function login() {
  loading.value = true;

  try {
    let success = false;

    if (usePassword.value) {
      const { valid } = await form.value.validate();
      if (valid) {
        success = await store.login(username.value, password.value);
      }
    } else {
      success = await store.login();
    }
    loginErr.value = !success;
    if (success) {
      router.push({ name: "/" });
    }
  } catch {
    loginErr.value = true;
  }
  loading.value = false;
}
</script>
