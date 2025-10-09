<script setup lang="ts">
import { ref, onMounted, inject } from 'vue'
import { useRoute, useRouter } from 'vue-router/auto'
import { useDisplay } from 'vuetify'
import type Keycloak from 'keycloak-js'

const keycloak = inject<Keycloak>('keycloak')
const router = useRouter()
const { mdAndUp } = useDisplay()

// 🧭 Haversine helper for distance
function haversine(lat1: number, lon1: number, lat2: number, lon2: number): number {
  const R = 6371e3
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

interface User {
  id: string
  username: string
  email: string
  firstName: string
  lastName: string
  roles: string[]
}

const { id } = useRoute('/UserRoutes/[id]').params
const routes = ref<RouteRecord[]>([])
const user = ref<User | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

const headers = [
  { title: 'Created', value: 'created_at', sortable: true },
  { title: 'Points', value: 'points_count', sortable: true },
  { title: 'Distance (m)', value: 'total_distance', sortable: true },
  { title: 'Duration (min)', value: 'total_duration', sortable: true },
  { title: 'Actions', value: 'actions', sortable: false },
]

/* -----------------------------------------
   🧠 Fetch the user’s profile from Keycloak
----------------------------------------- */
async function fetchUser() {
  try {
    const res = await fetch(`${import.meta.env.VITE_API_BASE}/users/${id}`, {
      headers: { Authorization: `Bearer ${keycloak?.token}` },
    })
    if (!res.ok) throw new Error(await res.text())
    user.value = await res.json()
  } catch (err) {
    console.error('Failed to fetch user info:', err)
  }
}

/* -----------------------------------------
   📦 Fetch all routes for this user
----------------------------------------- */
async function fetchRoutes() {
  loading.value = true
  error.value = null
  try {
    const res = await fetch(`${import.meta.env.VITE_API_BASE}/routes/${id}`, {
      headers: { Authorization: `Bearer ${keycloak?.token}` },
    })
    if (!res.ok) throw new Error(await res.text())
    const data = await res.json()

    routes.value = data.map((r: any) => {
      let points: RoutePoint[] = []
      try {
        points = JSON.parse(r.points)
      } catch {
        console.warn('Failed to parse points for route', r.id)
      }

      let totalDist = 0
      let totalDur = 0
      for (let i = 0; i < points.length - 1; i++) {
        const a = points[i]
        const b = points[i + 1]
        if (a && b) {
          totalDist += haversine(a.lat, a.lng, b.lat, b.lng)
          totalDur += b.timestamp - a.timestamp
        }
      }

      return {
        id: r.id,
        created_at: r.created_at,
        points_count: points.length,
        total_distance: Math.round(totalDist),
        total_duration: Number((totalDur / 60000).toFixed(1)), // minutes
      }
    })
  } catch (err: any) {
    console.error('Failed to fetch user routes:', err)
    error.value = err.message
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await Promise.all([fetchUser(), fetchRoutes()])
})

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
      <v-toolbar-title :class="{ 'v-toolbar-title-mobile': !mdAndUp }">
        {{ user ? `Routes for ${user.lastName}, ${user.firstName}` : 'User Routes' }}
      </v-toolbar-title>
      <v-spacer v-if="mdAndUp"></v-spacer>
      <v-btn color="primary" @click="$router.back()">← Back</v-btn>
    </v-toolbar>

    <!-- 💻 Desktop / Tablet -->
    <v-card v-if="mdAndUp" class="mt-4">
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

    <!-- 📱 Mobile -->
    <v-row v-else dense class="mt-2">
      <v-col
        v-for="(r, i) in routes"
        :key="r.id"
        cols="12"
        class="py-1"
      >
        <v-card
          :class="i % 2 === 0 ? 'bg-light' : 'bg-lighter'"
          class="pa-3 elevation-1"
        >
          <div class="text-subtitle-1 mb-1">
            📅 {{ new Date(r.created_at).toLocaleString() }}
          </div>
          <div class="text-body-2">
            <b>Points:</b> {{ r.points_count }}<br />
            <b>Distance:</b> {{ r.total_distance.toLocaleString() }} m<br />
            <b>Duration:</b> {{ r.total_duration }} min
          </div>
          <v-btn
            color="primary"
            block
            class="mt-3"
            @click="goToVolunteerRoute(r.id)"
          >
            View on Map
          </v-btn>
        </v-card>
      </v-col>
    </v-row>

    <v-alert v-if="error" type="error" class="mt-4">
      {{ error }}
    </v-alert>
  </v-container>
</template>

<style scoped>
.v-toolbar-title {
  font-weight: 600;
}

.v-toolbar-title-mobile { 
    font-size: 15px;
}

.bg-light {
  background-color: #f8f9fa;
}

.bg-lighter {
  background-color: #ffffff;
}

.v-card {
  border-radius: 12px;
  transition: transform 0.1s ease;
}

.v-card:hover {
  transform: scale(1.01);
}
</style>
