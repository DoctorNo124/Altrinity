<script setup lang="ts">
import { ref, inject, onMounted } from 'vue'
import { useRoute } from 'vue-router/auto'
import RouteMap from '@/components/RouteMap.vue'
import type Keycloak from 'keycloak-js'
import type { RoutePoint } from '@/models/RoutePoint'

const keycloak = inject<Keycloak>('keycloak')
const { routeId } = useRoute('/VolunteerRoute/[routeId]').params
const points = ref([] as RoutePoint[])
const userId = ref<string>()
const createdAt = ref<string>()

onMounted(async () => {
  const res = await fetch(`${import.meta.env.VITE_API_BASE}/route/${routeId}`, {
    headers: { Authorization: `Bearer ${keycloak?.token}` },
  })
  const data = await res.json()
  userId.value = data.user_id;
  createdAt.value = data.created_at;
  points.value = (data.points_parsed || []) as RoutePoint[];
})
</script>

<template>
    <RouteMap :route-points="points" :user-id="userId!" :created-at="createdAt!"/>
</template>
