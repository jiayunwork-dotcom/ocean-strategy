import type { GameState } from '~/types'

interface WebSocketMessage {
  type: string
  data: any
  player_id?: string
}

export function useGameWebSocket(gameId: string, playerId: string) {
  const config = useRuntimeConfig()
  const wsBase = config.public.wsBase
  const gameStore = useGameStore()

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
          gameStore.setGameState(msg.data as GameState)
        }

        if (msg.type === 'treaty_proposed' || 
            msg.type === 'proposal_responded' || 
            msg.type === 'treaty_broken' ||
            msg.type === 'battle_occurred' ||
            msg.type === 'ship_moved' ||
            msg.type === 'player_joined' ||
            msg.type === 'order_placed' ||
            msg.type === 'order_cancelled' ||
            msg.type === 'auction_created' ||
            msg.type === 'bid_placed' ||
            msg.type === 'futures_created' ||
            msg.type === 'futures_accepted' ||
            msg.type === 'futures_cancelled' ||
            msg.type === 'futures_margin_added') {
          refreshGameState()
        }

        if (msg.type === 'futures_margin_call' ||
            msg.type === 'futures_liquidated' ||
            msg.type === 'futures_settled') {
          refreshGameState()
          if (msg.player_id === playerId) {
            showNotification(msg)
          }
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

  const refreshGameState = async () => {
    const { getGame } = useGameApi()
    try {
      const state = await getGame(gameId)
      gameState.value = state
      gameStore.setGameState(state)
    } catch (e) {
      console.error('Failed to refresh game state:', e)
    }
  }

  const showNotification = (msg: WebSocketMessage) => {
    let title = ''
    let message = ''
    let type = 'info'

    switch (msg.type) {
      case 'futures_margin_call':
        title = '⚠️ 保证金追缴通知'
        message = '您的期货合约保证金不足，请在下一决策阶段追加保证金，否则将被强制平仓！'
        type = 'warning'
        break
      case 'futures_liquidated':
        title = '💥 强制平仓'
        message = '您的期货合约因保证金不足已被强制平仓！'
        type = 'error'
        break
      case 'futures_settled':
        title = '📊 期货结算'
        const pnl = msg.data?.pnl || 0
        const pnlText = pnl >= 0 ? `盈利 ${pnl}` : `亏损 ${Math.abs(pnl)}`
        message = `您的期货合约已完成结算，${pnlText} 金币`
        type = pnl >= 0 ? 'success' : 'error'
        break
    }

    if (title && message) {
      alert(`${title}\n\n${message}`)
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
    disconnect,
    refreshGameState
  }
}
