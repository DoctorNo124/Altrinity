<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type Keycloak from 'keycloak-js'

const keycloak = inject<Keycloak>('keycloak')
interface User {
  id: string
  username: string
  email: string
  roles: string[]
}

const users = ref<User[]>([])
const loading = ref(false)

// Fetch users from Go API
async function fetchUsers() {
  loading.value = true
  try {
    const res = await fetch(`${import.meta.env.VITE_API_BASE}/users`, {
      headers: {
        Authorization: `Bearer ${keycloak?.token}`,
      },
    })
    users.value = await res.json()
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

onMounted(fetchUsers)
</script>

<template>
  <v-data-table
    :items="users"
    :loading="loading"
    :headers="[
      { title: 'Username', key: 'username' },
      { title: 'Email', key: 'email' },
      { title: 'Roles', key: 'roles' },
      { title: 'Actions', key: 'actions', sortable: false }
    ]"
  >
    <template #item.roles="{ item }">
      {{ item.roles.join(', ') }}
    </template>

    <template #item.actions="{ item }">
      <v-btn
        v-if="item.roles.includes('pending')"
        color="primary"
        @click="approveUser(item)"
      >
        Approve
      </v-btn>
    </template>
  </v-data-table>
</template>