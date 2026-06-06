<template>
  <div class="action-bar">
    <div class="action-bar-left">
      <div v-if="gameState?.game" class="turn-info">
        <span class="turn-label">回合</span>
        <span class="turn-value">{{ gameState.game.current_turn }}</span>
        <span class="turn-divider">/</span>
        <span class="turn-max">{{ gameState.game.max_turns }}</span>
        <span class="phase-badge">{{ PHASE_NAMES[gameState.game.phase] }}</span>
      </div>
    </div>
    
    <div class="action-bar-right">
      <button
        v-if="gameState?.game?.status === 'waiting' && isHost"
        class="action-btn start-btn"
        @click="startGame"
      >
        ▶️ 开始游戏
      </button>
      
      <button
        v-if="gameState?.game?.status === 'playing'"
        class="action-btn next-btn"
        @click="nextPhase"
      >
        ⏭️ 下一阶段
      </button>
      
      <button
        class="action-btn menu-btn"
        @click="$emit('toggleMenu')"
      >
        ☰ 菜单
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { PHASE_NAMES } from '~/types'

const props = defineProps<{
  isHost?: boolean
}>()

defineEmits<{
  (e: 'toggleMenu'): void
}>()

const gameStore = useGameStore()
const route = useRoute()
const gameApi = useGameApi()

const gameState = computed(() => gameStore.gameState)

async function startGame() {
  try {
    await gameApi.startGame(route.params.id as string)
    const state = await gameApi.getGame(route.params.id as string)
    gameStore.setGameState(state)
  } catch (e) {
    console.error('Start game failed:', e)
  }
}

async function nextPhase() {
  try {
    const state = await gameApi.nextPhase(route.params.id as string)
    gameStore.setGameState(state)
  } catch (e) {
    console.error('Next phase failed:', e)
  }
}
</script>

<style scoped>
.action-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: rgba(15, 23, 42, 0.98);
  border-bottom: 1px solid rgba(59, 130, 246, 0.3);
}

.turn-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.turn-label {
  color: #94a3b8;
  font-size: 14px;
}

.turn-value {
  color: white;
  font-size: 18px;
  font-weight: bold;
}

.turn-divider {
  color: #64748b;
}

.turn-max {
  color: #64748b;
  font-size: 14px;
}

.phase-badge {
  margin-left: 12px;
  padding: 4px 12px;
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 500;
}

.action-bar-right {
  display: flex;
  gap: 10px;
}

.action-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.start-btn {
  background: #22c55e;
  color: white;
}

.start-btn:hover {
  background: #16a34a;
}

.next-btn {
  background: #3b82f6;
  color: white;
}

.next-btn:hover {
  background: #2563eb;
}

.menu-btn {
  background: rgba(71, 85, 105, 0.8);
  color: white;
}

.menu-btn:hover {
  background: rgba(71, 85, 105, 1);
}
</style>
