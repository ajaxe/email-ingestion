<template>
  <form
    ref="formRef"
    method="POST"
    :action="logoutPath"
    @submit="onSubmit"
  >
    <input type="hidden" :name="csrfFormFieldName" :value="csrfToken ?? ''" />
    <slot :submit="submit" :submitting="submitting" :disabled="isDisabled">
      <button type="submit" :disabled="isDisabled">
        {{ submitting ? "Signing out..." : "Sign out" }}
      </button>
    </slot>
  </form>
</template>
<script setup>
import { computed, ref } from "vue";
import { storeToRefs } from "pinia";
import { useAuthStore } from "@/stores/auth.js";

const store = useAuthStore();
const { csrfToken, csrfFormFieldName, logoutPath } = storeToRefs(store);

const formRef = ref(null);
const submitting = ref(false);
const isDisabled = computed(() => !csrfToken.value || submitting.value);

function onSubmit() {
  submitting.value = true;
}

function submit() {
  if (isDisabled.value || !formRef.value) return;
  submitting.value = true;
  formRef.value.submit();
}
</script>
