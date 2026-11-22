<template>
  <v-card class="mx-auto pa-12 pb-8" elevation="3" width="448" rounded="lg">
    <AppName />
    
    <v-btn
      class="mb-8 mt-4"
      color="blue"
      size="large"
      variant="tonal"
      block
      href="/app/login"
    >
      <template #append>
        <v-icon icon="mdi-open-in-new" />
      </template>
      Log In using Idp
    </v-btn>
  </v-card>
</template>

<script setup>
import AppName from "./AppName.vue";
</script>
