import type { GameState } from '~/types'

interface WebSocketMessage {
  type: string
  data: any
  player_id?: string
}

export function useGameWebSocket(gameId: string, playerId: string) {
  const config = useRuntimeConfig()
  const wsBase = config.public.wsBase
  
  const socket = ref<WebSocket | null>(null)
  const isConnected = ref(false)
  const gameState = ref<GameState | null>(null)
  const messages = ref<WebSocketMessage[]>([])

  const connect = () => {
    if (socket.value) {
      socket.value.close()
    }

    const wsUrl = `${wsBase}/${gameId}/${playerId}`
    socket.value = new WebSocket(wsUrl)

    socket.value.onopen = () => {
      isConnected.value = true
      console.log('WebSocket connected')
    }

    socket.value.onmessage = (event) => {
      try {
        const msg: WebSocketMessage = JSON.parse(event.data)
        messages.value.push(msg)
        
        if (msg.type === 'phase_changed' || msg.type === 'game_started') {
          gameState.value = msg.data as GameState
        }
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e)
      }
    }

    socket.value.onclose = () => {
      isConnected.value = false
      console.log('WebSocket disconnected')
      setTimeout(connect, 3000)
    }

    socket.value.onerror = (error) => {
      console.error('WebSocket error:', error)
    }
  }

  const disconnect = () => {
    if (socket.value) {
      socket.value.close()
      socket.value = null
    }
    isConnected.value = false
  }

  const send = (type: string, data: any) => {
    if (socket.value && isConnected.value) {
      socket.value.send(JSON.stringify({ type, data }))
    }
  }

  onMounted(() => {
    connect()
  })

  onUnmounted(() => {
    disconnect()
  })

  return {
    isConnected,
    gameState,
    messages,
    send,
    connect,
    disconnect
  }
}
