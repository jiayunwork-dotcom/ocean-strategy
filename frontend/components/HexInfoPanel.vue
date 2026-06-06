<template>
  <div class="hex-info-panel">
    <div class="panel-header">
      <h3 class="text-lg font-bold text-white">海域信息</h3>
    </div>
    
    <div v-if="selectedHexData" class="hex-info">
      <div class="hex-coords text-sm text-gray-400 mb-3">
        坐标: ({{ selectedHexData.q }}, {{ selectedHexData.r }})
      </div>
      
      <div class="terrain-info mb-4">
        <span class="terrain-badge" :style="{ backgroundColor: TERRAIN_COLORS[selectedHexData.terrain] }">
          {{ TERRAIN_NAMES[selectedHexData.terrain] }}
        </span>
      </div>
      
      <div v-if="selectedHexData.resources && Object.keys(selectedHexData.resources).length > 0" class="resources-section mb-4">
        <h4 class="text-sm font-semibold text-gray-300 mb-2">资源储量</h4>
        <div class="resource-list space-y-1">
          <div v-for="(amount, resource) in selectedHexData.resources" :key="resource" class="resource-item">
            <span class="resource-name">{{ RESOURCE_NAMES[resource as ResourceType] }}</span>
            <span class="resource-amount">{{ amount }}</span>
          </div>
        </div>
      </div>
      
      <div class="env-stats space-y-2 text-sm">
        <div class="flex justify-between">
          <span class="text-gray-400">生态健康</span>
          <div class="health-bar">
            <div class="health-fill" :style="{ width: selectedHexData.ecological_health + '%', backgroundColor: getHealthColor(selectedHexData.ecological_health) }"></div>
          </div>
        </div>
        <div class="flex justify-between">
          <span class="text-gray-400">污染程度</span>
          <div class="health-bar">
            <div class="health-fill" :style="{ width: selectedHexData.pollution + '%', backgroundColor: getPollutionColor(selectedHexData.pollution) }"></div>
          </div>
        </div>
        <div class="flex justify-between" v-if="selectedHexData.has_current">
          <span class="text-gray-400">洋流方向</span>
          <span class="text-blue-400">➤ {{ selectedHexData.current_dir }}</span>
        </div>
      </div>
      
      <div v-if="selectedHexData.is_eez" class="eez-info mt-4 p-2 bg-blue-900/30 rounded">
        <span class="text-blue-300 text-sm">📌 专属经济区</span>
      </div>
    </div>
    
    <div v-else class="no-selection">
      <p class="text-gray-400 text-sm text-center py-8">点击地图上的六角格查看详情</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { ResourceType } from '~/types'
import { TERRAIN_COLORS, TERRAIN_NAMES, RESOURCE_NAMES } from '~/types'

const gameStore = useGameStore()

const selectedHexData = computed(() => {
  if (!gameStore.selectedHex) return null
  return gameStore.getHexAt(gameStore.selectedHex.q, gameStore.selectedHex.r)
})

function getHealthColor(health: number): string {
  if (health > 70) return '#22c55e'
  if (health > 40) return '#eab308'
  return '#ef4444'
}

function getPollutionColor(pollution: number): string {
  if (pollution > 70) return '#ef4444'
  if (pollution > 40) return '#eab308'
  return '#22c55e'
}
</script>

<style scoped>
.hex-info-panel {
  background: rgba(15, 23, 42, 0.95);
  border-radius: 12px;
  padding: 16px;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.panel-header {
  margin-bottom: 12px;
}

.terrain-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 20px;
  color: white;
  font-size: 13px;
  font-weight: 500;
}

.resource-list {
  background: rgba(30, 41, 59, 0.6);
  border-radius: 8px;
  padding: 8px;
}

.resource-item {
  display: flex;
  justify-content: space-between;
  padding: 4px 0;
}

.resource-name {
  color: #cbd5e1;
  font-size: 13px;
}

.resource-amount {
  color: #f1c40f;
  font-weight: 600;
  font-size: 13px;
}

.health-bar {
  width: 80px;
  height: 8px;
  background: #374151;
  border-radius: 4px;
  overflow: hidden;
  align-self: center;
}

.health-fill {
  height: 100%;
  transition: width 0.3s ease;
}

.no-selection {
  min-height: 150px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
