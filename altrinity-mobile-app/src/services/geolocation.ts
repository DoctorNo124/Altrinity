import { Geolocation, type Position } from '@capacitor/geolocation'

const USE_MOCK = true

let mockActive = false
let mockTimeout: ReturnType<typeof setTimeout> | null = null
let lastMockId: string | number | null = null
let lastMockPos: Position | null = null

// 🌍 Start near downtown San Francisco
const START_LAT = 37.7858
const START_LNG = -122.4064

interface MockPosition extends Position {
  coords: {
    latitude: number
    longitude: number
    accuracy: number
    altitudeAccuracy: number | null
    altitude: number | null
    speed: number | null
    heading: number | null
  }
  timestamp: number
}

// 🚶 Linear route pattern parameters
const DIRECTION_LAT_DELTA = 0.00005 // ~5m per step north/south
const DIRECTION_LNG_DELTA = -0.00015 // ~15m per step west/east

function randomDrift() {
  return (Math.random() - 0.5) * 0.00003 // ±3m
}

function randomDelay() {
  return 1000 + Math.random() * 7000 // 1s–8s
}

/**
 * 🔄 Starts a continuous mock or real geolocation watcher.
 */
export async function startWatch(
  callback: (pos: Position) => void
): Promise<string | number | null> {
  if (USE_MOCK) {
    let lat = START_LAT
    let lng = START_LNG
    mockActive = true

    const emitNext = () => {
      if (!mockActive) return

      const shouldMove = Math.random() > 0.3 // sometimes pause
      if (shouldMove) {
        lat += DIRECTION_LAT_DELTA + randomDrift()
        lng += DIRECTION_LNG_DELTA + randomDrift()
      }

      const fakePos: MockPosition = {
        coords: {
          latitude: lat,
          longitude: lng,
          accuracy: 5,
          altitudeAccuracy: null,
          altitude: null,
          speed: shouldMove ? 1 + Math.random() : 0,
          heading: shouldMove ? 270 : null, // roughly westward
        },
        timestamp: Date.now(),
      }

      // 🔁 save latest mock
      lastMockPos = fakePos
      callback(fakePos)

      mockTimeout = setTimeout(emitNext, randomDelay())
    }

    emitNext()
    lastMockId = 'mock'
    return lastMockId
  }

  // 📡 Real GPS mode
  const id = await Geolocation.watchPosition(
    { enableHighAccuracy: true },
    (pos, err) => {
      if (err) {
        console.error('Geolocation error:', err)
        return
      }
      if (!pos || pos.coords.accuracy > 30) return
      lastMockPos = pos
      callback(pos)
    }
  )

  lastMockId = id
  return id
}

/**
 * ✅ Clears active watcher (mock or real)
 */
export async function clearWatch(id: string | number | null): Promise<void> {
  if (!id) return
  if (USE_MOCK && mockTimeout) {
    clearTimeout(mockTimeout)
    mockTimeout = null
    mockActive = false
    console.log('✅ Cleared mock watcher')
  } else {
    await Geolocation.clearWatch({ id: id + '' })
    console.log('✅ Cleared real geolocation watcher')
  }
}

/**
 * 📍 Unified getCurrentPosition — mock-aware
 */
export async function getCurrentPosition(): Promise<Position> {
  if (USE_MOCK) {
    if (lastMockPos) return lastMockPos

    // fallback first fix
    const fake: Position = {
      coords: {
        latitude: START_LAT,
        longitude: START_LNG,
        accuracy: 5,
        altitudeAccuracy: null,
        altitude: null,
        speed: 0,
        heading: null,
      },
      timestamp: Date.now(),
    }
    lastMockPos = fake
    return fake
  }

  return await Geolocation.getCurrentPosition({ enableHighAccuracy: true })
}
