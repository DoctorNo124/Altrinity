<template>
  <v-app>
    <v-app-bar class="d-flex flex-row justify-center bg-blue">
      <v-btn><router-link to="/">Home</router-link></v-btn>
      <v-btn><router-link to="/Map">Map</router-link></v-btn>
      <v-btn><router-link v-if="roles.includes('admin')" to="/Admin">Admin</router-link></v-btn>
      <v-spacer></v-spacer>
      <v-btn @click="keycloak?.logout()">Logout</v-btn>
    </v-app-bar>
    <router-view />
  </v-app>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import { useApi } from '@/composables/keycloak-api'
import type Keycloak from 'keycloak-js'
import type { RouteRecordRaw } from 'vue-router'
  
const keycloak = inject<Keycloak>('keycloak')

const api = useApi()
const userMsg = ref('')
const roles = computed(() => keycloak?.tokenParsed?.realm_access?.roles || []);

onMounted(async () => {
  userMsg.value = (await api.getUserData()).msg
})
</script>
