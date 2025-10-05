<template>
  <v-container fluid>
    <h2>Your Canvassing Map</h2>
    <div id="map" style="height: 80vh;"></div>
  </v-container>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import L, { Map as LeafletMap} from 'leaflet'
import Keycloak from 'keycloak-js'

const keycloak = inject<Keycloak>('keycloak');
const map = ref<LeafletMap>();
const marker = ref<L.Marker | null>(null);

onMounted(async () => {
  map.value = L.map('map').setView([40.0, -83.0], 13);
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map.value);

  if (!navigator.geolocation) {
    alert('Geolocation is not supported by your browser');
    return;
  }

  // Watch the user’s location and send updates to backend
  navigator.geolocation.watchPosition(async (pos) => {
    const lat = pos.coords.latitude;
    const lng = pos.coords.longitude;

    // Update map marker
    if (marker.value) marker.value.setLatLng([lat, lng]);
    else marker.value = L.marker([lat, lng]).addTo(map.value!).bindPopup('You are here');

    // Send location to Go API
    try {
      if(keycloak) { 
        await fetch(`${import.meta.env.VITE_API_BASE}/positions`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${keycloak.token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ lat, lng }),
        });
      }
    } catch (e) {
      console.error('Failed to send position', e);
    }
  }, console.error, { enableHighAccuracy: true });
});
</script>