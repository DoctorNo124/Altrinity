import { App, type URLOpenListenerEvent } from '@capacitor/app'
import { useRouter, useRoute } from 'vue-router'
import Keycloak from 'keycloak-js'
import { Browser } from '@capacitor/browser'
import qs from 'qs';
import { useAuthStore } from '@/stores/auth';
import { openLogin } from '@/utils'

export function useAppUrlOpenListener() {
  
  async function completeLogin(kc: Keycloak, code: string) {
    const auth = useAuthStore();
    const tokenUrl =
      `${kc.authServerUrl}realms/${kc.realm}/protocol/openid-connect/token`;

    const body = qs.stringify({
      grant_type: 'authorization_code',
      client_id: kc.clientId,
      code,
      redirect_uri: 'altrinity://login',
    });

    const resp = await fetch(tokenUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body,
    });
    const data = await resp.json();
    console.log('[TOKEN RESPONSE]', data);

    kc.token = data.access_token;
    kc.refreshToken = data.refresh_token;
    kc.idToken = data.id_token;
    kc.authenticated = true;

    auth.setToken(kc.token!, kc.refreshToken)
  }

  function registerAppUrlListener(keycloak: Keycloak) {
    console.log('[Keycloak] Registering appUrlOpen listener')

    App.addListener('appUrlOpen', async (event) => {
      try {
        const auth = useAuthStore();
        console.log('[Deep link fired]', event.url)

        // Use both host and pathname for robust matching
        const url = new URL(event.url)
        console.log('[URL breakdown]', {
          href: url.href,
          protocol: url.protocol,
          host: url.host,
          hostname: url.hostname,
          pathname: url.pathname,
          hash: url.hash,
        })

        const path = url.pathname.replace(/^\//, '') // 'login' or 'logout'

        /* -------------------- LOGIN HANDLER -------------------- */
        if (
          path === 'login' ||
          url.host === 'login' ||
          event.url.startsWith('altrinity://login')
        ) {
          console.log('[Deep link → login detected]')
          await Browser.close()

          // Handle both ? and # fragments safely
          const search = url.search?.startsWith('?') ? url.search.slice(1) : ''
          const hash = event.url.split('#')[1] || ''
          const params = new URLSearchParams(search || hash)

          const code = params.get('code')
          const state = params.get('state')
          console.log('[Parsed params]', { code, state })

          if (!code) {
            console.warn('[No code param – cannot finish login]')
            return
          }

          await completeLogin(keycloak, code)
          console.log('[LOGIN SUCCESS]', keycloak.token)
          return
        }

        /* -------------------- LOGOUT HANDLER -------------------- */
        if (
          path === 'logout' ||
          url.host === 'logout' ||
          event.url.startsWith('altrinity://logout')
        ) {
          console.log('[Deep link → logout detected]')
          keycloak.token = undefined
          keycloak.refreshToken = undefined
          keycloak.idToken = undefined
          keycloak.authenticated = false
          auth.clear()
          await openLogin(keycloak)
          return
        }
      } catch (err) {
        console.error('[Deep link error]', err)
      }
    })
  }
  return { registerAppUrlListener }
}