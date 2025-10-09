<script setup lang="ts">
import { ref, inject, onMounted } from 'vue'
import { useRoute } from 'vue-router/auto'
import RouteMap from '@/components/RouteMap.vue'
import type Keycloak from 'keycloak-js'
import type { RoutePoint } from '@/models/RoutePoint'

const keycloak = inject<Keycloak>('keycloak')
const { routeId } = useRoute('/VolunteerRoute/[routeId]').params
const points = ref([] as RoutePoint[])

onMounted(async () => {
  const res = await fetch(`${import.meta.env.VITE_API_BASE}/route/${routeId}`, {
    headers: { Authorization: `Bearer ${keycloak?.token}` },
  })
  const data = await res.json()
  points.value = (data.points_parsed || []) as RoutePoint[];
})
</script>

<template>
  <v-container>
    <h2>Route Details</h2>
    <RouteMap :route-points="points" />
  </v-container>
</template>
