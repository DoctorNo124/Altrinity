import { createApp } from 'vue'
import Keycloak from 'keycloak-js'
import App from './App.vue'
// Vuetify
import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import { registerPlugins } from '@/plugins'
// Styles
import 'unfonts.css'
import { setKeycloak } from './router'
import 'leaflet/dist/leaflet.css'

// Define a Keycloak type (keycloak-js has partial typing support)
let keycloak = new Keycloak({
  url: 'https://auth.altrinitytech.com',   // Keycloak base URL
  realm: 'altrinity',                // your realm
  clientId: 'vue-frontend',       // frontend client
})

const vuetify = createVuetify({
  components,
  directives,
})

keycloak.init(
  { 
    onLoad: 'login-required',
  }).then((authenticated) => {
  setKeycloak(keycloak)
  if (!authenticated) {
    window.location.reload()
  } else {
    console.log('Authenticated ✅')

    const app = createApp(App);
    registerPlugins(app)

    // provide keycloak instance to all components
    app.provide('keycloak', keycloak)

    app.mount('#app')
  }

  // Optionally refresh token periodically
  setInterval(() => {
    keycloak.updateToken(60).catch(() => {
      console.error('Failed to refresh token')
    })
  }, 6000)
}).catch((err) => {
  console.error('Keycloak init error:', err)
})
