<template>
  <div class="fleet-panel">
    <div class="panel-header">
      <h3 class="text-lg font-bold text-white">船队调度</h3>
    </div>
    
    <div v-if="playerShips.length > 0" class="ship-list space-y-2">
      <div
        v-for="ship in playerShips"
        :key="ship.id"
        class="ship-item"
        :class="{ 'selected': selectedShip?.id === ship.id }"
        @click="selectShip(ship)"
      >
        <div class="ship-icon-wrapper">
          <span class="ship-icon">{{ getShipIcon(ship.type) }}</span>
        </div>
        <div class="ship-info">
          <div class="ship-name">{{ SHIP_NAMES[ship.type] }}</div>
          <div class="ship-stats">
            <span class="stat">❤️ {{ ship.health }}/{{ ship.max_health }}</span>
            <span class="stat">⛽ {{ ship.fuel }}/{{ ship.max_fuel }}</span>
          </div>
          <div class="ship-position text-xs text-gray-400">
            位置: ({{ ship.hex_q }}, {{ ship.hex_r }})
          </div>
        </div>
      </div>
    </div>
    
    <div v-else class="no-ships">
      <p class="text-gray-400 text-sm text-center py-6">暂无船只</p>
      <p class="text-gray-500 text-xs text-center">在港口建造船只</p>
    </div>
    
    <div v-if="selectedShip" class="ship-actions mt-4 pt-4 border-t border-gray-700">
      <h4 class="text-sm font-semibold text-gray-300 mb-3">操作</h4>
      <div class="action-buttons space-y-2">
        <button
          v-if="selectedShip.type === 'explorer'"
          class="action-btn primary w-full"
          @click="exploreHex"
        >
          🔍 探测当前海域
        </button>
        <button
          class="action-btn secondary w-full"
          @click="showMoveDialog = true"
        >
          🗺️ 移动船只
        </button>
      </div>
    </div>
    
    <Teleport to="body">
      <div v-if="showMoveDialog" class="modal-overlay" @click.self="showMoveDialog = false">
        <div class="modal-content">
          <h3 class="text-lg font-bold text-white mb-4">移动船只</h3>
          <p class="text-gray-300 text-sm mb-4">在地图上点击目标位置</p>
          <div class="flex gap-2 justify-end">
            <button class="action-btn secondary" @click="showMoveDialog = false">取消</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { SHIP_NAMES } from '~/types'
import type { Ship } from '~/types'

const gameStore = useGameStore()
const route = useRoute()
const gameApi = useGameApi()

const playerShips = computed(() => gameStore.playerShips)
const selectedShip = computed(() => gameStore.selectedShip)

const showMoveDialog = ref(false)

function getShipIcon(type: string): string {
  const icons: Record<string, string> = {
    explorer: '🔍',
    constructor: '🔧',
    transport: '📦',
    escort: '⚔️'
  }
  return icons[type] || '🚢'
}

function selectShip(ship: Ship) {
  gameStore.selectShip(ship)
}

async function exploreHex() {
  if (!selectedShip.value) return
  try {
    await gameApi.explore(route.params.id as string, selectedShip.value.id)
    const state = await gameApi.getGame(route.params.id as string)
    gameStore.setGameState(state)
  } catch (e) {
    console.error('Explore failed:', e)
  }
}
</script>

<style scoped>
.fleet-panel {
  background: rgba(15, 23, 42, 0.95);
  border-radius: 12px;
  padding: 16px;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.panel-header {
  margin-bottom: 12px;
}

.ship-item {
  display: flex;
  gap: 12px;
  padding: 10px;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  border: 2px solid transparent;
}

.ship-item:hover {
  background: rgba(30, 41, 59, 0.9);
}

.ship-item.selected {
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.2);
}

.ship-icon-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: rgba(59, 130, 246, 0.2);
  border-radius: 8px;
}

.ship-icon {
  font-size: 20px;
}

.ship-info {
  flex: 1;
}

.ship-name {
  font-weight: 600;
  color: white;
  font-size: 14px;
}

.ship-stats {
  display: flex;
  gap: 12px;
  margin-top: 4px;
}

.ship-stats .stat {
  font-size: 11px;
  color: #94a3b8;
}

.action-btn {
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  border: none;
}

.action-btn.primary {
  background: #3b82f6;
  color: white;
}

.action-btn.primary:hover {
  background: #2563eb;
}

.action-btn.secondary {
  background: rgba(71, 85, 105, 0.8);
  color: white;
}

.action-btn.secondary:hover {
  background: rgba(71, 85, 105, 1);
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: #1e293b;
  padding: 24px;
  border-radius: 12px;
  min-width: 300px;
}

.no-ships {
  min-height: 100px;
}
</style>
