<template>
  <v-app>
    <!-- App Bar -->
    <v-app-bar color="blue" density="comfortable" app>
      <v-btn><router-link to="/">Home</router-link></v-btn>

      <v-btn v-if="roles.includes('volunteer')">
        <router-link to="/VolunteerMap">Map</router-link>
      </v-btn>

      <v-btn v-if="roles.includes('admin')">
        <router-link to="/Admin">Admin</router-link>
      </v-btn>

      <v-spacer></v-spacer>

      <!-- 🔸 Pending routes indicator (only shows if > 0) -->
      <v-chip
        v-if="offlineCount > 0"
        color="orange"
        text-color="white"
        label
        size="small"
        class="mr-3"
      >
        {{ offlineCount }} pending route<span v-if="offlineCount > 1">s</span>
      </v-chip>

      <v-btn
        v-if="roles.length > 0"
        @click="keycloak?.logout({ redirectUri: 'altrinity://logout' })"
      >
        Logout
      </v-btn>

      <v-btn v-else @click="openLogin(keycloak!)">Login</v-btn>
    </v-app-bar>

    <router-view />
  </v-app>
</template>

<script setup lang="ts">
import { inject, onMounted, onUnmounted } from 'vue'
import type Keycloak from 'keycloak-js'
import { useAuthStore } from '@/stores/auth'
import { storeToRefs } from 'pinia'
import { openLogin } from './utils'
import { useRouteQueue } from '@/composables/useRouteQueue'
import { Network } from '@capacitor/network'

const keycloak = inject<Keycloak>('keycloak')
const authStore = useAuthStore()
const { roles } = storeToRefs(authStore)
const { offlineCount, flushQueue } = useRouteQueue()

onMounted(async () => {
  // Login first to ensure token is available
  await openLogin(keycloak!)
  // Check immediately if we can flush again (if we just regained network)
  const status = await Network.getStatus()
  if (status.connected && authStore.token) {
    await flushQueue();
  }

  setInterval(async () => {
    if(keycloak) { 
      try {
        const refreshed = await keycloak.updateToken(30)
        if (refreshed) {
          console.log('🔄 Token refreshed')
          authStore.setToken(keycloak.token!, keycloak.refreshToken)
        }
      } catch (err) {
        console.warn('⚠️ Token refresh failed, forcing re-login', err)
        await openLogin(keycloak!)
      }
    }
}, 60_000)


})

</script>

<style>
.v-main {
  /* subtract app bar height from viewport */
  height: 100vh;
  overflow: hidden;

  /* ✅ combine fallback + iOS safe area */
  padding-top: calc(var(--v-layout-top) + max(env(safe-area-inset-top, 0px), 8px)) !important;
}

.v-app-bar {
  /* ✅ top notch-safe padding (iOS) + fallback */
  padding-top: max(env(safe-area-inset-top, 0px), 8px);
}
</style>
