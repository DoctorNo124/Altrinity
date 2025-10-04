import { inject } from 'vue'
import type Keycloak from 'keycloak-js'

export function useApi() {
    const keycloak = inject<Keycloak>('keycloak')
    async function fetchWithAuth(url: string, options: RequestInit = {}) {
    const res = await fetch(url, {
            ...options,
            headers: {
            ...(options.headers || {}),
            Authorization: `Bearer ${keycloak?.token}`,
            },
        })

        if (!res.ok) {
            // Force logout if unauthorized/forbidden
            if (res.status === 401 || res.status === 403) {
            alert('Session expired or not authorized.')
            keycloak?.logout({ redirectUri: window.location.origin })
            }
            throw new Error(`API call failed: ${res.status}`)
        }

        return res.json()
    }

    function getUserRoles(): string[] {
        if (!keycloak?.tokenParsed) return []
        const realmAccess = keycloak.tokenParsed['realm_access']
        return realmAccess?.roles || []
    }

    async function getUserData() : Promise<{ msg : string }> {
        return (await fetchWithAuth('https://api.altrinitytech.com/api/user'));
    }

    async function getAdminData() : Promise<{ msg : string }>{
        return (await fetchWithAuth('https://api.altrinitytech.com/api/admin'));
    }

    return { getUserData, getAdminData, getUserRoles }
}
