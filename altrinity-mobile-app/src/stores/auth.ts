import { ref, computed } from 'vue'
import { defineStore } from 'pinia'

export const useAuthStore = defineStore('auth', () => {
  // 🔐 State
  const token = ref<string | null>(null)
  const refreshToken = ref<string | null>(null)
  const tokenParsed = ref<any | null>(null)
  const authenticated = ref(false)

  // 🧠 Getters
  const roles = computed(() => tokenParsed.value?.realm_access?.roles || [])
  const username = computed(() => tokenParsed.value?.preferred_username || null)
  const userId = computed(() => tokenParsed.value?.sub ?? '')

  // ⚙️ Actions
  function setToken(newToken: string, newRefreshToken?: string) {
    token.value = newToken
    refreshToken.value = newRefreshToken || null
    try {
      const payload = newToken.split('.')[1]
      if(payload) {
        tokenParsed.value = JSON.parse(
            atob(payload.replace(/-/g, '+').replace(/_/g, '/'))
        )
      }
    } catch (err) {
      console.error('[decode failed]', err)
      tokenParsed.value = null
    }
    authenticated.value = true
  }

  function clear() {
    token.value = null
    refreshToken.value = null
    tokenParsed.value = null
    authenticated.value = false
  }

  return {
    // state
    token,
    refreshToken,
    tokenParsed,
    authenticated,
    // getters
    roles,
    username,
    userId,
    // actions
    setToken,
    clear,
  }
})
