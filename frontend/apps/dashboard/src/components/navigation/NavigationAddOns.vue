<template>
  <v-list density="compact" nav>
    <v-list-item
      prepend-avatar="https://randomuser.me/api/portraits/women/85.jpg"
      title="Sandra Adams"
      subtitle="sandra_a88@gmail.com"
    >
      <template #append>
        <v-btn
          class=""
          icon="mdi-pin-outline"
          variant="text"
          size="x-small"
        ></v-btn>
      </template>
    </v-list-item>

    <v-list-item prepend-icon="mdi-logout" title="Logout" @click="() => ''">
    </v-list-item>
  </v-list>
</template>
