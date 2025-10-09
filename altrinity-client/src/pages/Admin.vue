<script setup lang="ts">
import { ref, onMounted, inject } from 'vue'
import { useRouter } from 'vue-router/auto'
import { useDisplay } from 'vuetify'
import type Keycloak from 'keycloak-js'

const keycloak = inject<Keycloak>('keycloak')
const router = useRouter()
const { mdAndUp } = useDisplay()

interface User {
  id: string
  username: string
  email: string
  roles: string[]
  firstName: string
  lastName: string
}

const users = ref<User[]>([])
const loading = ref(false)

async function fetchUsers() {
  loading.value = true
  try {
    const res = await fetch(`${import.meta.env.VITE_API_BASE}/users`, {
      headers: { Authorization: `Bearer ${keycloak?.token}` },
    })
    const data = await res.json()
    users.value = data.filter((u: User) => !u.roles.includes('admin'))
  } catch (err) {
    console.error('Failed to fetch users:', err)
  } finally {
    loading.value = false
  }
}

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
  router.push({ name: '/VolunteerRoute/[userId]', params: { userId: user.id } })
}

function goToUserRoutes(user: User) {
  router.push({ name: '/UserRoutes/[id]', params: { id: user.id } })
}

onMounted(fetchUsers)
</script>

<template>
  <v-container class="pa-4">
    <v-progress-circular
      v-if="loading"
      indeterminate
      color="primary"
      class="mx-auto my-8 d-block"
      size="48"
    />

    <!-- 💻 Desktop / Tablet View -->
    <v-toolbar-title class="mb-2">
      Volunteers
    </v-toolbar-title>
    <v-data-table
      v-if="mdAndUp"
      :items="users"
      :loading="loading"
      :headers="[
        { title: 'Last Name', key: 'lastName' },
        { title: 'First Name', key: 'firstName' },
        { title: 'Username', key: 'username' },
        { title: 'Email', key: 'email' },
        { title: 'Actions', key: 'actions', sortable: false },
      ]"
      class="elevation-1"
    >
      <template #item.actions="{ item }">
        <v-btn
          v-if="!item.roles.includes('volunteer')"
          class="mr-2"
          color="primary"
          @click="approveUser(item)"
        >
          Approve User
        </v-btn>
        <v-btn class="mr-2" color="secondary" @click="goToVolunteerRoute(item)">
          Latest Route
        </v-btn>
        <v-btn color="secondary" variant="tonal" @click="goToUserRoutes(item)">
          All Routes
        </v-btn>
      </template>
    </v-data-table>

    <!-- 📱 Mobile View (Card List) -->
    <v-row v-else dense>
      <v-col
        v-for="(user, i) in users"
        :key="user.id"
        cols="12"
        class="py-1"
      >
        <v-card
          :class="i % 2 === 0 ? 'bg-light' : 'bg-lighter'"
          class="pa-3 elevation-1"
        >
          <v-row align="center">
            <v-col cols="12" sm="6">
              <div class="text-h6">
                {{ user.lastName }}, {{ user.firstName }}
              </div>
              <div class="text-body-2 text-medium-emphasis">
                {{ user.email }}
              </div>
              <div class="text-caption text-disabled mt-1">
                @{{ user.username }}
              </div>
            </v-col>

            <v-col
              cols="12"
              sm="6"
              class="d-flex flex-wrap justify-end ga-2 mt-2 mt-sm-0"
            >
              <v-btn
                v-if="!user.roles.includes('volunteer')"
                color="primary"
                size="small"
                @click="approveUser(user)"
              >
                Approve
              </v-btn>
              <v-btn
                color="secondary"
                size="small"
                @click="goToVolunteerRoute(user)"
              >
                Latest Route
              </v-btn>
              <v-btn
                color="secondary"
                size="small"
                variant="tonal"
                @click="goToUserRoutes(user)"
              >
                All Routes
              </v-btn>
            </v-col>
          </v-row>
        </v-card>
      </v-col>
    </v-row>

    <div v-if="!loading && users.length === 0" class="text-center mt-8">
      <v-icon icon="mdi-account-off" size="36" class="mb-2" />
      <div class="text-body-2 text-medium-emphasis">No volunteers found.</div>
    </div>
  </v-container>
</template>

<style scoped>
.bg-light {
  background-color: #f8f9fa;
}

.bg-lighter {
  background-color: #ffffff;
}

.v-card {
  border-radius: 12px;
  transition: transform 0.1s ease;
}

.v-card:hover {
  transform: scale(1.01);
}
</style>
