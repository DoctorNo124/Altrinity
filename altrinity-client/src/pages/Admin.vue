<script setup lang="ts">
import { ref, onMounted, inject } from 'vue'
import { useRouter } from 'vue-router/auto'
import type Keycloak from 'keycloak-js'

const keycloak = inject<Keycloak>('keycloak')
const router = useRouter()

interface User {
  id: string
  username: string
  email: string
  roles: string[]
}

const users = ref<User[]>([])
const loading = ref(false)

async function fetchUsers() {
  loading.value = true
  try {
    const res = await fetch(`${import.meta.env.VITE_API_BASE}/users`, {
      headers: {
        Authorization: `Bearer ${keycloak?.token}`,
      },
    })
    const data = await res.json()
    // ✅ Filter to show only volunteers
    users.value = data.filter((u: User) => !u.roles.includes('admin'))
  } catch (err) {
    console.error('Failed to fetch users:', err)
  } finally {
    loading.value = false
  }
}

// Approve user: pending -> volunteer
async function approveUser(user: User) {
  await fetch(`${import.meta.env.VITE_API_BASE}/${user.id}/role`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${keycloak?.token}`,
    },
    body: JSON.stringify({ role: 'volunteer' }),
  })
  await fetchUsers()
}

function goToVolunteerRoute(user: User) {
  router.push({
    name: '/VolunteerRoute/[userId]',
    params: { userId: user.id },
  })
}

function goToUserRoutes(user: User) {
  router.push({
    name: '/UserRoutes/[id]',
    params: { id: user.id },
  })
}



onMounted(fetchUsers)
</script>

<template>
  <v-container>
    <v-data-table
      :items="users"
      :loading="loading"
      :headers="[
        { title: 'Username', key: 'username' },
        { title: 'Email', key: 'email' },
        { title: 'Actions', key: 'actions', sortable: false }
      ]"
    >
      <template #item.actions="{ item }">
        <v-btn v-if="!item.roles.includes('volunteer')" class="mr-2" color="primary" @click="approveUser(item)">
          Approve User
        </v-btn>
        <v-btn class="mr-2" color="primary" @click="goToVolunteerRoute(item)">
          View Latest Route
        </v-btn>
        <v-btn color="primary" @click="goToUserRoutes(item)">
          View All Routes
        </v-btn>
      </template>
    </v-data-table>
  </v-container>
</template>
