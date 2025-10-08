import { Geolocation, type Position } from '@capacitor/geolocation'

const USE_MOCK = true

let mockActive = false
let mockTimeout: ReturnType<typeof setTimeout> | null = null
let lastMockId: string | number | null = null

// Start somewhere nice (San Francisco)
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

// 🚶 Linear path direction — these define the “street”
const DIRECTION_LAT_DELTA = 0.00005  // small drift north/south (~5m per step)
const DIRECTION_LNG_DELTA = -0.00015 // west/east movement (~15m per step)

// light side-to-side drift (GPS noise)
function randomDrift() {
  return (Math.random() - 0.5) * 0.00003 // ±3m sideways drift
}

// random delay to simulate different pause durations
function randomDelay() {
  return 1000 + Math.random() * 7000 // 1s – 8s
}

export async function startWatch(
  callback: (pos: Position) => void
): Promise<string | number | null> {
  if (USE_MOCK) {
    let lat = START_LAT
    let lng = START_LNG
    mockActive = true

    const emitNext = () => {
      if (!mockActive) return

      // 70% chance to move
      const shouldMove = Math.random() > 0.3

      if (shouldMove) {
        // Move roughly in one direction, with some drift
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
          heading: shouldMove ? 270 : null, // roughly westbound
        },
        timestamp: Date.now(),
      }

      callback(fakePos)

      // Next position after random delay
      mockTimeout = setTimeout(emitNext, randomDelay())
    }

    emitNext()
    lastMockId = 'mock'
    return lastMockId
  }

  // Real native geolocation
  const id = await Geolocation.watchPosition(
    { enableHighAccuracy: true },
    (pos, err) => {
      if (err) {
        console.error('Geolocation error:', err)
        return
      }
      if (!pos) return
      if (pos.coords.accuracy > 30) return
      callback(pos)
    }
  )

  lastMockId = id
  return id
}

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
