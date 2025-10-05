<template>
  <v-container fluid>
    <h2>Command Hub</h2>
    <div id="map" style="height: 80vh;"></div>
  </v-container>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import L, { Map as LeafletMap} from 'leaflet'
import Keycloak from 'keycloak-js'

interface VolunteerPosition {
  volunteerId: string
  fullName: string
  lat: number
  lng: number
}
const keycloak = inject<Keycloak>('keycloak');
const map = ref<LeafletMap>();
const markers = new Map<string, L.Marker>();
const ws = ref<WebSocket | null>(null)

onMounted(async () => {
  map.value = L.map('map').setView([40.0, -83.0], 12);
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map.value);

  if(keycloak) { 
        // Load initial positions from the API
        const resp = await fetch(`${import.meta.env.VITE_API_BASE}/positions`, {
            headers: { 'Authorization': `Bearer ${keycloak.token}` },
        });
        const initial = await resp.json();
        initial.forEach((pos: VolunteerPosition) => updateMarker(pos));

        // Connect WebSocket for live updates
        ws.value = new WebSocket(`${import.meta.env.VITE_WS_BASE}/ws/positions?token=${keycloak.token}`);

        ws.value.onmessage = (msg) => {
            const data = JSON.parse(msg.data) as VolunteerPosition;
            updateMarker(data);
        };
    }
});

onBeforeUnmount(() => {
  if (ws.value) ws.value.close()
})

// --- Core logic ---
function updateMarker(vol: VolunteerPosition) {
  if (!map.value) return

  let marker = markers.get(vol.volunteerId)
  const latLng = [vol.lat, vol.lng] as [number, number]
  if (!marker) {
    // Create new marker
    marker = L.marker(latLng).addTo(map.value)
    marker.bindTooltip(vol.fullName, {
      permanent: false,
      direction: 'top',
      offset: L.point(0, -10),
    })
    markers.set(vol.volunteerId, marker)
  } else {
    // Update existing marker position
    marker.setLatLng(latLng)
  }

  // Optional: update tooltip text in case name changed
  marker.getTooltip()?.setContent(vol.fullName)
}
</script>