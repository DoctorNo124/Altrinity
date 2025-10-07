;(window as any).cordova = undefined;
;(window as any).Cordova = undefined;
console.log('[Startup] Cordova cleared:', (window as any).cordova, (window as any).Cordova)

import { createApp } from 'vue'


import * as KeycloakModule from 'keycloak-js';
console.log('[DEBUG] Raw keycloak-js module:', KeycloakModule);
import App from './App.vue'
import 'vuetify/styles'
import { registerPlugins } from '@/plugins'
import 'unfonts.css'
import 'leaflet/dist/leaflet.css'
import { useAppUrlOpenListener } from './composables/useAppUrlOpenListener'
import { Capacitor } from '@capacitor/core'

const Keycloak = (KeycloakModule as any).default; // <-- unwrap the real constructor

const keycloak = new Keycloak({
  url: import.meta.env.VITE_KEYCLOAK_URL,
  realm: import.meta.env.VITE_KEYCLOAK_REALM,
  clientId: 'altrinity-mobile',
});

async function initApp() {
  try {
    console.log('[INITAPP] starting...');
    console.log('[Secure Context?]', window.isSecureContext);
    console.log('[Location]', window.location.origin);

    const authenticated = await keycloak.init({
      adapter: 'default',           // 👈  override Cordova mode
      pkceMethod: false,            // avoid secure-context crypto issues
      checkLoginIframe: false,
      enableLogging: true,
    })
    console.log('pee pee poop poop')
    const { registerAppUrlListener } = useAppUrlOpenListener();
    registerAppUrlListener(keycloak);

    const app = createApp(App);
    registerPlugins(app);
    app.provide('keycloak', keycloak);
    app.mount('#app');
    console.log('[INITAPP] app mounted');
  } catch (err) {
    console.error(err);
  }
}

initApp()