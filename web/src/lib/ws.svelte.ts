import { browser } from '$app/environment'

export type SensorData = {
  temperature: number
  humidity: number
  timestamp: string
}

export type DailyStats = {
  min_temp: number
  max_temp: number
  min_hum: number
  max_hum: number
}

type WsState = {
  connected: boolean
  latest: SensorData | null
  history: SensorData[]
  stats: DailyStats
}

const MAX_HISTORY = 60 // 60 seconds of data on chart

function createWsStore() {
  let state = $state<WsState>({
    connected: false,
    latest: null,
    history: [],
    stats: {
      min_temp: 0,
      max_temp: 0,
      min_hum: 0,
      max_hum: 0
    }
  })

  let socket: WebSocket | null = null
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null

  async function fetchStats() {
    try {
      const res = await fetch('http://localhost:8080/api/stats')
      const data = await res.json()
      state.stats = data
    } catch (e) {
      console.error('Failed to fetch stats', e)
    }
  }

  function connect() {
    if (!browser) return

    socket = new WebSocket('ws://localhost:8080/ws')

    socket.onopen = () => {
      state.connected = true
      fetchStats()
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
        reconnectTimer = null
      }
    }

    socket.onmessage = (event) => {
      const msg = JSON.parse(event.data)
      if (msg.type === 'history') {
        state.history = msg.data.slice(-MAX_HISTORY)
        if (state.history.length > 0) {
          state.latest = state.history[state.history.length - 1]
        }
      } else if (msg.type === 'live') {
        const data: SensorData = msg.data
        state.latest = data
        state.history = [...state.history.slice(-(MAX_HISTORY - 1)), data]

        // Update real-time stats
        if (state.stats.min_temp === 0 || data.temperature < state.stats.min_temp) state.stats.min_temp = data.temperature
        if (data.temperature > state.stats.max_temp) state.stats.max_temp = data.temperature
        if (state.stats.min_hum === 0 || data.humidity < state.stats.min_hum) state.stats.min_hum = data.humidity
        if (data.humidity > state.stats.max_hum) state.stats.max_hum = data.humidity
      }
    }

    socket.onclose = () => {
      state.connected = false
      // auto-reconnect after 2s
      reconnectTimer = setTimeout(connect, 2000)
    }

    socket.onerror = () => {
      socket?.close()
    }
  }

  connect()

  return {
    get connected() { return state.connected },
    get latest() { return state.latest },
    get history() { return state.history },
    get stats() { return state.stats },
  }
}

export const ws = createWsStore()
