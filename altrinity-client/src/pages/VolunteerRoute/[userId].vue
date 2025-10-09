<script setup lang="ts">
import { ref, inject, onMounted } from 'vue'
import { useRoute } from 'vue-router/auto'
import RouteMap from '@/components/RouteMap.vue'
import type Keycloak from 'keycloak-js'
import type { RoutePoint } from '@/models/RoutePoint'

const keycloak = inject<Keycloak>('keycloak')
const { userId } = useRoute('/VolunteerRoute/[userId]').params
const points = ref([] as RoutePoint[])
const createdAt = ref<string>()

onMounted(async () => {
  const res = await fetch(`${import.meta.env.VITE_API_BASE}/latest-route/${userId}`, {
    headers: { Authorization: `Bearer ${keycloak?.token}` },
  })
  const data = await res.json()
  createdAt.value = data.created_at;
  points.value = (data.points_parsed || []) as RoutePoint[];
})
</script>

<template>
  <RouteMap :route-points="points" :user-id="userId!" :created-at="createdAt!"/>
</template>
