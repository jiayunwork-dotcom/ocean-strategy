<template>
  <div class="tech-panel">
    <div class="panel-header">
      <h3 class="text-lg font-bold text-white">科技研发</h3>
    </div>
    
    <div class="tech-categories space-y-4">
      <div v-for="category in categories" :key="category.key" class="tech-category">
        <div class="category-header" @click="toggleCategory(category.key)">
          <span class="category-icon">{{ category.icon }}</span>
          <span class="category-name">{{ category.name }}</span>
          <span class="toggle-icon">{{ expandedCategories.includes(category.key) ? '▼' : '▶' }}</span>
        </div>
        
        <div v-if="expandedCategories.includes(category.key)" class="tech-list space-y-2 mt-2">
          <div
            v-for="tech in getTechsByCategory(category.key)"
            :key="tech.id"
            class="tech-item"
            :class="{ 'researching': isResearching(tech.id), 'completed': isCompleted(tech.id), 'locked': !canResearch(tech.id) }"
          >
            <div class="tech-info">
              <span class="tech-name">{{ tech.name }}</span>
              <p class="tech-desc">{{ tech.description }}</p>
              <div class="tech-meta">
                <span class="tech-cost">💰 {{ tech.cost }}</span>
                <span class="tech-turns">⏱️ {{ tech.turns }} 回合</span>
              </div>
            </div>
            <div class="tech-action">
              <button
                v-if="!isCompleted(tech.id) && !isResearching(tech.id)"
                class="research-btn"
                :disabled="!canResearch(tech.id)"
                @click="startResearch(tech.id)"
              >
                研发
              </button>
              <span v-else-if="isResearching(tech.id)" class="researching-text">
                研发中 ({{ getResearchTurnsLeft(tech.id) }}回合)
              </span>
              <span v-else class="completed-text">✓ 已完成</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Technology, TechnologyCategory } from '~/types'

const gameStore = useGameStore()
const route = useRoute()
const gameApi = useGameApi()

const { data: allTechs } = useAsyncData('technologies', () => gameApi.getTechnologies())

const expandedCategories = ref<string[]>(['extraction', 'ecology', 'military'])

const categories = [
  { key: 'extraction' as TechnologyCategory, name: '开采科技', icon: '⛏️' },
  { key: 'ecology' as TechnologyCategory, name: '环保科技', icon: '🌿' },
  { key: 'military' as TechnologyCategory, name: '军事科技', icon: '⚔️' },
]

function getTechsByCategory(category: string): Technology[] {
  return (allTechs.value || []).filter(t => t.category === category)
}

function isCompleted(techId: string): boolean {
  return gameStore.playerTechs.some(t => t.tech_id === techId && t.completed)
}

function isResearching(techId: string): boolean {
  return gameStore.playerTechs.some(t => t.tech_id === techId && t.researching)
}

function getResearchTurnsLeft(techId: string): number {
  const tech = gameStore.playerTechs.find(t => t.tech_id === techId)
  return tech?.turns_left || 0
}

function canResearch(techId: string): boolean {
  const tech = (allTechs.value || []).find(t => t.id === techId)
  if (!tech || !gameStore.currentPlayer) return false
  if (isCompleted(techId) || isResearching(techId)) return false
  if (gameStore.currentPlayer.money < tech.cost) return false
  
  for (const prereq of tech.prerequisites) {
    if (!isCompleted(prereq)) return false
  }
  
  return true
}

function toggleCategory(key: string) {
  const index = expandedCategories.value.indexOf(key)
  if (index > -1) {
    expandedCategories.value.splice(index, 1)
  } else {
    expandedCategories.value.push(key)
  }
}

async function startResearch(techId: string) {
  if (!gameStore.currentPlayer) return
  
  try {
    await gameApi.startResearch(route.params.id as string, gameStore.currentPlayer.id, techId)
    const state = await gameApi.getGame(route.params.id as string)
    gameStore.setGameState(state)
  } catch (e) {
    console.error('Research failed:', e)
  }
}
</script>

<style scoped>
.tech-panel {
  background: rgba(15, 23, 42, 0.95);
  border-radius: 12px;
  padding: 16px;
  border: 1px solid rgba(59, 130, 246, 0.3);
  max-height: 400px;
  overflow-y: auto;
}

.panel-header {
  margin-bottom: 12px;
}

.category-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(30, 41, 59, 0.8);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.category-header:hover {
  background: rgba(30, 41, 59, 1);
}

.category-icon {
  font-size: 18px;
}

.category-name {
  flex: 1;
  color: white;
  font-weight: 500;
  font-size: 14px;
}

.toggle-icon {
  color: #94a3b8;
  font-size: 10px;
}

.tech-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  background: rgba(30, 41, 59, 0.5);
  border-radius: 8px;
  border-left: 3px solid #3b82f6;
}

.tech-item.completed {
  border-left-color: #22c55e;
  opacity: 0.7;
}

.tech-item.researching {
  border-left-color: #eab308;
}

.tech-item.locked {
  border-left-color: #475569;
  opacity: 0.5;
}

.tech-info {
  flex: 1;
}

.tech-name {
  color: white;
  font-weight: 500;
  font-size: 14px;
}

.tech-desc {
  color: #94a3b8;
  font-size: 12px;
  margin-top: 4px;
}

.tech-meta {
  display: flex;
  gap: 12px;
  margin-top: 6px;
}

.tech-cost,
.tech-turns {
  color: #f1c40f;
  font-size: 11px;
}

.research-btn {
  padding: 6px 12px;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.research-btn:hover:not(:disabled) {
  background: #2563eb;
}

.research-btn:disabled {
  background: #475569;
  cursor: not-allowed;
}

.researching-text {
  color: #eab308;
  font-size: 12px;
}

.completed-text {
  color: #22c55e;
  font-size: 12px;
}
</style>
