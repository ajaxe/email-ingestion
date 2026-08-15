<template>
  <v-container class="fill-height d-flex justify-center align-center">
    <LoginCard />
  </v-container>
</template>

<route lang="yaml">
meta:
  requiresAuth: false
  layout: NoLayout
  title: Login
</route>

<script setup>
import LoginCard from "@/components/LoginCard.vue";
</script>
