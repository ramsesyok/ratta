<script setup>
import { ref } from 'vue'

const clickCount = ref(0)
</script>

<template>
  <v-app>
    <v-main>
      <v-container class="py-8">
        <v-card class="mx-auto" max-width="720" rounded="lg">
          <v-card-title class="text-h5">ratta</v-card-title>
          <v-card-subtitle>
            Issue tracking for contractor and vendor collaboration.
          </v-card-subtitle>
          <v-card-text class="text-body-1">
            This is the Phase 0 scaffold. Wails, Vue 3, Vuetify, and Pinia are
            wired and ready for feature work.
          </v-card-text>
          <v-card-actions class="align-center">
            <v-btn
              color="teal-darken-2"
              variant="flat"
              @click="clickCount++"
            >
              Demo Action
            </v-btn>
            <span class="text-caption">Clicked {{ clickCount }} times</span>
          </v-card-actions>
        </v-card>
      </v-container>
    </v-main>
  </v-app>
</template>
