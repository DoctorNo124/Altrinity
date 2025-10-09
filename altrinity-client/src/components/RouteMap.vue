<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import L, { type LatLngExpression } from 'leaflet'
import 'leaflet/dist/leaflet.css'
import type { RoutePoint } from '@/models/RoutePoint';


const map = ref<L.Map>()
const props = defineProps<{
  routePoints: RoutePoint[]
}>()

/* ---------------------------
   🗺️ Map setup
--------------------------- */
function initMap() {
  map.value = L.map('map').setView([37.7858, -122.4064], 15)
  L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
    maxZoom: 19,
    crossOrigin: true,
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map.value!);
}

/* ---------------------------
   🎨 Draw heatmap route
--------------------------- */
function drawRoute(routePoints: RoutePoint[]) {
  if (!map.value || routePoints.length < 2) { 
    return
  } 

  const coords = routePoints.map(p => [p.lat, p.lng] as [number, number])
  const durations = routePoints.map(p => p.duration ?? 0)
  const max = Math.max(...durations)
  const min = Math.min(...durations)
  const range = Math.max(max - min, 1000)

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
   🏁 Start / End labels with offset + connector
--------------------------- */
function addStartEndMarkers(coords: LatLngExpression[]) {
  if (!map.value || coords.length < 2) return

  const [start, end] = [coords[0]!, coords[coords.length - 1]!]

  // Offset markers by small lat/lng deltas
  const OFFSET = 0.0002 // ≈ 20 meters
  const startOffset: LatLngExpression = [
    (start as [number, number])[0] + OFFSET,
    (start as [number, number])[1] + OFFSET,
  ]
  const endOffset: LatLngExpression = [
    (end as [number, number])[0] - OFFSET,
    (end as [number, number])[1] - OFFSET,
  ]

  // Connector lines (thin and gray)
  L.polyline([start, startOffset], {
    color: '#555',
    weight: 1.5,
    dashArray: '3, 3',
  }).addTo(map.value!)
  L.polyline([end, endOffset], {
    color: '#555',
    weight: 1.5,
    dashArray: '3, 3',
  }).addTo(map.value!)

  // DivIcons with styled labels (slightly offset)
  const startIcon = L.divIcon({
    className: 'route-label start-label',
    html: '<div class="label-text">Start</div>',
    iconSize: [40, 20],
    iconAnchor: [20, 20],
  })

  const endIcon = L.divIcon({
    className: 'route-label end-label',
    html: '<div class="label-text">End</div>',
    iconSize: [40, 20],
    iconAnchor: [20, 20],
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
      "></div>
      <div style="
        display: flex;
        justify-content: space-between;
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
   🚀 Lifecycle
--------------------------- */
onMounted(() => {
  initMap()
})

watch(() => props.routePoints, (newRoutePoints) => { 
    drawRoute(newRoutePoints);
});
</script>

<template>
  <v-container>
    <h2>Volunteer Route</h2>
    <div id="map" style="height: 80vh; border-radius: 12px; overflow: hidden;"></div>
  </v-container>
</template>

<style>
.heat-legend {
  background: white;
  padding: 6px 8px;
  border-radius: 8px;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.25);
  font-family: system-ui, sans-serif;
}

/* Label styles for Start/End markers */
.route-label {
  background: transparent;
  border: none;
}
.label-text {
  background-color: white;
  border: 2px solid #1976d2;
  border-radius: 6px;
  padding: 2px 6px;
  font-size: 12px;
  font-weight: 600;
  color: #1976d2;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
}
.end-label .label-text {
  border-color: #e53935;
  color: #e53935;
}
</style>
