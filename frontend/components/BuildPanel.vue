<template>
  <div class="build-panel">
    <div class="panel-header">
      <h3 class="text-lg font-bold text-white">建造设施</h3>
    </div>
    
    <div v-if="selectedHexData" class="build-options space-y-2">
      <div v-for="facility in availableFacilities" :key="facility.type" class="build-option">
        <div class="facility-info">
          <span class="facility-icon">{{ facility.icon }}</span>
          <div class="facility-details">
            <span class="facility-name">{{ facility.name }}</span>
            <span class="facility-cost">💰 {{ facility.cost }}</span>
          </div>
        </div>
        <button
          class="build-btn"
          :disabled="!canBuild(facility.type)"
          @click="buildFacility(facility.type)"
        >
          建造
        </button>
      </div>
      
      <div v-if="availableFacilities.length === 0" class="no-options text-center py-4">
        <p class="text-gray-400 text-sm">该地形无法建造设施</p>
      </div>
    </div>
    
    <div v-else class="no-selection">
      <p class="text-gray-400 text-sm text-center py-6">选择一个六角格查看建造选项</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { FACILITY_NAMES, type HexTerrain, type FacilityType } from '~/types'

const gameStore = useGameStore()
const route = useRoute()
const gameApi = useGameApi()

const selectedHexData = computed(() => {
  if (!gameStore.selectedHex) return null
  return gameStore.getHexAt(gameStore.selectedHex.q, gameStore.selectedHex.r)
})

const facilityConfigs = [
  { type: 'drilling' as FacilityType, name: FACILITY_NAMES.drilling, icon: '🏗️', cost: 5000, terrains: ['shallow', 'deep'] },
  { type: 'mine' as FacilityType, name: FACILITY_NAMES.mine, icon: '⛏️', cost: 8000, terrains: ['deep', 'trench', 'vent'] },
  { type: 'tidal' as FacilityType, name: FACILITY_NAMES.tidal, icon: '⚡', cost: 3000, terrains: [] },
  { type: 'farm' as FacilityType, name: FACILITY_NAMES.farm, icon: '🐟', cost: 4000, terrains: ['reef'] },
  { type: 'port' as FacilityType, name: FACILITY_NAMES.port, icon: '⚓', cost: 10000, terrains: ['shallow'] },
]

const availableFacilities = computed(() => {
  if (!selectedHexData.value) return []
  
  return facilityConfigs.filter(f => {
    if (f.type === 'tidal') {
      return selectedHexData.value?.has_current
    }
    return f.terrains.includes(selectedHexData.value?.terrain || '')
  })
})

function canBuild(type: FacilityType): boolean {
  if (!selectedHexData.value || !gameStore.currentPlayer) return false
  if (selectedHexData.value.facility) return false
  
  const config = facilityConfigs.find(f => f.type === type)
  if (!config) return false
  
  return gameStore.currentPlayer.money >= config.cost
}

async function buildFacility(type: FacilityType) {
  if (!selectedHexData.value || !gameStore.currentPlayer) return
  
  try {
    await gameApi.buildFacility(
      route.params.id as string,
      gameStore.currentPlayer.id,
      selectedHexData.value.q,
      selectedHexData.value.r,
      type
    )
    
    const state = await gameApi.getGame(route.params.id as string)
    gameStore.setGameState(state)
  } catch (e) {
    console.error('Build failed:', e)
  }
}
</script>

<style scoped>
.build-panel {
  background: rgba(15, 23, 42, 0.95);
  border-radius: 12px;
  padding: 16px;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.panel-header {
  margin-bottom: 12px;
}

.build-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  background: rgba(30, 41, 59, 0.6);
  border-radius: 8px;
}

.facility-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.facility-icon {
  font-size: 24px;
}

.facility-details {
  display: flex;
  flex-direction: column;
}

.facility-name {
  color: white;
  font-size: 14px;
  font-weight: 500;
}

.facility-cost {
  color: #f1c40f;
  font-size: 12px;
}

.build-btn {
  padding: 6px 14px;
  background: #22c55e;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.build-btn:hover:not(:disabled) {
  background: #16a34a;
}

.build-btn:disabled {
  background: #475569;
  cursor: not-allowed;
  opacity: 0.6;
}

.no-selection {
  min-height: 150px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
