<template>
  <v-container fluid>
    <h2>Command Hub</h2>
    <div id="map" style="height: 80vh;"></div>
  </v-container>
</template>

<script lang="ts" setup>
import { onMounted, onBeforeUnmount, ref, inject } from 'vue'
import L, { Map as LeafletMap } from 'leaflet'
import 'leaflet.heat'
import Keycloak from 'keycloak-js'

interface VolunteerPosition {
  volunteerId: string
  fullName: string
  lat: number
  lng: number
  weight?: number
}

const keycloak = inject<Keycloak>('keycloak')
const map = ref<LeafletMap>()
const markers = new Map<string, L.Marker>()
const polylines = new Map<string, L.Polyline>()
const ws = ref<WebSocket | null>(null)
let heatLayer: any

onMounted(async () => {
  map.value = L.map('map').setView([40.0, -83.0], 12)
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map.value)

  // Heatmap layer
  heatLayer = (L as any).heatLayer([], { radius: 25, maxZoom: 15 }).addTo(map.value)

  if (keycloak) {
    await loadExistingRoutes()
    connectWebSocket()
  }
})

onBeforeUnmount(() => {
  if (ws.value) ws.value.close()
})

// --- Load full route history from API ---
async function loadExistingRoutes() {
  try {
    const resp = await fetch(`${import.meta.env.VITE_API_BASE}/route/positions`, {
      headers: { Authorization: `Bearer ${keycloak?.token}` },
    })
    const allPositions: VolunteerPosition[] = await resp.json()
    if (!map.value || !allPositions) return

    // Group by volunteerId
    const grouped: Record<string, VolunteerPosition[]> = {}
    for (const pos of allPositions) {
      if (!grouped[pos.volunteerId]) grouped[pos.volunteerId] = []
      grouped[pos.volunteerId]?.push(pos)
    }

    // Draw each volunteer’s full route + heatmap
    for (const [volunteerId, positions] of Object.entries(grouped)) {
      const latlngs = positions.map((p) => [p.lat, p.lng] as [number, number])
      console.log(latlngs);
      const polyline = L.polyline(latlngs, { color: 'blue', weight: 4 }).addTo(map.value!)
      polylines.set(volunteerId, polyline)
      positions.forEach((p) => heatLayer.addLatLng([p.lat, p.lng, p.weight || 0.5]))
      updateMarker(positions[positions.length - 1]!)
    }

    // Fit map to all current data
    const allLatLngs = Object.values(grouped).flat().map((p) => [p.lat, p.lng] as [number, number])
    if (allLatLngs.length > 0) map.value.fitBounds(L.latLngBounds(allLatLngs))
  } catch (err) {
    console.error('Failed to load routes', err)
  }
}

// --- WebSocket live updates ---
function connectWebSocket() {
  ws.value = new WebSocket(`${import.meta.env.VITE_WS_BASE}/ws/positions?token=${keycloak?.token}`)

  ws.value.onmessage = (msg) => {
    const pos = JSON.parse(msg.data) as VolunteerPosition
    updateMarker(pos)
    extendRoute(pos)
    if(!heatLayer || !map.value) return
    heatLayer.addLatLng([pos.lat, pos.lng, pos.weight || 0.5])
  }
}

// --- Marker updates ---
function updateMarker(vol: VolunteerPosition) {
  if (!map.value) return
  let marker = markers.get(vol.volunteerId)
  const latLng = [vol.lat, vol.lng] as [number, number]
  if (!marker) {
    marker = L.marker(latLng).addTo(map.value)
    marker.bindTooltip(vol.fullName, { permanent: false, direction: 'top', offset: L.point(0, -10) })
    markers.set(vol.volunteerId, marker)
  } else {
    marker.setLatLng(latLng)
  }
  marker.getTooltip()?.setContent(vol.fullName)
}

// --- Extend existing route line ---
function extendRoute(pos: VolunteerPosition) {
    let line = polylines.get(pos.volunteerId)
    if (!line) {
        line = L.polyline([], { color: 'blue', weight: 4 }).addTo(map.value!)
        polylines.set(pos.volunteerId, line)
    }
    const latLng = [pos.lat, pos.lng] as [number, number]
    line.addLatLng(latLng)
}
</script>