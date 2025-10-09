<template>
  <v-app>
    <v-app-bar class="d-flex flex-row justify-center bg-blue">
      <v-btn><router-link to="/">Home</router-link></v-btn>
      <v-btn v-if="roles.includes('volunteer')"><router-link to="/VolunteerMap" >Map</router-link></v-btn>
      <!-- <v-btn v-if="roles.includes('admin')"><router-link to="/CommandHub">Command Hub</router-link></v-btn> -->
      <v-btn v-if="roles.includes('admin')"><router-link to="/Admin">Manage Volunteers</router-link></v-btn>
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
  
const keycloak = inject<Keycloak>('keycloak')

const api = useApi()
const userMsg = ref('')
const roles = computed(() => keycloak?.tokenParsed?.realm_access?.roles || []);

onMounted(async () => {
  userMsg.value = (await api.getUserData()).msg
})
</script>
