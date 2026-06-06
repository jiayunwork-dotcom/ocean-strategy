<template>
  <Teleport to="body">
    <div v-if="showModal" class="modal-overlay" @click.self="handleClose">
      <div class="modal-content">
        <div class="modal-header">
          <h3 class="text-xl font-bold text-white">外交提议</h3>
          <button class="close-btn" @click="handleClose">×</button>
        </div>

        <div v-if="pendingProposals.length === 0" class="no-proposals">
          <p class="text-gray-400">暂无待处理的提议</p>
        </div>

        <div v-else class="proposal-list">
          <div
            v-for="proposal in pendingProposals"
            :key="proposal.id"
            class="proposal-item"
          >
            <div class="proposal-info">
              <div class="proposal-header">
                <span class="proposal-type">{{ getTreatyName(proposal.treaty_type) }}</span>
                <span class="proposal-from">来自: {{ getPlayerName(proposal.from_player_id) }}</span>
              </div>
              <p class="proposal-desc">{{ getProposalDescription(proposal.treaty_type) }}</p>
            </div>

            <div class="proposal-actions">
              <button class="action-btn accept-btn" @click="respondToProposal(proposal.id, true)">
                接受
              </button>
              <button class="action-btn reject-btn" @click="respondToProposal(proposal.id, false)">
                拒绝
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import type { DiplomaticProposal } from '~/types'
import { TREATY_TYPE_NAMES } from '~/types'

const gameStore = useGameStore()
const { respondToProposal: apiRespondToProposal, getGame: apiGetGame } = useGameApi()
const route = useRoute()

const showModal = computed(() => gameStore.showProposalModal)
const pendingProposals = computed(() => gameStore.myPendingProposals)

function getPlayerName(playerId: string): string {
  if (!gameStore.gameState) return '未知玩家'
  return gameStore.gameState.players[playerId]?.name || '未知玩家'
}

function getTreatyName(treatyType: string): string {
  return TREATY_TYPE_NAMES[treatyType as keyof typeof TREATY_TYPE_NAMES] || '条约'
}

function getProposalDescription(treatyType: string): string {
  switch (treatyType) {
    case 'nap':
      return '互不侵犯条约：双方承诺不主动攻击对方，保持和平共处。'
    case 'alliance':
      return '军事同盟：双方共享已探索海域的视野，共同对抗敌人。'
    default:
      return ''
  }
}

function handleClose() {
  gameStore.closeProposalModal()
}

async function respondToProposal(proposalId: string, accept: boolean) {
  if (!gameStore.currentPlayer) return
  const gameId = route.params.id as string

  try {
    await apiRespondToProposal(gameId, proposalId, gameStore.currentPlayer.id, accept)
    const updatedState = await apiGetGame(gameId)
    gameStore.setGameState(updatedState)

    if (gameStore.myPendingProposals.length === 0) {
      gameStore.closeProposalModal()
    }
  } catch (error: any) {
    console.error('Failed to respond to proposal:', error)
    alert(error.data?.error || '回应提议失败')
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.75);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal-content {
  background: linear-gradient(135deg, #1e293b, #0f172a);
  border-radius: 16px;
  padding: 24px;
  min-width: 400px;
  max-width: 500px;
  border: 1px solid rgba(59, 130, 246, 0.3);
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid #374151;
}

.close-btn {
  background: none;
  border: none;
  color: #94a3b8;
  font-size: 24px;
  cursor: pointer;
  padding: 4px;
  line-height: 1;
  transition: color 0.2s;
}

.close-btn:hover {
  color: white;
}

.no-proposals {
  text-align: center;
  padding: 40px 0;
}

.proposal-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.proposal-item {
  background: rgba(30, 41, 59, 0.8);
  border-radius: 12px;
  padding: 16px;
  border: 1px solid rgba(59, 130, 246, 0.2);
}

.proposal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.proposal-type {
  font-size: 16px;
  font-weight: 600;
  color: #60a5fa;
}

.proposal-from {
  font-size: 12px;
  color: #94a3b8;
}

.proposal-desc {
  font-size: 13px;
  color: #cbd5e1;
  margin-bottom: 16px;
  line-height: 1.5;
}

.proposal-actions {
  display: flex;
  gap: 12px;
}

.action-btn {
  flex: 1;
  padding: 10px 16px;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.accept-btn {
  background: linear-gradient(135deg, #22c55e, #16a34a);
  color: white;
}

.accept-btn:hover {
  background: linear-gradient(135deg, #16a34a, #15803d);
  transform: translateY(-1px);
}

.reject-btn {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.reject-btn:hover {
  background: rgba(239, 68, 68, 0.3);
}
</style>
