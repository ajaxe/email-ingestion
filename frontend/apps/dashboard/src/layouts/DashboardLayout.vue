<template>
  <v-layout class="rounded rounded-md">
    <v-navigation-drawer expand-on-hover permanent>
      <NavigationLinks />

      <template #prepend>
        <v-list>
          <v-list-item
            prepend-icon="mdi-email-multiple"
            title="Email Ingestion Gateway"
            lines="two"
          >
          </v-list-item>
        </v-list>

        <v-divider></v-divider>
      </template>

      <template #append>
        <v-divider></v-divider>

        <NavigationAddOns />
      </template>
    </v-navigation-drawer>

    <v-main class="d-flex align-center justify-center">
      <v-container fluid>
        <div>
          <AppSelector />
        </div>

        <v-sheet rounded="lg" width="100%">
          <slot />
        </v-sheet>
      </v-container>
    </v-main>
  </v-layout>
</template>
<script setup>
import NavigationLinks from "@/components/navigation/NavigationLinks.vue";
import NavigationAddOns from "@/components/navigation/NavigationAddOns.vue";
import AppSelector from "@/components/AppSelector.vue";
</script>
