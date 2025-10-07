<template>
  <v-app>
    <v-app-bar class="mt-12 d-flex flex-row justify-center bg-blue">
      <v-btn><router-link to="/">Home</router-link></v-btn>
      <v-btn v-if="roles.includes('volunteer')"><router-link to="/VolunteerMap" >Map</router-link></v-btn>
      <v-btn v-if="roles.includes('admin')"><router-link to="/CommandHub">Command Hub</router-link></v-btn>
      <v-btn v-if="roles.includes('admin')"><router-link to="/Admin">Admin</router-link></v-btn>
      <v-spacer></v-spacer>
      <v-btn v-if="roles.length > 0" @click="keycloak?.logout({ redirectUri: 'altrinity://logout' })">Logout</v-btn>
      <v-btn v-else @click="openLogin(keycloak!)">Login</v-btn>
    </v-app-bar>
    <router-view />
  </v-app>
</template>

<script lang="ts" setup>
import type Keycloak from 'keycloak-js'
import { useAuthStore } from '@/stores/auth'
import { storeToRefs } from 'pinia'
import { openLogin } from './utils'
  
const keycloak = inject<Keycloak>('keycloak')

const auth = useAuthStore()
const { roles } = storeToRefs(auth)

onMounted(async () => { 
  await openLogin(keycloak!);
})
</script>
