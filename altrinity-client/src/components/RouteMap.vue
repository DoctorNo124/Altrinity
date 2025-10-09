<script setup lang="ts">
import { ref, onMounted, watch, inject } from 'vue'
import L, { type LatLngExpression } from 'leaflet'
import 'leaflet/dist/leaflet.css'
import type { RoutePoint } from '@/models/RoutePoint'
import type Keycloak from 'keycloak-js'

const keycloak = inject<Keycloak>('keycloak')

const map = ref<L.Map>()
const userFullName = ref<string>('') // 🧍 overlay header text

const props = defineProps<{
  routePoints: RoutePoint[]
  userId: string
  createdAt: string
}>()

/* ---------------------------
   Fetch user full name
--------------------------- */
async function fetchUserFullName(userId: string) {
  try {
    const res = await fetch(`${import.meta.env.VITE_API_BASE}/users/${userId}`, {
      headers: { Authorization: `Bearer ${keycloak?.token}` },
    })
    if (!res.ok) throw new Error(await res.text())
    const data = await res.json()
    userFullName.value = `${data.lastName}, ${data.firstName}`
  } catch (err) {
    console.error('Failed to fetch user name:', err)
    userFullName.value = 'Unknown User'
  }
}

/* ---------------------------
   🗺️ Map setup
--------------------------- */
function initMap() {
  map.value = L.map('map').setView([37.7858, -122.4064], 15)
  L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    crossOrigin: true,
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map.value!)
}

/* ---------------------------
   🎨 Draw heatmap route
--------------------------- */
function drawRoute(routePoints: RoutePoint[]) {
  if (!map.value || routePoints.length < 2) return

  const coords = routePoints.map(p => [p.lat, p.lng] as [number, number])
  const durations = routePoints.map(p => p.duration ?? 0)
  const max = Math.max(...durations)
  const min = Math.min(...durations)
  const range = Math.max(max - min, 1000)

  // Draw color segments
  for (let i = 0; i < coords.length - 1; i++) {
    const d = durations[i] ?? min
    const t = (d - min) / range
    const color = `hsl(${240 - 240 * Math.min(Math.max(t, 0), 1)}, 100%, 55%)`
    L.polyline([coords[i]!, coords[i + 1]!], { color, weight: 5 }).addTo(map.value!)
  }

  const bounds = L.latLngBounds(coords as [number, number][])
  map.value.fitBounds(bounds, { padding: [40, 40] })

  addStartEndMarkers(coords)
  addHeatLegend(min, max)
}

/* ---------------------------
   🏁 Start / End markers
--------------------------- */
function addStartEndMarkers(coords: LatLngExpression[]) {
  if (!map.value || coords.length < 2) return
  const [start, end] = [coords[0]!, coords[coords.length - 1]!]
  const OFFSET = 0.0002
  const startOffset: LatLngExpression = [
    (start as [number, number])[0] + OFFSET,
    (start as [number, number])[1] + OFFSET,
  ]
  const endOffset: LatLngExpression = [
    (end as [number, number])[0] - OFFSET,
    (end as [number, number])[1] - OFFSET,
  ]

  L.polyline([start, startOffset], { color: '#555', weight: 1.5, dashArray: '3, 3' }).addTo(map.value!)
  L.polyline([end, endOffset], { color: '#555', weight: 1.5, dashArray: '3, 3' }).addTo(map.value!)

  const startIcon = L.divIcon({
    className: 'route-label start-label',
    html: '<div class="label-text">Start</div>',
    iconSize: [48, 28],
    iconAnchor: [24, 22],
  })
  const endIcon = L.divIcon({
    className: 'route-label end-label',
    html: '<div class="label-text">End</div>',
    iconSize: [48, 28],
    iconAnchor: [24, 22],
  })
  L.marker(startOffset, { icon: startIcon }).addTo(map.value!)
  L.marker(endOffset, { icon: endIcon }).addTo(map.value!)
}

/* ---------------------------
   🔥 Heat legend
--------------------------- */
function formatDuration(ms: number): string {
  if (ms < 1000) return `${ms} ms`
  const s = ms / 1000
  if (s < 60) return `${s.toFixed(1)} s`
  const m = s / 60
  return `${m.toFixed(1)} min`
}

function addHeatLegend(min: number, max: number) {
  if (!map.value) return
  const existing = document.querySelector('.heat-legend')
  if (existing) existing.remove()

  const legend = new L.Control({ position: 'bottomright' })
  legend.onAdd = () => {
    const div = L.DomUtil.create('div', 'heat-legend leaflet-bar')
    div.innerHTML = `
      <div style="
        background: linear-gradient(to right,
          hsl(240, 100%, 50%),
          hsl(180, 100%, 50%),
          hsl(120, 100%, 50%),
          hsl(60, 100%, 50%),
          hsl(0, 100%, 50%)
        );
        height: 14px;
        width: 180px;
        border-radius: 6px;
        margin-bottom: 6px;
        border: 1px solid rgba(0,0,0,0.2);
      "></div>
      <div style="
        display: flex;
        justify-content: space-between;
        align-items: center;
        font-size: 11px;
        color: #222;
      ">
        <span>${formatDuration(min)}</span>
        <span>Short Stay → Long Stay</span>
        <span>${formatDuration(max)}</span>
      </div>
    `
    return div
  }
  legend.addTo(map.value!)
}

/* ---------------------------
   🚀 Lifecycle & Watch
--------------------------- */
onMounted(() => initMap())

watch(
  () => props.routePoints,
  async newPoints => {
    if (!newPoints?.length) return
    await fetchUserFullName(props.userId)
    drawRoute(newPoints)
  },
  { immediate: true }
)
</script>

<template>
  <div class="map-container">
    <div id="map"></div>

    <!-- 🧾 Overlay header -->
    <div class="map-header">
      <div class="user-name">{{ userFullName }}</div>
      <div class="route-date">
        {{ new Date(props.createdAt).toLocaleString(undefined, {
          dateStyle: 'medium',
          timeStyle: 'short'
        }) }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.map-container {
  position: relative;
  height: 100%;
}

#map {
  width: 100%;
  height: 100%;
  z-index: 0;
  border-radius: 12px;
  overflow: hidden;
}

/* 🌍 Overlay Header */
.map-header {
  position: absolute;
  top: 16px;
  right: 16px;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(4px);
  border-radius: 10px;
  padding: 10px 14px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.25);
  z-index: 9999;
  font-family: system-ui, sans-serif;
}

.user-name {
  font-weight: 600;
  font-size: 16px;
  color: #1e293b;
}

.route-date {
  font-size: 13px;
  color: #475569;
}

</style>

<!-- ✅ Unscoped block so Leaflet-injected icons get styles -->
<style>
.route-label {
  background: transparent;
  border: none;
}

.route-label .label-text {
  background-color: white;
  border: 2px solid #1976d2;
  border-radius: 6px;
  padding: 4px 8px;
  font-size: 13px;
  font-weight: 700;
  color: #1976d2;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
  text-align: center;
  white-space: nowrap;
}

.route-label.end-label .label-text {
  border-color: #e53935;
  color: #e53935;
}

/* Heat legend */
.heat-legend {
  background-color: #fff !important;
  opacity: 1 !important;
  border-radius: 8px;
  border: 1px solid rgba(0, 0, 0, 0.15);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
  padding: 6px 8px;
  z-index: 9999;
  font-family: system-ui, sans-serif;
  font-size: 12px;
  color: #222;
}

.heat-legend div {
  background-clip: padding-box;
}
</style>
