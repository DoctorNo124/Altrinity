<script setup lang="ts">
import { onMounted, ref } from 'vue'
import L, { Map as LeafletMap, Marker} from 'leaflet'

const map = ref<LeafletMap>()
let userMarker: Marker | null = null

onMounted(() => {
  map.value = L.map('map').setView([39.95, -83.0], 13)

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; OpenStreetMap contributors'
  }).addTo(map.value)

  // Watch the user's location
  if (navigator.geolocation) {
    navigator.geolocation.watchPosition(
      pos => {
        const { latitude, longitude } = pos.coords
        const coords: [number, number] = [latitude, longitude]

        // Add or update marker
        if (!userMarker) {
          userMarker = L.marker(coords).addTo(map.value!)
          userMarker.bindPopup("You are here")
        } else {
          userMarker.setLatLng(coords)
        }

        // Optionally re-center map
        map.value?.setView(coords, 15)
      },
      err => {
        console.error("Geolocation error:", err)
      },
      { enableHighAccuracy: true }
    )
  } else {
    console.warn("Geolocation not supported")
  }
})
</script>

<template>
  <div id="map" style="height: 100vh; width: 100%;"></div>
</template>

<style lang="css">
</style>