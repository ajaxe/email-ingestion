<template>
  <div class="d-flex align-center gap-3">
    <v-avatar
      :image="logo"
      variant="flat"
      size="40"
      class="elevation-2"
    >
    </v-avatar>
    <div class="ms-2">
      <div class="text-h6 font-weight-bold line-height-tight">
        Email Ingestion
      </div>
      <div
        class="text-caption text-medium-emphasis text-uppercase tracking-wider"
      >
        Gateway Console
      </div>
    </div>
  </div>
</template>
<script setup>
const logo = "/favicon.svg"
</script>