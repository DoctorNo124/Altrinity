<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Geolocation, type Position } from '@capacitor/geolocation'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { startWatch, clearWatch } from '@/services/geolocation'
// ✅ Fix default marker paths (Vite can't auto-resolve them)
import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png'
import markerIcon from 'leaflet/dist/images/marker-icon.png'
import markerShadow from 'leaflet/dist/images/marker-shadow.png'
import { useAuthStore } from '@/stores/auth'

L.Icon.Default.mergeOptions({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow,
})

const authStore = useAuthStore();
const map = ref<L.Map>()
const routePoints = ref<{ lat: number; lng: number; timestamp: number }[]>([])
const marker = ref<L.Marker | null>(null)
const watchId = ref<string | number | null>(null)
const recording = ref(false)
const routeLayers = ref<L.LayerGroup<L.Polyline<any>>>()
const hasCenteredOnce = ref(false) // prevents re-centering every GPS update


async function initMap() {
  // Initial placeholder position (center on 0,0 until we have a fix)
  map.value = L.map('map').setView([0, 0], 2)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png').addTo(map.value)
  routeLayers.value = L.layerGroup<L.Polyline<any>>().addTo(map.value)

  // Ask for permission and start live tracking
  await Geolocation.requestPermissions()
  startLocationWatcher()
}

async function startLocationWatcher() {
  watchId.value = await startWatch(pos => {
    updateCurrentMarker(pos)
    if (recording.value) recordRoutePoint(pos)
  })
}

function updateCurrentMarker(pos: Position) {
  const { latitude, longitude, accuracy } = pos.coords
  if (!map.value) return

  // Create or move marker
  if (!marker.value) {
    marker.value = L.marker([latitude, longitude]).addTo(map.value).bindPopup('You are here')
  } else {
    marker.value.setLatLng([latitude, longitude])
  }

  // Center map the first time we get a valid GPS fix
  if (!hasCenteredOnce.value) {
    map.value.setView([latitude, longitude], 16) // zoom 16 ~ street level
    hasCenteredOnce.value = true
  }
}

function recordRoutePoint(pos: Position) {
  const { latitude, longitude, accuracy } = pos.coords
  if (accuracy > 30) return
  const timestamp = Date.now()
  routePoints.value.push({ lat: latitude, lng: longitude, timestamp })
  redrawRoute()
}

function redrawRoute() {
  if (!map.value || routePoints.value.length < 2) return
  routeLayers.value?.clearLayers()

  const coords = routePoints.value.map(p => [p.lat, p.lng])
  const durations = calcDurations(routePoints.value)
  const segments = colorSegments(coords, durations)
  segments.forEach(seg => L.polyline(seg.coords, { color: seg.color, weight: 5 }).addTo(routeLayers.value!))
}

function calcDurations(points: any[]) {
  const durations: number[] = []
  for (let i = 0; i < points.length - 1; i++) {
    const timeSpent = points[i + 1].timestamp - points[i].timestamp
    durations.push(timeSpent)
  }
  return durations
}

function colorSegments(coords: any[], durations: number[]) {
  const max = Math.max(...durations)
  const min = Math.min(...durations)
  const segments = []
  for (let i = 0; i < coords.length - 1; i++) {
    const t = (durations[i]! - min) / (max - min || 1)
    const color = `hsl(${240 - 240 * t}, 100%, 50%)` // blue→red gradient
    segments.push({ coords: [coords[i], coords[i + 1]], color })
  }
  return segments
}

function startRecording() {
  if (recording.value) return
  recording.value = true
  routePoints.value = []
}

async function stopRecording() {
  if (!recording.value) return
  recording.value = false
  await uploadRoute()
}

async function uploadRoute() {
  const data = routePoints.value.map(p => ({
    lat: p.lat,
    lng: p.lng,
    timestamp: p.timestamp
  }))

  try {
    const res = await fetch(`${import.meta.env.VITE_API_BASE}/routes`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${authStore.token}`,
      },
      body: JSON.stringify({ user_id: 'current-user-id', route: data })
    })
    if (!res.ok) throw new Error(await res.text())
    alert('Route uploaded successfully!')
  } catch (err) {
    console.error('Upload failed:', err)
    alert('Upload failed — check network or token.')
  }
}

onMounted(() => {
  setTimeout(initMap, 0)
})

onUnmounted(async () => {
  if (watchId.value) await clearWatch(watchId.value)
})
</script>

<template>
  <div id="map-container">
    <div id="map"></div>
    <div class="controls">
      <v-btn color="primary" @click="startRecording" :disabled="recording">
        Start Recording
      </v-btn>
      <v-btn color="red" @click="stopRecording" :disabled="!recording">
        Stop Recording
      </v-btn>
    </div>
  </div>
</template>

<style scoped>
#map-container {
  position: relative;
  height: 100vh;
  display: flex; 
  justify-content: center;
  align-items: center;
}

#map {
  width: 100%;
  height: 100%;
  z-index: 0;
}

.leaflet-container {
  z-index: 0 !important;
}

.controls {
  position: absolute;
  bottom: 17vh;
  z-index: 9999;
  display: flex;
  gap: 16px;
}
</style>