<template>
  <div class="player-panel">
    <div class="panel-header">
      <h3 class="text-lg font-bold text-white mb-3">玩家信息</h3>
    </div>
    
    <div v-if="currentPlayer" class="player-info">
      <div class="player-name-row">
        <div class="player-color-dot" :style="{ backgroundColor: currentPlayer.color }"></div>
        <span class="player-name">{{ currentPlayer.name }}</span>
      </div>
      
      <div class="stat-grid">
        <div class="stat-item">
          <span class="stat-icon">💰</span>
          <div class="stat-content">
            <span class="stat-label">资金</span>
            <span class="stat-value">{{ formatNumber(currentPlayer.money) }}</span>
          </div>
        </div>
        
        <div class="stat-item">
          <span class="stat-icon">⭐</span>
          <div class="stat-content">
            <span class="stat-label">声誉</span>
            <span class="stat-value">{{ currentPlayer.reputation }}</span>
          </div>
        </div>
        
        <div class="stat-item">
          <span class="stat-icon">🏗️</span>
          <div class="stat-content">
            <span class="stat-label">设施</span>
            <span class="stat-value">{{ playerFacilities.length }}</span>
          </div>
        </div>
        
        <div class="stat-item">
          <span class="stat-icon">🚢</span>
          <div class="stat-content">
            <span class="stat-label">船只</span>
            <span class="stat-value">{{ playerShips.length }}</span>
          </div>
        </div>
      </div>
    </div>
    
    <div v-else class="no-player">
      <p class="text-gray-400 text-sm">等待加入游戏...</p>
    </div>
    
    <div v-if="gameState?.game" class="game-info mt-4 pt-4 border-t border-gray-700">
      <h4 class="text-sm font-semibold text-gray-300 mb-2">游戏状态</h4>
      <div class="space-y-1 text-sm">
        <div class="flex justify-between">
          <span class="text-gray-400">回合</span>
          <span class="text-white">{{ gameState.game.current_turn }} / {{ gameState.game.max_turns }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-gray-400">阶段</span>
          <span class="text-emerald-400">{{ PHASE_NAMES[gameState.game.phase] }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-gray-400">状态</span>
          <span :class="gameStatusClass">{{ gameState.game.status }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { PHASE_NAMES } from '~/types'

const gameStore = useGameStore()

const currentPlayer = computed(() => gameStore.currentPlayer)
const playerFacilities = computed(() => gameStore.playerFacilities)
const playerShips = computed(() => gameStore.playerShips)
const gameState = computed(() => gameStore.gameState)

const gameStatusClass = computed(() => {
  const status = gameState.value?.game?.status
  if (status === 'playing') return 'text-green-400'
  if (status === 'finished') return 'text-red-400'
  return 'text-yellow-400'
})

function formatNumber(num: number): string {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return num.toString()
}
</script>

<style scoped>
.player-panel {
  background: rgba(15, 23, 42, 0.95);
  border-radius: 12px;
  padding: 16px;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.panel-header {
  margin-bottom: 12px;
}

.player-name-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.player-color-dot {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.3);
}

.player-name {
  font-size: 18px;
  font-weight: 600;
  color: white;
}

.stat-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 10px;
  background: rgba(30, 41, 59, 0.8);
  padding: 10px;
  border-radius: 8px;
}

.stat-icon {
  font-size: 20px;
}

.stat-content {
  display: flex;
  flex-direction: column;
}

.stat-label {
  font-size: 11px;
  color: #94a3b8;
}

.stat-value {
  font-size: 14px;
  font-weight: 600;
  color: white;
}

.no-player {
  text-align: center;
  padding: 20px 0;
}
</style>
