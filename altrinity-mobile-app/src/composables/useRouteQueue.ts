import { ref, watch } from 'vue'
import { Network } from '@capacitor/network'
import { Preferences } from '@capacitor/preferences'
import { useAuthStore } from '@/stores/auth'

const offlineCount = ref(0)
const queue = ref<any[]>([])
let initialized = false

async function loadQueue() {
  const stored = await Preferences.get({ key: 'routeQueue' })
  queue.value = stored.value ? JSON.parse(stored.value) : []
  offlineCount.value = queue.value.length
}

async function saveQueue() {
  await Preferences.set({ key: 'routeQueue', value: JSON.stringify(queue.value) })
  offlineCount.value = queue.value.length
}

async function enqueue(item: any) {
  queue.value.push(item)
  await saveQueue()
}

async function flushQueue() {
  const auth = useAuthStore()
  const baseUrl = import.meta.env.VITE_API_BASE

  if (queue.value.length === 0) return
  if (!auth.token) return console.warn('⚠️ No token for flushQueue')

  console.log(`📤 Flushing ${queue.value.length} routes...`)

  const remaining = []

  for (const item of queue.value) {
    try {
      const res = await fetch(`${baseUrl}/routes`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify(item),
      })
      if (!res.ok) throw new Error(await res.text())
      console.log('✅ Route sent successfully')
    } catch (err) {
      console.error('❌ Failed to send route:', err)
      remaining.push(item)
    }
  }

  queue.value = remaining
  await saveQueue()
}

function initQueueSync() {
  if (initialized) return
  initialized = true
  loadQueue()

  Network.addListener('networkStatusChange', async (status) => {
    if (status.connected) {
      console.log('📶 Online — auto-flushing routes')
      await flushQueue()
    }
  })
}

export function useRouteQueue() {
  initQueueSync()
  return { enqueue, flushQueue, offlineCount }
}
