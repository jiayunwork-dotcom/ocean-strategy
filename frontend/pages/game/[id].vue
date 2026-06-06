<template>
  <div class="game-page">
    <ActionBar :is-host="isHost" />
    
    <div class="game-content">
      <div class="left-panel">
        <PlayerPanel />
        <HexInfoPanel />
        <DiplomacyPanel />
      </div>
      
      <div class="map-area">
        <HexMap
          v-if="gameState?.game"
          :radius="gameState.game.map_radius"
          :hex-size="35"
          @hex-click="handleHexClick"
        />
        <div v-else class="loading-map">
          <p class="text-gray-400">加载地图中...</p>
        </div>
      </div>
      
      <div class="right-panel">
        <div class="panel-tabs">
          <button
            v-for="tab in rightPanelTabs"
            :key="tab.id"
            @click="activeRightPanel = tab.id"
            :class="['panel-tab-btn', { active: activeRightPanel === tab.id }]"
          >
            {{ tab.name }}
          </button>
        </div>
        
        <div v-if="activeRightPanel === 'build'">
          <BuildPanel />
          <FleetPanel />
          <TechPanel />
        </div>
        
        <div v-if="activeRightPanel === 'market'">
          <MarketPanel
            :game-state="gameState"
            :current-player-id="currentPlayerId"
            @place-order="handlePlaceOrder"
            @cancel-order="handleCancelOrder"
            @create-futures="handleCreateFutures"
            @accept-futures="handleAcceptFutures"
            @cancel-futures="handleCancelFutures"
            @add-margin="handleAddMargin"
          />
          <AuctionPanel
            :game-state="gameState"
            :current-player-id="currentPlayerId"
            @create-auction="handleCreateAuction"
            @place-bid="handlePlaceBid"
          />
        </div>
      </div>
    </div>
    
    <ProposalModal v-if="showProposalModal" @close="closeProposalModal" />
  </div>
</template>

<script setup lang="ts">
import type { OrderType, ResourceType, AuctionItemType, FuturesContract } from '~/types'

const route = useRoute()
const gameApi = useGameApi()
const gameStore = useGameStore()

const isHost = ref(false)
const activeRightPanel = ref('build')
const showProposalModal = ref(false)

const gameState = computed(() => gameStore.gameState)
const currentPlayerId = computed(() => gameStore.currentPlayerId)

const rightPanelTabs = [
  { id: 'build', name: '建造' },
  { id: 'market', name: '市场' },
]

onMounted(async () => {
  const gameId = route.params.id as string
  const playerId = localStorage.getItem('playerId')
  
  if (playerId) {
    gameStore.setCurrentPlayer(playerId)
  }
  
  try {
    const state = await gameApi.getGame(gameId)
    gameStore.setGameState(state)
    
    if (playerId) {
      useGameWebSocket(gameId, playerId)
    }
  } catch (e) {
    console.error('Failed to load game:', e)
  }
})

function handleHexClick(q: number, r: number) {
  gameStore.selectHex(q, r)
}

async function handlePlaceOrder(order: { type: OrderType; resource: ResourceType; quantity: number; price: number }) {
  const gameId = route.params.id as string
  const playerId = gameStore.currentPlayerId
  if (!playerId) return
  
  try {
    await gameApi.placeOrder(gameId, playerId, order.type, order.resource, order.quantity, order.price)
    const state = await gameApi.getGame(gameId)
    gameStore.setGameState(state)
  } catch (e) {
    console.error('Failed to place order:', e)
    alert('挂单失败')
  }
}

async function handleCancelOrder(orderId: string) {
  const gameId = route.params.id as string
  const playerId = gameStore.currentPlayerId
  if (!playerId) return
  
  try {
    await gameApi.cancelOrder(gameId, playerId, orderId)
    const state = await gameApi.getGame(gameId)
    gameStore.setGameState(state)
  } catch (e) {
    console.error('Failed to cancel order:', e)
    alert('取消订单失败')
  }
}

async function handleCreateAuction(data: { itemType: AuctionItemType; itemId: string; startingPrice: number }) {
  const gameId = route.params.id as string
  const playerId = gameStore.currentPlayerId
  if (!playerId) return
  
  try {
    await gameApi.createAuction(gameId, playerId, data.itemType, data.itemId, data.startingPrice)
    const state = await gameApi.getGame(gameId)
    gameStore.setGameState(state)
  } catch (e) {
    console.error('Failed to create auction:', e)
    alert('发起拍卖失败')
  }
}

async function handlePlaceBid(auctionId: string, amount: number) {
  const gameId = route.params.id as string
  const playerId = gameStore.currentPlayerId
  if (!playerId) return
  
  try {
    await gameApi.placeBid(gameId, playerId, auctionId, amount)
    const state = await gameApi.getGame(gameId)
    gameStore.setGameState(state)
  } catch (e) {
    console.error('Failed to place bid:', e)
    alert('出价失败')
  }
}

async function handleCreateFutures(contract: { resource: ResourceType; quantity: number; contract_price: number; delivery_turn: number }) {
  const gameId = route.params.id as string
  const playerId = gameStore.currentPlayerId
  if (!playerId) return
  
  try {
    await gameApi.createFuturesContract(gameId, playerId, contract.resource, contract.quantity, contract.contract_price, contract.delivery_turn)
    const state = await gameApi.getGame(gameId)
    gameStore.setGameState(state)
  } catch (e: any) {
    console.error('Failed to create futures contract:', e)
    alert(e.data?.error || '创建期货合约失败')
  }
}

async function handleAcceptFutures(contractId: string) {
  const gameId = route.params.id as string
  const playerId = gameStore.currentPlayerId
  if (!playerId) return
  
  try {
    await gameApi.acceptFuturesContract(gameId, playerId, contractId)
    const state = await gameApi.getGame(gameId)
    gameStore.setGameState(state)
  } catch (e: any) {
    console.error('Failed to accept futures contract:', e)
    alert(e.data?.error || '接受期货合约失败')
  }
}

async function handleCancelFutures(contractId: string) {
  const gameId = route.params.id as string
  const playerId = gameStore.currentPlayerId
  if (!playerId) return
  
  try {
    await gameApi.cancelFuturesContract(gameId, playerId, contractId)
    const state = await gameApi.getGame(gameId)
    gameStore.setGameState(state)
  } catch (e: any) {
    console.error('Failed to cancel futures contract:', e)
    alert(e.data?.error || '取消期货合约失败')
  }
}

async function handleAddMargin(contractId: string, amount: number) {
  const gameId = route.params.id as string
  const playerId = gameStore.currentPlayerId
  if (!playerId) return
  
  try {
    await gameApi.addFuturesMargin(gameId, playerId, contractId, amount)
    const state = await gameApi.getGame(gameId)
    gameStore.setGameState(state)
  } catch (e: any) {
    console.error('Failed to add margin:', e)
    alert(e.data?.error || '追加保证金失败')
  }
}

function closeProposalModal() {
  gameStore.closeProposalModal()
}
</script>

<style scoped>
.game-page {
  width: 100vw;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #0a1628;
  overflow: hidden;
}

.game-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.left-panel {
  width: 280px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow-y: auto;
}

.map-area {
  flex: 1;
  padding: 12px;
  overflow: auto;
}

.right-panel {
  width: 320px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow-y: auto;
}

.loading-map {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.panel-tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 4px;
}

.panel-tab-btn {
  flex: 1;
  padding: 8px 12px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  color: #94a3b8;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.panel-tab-btn:hover {
  background: #334155;
}

.panel-tab-btn.active {
  background: #0ea5e9;
  border-color: #0ea5e9;
  color: white;
}
</style>
