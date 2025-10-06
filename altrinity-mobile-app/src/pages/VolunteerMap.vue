<template>
  <v-container fluid>
    <h2>Your Canvassing Map</h2>
    <div id="map" style="height: 80vh;"></div>
    <v-btn @click="startRecording" class="mr-2">Start Recording Route</v-btn>
    <v-btn @click="stopRecording">Stop Recvording Route</v-btn>
  </v-container>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import L, { Map as LeafletMap} from 'leaflet'
import Keycloak from 'keycloak-js'
import { Geolocation, type Position } from '@capacitor/geolocation'

const keycloak = inject<Keycloak>('keycloak');
const map = ref<LeafletMap>();
const marker = ref<L.Marker | null>(null);

let watchId: string | null = null

async function startRecording() {
  // Request permissions first
  const perm = await Geolocation.requestPermissions()
  if (perm.location === 'denied') {
    alert('Location access denied. Please enable it in settings.')
    return
  }

  // Start watching
  watchId = await Geolocation.watchPosition(
    { enableHighAccuracy: true, timeout: 10000, maximumAge: 5000 },
    async (pos: Position | null, err?: any) => {
      if (err) {
        console.error('Geolocation error:', err)
        return
      }
      if (!pos) return

      const { latitude: lat, longitude: lng, accuracy } = pos.coords
      if (accuracy > 30) return

      // Update map marker
      if (marker.value) marker.value.setLatLng([lat, lng])
      else marker.value = L.marker([lat, lng])
        .addTo(map.value!)
        .bindPopup('You are here')

      // Send location to backend
      try {
        if (keycloak) {
          await fetch(`${import.meta.env.VITE_API_BASE}/positions`, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${keycloak.token}`,
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({ lat, lng }),
          })
        }
      } catch (e) {
        console.error('Failed to send position', e)
      }
    }
  )
}

async function stopRecording() {
  if (watchId) {
    Geolocation.clearWatch({ id: watchId })
    watchId = null
    try {
      if(keycloak && keycloak.tokenParsed) { 
        await fetch(`${import.meta.env.VITE_API_BASE}/route/complete`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${keycloak.token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ userId: keycloak.tokenParsed["user_id"] }),
        });
      }
    } catch (e) {
      console.error('Failed to complete route', e);
    }
  }
}


onMounted(async () => {
  map.value = L.map('map').setView([40.0, -83.0], 13);
  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors',
  }).addTo(map.value);
});
</script>