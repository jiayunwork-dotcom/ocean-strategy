<template>
  <div class="auction-panel">
    <div class="panel-header">
      <h3 class="text-lg font-bold text-white mb-2">拍卖行</h3>
      <div class="tab-buttons">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="activeTab = tab.id"
          :class="['tab-btn', { active: activeTab === tab.id }]"
        >
          {{ tab.name }}
        </button>
      </div>
    </div>

    <div class="panel-content">
      <div v-if="activeTab === 'list'" class="tab-content">
        <div class="auctions-list">
          <div
            v-for="auction in activeAuctions"
            :key="auction.id"
            class="auction-card"
          >
            <div class="auction-header">
              <span class="item-type-badge">{{ AUCTION_ITEM_TYPE_NAMES[auction.item.item_type] }}</span>
              <span class="auction-status" :class="auction.status">
                {{ AUCTION_STATUS_NAMES[auction.status] }}
              </span>
            </div>
            <div class="auction-item-name">{{ auction.item.item_name }}</div>
            <div class="auction-info">
              <div class="info-row">
                <span class="label">起拍价</span>
                <span class="value">{{ auction.starting_price }} 金币</span>
              </div>
              <div class="info-row">
                <span class="label">当前出价</span>
                <span class="value highlight">{{ auction.current_bid }} 金币</span>
              </div>
              <div class="info-row">
                <span class="label">剩余回合</span>
                <span class="value">{{ remainingTurns(auction) }} 回合</span>
              </div>
              <div class="info-row" v-if="auction.current_bidder">
                <span class="label">最高出价者</span>
                <span class="value">{{ getBidderName(auction.current_bidder) }}</span>
              </div>
            </div>
            <div v-if="canBid(auction)" class="auction-actions">
              <div class="bid-input-row">
                <input
                  v-model.number="bidAmounts[auction.id]"
                  type="number"
                  :min="minBid(auction)"
                  class="bid-input"
                  placeholder="输入出价"
                />
                <button
                  @click="placeBid(auction.id)"
                  :disabled="!canPlaceBid(auction)"
                  class="bid-btn"
                >
                  出价
                </button>
              </div>
              <div class="min-bid-hint">
                最低出价：{{ minBid(auction) }} 金币
              </div>
            </div>
            <div v-else-if="auction.seller_id === currentPlayerId" class="my-auction-hint">
              这是你发起的拍卖
            </div>
          </div>
          <div v-if="activeAuctions.length === 0" class="empty-state">
            暂无进行中的拍卖
          </div>
        </div>
      </div>

      <div v-if="activeTab === 'create'" class="tab-content">
        <div class="create-auction-form">
          <div class="form-group">
            <label>物品类型</label>
            <div class="type-selector">
              <button
                v-for="type in auctionableTypes"
                :key="type.id"
                :class="['type-btn', { active: newAuction.itemType === type.id }]"
                @click="newAuction.itemType = type.id"
              >
                {{ type.name }}
              </button>
            </div>
          </div>
          <div class="form-group">
            <label>选择物品</label>
            <select v-model="newAuction.itemId" class="form-select">
              <option value="">请选择</option>
              <option v-for="item in availableItems" :key="item.id" :value="item.id">
                {{ item.name }}
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>起拍价（金币）</label>
            <input
              v-model.number="newAuction.startingPrice"
              type="number"
              min="1"
              class="form-input"
              placeholder="输入起拍价"
            />
          </div>
          <div class="auction-info-note">
            <p>拍卖将持续 3 个回合</p>
            <p>每次出价需至少高于当前价 10%</p>
            <p>流拍物品将退还卖家</p>
          </div>
          <button
            @click="createAuction"
            :disabled="!canCreateAuction"
            class="create-btn"
          >
            发起拍卖
          </button>
        </div>
      </div>

      <div v-if="activeTab === 'history'" class="tab-content">
        <div class="auctions-list">
          <div
            v-for="auction in finishedAuctions"
            :key="auction.id"
            :class="['auction-card', 'finished']"
          >
            <div class="auction-header">
              <span class="item-type-badge">{{ AUCTION_ITEM_TYPE_NAMES[auction.item.item_type] }}</span>
              <span class="auction-status" :class="auction.status">
                {{ AUCTION_STATUS_NAMES[auction.status] }}
              </span>
            </div>
            <div class="auction-item-name">{{ auction.item.item_name }}</div>
            <div class="auction-info">
              <div class="info-row">
                <span class="label">成交价</span>
                <span class="value highlight">{{ auction.current_bid }} 金币</span>
              </div>
              <div class="info-row" v-if="auction.status === 'finished' && auction.current_bidder">
                <span class="label">买家</span>
                <span class="value">{{ getBidderName(auction.current_bidder) }}</span>
              </div>
            </div>
          </div>
          <div v-if="finishedAuctions.length === 0" class="empty-state">
            暂无历史拍卖
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import type { Auction, AuctionItemType, GameState } from '~/types'
import { AUCTION_ITEM_TYPE_NAMES, AUCTION_STATUS_NAMES } from '~/types'

const props = defineProps<{
  gameState: GameState | null
  currentPlayerId: string | null
}>()

const emit = defineEmits<{
  (e: 'createAuction', data: { itemType: AuctionItemType; itemId: string; startingPrice: number }): void
  (e: 'placeBid', auctionId: string, amount: number): void
}>()

const tabs = [
  { id: 'list', name: '拍卖中' },
  { id: 'create', name: '发起拍卖' },
  { id: 'history', name: '历史' },
]

const auctionableTypes = [
  { id: 'tech' as AuctionItemType, name: '科技' },
  { id: 'ship' as AuctionItemType, name: '船只' },
]

const activeTab = ref('list')
const bidAmounts = reactive<Record<string, number>>({})

const newAuction = reactive({
  itemType: 'tech' as AuctionItemType,
  itemId: '',
  startingPrice: 1000,
})

const allAuctions = computed(() => {
  return props.gameState?.auctions || []
})

const activeAuctions = computed(() => {
  return allAuctions.value.filter(a => a.status === 'active')
})

const finishedAuctions = computed(() => {
  return allAuctions.value.filter(a => a.status !== 'active')
})

const myCompletedTechs = computed(() => {
  if (!props.gameState?.techs || !props.currentPlayerId) return []
  return props.gameState.techs.filter(
    t => t.player_id === props.currentPlayerId && t.completed
  ).map(t => ({
    id: t.tech_id,
    name: getTechName(t.tech_id)
  }))
})

const myShips = computed(() => {
  if (!props.gameState?.ships || !props.currentPlayerId) return []
  const frozenShips = props.gameState.frozen_ships || {}
  return props.gameState.ships
    .filter(s => s.owner_id === props.currentPlayerId && !frozenShips[s.id])
    .map(s => ({
      id: s.id,
      name: getShipTypeName(s.type)
    }))
})

const availableItems = computed(() => {
  if (newAuction.itemType === 'tech') {
    return myCompletedTechs.value
  } else if (newAuction.itemType === 'ship') {
    return myShips.value
  }
  return []
})

const canCreateAuction = computed(() => {
  return (
    newAuction.itemId !== '' &&
    newAuction.startingPrice > 0 &&
    props.gameState?.game.phase === 'decision'
  )
})

function remainingTurns(auction: Auction): number {
  if (!props.gameState) return 0
  const elapsed = props.gameState.game.current_turn - auction.start_turn
  return Math.max(0, auction.duration - elapsed)
}

function minBid(auction: Auction): number {
  if (!auction.current_bidder) {
    return auction.starting_price
  }
  return Math.ceil(auction.current_bid * 1.1)
}

function canBid(auction: Auction): boolean {
  if (!props.currentPlayerId) return false
  if (auction.seller_id === props.currentPlayerId) return false
  if (auction.status !== 'active') return false
  if (props.gameState?.game.phase !== 'decision') return false
  return true
}

function canPlaceBid(auction: Auction): boolean {
  const amount = bidAmounts[auction.id] || 0
  return amount >= minBid(auction) && canBid(auction)
}

function getBidderName(bidderId: string): string {
  if (!props.gameState?.players) return '未知'
  const player = props.gameState.players[bidderId]
  return player ? player.name : '未知'
}

function getTechName(techId: string): string {
  const techNames: Record<string, string> = {
    'deep_drilling_1': '深海钻探 I',
    'deep_drilling_2': '深海钻探 II',
    'trench_mining': '海沟采矿技术',
    'eco_friendly_1': '环保技术 I',
    'eco_friendly_2': '环保技术 II',
    'sustainable_farm': '可持续养殖',
    'ship_armor': '舰船装甲',
    'ship_weapons': '舰载武器',
    'typhoon_resist': '抗台风技术',
  }
  return techNames[techId] || techId
}

function getShipTypeName(type: string): string {
  const names: Record<string, string> = {
    explorer: '勘探船',
    constructor: '工程船',
    transport: '运输船',
    escort: '护卫舰',
  }
  return names[type] || type
}

function placeBid(auctionId: string) {
  const amount = bidAmounts[auctionId] || 0
  if (amount <= 0) return
  emit('placeBid', auctionId, amount)
  bidAmounts[auctionId] = 0
}

function createAuction() {
  if (!canCreateAuction.value) return
  emit('createAuction', {
    itemType: newAuction.itemType,
    itemId: newAuction.itemId,
    startingPrice: newAuction.startingPrice,
  })
  newAuction.itemId = ''
  newAuction.startingPrice = 1000
}
</script>

<style scoped>
.auction-panel {
  background: rgba(15, 23, 42, 0.95);
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 12px;
  max-height: 500px;
  overflow-y: auto;
}

.panel-header h3 {
  color: #f1f5f9;
}

.tab-buttons {
  display: flex;
  gap: 4px;
  margin-bottom: 12px;
}

.tab-btn {
  flex: 1;
  padding: 6px 8px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 4px;
  color: #94a3b8;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
}

.tab-btn:hover {
  background: #334155;
}

.tab-btn.active {
  background: #a855f7;
  border-color: #a855f7;
  color: white;
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.auctions-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.auction-card {
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 12px;
}

.auction-card.finished {
  opacity: 0.7;
}

.auction-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.item-type-badge {
  padding: 2px 8px;
  background: rgba(168, 85, 247, 0.2);
  color: #c084fc;
  border-radius: 10px;
  font-size: 10px;
  font-weight: 600;
}

.auction-status {
  padding: 2px 8px;
  border-radius: 10px;
  font-size: 10px;
  font-weight: 600;
}

.auction-status.active {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
}

.auction-status.finished {
  background: rgba(59, 130, 246, 0.2);
  color: #3b82f6;
}

.auction-status.expired {
  background: rgba(107, 114, 128, 0.2);
  color: #9ca3af;
}

.auction-item-name {
  color: #f1f5f9;
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 10px;
}

.auction-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 10px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.info-row .label {
  color: #64748b;
}

.info-row .value {
  color: #e2e8f0;
}

.info-row .value.highlight {
  color: #fbbf24;
  font-weight: 600;
}

.auction-actions {
  border-top: 1px solid #334155;
  padding-top: 10px;
}

.bid-input-row {
  display: flex;
  gap: 6px;
}

.bid-input {
  flex: 1;
  padding: 6px 10px;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 4px;
  color: #f1f5f9;
  font-size: 12px;
}

.bid-input:focus {
  outline: none;
  border-color: #a855f7;
}

.bid-btn {
  padding: 6px 14px;
  background: linear-gradient(135deg, #a855f7, #9333ea);
  border: none;
  border-radius: 4px;
  color: white;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.bid-btn:hover:not(:disabled) {
  opacity: 0.9;
}

.bid-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.min-bid-hint {
  font-size: 10px;
  color: #64748b;
  margin-top: 4px;
  text-align: right;
}

.my-auction-hint {
  text-align: center;
  padding: 8px;
  background: rgba(168, 85, 247, 0.1);
  border-radius: 4px;
  color: #c084fc;
  font-size: 11px;
}

.empty-state {
  text-align: center;
  padding: 30px;
  color: #64748b;
  font-size: 12px;
}

.create-auction-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.form-group label {
  color: #94a3b8;
  font-size: 12px;
}

.type-selector {
  display: flex;
  gap: 4px;
}

.type-btn {
  flex: 1;
  padding: 8px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 4px;
  color: #94a3b8;
  font-size: 12px;
  cursor: pointer;
}

.type-btn.active {
  border-color: #a855f7;
  background: rgba(168, 85, 247, 0.1);
  color: #c084fc;
}

.form-input,
.form-select {
  padding: 8px 10px;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 4px;
  color: #f1f5f9;
  font-size: 13px;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  border-color: #a855f7;
}

.auction-info-note {
  background: #0f172a;
  border-radius: 6px;
  padding: 10px;
}

.auction-info-note p {
  color: #64748b;
  font-size: 11px;
  margin: 2px 0;
}

.create-btn {
  padding: 10px;
  background: linear-gradient(135deg, #a855f7, #9333ea);
  border: none;
  border-radius: 6px;
  color: white;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.create-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(168, 85, 247, 0.4);
}

.create-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
