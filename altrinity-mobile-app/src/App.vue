<template>
  <v-app>
    <!-- App Bar -->
    <v-app-bar color="blue" density="comfortable" app>
      <v-btn><router-link to="/">Home</router-link></v-btn>
      <v-btn v-if="roles.includes('volunteer')">
        <router-link to="/VolunteerMap">Map</router-link>
      </v-btn>
      <v-btn v-if="roles.includes('admin')">
        <router-link to="/CommandHub">Command Hub</router-link>
      </v-btn>
      <v-btn v-if="roles.includes('admin')">
        <router-link to="/Admin">Admin</router-link>
      </v-btn>
      <v-spacer></v-spacer>
      <v-btn v-if="roles.length > 0" @click="keycloak?.logout({ redirectUri: 'altrinity://logout' })">
        Logout
      </v-btn>
      <v-btn v-else @click="openLogin(keycloak!)">Login</v-btn>
    </v-app-bar>

    <!-- Main content (router-view + map pages, etc.) -->
    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import type Keycloak from 'keycloak-js'
import { useAuthStore } from '@/stores/auth'
import { storeToRefs } from 'pinia'
import { openLogin } from './utils'
import { inject, onMounted } from 'vue'

const keycloak = inject<Keycloak>('keycloak')
const auth = useAuthStore()
const { roles } = storeToRefs(auth)

onMounted(async () => {
  await openLogin(keycloak!)
})
</script>

<style>
/* Optional: ensure router-view fills full height minus app bar */
.v-main {
  height: calc(100vh - 64px); /* adjust if app bar is dense or prominent */
  overflow: hidden;
}

.v-app-bar {
  padding-top: env(safe-area-inset-top);
}

.v-main {
  padding-top: env(safe-area-inset-top);
}
</style>
