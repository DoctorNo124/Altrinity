/**
 * router/index.ts
 *
 * Automatic routes for `./src/pages/*.vue`
 */

// Composables
import { createRouter, createWebHistory } from 'vue-router'
import { setupLayouts } from 'virtual:generated-layouts'
import { routes } from 'vue-router/auto-routes'
import type { RouteLocationRaw } from 'vue-router'
import type Keycloak from 'keycloak-js'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: setupLayouts(routes),
})

let keycloak: Keycloak | null = null
export function setKeycloak(kc: Keycloak) {
  keycloak = kc
}

function getRoles(): string[] {
  if (!keycloak?.tokenParsed) return []
  const ra = keycloak.tokenParsed['realm_access']
  return ra?.roles || []
}


// Workaround for https://github.com/vitejs/vite/issues/11804
router.onError((err, to) => {
  if (err?.message?.includes?.('Failed to fetch dynamically imported module')) {
    if (localStorage.getItem('vuetify:dynamic-reload')) {
      console.error('Dynamic import error, reloading page did not fix it', err)
    } else {
      console.log('Reloading page to fix dynamic import error')
      localStorage.setItem('vuetify:dynamic-reload', 'true')
      location.assign(to.fullPath)
    }
  } else {
    console.error(err)
  }
})

router.isReady().then(() => {
  localStorage.removeItem('vuetify:dynamic-reload')
})

router.beforeEach((to, from, next) => {
  const roles = getRoles();

  if (roles.includes('pending') && to.path !== '/PendingApproval') {
    // force pending users to "Pending Approval" page
    return next('/PendingApproval');
  }

  next()
})


export default router
