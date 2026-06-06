<template>
  <div class="game-page">
    <ActionBar :is-host="isHost" />
    
    <div class="game-content">
      <div class="left-panel">
        <PlayerPanel />
        <HexInfoPanel />
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
        <BuildPanel />
        <FleetPanel />
        <TechPanel />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
const gameApi = useGameApi()
const gameStore = useGameStore()

const isHost = ref(false)

onMounted(async () => {
  const gameId = route.params.id as string
  const playerId = localStorage.getItem('playerId')
  
  if (playerId) {
    gameStore.setCurrentPlayer(playerId)
  }
  
  try {
    const state = await gameApi.getGame(gameId)
    gameStore.setGameState(state)
  } catch (e) {
    console.error('Failed to load game:', e)
  }
})

function handleHexClick(q: number, r: number) {
  console.log('Hex clicked:', q, r)
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
  width: 300px;
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
</style>
