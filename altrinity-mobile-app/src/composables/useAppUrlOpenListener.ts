import { App, type URLOpenListenerEvent } from '@capacitor/app'
import { useRouter, useRoute } from 'vue-router'
import Keycloak from 'keycloak-js'
import { Browser } from '@capacitor/browser'
import qs from 'qs';
import { useAuthStore } from '@/stores/auth';
import { openLogin } from '@/utils'

export function useAppUrlOpenListener() {
  const router = useRouter()
  const route = useRoute()


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
      const auth = useAuthStore();
      console.log('[Deep link fired]', event.url);

      const url = new URL(event.url);
      if(url.host === 'login') {
        await Browser.close();
        const hash = url.hash.startsWith('#') ? url.hash.slice(1) : url.hash;
        const params = new URLSearchParams(hash);

        const code = params.get('code');
        const state = params.get('state');
        console.log('[Parsed params]', { code, state });

        if (!code) {
          console.warn('[No code param – cannot finish login]');
          return;
        }

        await completeLogin(keycloak, code);
        console.log('[LOGIN SUCCESS]', keycloak.token);
        return;
      }
      if (url.host === 'logout') { 
        console.log('[Deep link logout]')
        keycloak.token = undefined;
        keycloak.refreshToken = undefined;
        keycloak.idToken = undefined;
        keycloak.authenticated = false;
        auth.clear()
        await openLogin(keycloak);
        return;
      }
    });  
  }

  return { registerAppUrlListener }
}