import type Keycloak from 'keycloak-js'
import { Browser } from '@capacitor/browser'

export async function openLogin(keycloak: Keycloak) {
  if(keycloak) { 
    const loginUrl = await keycloak.createLoginUrl({
      redirectUri: 'altrinity://login',
    })
    console.log('[LOGIN URL]', loginUrl)
    await Browser.open({ url: loginUrl })
  }
}
