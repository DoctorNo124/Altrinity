<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import 'leaflet.offline'
import markerIcon2x from 'leaflet/dist/images/marker-icon-2x.png'
import markerIcon from 'leaflet/dist/images/marker-icon.png'
import markerShadow from 'leaflet/dist/images/marker-shadow.png'
import { startWatch, clearWatch, getCurrentPosition } from '@/services/geolocation'
import { useAuthStore } from '@/stores/auth'
import { useRouteQueue } from '@/composables/useRouteQueue'
import type { Position } from '@capacitor/geolocation'

L.Icon.Default.mergeOptions({
  iconRetinaUrl: markerIcon2x,
  iconUrl: markerIcon,
  shadowUrl: markerShadow,
})

const auth = useAuthStore()
const { enqueue, offlineCount, flushQueue } = useRouteQueue()

const map = ref<L.Map>()
const routePoints = ref<{ lat: number; lng: number; timestamp: number; duration: number }[]>([])
const marker = ref<L.Marker | null>(null)
const watchId = ref<string | number | null>(null)
const recording = ref(false)
const routeLayers = ref<L.LayerGroup<L.Polyline<any>>>()
const hasCenteredOnce = ref(false)

/* ---------------------------
   🗺️ MAP INITIALIZATION
--------------------------- */
async function initMap() {
  let pos: Position | null = null
  try {
    pos = await getCurrentPosition()
    console.log('📍 Initial position', pos.coords.latitude, pos.coords.longitude)
  } catch (err) {
    console.warn('⚠️ Could not get initial position, defaulting to SF:', err)
  }

  const lat = pos?.coords.latitude ?? 37.7858
  const lng = pos?.coords.longitude ?? -122.4064
  const zoom = 16

  map.value = L.map('map', { zoomControl: false }).setView([lat, lng], zoom)

  const tileLayer = (L as any).tileLayer.offline(
    'https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
    {
      subdomains: 'abc',
      minZoom: 3,
      maxZoom: 19,
      attribution: '&copy; OpenStreetMap contributors',
    }
  )
  tileLayer.addTo(map.value)
  routeLayers.value = L.layerGroup<L.Polyline<any>>().addTo(map.value)

  tileLayer.on('storagesize', (e: any) =>
    console.log('🧱 Tile cache size:', e.storagesize)
  )

  await nextTick()
  map.value.invalidateSize()

  if (pos) {
    updateCurrentMarker(pos)
    hasCenteredOnce.value = true
  }

  await startLocationWatcher()
}

/* ---------------------------
   📡 GEOLOCATION WATCHER
--------------------------- */
async function startLocationWatcher() {
  watchId.value = await startWatch(pos => {
    updateCurrentMarker(pos)
    if (recording.value) recordRoutePoint(pos)
  })
}

function updateCurrentMarker(pos: Position) {
  const { latitude, longitude } = pos.coords
  if (!map.value) return

  if (!marker.value) {
    marker.value = L.marker([latitude, longitude])
      .addTo(map.value)
      .bindPopup('You are here')
  } else {
    marker.value.setLatLng([latitude, longitude])
  }

  if (!hasCenteredOnce.value) {
    map.value.setView([latitude, longitude], 16)
    map.value.invalidateSize()
    hasCenteredOnce.value = true
  }
}

/* ---------------------------
   🎯 RECORDING POINTS
--------------------------- */
function recordRoutePoint(pos: Position) {
  const { latitude, longitude, accuracy } = pos.coords
  if (accuracy > 30) return

  const timestamp = Date.now()
  const last = routePoints.value[routePoints.value.length - 1];
  const duration = last ? timestamp - last.timestamp : 0

  routePoints.value.push({
    lat: latitude,
    lng: longitude,
    timestamp,
    duration,
  })

  console.log(`📍 Recorded point (${latitude}, ${longitude}) duration=${duration}ms`)
  redrawRoute()
}

/* ---------------------------
   🎨 REDRAW HEATMAP ROUTE
--------------------------- */
function redrawRoute() {
  if (!map.value || routePoints.value.length < 2) return
  routeLayers.value?.clearLayers()

  const coords = routePoints.value.map(p => [p.lat, p.lng] as [number,number])
  const durations = routePoints.value.map(p => p.duration)
  const max = Math.max(...durations)
  const min = Math.min(...durations)

  for (let i = 0; i < coords.length - 1; i++) {
    const d = durations[i]
    const t = (d! - min) / (max - min || 1)
    const color = `hsl(${240 - 240 * t}, 100%, 50%)` // blue → red
    L.polyline([coords[i]!, coords[i + 1]!], { color, weight: 5 }).addTo(routeLayers.value!)
  }
}

/* ---------------------------
   ⏯️ START / STOP RECORDING
--------------------------- */
function startRecording() {
  if (recording.value) return
  recording.value = true
  routePoints.value = []
  console.log('🎬 Recording started')
}

async function stopRecording() {
  if (!recording.value) return
  recording.value = false

  console.log('⏹️ Stopping recording...')
  if (watchId.value) {
    await clearWatch(watchId.value)
    watchId.value = null
  }

  const snapshot = [...routePoints.value]
  console.log(`📦 Final route length: ${snapshot.length} points`)

  if (snapshot.length < 2) {
    alert('⚠️ Not enough data points recorded.')
    return
  }

  // ensure last point has 0 duration for clean JSON structure
  snapshot[snapshot.length - 1]!.duration = 0
  await uploadRoute(snapshot)
}

/* ---------------------------
   ☁️ UPLOAD OR QUEUE ROUTE
--------------------------- */
async function uploadRoute(snapshot: { lat: number; lng: number; timestamp: number; duration: number }[]) {
  try {
    const res = await fetch(`${import.meta.env.VITE_API_BASE}/routes`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${auth.token}`,
      },
      body: JSON.stringify({ route: snapshot }),
    })
    if (!res.ok) throw new Error(await res.text())
    console.log('✅ Uploaded route successfully.')
    alert('✅ Route uploaded successfully!')
    await flushQueue()
  } catch (err) {
    console.error('Upload failed, adding to offline queue:', err)
    await enqueue({ route: snapshot, createdAt: Date.now() })
    alert('⚠️ Offline — route saved locally!')
  }
}

/* ---------------------------
   🧹 LIFECYCLE HOOKS
--------------------------- */
onMounted(() => setTimeout(initMap, 0))
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

    <div class="offline-indicator" v-if="offlineCount > 0">
      <v-chip color="orange" text-color="white" label>
        {{ offlineCount }} pending route<span v-if="offlineCount > 1">s</span>
      </v-chip>
    </div>
  </div>
</template>

<style scoped>
#map-container {
  position: relative;
  height: 100%;
  padding-top: 10px;
}

#map {
  width: 100%;
  height: 100%;
  z-index: 0;
  background: #c8d6e5;
}

.controls {
  position: absolute;
  bottom: 5vh;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 10px;
  z-index: 9999;
}

.offline-indicator {
  position: absolute;
  top: 10px;
  right: 10px;
  z-index: 9999;
}
</style>
