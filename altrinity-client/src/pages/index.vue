<script setup lang="ts">
import type Keycloak from 'keycloak-js'

const keycloak = inject<Keycloak>('keycloak')

const roles = computed(() => keycloak?.tokenParsed?.realm_access?.roles || []);
</script>

<template>
  <v-container class="py-10 text-center home-container">
    <v-row justify="center">
      <v-col cols="12" md="8" lg="6">
        <v-card elevation="3" class="pa-6 rounded-xl">
          <v-card-title class="text-h5 font-weight-bold text-primary mb-4">
            Welcome to Altrinity
          </v-card-title>

          <v-card-text class="text-body-1">
            <p>
              <strong>Altrinity</strong> is a volunteer coordination and field tracking platform
              designed to empower local outreach teams. It helps administrators manage volunteers,
              review their routes, and collect location-based data securely — all in real-time.
            </p>

            <p>
              Volunteers can record their walking routes, even offline, and automatically upload them
              when connectivity is restored. Administrators can review heat-mapped paths to analyze
              coverage and effectiveness across assigned regions.
            </p>

            <v-divider class="my-4" />

            <!-- Role-based content -->
            <div v-if="roles.includes('admin')">
              <p>
                You are logged in as an <strong>Administrator</strong>.
                You can manage users, approve pending volunteers, and review collected route data.
              </p>
              <!-- <v-btn color="primary" class="mt-3" to="/CommandHub">
                Go to Command Hub
              </v-btn> -->
              <v-btn variant="outlined" class="mt-3 ml-2" to="/Admin">
                Manage Volunteers
              </v-btn>
            </div>

            <div v-else-if="roles.includes('volunteer')">
              <p>
                You are logged in as a <strong>Volunteer</strong>.
                Use the map to record your walking route. You can record offline — your route will
                be uploaded when you're back online.
              </p>
              <v-btn color="primary" class="mt-3" to="/VolunteerMap">
                Open Volunteer Map
              </v-btn>
            </div>

            <div v-else-if="roles.includes('pending')">
              <p>
                Your volunteer account is currently <strong>pending approval</strong>.
                Once approved by an administrator, you’ll gain access to route recording features.
              </p>
            </div>

            <div v-else>
              <p>
                Please log in to continue. Once authenticated, you’ll see your volunteer dashboard
                or admin tools depending on your role.
              </p>
              <v-btn color="primary" class="mt-3" to="/">
                Log In
              </v-btn>
            </div>
          </v-card-text>

          <v-divider class="my-4" />

          <v-card-subtitle class="text-caption text-grey-darken-1">
            Altrinity © {{ new Date().getFullYear() }} — Empowering volunteer coordination with
            transparency and data-driven insights.
          </v-card-subtitle>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<style scoped>
.home-container {
  background: linear-gradient(to bottom right, #f5f9ff, #eef3ff);
  min-height: calc(100vh - 64px);
}

.v-card {
  background-color: white;
}

.text-primary {
  color: #1976d2;
}
</style>
