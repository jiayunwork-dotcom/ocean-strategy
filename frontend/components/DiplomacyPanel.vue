<template>
  <div class="diplomacy-panel">
    <div class="panel-header">
      <h3 class="text-lg font-bold text-white mb-3">外交关系</h3>
    </div>

    <div v-if="otherPlayers.length === 0" class="no-players">
      <p class="text-gray-400 text-sm">暂无其他玩家</p>
    </div>

    <div v-else class="player-list">
      <div v-for="player in otherPlayers" :key="player.id" class="player-item">
        <div class="player-info">
          <div class="player-color-dot" :style="{ backgroundColor: player.color }"></div>
          <div class="player-details">
            <span class="player-name">{{ player.name }}</span>
            <div class="player-reputation">
              <span class="rep-icon">⭐</span>
              <span class="rep-value">{{ player.reputation }}</span>
              <span v-if="hasCooldown(player.id)" class="cooldown-badge">
                冷却 {{ getCooldown(player.id) }} 回合
              </span>
            </div>
          </div>
        </div>

        <div class="relation-status">
          <span
            class="relation-badge"
            :style="{ backgroundColor: getRelationColor(player.id) + '20', color: getRelationColor(player.id) }"
          >
            {{ getRelationName(player.id) }}
          </span>
        </div>

        <div class="action-buttons">
          <template v-if="getRelationStatus(player.id) === 'neutral'">
            <button
              class="action-btn propose-btn"
              :disabled="hasCooldown(currentPlayer?.id || '')"
              @click="proposeTreaty(player.id, 'nap')"
            >
              提议互不侵犯
            </button>
          </template>

          <template v-else-if="getRelationStatus(player.id) === 'nap'">
            <button
              class="action-btn propose-btn alliance-btn"
              :disabled="hasCooldown(currentPlayer?.id || '')"
              @click="proposeTreaty(player.id, 'alliance')"
            >
              提议军事同盟
            </button>
            <button class="action-btn break-btn" @click="breakTreaty(player.id)">
              撕毁条约
            </button>
          </template>

          <template v-else-if="getRelationStatus(player.id) === 'alliance'">
            <button class="action-btn break-btn" @click="breakTreaty(player.id)">
              撕毁条约
            </button>
          </template>

          <template v-else-if="getRelationStatus(player.id) === 'hostile'">
            <span class="text-red-400 text-xs">处于敌对状态</span>
          </template>
        </div>
      </div>
    </div>

    <div v-if="gameLogs.length > 0" class="game-logs mt-4 pt-4 border-t border-gray-700">
      <h4 class="text-sm font-semibold text-gray-300 mb-2">游戏日志</h4>
      <div class="log-list">
        <div v-for="log in gameLogs.slice(0, 5)" :key="log.id" class="log-item">
          <span class="log-turn">T{{ log.turn }}</span>
          <span class="log-message">{{ log.message }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { RELATION_STATUS_NAMES, RELATION_STATUS_COLORS } from '~/types'

const gameStore = useGameStore()
const { proposeTreaty: apiProposeTreaty, breakTreaty: apiBreakTreaty, getGame: apiGetGame } = useGameApi()
const route = useRoute()

const currentPlayer = computed(() => gameStore.currentPlayer)
const otherPlayers = computed(() => gameStore.otherPlayers)
const gameLogs = computed(() => gameStore.gameLogs)

function getRelationWithPlayer(playerId: string) {
  return gameStore.getRelationWithPlayer(playerId)
}

function getRelationStatus(playerId: string): string {
  const relation = getRelationWithPlayer(playerId)
  return relation?.status || 'neutral'
}

function getRelationName(playerId: string): string {
  const status = getRelationStatus(playerId) as any
  return RELATION_STATUS_NAMES[status] || '中立'
}

function getRelationColor(playerId: string): string {
  const status = getRelationStatus(playerId) as any
  return RELATION_STATUS_COLORS[status] || '#94a3b8'
}

function hasCooldown(playerId: string): boolean {
  return gameStore.hasCooldown(playerId)
}

function getCooldown(playerId: string): number {
  return gameStore.getCooldown(playerId)
}

async function proposeTreaty(toPlayerId: string, treatyType: string) {
  if (!currentPlayer.value) return
  const gameId = route.params.id as string

  try {
    await apiProposeTreaty(gameId, currentPlayer.value.id, toPlayerId, treatyType)
    const updatedState = await apiGetGame(gameId)
    gameStore.setGameState(updatedState)
  } catch (error: any) {
    console.error('Failed to propose treaty:', error)
    alert(error.data?.error || '发起条约提议失败')
  }
}

async function breakTreaty(otherPlayerId: string) {
  if (!currentPlayer.value) return
  const gameId = route.params.id as string

  if (!confirm('确定要撕毁条约吗？这将使双方进入敌对状态，并扣除20点声誉值。')) {
    return
  }

  try {
    await apiBreakTreaty(gameId, currentPlayer.value.id, otherPlayerId)
    const updatedState = await apiGetGame(gameId)
    gameStore.setGameState(updatedState)
  } catch (error: any) {
    console.error('Failed to break treaty:', error)
    alert(error.data?.error || '撕毁条约失败')
  }
}
</script>

<style scoped>
.diplomacy-panel {
  background: rgba(15, 23, 42, 0.95);
  border-radius: 12px;
  padding: 16px;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.panel-header {
  margin-bottom: 12px;
}

.no-players {
  text-align: center;
  padding: 20px 0;
}

.player-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.player-item {
  background: rgba(30, 41, 59, 0.8);
  border-radius: 8px;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.player-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.player-color-dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.3);
  flex-shrink: 0;
}

.player-details {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.player-name {
  font-size: 14px;
  font-weight: 600;
  color: white;
}

.player-reputation {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
}

.rep-icon {
  font-size: 12px;
}

.rep-value {
  color: #fbbf24;
  font-weight: 500;
}

.cooldown-badge {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
  margin-left: 4px;
}

.relation-status {
  display: flex;
  justify-content: flex-start;
}

.relation-badge {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.action-btn {
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.propose-btn {
  background: linear-gradient(135deg, #22c55e, #16a34a);
  color: white;
}

.propose-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #16a34a, #15803d);
}

.alliance-btn {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
}

.alliance-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
}

.break-btn {
  background: linear-gradient(135deg, #ef4444, #dc2626);
  color: white;
}

.break-btn:hover {
  background: linear-gradient(135deg, #dc2626, #b91c1c);
}

.game-logs {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #374151;
}

.log-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 150px;
  overflow-y: auto;
}

.log-item {
  display: flex;
  gap: 8px;
  font-size: 11px;
  color: #94a3b8;
}

.log-turn {
  color: #60a5fa;
  font-weight: 500;
  flex-shrink: 0;
}

.log-message {
  word-break: break-word;
}
</style>
