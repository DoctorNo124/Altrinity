<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router/auto'
import type Keycloak from 'keycloak-js'

const keycloak = inject<Keycloak>('keycloak')
const router = useRouter();

// Small haversine helper to approximate distance between coordinates
function haversine(lat1: number, lon1: number, lat2: number, lon2: number): number {
  const R = 6371e3 // meters
  const toRad = (deg: number) => (deg * Math.PI) / 180
  const φ1 = toRad(lat1)
  const φ2 = toRad(lat2)
  const Δφ = toRad(lat2 - lat1)
  const Δλ = toRad(lon2 - lon1)
  const a = Math.sin(Δφ / 2) ** 2 + Math.cos(φ1) * Math.cos(φ2) * Math.sin(Δλ / 2) ** 2
  return 2 * R * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
}

interface RoutePoint {
  lat: number
  lng: number
  duration: number
  timestamp: number
}

interface RouteRecord {
  id: string
  created_at: string
  points_count: number
  total_duration: number
  total_distance: number
}

const { id } = useRoute('/UserRoutes/[id]').params

const routes = ref<RouteRecord[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const headers = [
  { title: 'Created', value: 'created_at', sortable: true },
  { title: 'Points', value: 'points_count', sortable: true },
  { title: 'Distance (m)', value: 'total_distance', sortable: true },
  { title: 'Duration (min)', value: 'total_duration', sortable: true },
  { title: 'Actions', value: 'actions', sortable: false },
]

async function fetchRoutes() {
  loading.value = true
  error.value = null
  try {
    const res = await fetch(`${import.meta.env.VITE_API_BASE}/routes/${id}`, {
      headers: {
        Authorization: `Bearer ${keycloak!.token}`,
      },
    })
    if (!res.ok) throw new Error(await res.text())
    const data = await res.json()

    // Transform raw backend routes
    routes.value = data.map((r: any) => {
      let points: RoutePoint[] = []
      try {
        points = JSON.parse(r.points)
      } catch (e) {
        console.warn('Failed to parse route points for', r.id)
      }

      // Calculate totals
      let totalDist = 0
      let totalDur = 0
      for (let i = 0; i < points.length - 1; i++) {
        const a = points[i]
        const b = points[i + 1]
        if(a && b) { 
            totalDist += haversine(a.lat, a.lng, b.lat, b.lng)
            totalDur += b.timestamp - a.timestamp
        }
      }

      return {
        id: r.id,
        created_at: r.created_at,
        points_count: points.length,
        total_distance: Math.round(totalDist),
        total_duration: (totalDur / 60000).toFixed(1), // minutes
      }
    })
  } catch (err: any) {
    console.error('Failed to fetch user routes:', err)
    error.value = err.message
  } finally {
    loading.value = false
  }
}
onMounted(fetchRoutes)

function goToVolunteerRoute(routeId: string) { 
    router.push({
        name: '/VolunteerRoute/[routeId]',
        params: { routeId },
    })
}

</script>

<template>
  <v-container fluid>
    <v-toolbar flat color="transparent">
      <v-toolbar-title>User Routes</v-toolbar-title>
      <v-spacer></v-spacer>
      <v-btn color="primary" @click="$router.back()">← Back</v-btn>
    </v-toolbar>

    <v-card class="mt-4">
      <v-data-table
        :headers="headers"
        :items="routes"
        :loading="loading"
        :items-per-page="10"
        class="elevation-1"
      >
        <template #item.created_at="{ item }">
          {{ new Date(item.created_at).toLocaleString() }}
        </template>

        <template #item.total_distance="{ item }">
          {{ item.total_distance.toLocaleString() }}
        </template>

        <template #item.total_duration="{ item }">
          {{ item.total_duration }}
        </template>

        <template #item.actions="{ item }">
          <v-btn color="blue" variant="text" @click="goToVolunteerRoute(item.id)">
            View on Map
          </v-btn>
        </template>

        <template #no-data>
          <v-alert type="info">No routes found for this user.</v-alert>
        </template>
      </v-data-table>
    </v-card>

    <v-alert v-if="error" type="error" class="mt-4">
      {{ error }}
    </v-alert>
  </v-container>
</template>

<style scoped>
.v-toolbar-title {
  font-weight: 600;
}
</style>
