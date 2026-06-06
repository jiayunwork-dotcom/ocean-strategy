<template>
  <div class="market-panel">
    <div class="panel-header">
      <h3 class="text-lg font-bold text-white mb-2">资源市场</h3>
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
      <div v-if="activeTab === 'prices'" class="tab-content">
        <div class="resource-selector">
          <button
            v-for="res in resources"
            :key="res"
            @click="selectedResource = res"
            :class="['res-btn', { active: selectedResource === res }]"
          >
            {{ RESOURCE_NAMES[res] }}
          </button>
        </div>
        <div class="chart-container">
          <svg :width="chartWidth" :height="chartHeight" class="price-chart">
            <defs>
              <linearGradient id="lineGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                <stop offset="0%" style="stop-color:#3b82f6;stop-opacity:0.3" />
                <stop offset="100%" style="stop-color:#3b82f6;stop-opacity:0" />
              </linearGradient>
            </defs>
            <g class="grid">
              <line
                v-for="i in 5"
                :key="'h'+i"
                :x1="padding"
                :y1="padding + (i - 1) * (chartHeight - 2 * padding) / 4"
                :x2="chartWidth - padding"
                :y2="padding + (i - 1) * (chartHeight - 2 * padding) / 4"
                stroke="#374151"
                stroke-width="1"
              />
            </g>
            <polyline
              :points="chartPoints"
              fill="none"
              stroke="#3b82f6"
              stroke-width="2"
            />
            <polygon
              :points="areaPoints"
              fill="url(#lineGradient)"
            />
            <text
              v-for="(point, idx) in labeledPoints"
              :key="'label'+idx"
              :x="point.x"
              :y="point.y - 8"
              fill="#9ca3af"
              font-size="10"
              text-anchor="middle"
            >
              {{ point.price }}
            </text>
          </svg>
        </div>
        <div class="current-price-info">
          <span class="text-gray-400">当前价格：</span>
          <span class="text-yellow-400 font-bold">
            {{ currentPrice }} 金币
          </span>
        </div>
      </div>

      <div v-if="activeTab === 'orders'" class="tab-content">
        <div class="order-tabs">
          <button
            :class="['order-tab', { active: orderTypeTab === 'buy' }]"
            @click="orderTypeTab = 'buy'"
          >
            买单
          </button>
          <button
            :class="['order-tab', { active: orderTypeTab === 'sell' }]"
            @click="orderTypeTab = 'sell'"
          >
            卖单
          </button>
        </div>
        <div class="orders-list">
          <div
            v-for="order in filteredOrders"
            :key="order.id"
            :class="['order-item', order.order_type]"
          >
            <div class="order-header">
              <span class="order-type-badge">{{ ORDER_TYPE_NAMES[order.order_type] }}</span>
              <span class="order-resource">{{ RESOURCE_NAMES[order.resource] }}</span>
            </div>
            <div class="order-details">
              <div class="order-detail">
                <span class="label">单价</span>
                <span class="value">{{ order.price }}</span>
              </div>
              <div class="order-detail">
                <span class="label">数量</span>
                <span class="value">{{ order.remaining_qty }} / {{ order.quantity }}</span>
              </div>
              <div class="order-detail">
                <span class="label">状态</span>
                <span class="value">{{ ORDER_STATUS_NAMES[order.status] }}</span>
              </div>
            </div>
            <div v-if="order.player_id === currentPlayerId" class="order-actions">
              <button
                v-if="order.status === 'active' || order.status === 'partial'"
                @click="cancelOrder(order.id)"
                class="cancel-btn"
              >
                取消
              </button>
            </div>
          </div>
          <div v-if="filteredOrders.length === 0" class="empty-state">
            暂无挂单
          </div>
        </div>
      </div>

      <div v-if="activeTab === 'place'" class="tab-content">
        <div class="place-order-form">
          <div class="form-group">
            <label>订单类型</label>
            <div class="type-selector">
              <button
                :class="['type-btn', { active: newOrder.type === 'buy' }]"
                @click="newOrder.type = 'buy'"
              >
                买入
              </button>
              <button
                :class="['type-btn', { active: newOrder.type === 'sell' }]"
                @click="newOrder.type = 'sell'"
              >
                卖出
              </button>
            </div>
          </div>
          <div class="form-group">
            <label>资源类型</label>
            <select v-model="newOrder.resource" class="form-select">
              <option v-for="res in resources" :key="res" :value="res">
                {{ RESOURCE_NAMES[res] }}
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>数量</label>
            <input
              v-model.number="newOrder.quantity"
              type="number"
              min="1"
              class="form-input"
              placeholder="输入数量"
            />
          </div>
          <div class="form-group">
            <label>单价（金币）</label>
            <input
              v-model.number="newOrder.price"
              type="number"
              min="1"
              class="form-input"
              placeholder="输入单价"
            />
          </div>
          <div class="order-summary">
            <div class="summary-row">
              <span>总金额：</span>
              <span class="text-yellow-400">{{ newOrder.quantity * newOrder.price }} 金币</span>
            </div>
            <div class="summary-row">
              <span>手续费（5%）：</span>
              <span class="text-red-400">{{ Math.floor(newOrder.quantity * newOrder.price * 0.05) }} 金币</span>
            </div>
          </div>
          <button
            @click="placeOrder"
            :disabled="!canPlaceOrder"
            class="place-order-btn"
          >
            提交订单
          </button>
        </div>
      </div>

      <div v-if="activeTab === 'futures'" class="tab-content">
        <div class="futures-subtabs">
          <button
            :class="['subtab-btn', { active: futuresTab === 'market' }]"
            @click="futuresTab = 'market'"
          >
            市场合约
          </button>
          <button
            :class="['subtab-btn', { active: futuresTab === 'my' }]"
            @click="futuresTab = 'my'"
          >
            我的持仓
          </button>
          <button
            :class="['subtab-btn', { active: futuresTab === 'create' }]"
            @click="futuresTab = 'create'"
          >
            创建合约
          </button>
        </div>

        <div v-if="futuresTab === 'market'" class="futures-list">
          <div
            v-for="contract in openContracts"
            :key="contract.id"
            class="futures-card open"
          >
            <div class="futures-header">
              <span class="futures-resource">{{ RESOURCE_NAMES[contract.resource] }}</span>
              <span class="futures-status">{{ FUTURES_STATUS_NAMES[contract.status] }}</span>
            </div>
            <div class="futures-details">
              <div class="futures-detail">
                <span class="label">数量</span>
                <span class="value">{{ contract.quantity }}</span>
              </div>
              <div class="futures-detail">
                <span class="label">约定价</span>
                <span class="value">{{ contract.contract_price }}</span>
              </div>
              <div class="futures-detail">
                <span class="label">交割回合</span>
                <span class="value">{{ contract.delivery_turn }}</span>
              </div>
              <div class="futures-detail">
                <span class="label">保证金</span>
                <span class="value">{{ contract.initial_margin }}</span>
              </div>
            </div>
            <div class="futures-actions">
              <button
                v-if="contract.creator_id !== currentPlayerId"
                @click="acceptFutures(contract.id)"
                class="accept-btn"
                :disabled="gameState?.game.phase !== 'decision'"
              >
                接受(做多)
              </button>
              <button
                v-if="contract.creator_id === currentPlayerId"
                @click="cancelFutures(contract.id)"
                class="cancel-btn"
              >
                取消
              </button>
            </div>
          </div>
          <div v-if="openContracts.length === 0" class="empty-state">
            暂无市场合约
          </div>
        </div>

        <div v-if="futuresTab === 'my'" class="futures-list">
          <div
            v-for="contract in myContracts"
            :key="contract.id"
            :class="['futures-card', contract.status]"
          >
            <div class="futures-header">
              <span class="futures-resource">{{ RESOURCE_NAMES[contract.resource] }}</span>
              <div class="header-right">
                <span class="futures-role">{{ getContractRole(contract) }}</span>
                <span class="futures-status">{{ FUTURES_STATUS_NAMES[contract.status] }}</span>
              </div>
            </div>
            <div class="futures-details">
              <div class="futures-detail">
                <span class="label">数量</span>
                <span class="value">{{ contract.quantity }}</span>
              </div>
              <div class="futures-detail">
                <span class="label">约定价</span>
                <span class="value">{{ contract.contract_price }}</span>
              </div>
              <div class="futures-detail">
                <span class="label">交割回合</span>
                <span class="value">{{ contract.delivery_turn }}</span>
              </div>
              <div class="futures-detail">
                <span class="label">浮动盈亏</span>
                <span 
                  :class="['value', 'pnl', { profit: (contract.floating_pnl || 0) > 0, loss: (contract.floating_pnl || 0) < 0 }]"
                >
                  {{ (contract.floating_pnl || 0) > 0 ? '+' : '' }}{{ contract.floating_pnl || 0 }}
                </span>
              </div>
            </div>
            <div v-if="contract.status === 'active'" class="margin-status-bar">
              <span class="margin-label">保证金状态</span>
              <span 
                class="margin-indicator"
                :style="{ backgroundColor: getMarginColor(contract) }"
              >
                {{ MARGIN_STATUS_NAMES[contract.margin_status || 'safe'] }}
              </span>
            </div>
            <div class="futures-actions">
              <button
                v-if="contract.status === 'active' && contract.margin_status === 'danger'"
                @click="openAddMargin(contract.id)"
                class="margin-btn"
              >
                追加保证金
              </button>
              <button
                v-if="contract.status === 'settled' || contract.status === 'liquidated'"
                @click="showSettlement(contract)"
                class="detail-btn"
              >
                查看明细
              </button>
            </div>
          </div>
          <div v-if="myContracts.length === 0" class="empty-state">
            暂无持仓
          </div>
        </div>

        <div v-if="futuresTab === 'create'" class="create-futures-form">
          <div class="form-group">
            <label>资源类型</label>
            <select v-model="newFutures.resource" class="form-select">
              <option v-for="res in resources" :key="res" :value="res">
                {{ RESOURCE_NAMES[res] }}
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>数量</label>
            <input
              v-model.number="newFutures.quantity"
              type="number"
              min="1"
              class="form-input"
              placeholder="输入数量"
            />
          </div>
          <div class="form-group">
            <label>约定价格（金币）</label>
            <input
              v-model.number="newFutures.contract_price"
              type="number"
              min="1"
              class="form-input"
              placeholder="输入约定价格"
            />
          </div>
          <div class="form-group">
            <label>交割回合（{{ minDeliveryTurn }} - {{ maxDeliveryTurn }}）</label>
            <input
              v-model.number="newFutures.delivery_turn"
              type="number"
              :min="minDeliveryTurn"
              :max="maxDeliveryTurn"
              class="form-input"
              placeholder="输入交割回合"
            />
          </div>
          <div class="futures-summary">
            <div class="summary-row">
              <span>合约总价值：</span>
              <span class="text-yellow-400">{{ newFutures.quantity * newFutures.contract_price }} 金币</span>
            </div>
            <div class="summary-row">
              <span>所需保证金（20%）：</span>
              <span class="text-orange-400">{{ initialMargin }} 金币</span>
            </div>
            <div class="summary-row hint">
              <span>你作为做空方创建合约，若价格下跌则盈利</span>
            </div>
          </div>
          <button
            @click="createFuturesContract"
            :disabled="!canCreateFutures"
            class="create-futures-btn"
          >
            创建期货合约
          </button>
        </div>
      </div>
    </div>

    <div v-if="selectedContractForMargin" class="modal-overlay" @click.self="selectedContractForMargin = null">
      <div class="modal-content">
        <h3 class="modal-title">追加保证金</h3>
        <div class="form-group">
          <label>追加金额（金币）</label>
          <input
            v-model.number="addMarginAmount"
            type="number"
            min="1"
            class="form-input"
            placeholder="输入金额"
          />
        </div>
        <div class="modal-actions">
          <button @click="selectedContractForMargin = null" class="modal-btn secondary">
            取消
          </button>
          <button @click="submitAddMargin" class="modal-btn primary">
            确认追加
          </button>
        </div>
      </div>
    </div>

    <div v-if="showSettlementModal" class="modal-overlay" @click.self="showSettlementModal = false">
      <div class="modal-content">
        <h3 class="modal-title">结算明细</h3>
        <div v-if="selectedSettlement" class="settlement-details">
          <div class="detail-row">
            <span>资源：</span>
            <span>{{ RESOURCE_NAMES[selectedSettlement.resource] }}</span>
          </div>
          <div class="detail-row">
            <span>数量：</span>
            <span>{{ selectedSettlement.quantity }}</span>
          </div>
          <div class="detail-row">
            <span>约定价格：</span>
            <span>{{ selectedSettlement.contract_price }} 金币</span>
          </div>
          <div class="detail-row">
            <span>结算价格：</span>
            <span>{{ selectedSettlement.settlement_price }} 金币</span>
          </div>
          <div class="detail-row">
            <span>结算类型：</span>
            <span :class="{ danger: selectedSettlement.status === 'liquidated' }">
              {{ selectedSettlement.status === 'liquidated' ? '爆仓平仓' : '正常交割' }}
            </span>
          </div>
          <div class="detail-row highlight">
            <span>你的盈亏：</span>
            <span 
              :class="{ profit: (selectedSettlement.floating_pnl || 0) > 0, loss: (selectedSettlement.floating_pnl || 0) < 0 }"
            >
              {{ (selectedSettlement.floating_pnl || 0) > 0 ? '+' : '' }}{{ selectedSettlement.floating_pnl || 0 }} 金币
            </span>
          </div>
        </div>
        <div class="modal-actions">
          <button @click="showSettlementModal = false" class="modal-btn primary">
            确认
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue'
import type { ResourceType, MarketOrder, OrderType, GameState, FuturesContract, FuturesData, FuturesContractStatus, MarginStatus } from '~/types'
import { RESOURCE_NAMES, ORDER_TYPE_NAMES, ORDER_STATUS_NAMES, FUTURES_STATUS_NAMES, MARGIN_STATUS_COLORS, MARGIN_STATUS_NAMES } from '~/types'
import { useGameApi } from '~/composables/useGameApi'

const props = defineProps<{
  gameState: GameState | null
  currentPlayerId: string | null
}>()

const emit = defineEmits<{
  (e: 'placeOrder', order: { type: OrderType; resource: ResourceType; quantity: number; price: number }): void
  (e: 'cancelOrder', orderId: string): void
  (e: 'createFutures', contract: { resource: ResourceType; quantity: number; contract_price: number; delivery_turn: number }): void
  (e: 'acceptFutures', contractId: string): void
  (e: 'cancelFutures', contractId: string): void
  (e: 'addMargin', contractId: string, amount: number): void
}>()

const { getFuturesData } = useGameApi()

const futuresData = ref<FuturesData | null>(null)
const futuresTab = ref<'market' | 'my' | 'create'>('market')
const showSettlementModal = ref(false)
const selectedSettlement = ref<any>(null)

const newFutures = ref({
  resource: 'oil' as ResourceType,
  quantity: 10,
  contract_price: 50,
  delivery_turn: 0
})

const addMarginAmount = ref(100)
const selectedContractForMargin = ref<string | null>(null)

const tabs = [
  { id: 'prices', name: '价格走势' },
  { id: 'orders', name: '挂单列表' },
  { id: 'place', name: '挂单' },
  { id: 'futures', name: '期货' },
]

const currentTurn = computed(() => props.gameState?.game.current_turn || 0)

const minDeliveryTurn = computed(() => currentTurn.value + 3)
const maxDeliveryTurn = computed(() => currentTurn.value + 10)

const loadFuturesData = async () => {
  if (!props.gameState?.game.id || !props.currentPlayerId) return
  try {
    futuresData.value = await getFuturesData(props.gameState.game.id, props.currentPlayerId)
  } catch (e) {
    console.error('Failed to load futures data:', e)
  }
}

watch(() => props.gameState?.game.current_turn, () => {
  loadFuturesData()
})

watch(() => props.gameState?.game.id, () => {
  loadFuturesData()
})

onMounted(() => {
  if (newFutures.value.delivery_turn === 0) {
    newFutures.value.delivery_turn = minDeliveryTurn.value
  }
  loadFuturesData()
})

const openContracts = computed(() => {
  return futuresData.value?.contracts.filter(c => c.status === 'open')
    .sort((a, b) => a.delivery_turn - b.delivery_turn) || []
})

const myContracts = computed(() => {
  if (!props.currentPlayerId || !futuresData.value) return []
  return futuresData.value.contracts.filter(c => 
    c.creator_id === props.currentPlayerId || c.accepter_id === props.currentPlayerId
  ).sort((a, b) => {
    const statusOrder = { active: 0, open: 1, settled: 2, liquidated: 3, cancelled: 4 }
    return (statusOrder[a.status as keyof typeof statusOrder] || 0) - (statusOrder[b.status as keyof typeof statusOrder] || 0)
  })
})

const canCreateFutures = computed(() => {
  return (
    props.gameState?.game.phase === 'decision' &&
    newFutures.value.quantity > 0 &&
    newFutures.value.contract_price > 0 &&
    newFutures.value.delivery_turn >= minDeliveryTurn.value &&
    newFutures.value.delivery_turn <= maxDeliveryTurn.value
  )
})

const initialMargin = computed(() => {
  return Math.floor(newFutures.value.quantity * newFutures.value.contract_price * 0.2)
})

function createFuturesContract() {
  if (!canCreateFutures.value) return
  emit('createFutures', { ...newFutures.value })
  newFutures.value.quantity = 10
}

function acceptFutures(contractId: string) {
  emit('acceptFutures', contractId)
}

function cancelFutures(contractId: string) {
  emit('cancelFutures', contractId)
}

function openAddMargin(contractId: string) {
  selectedContractForMargin.value = contractId
  addMarginAmount.value = 100
}

function submitAddMargin() {
  if (!selectedContractForMargin.value || addMarginAmount.value <= 0) return
  emit('addMargin', selectedContractForMargin.value, addMarginAmount.value)
  selectedContractForMargin.value = null
}

function getContractRole(contract: FuturesContract): string {
  if (contract.creator_id === props.currentPlayerId) {
    return '做空方'
  } else if (contract.accepter_id === props.currentPlayerId) {
    return '做多方'
  }
  return ''
}

function getMarginColor(contract: FuturesContract): string {
  if (!contract.margin_status) return MARGIN_STATUS_COLORS.safe
  return MARGIN_STATUS_COLORS[contract.margin_status]
}

function showSettlement(settlement: any) {
  selectedSettlement.value = settlement
  showSettlementModal.value = true
}

const resources: ResourceType[] = ['oil', 'gas', 'manganese', 'sulfide', 'biomaterial']

const activeTab = ref('prices')
const orderTypeTab = ref<'buy' | 'sell'>('buy')
const selectedResource = ref<ResourceType>('oil')

const newOrder = ref({
  type: 'buy' as OrderType,
  resource: 'oil' as ResourceType,
  quantity: 10,
  price: 50,
})

const chartWidth = 260
const chartHeight = 180
const padding = 30

const currentPrice = computed(() => {
  if (!props.gameState?.current_prices) return 0
  return props.gameState.current_prices[selectedResource.value] || 0
})

const priceHistory = computed(() => {
  if (!props.gameState?.price_history) return []
  return props.gameState.price_history
    .filter(p => p.resource === selectedResource.value)
    .sort((a, b) => a.turn - b.turn)
    .slice(-20)
})

const chartPoints = computed(() => {
  const history = priceHistory.value
  if (history.length === 0) return ''

  const minPrice = Math.min(...history.map(p => p.price)) * 0.9
  const maxPrice = Math.max(...history.map(p => p.price)) * 1.1
  const priceRange = maxPrice - minPrice || 1

  const points = history.map((entry, idx) => {
    const x = padding + idx * (chartWidth - 2 * padding) / Math.max(history.length - 1, 1)
    const y = chartHeight - padding - ((entry.price - minPrice) / priceRange) * (chartHeight - 2 * padding)
    return `${x},${y}`
  })

  return points.join(' ')
})

const labeledPoints = computed(() => {
  const history = priceHistory.value
  if (history.length === 0) return []

  const minPrice = Math.min(...history.map(p => p.price)) * 0.9
  const maxPrice = Math.max(...history.map(p => p.price)) * 1.1
  const priceRange = maxPrice - minPrice || 1

  return history
    .filter((_, idx) => idx % 4 === 0 || idx === history.length - 1)
    .map((entry, i) => {
      const idx = history.indexOf(entry)
      const x = padding + idx * (chartWidth - 2 * padding) / Math.max(history.length - 1, 1)
      const y = chartHeight - padding - ((entry.price - minPrice) / priceRange) * (chartHeight - 2 * padding)
      return { x, y, price: entry.price }
    })
})

const areaPoints = computed(() => {
  const history = priceHistory.value
  if (history.length === 0) return ''

  const minPrice = Math.min(...history.map(p => p.price)) * 0.9
  const maxPrice = Math.max(...history.map(p => p.price)) * 1.1
  const priceRange = maxPrice - minPrice || 1

  const points: string[] = []
  points.push(`${padding},${chartHeight - padding}`)
  
  history.forEach((entry, idx) => {
    const x = padding + idx * (chartWidth - 2 * padding) / Math.max(history.length - 1, 1)
    const y = chartHeight - padding - ((entry.price - minPrice) / priceRange) * (chartHeight - 2 * padding)
    points.push(`${x},${y}`)
  })
  
  points.push(`${chartWidth - padding},${chartHeight - padding}`)
  return points.join(' ')
})

const allOrders = computed(() => {
  if (!props.gameState?.market_orders) return []
  return props.gameState.market_orders.filter(o => o.status === 'active' || o.status === 'partial')
})

const filteredOrders = computed(() => {
  return allOrders.value
    .filter(o => o.order_type === orderTypeTab.value)
    .sort((a, b) => {
      if (orderTypeTab.value === 'buy') {
        return b.price - a.price
      }
      return a.price - b.price
    })
})

const canPlaceOrder = computed(() => {
  return (
    newOrder.value.quantity > 0 &&
    newOrder.value.price > 0 &&
    props.gameState?.game.phase === 'decision'
  )
})

function placeOrder() {
  if (!canPlaceOrder.value) return
  emit('placeOrder', { ...newOrder.value })
  newOrder.value.quantity = 10
}

function cancelOrder(orderId: string) {
  emit('cancelOrder', orderId)
}
</script>

<style scoped>
.market-panel {
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
  background: #3b82f6;
  border-color: #3b82f6;
  color: white;
}

.tab-content {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.resource-selector {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.res-btn {
  padding: 4px 8px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 4px;
  color: #94a3b8;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s;
}

.res-btn:hover {
  background: #334155;
}

.res-btn.active {
  background: #10b981;
  border-color: #10b981;
  color: white;
}

.chart-container {
  background: #0f172a;
  border-radius: 6px;
  padding: 8px;
  overflow: hidden;
}

.price-chart {
  width: 100%;
  height: auto;
}

.current-price-info {
  text-align: center;
  padding: 8px;
  background: #1e293b;
  border-radius: 6px;
  font-size: 14px;
}

.order-tabs {
  display: flex;
  gap: 4px;
}

.order-tab {
  flex: 1;
  padding: 6px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 4px;
  color: #94a3b8;
  font-size: 12px;
  cursor: pointer;
}

.order-tab.active {
  background: #0ea5e9;
  border-color: #0ea5e9;
  color: white;
}

.orders-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 300px;
  overflow-y: auto;
}

.order-item {
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  padding: 10px;
}

.order-item.buy {
  border-left: 3px solid #10b981;
}

.order-item.sell {
  border-left: 3px solid #ef4444;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.order-type-badge {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 10px;
  font-weight: bold;
}

.order-item.buy .order-type-badge {
  background: rgba(16, 185, 129, 0.2);
  color: #10b981;
}

.order-item.sell .order-type-badge {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.order-resource {
  color: #f1f5f9;
  font-size: 12px;
  font-weight: 500;
}

.order-details {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 6px;
}

.order-detail {
  display: flex;
  flex-direction: column;
}

.order-detail .label {
  color: #64748b;
  font-size: 10px;
}

.order-detail .value {
  color: #e2e8f0;
  font-size: 12px;
  font-weight: 500;
}

.order-actions {
  margin-top: 8px;
  text-align: right;
}

.cancel-btn {
  padding: 4px 10px;
  background: #7f1d1d;
  border: none;
  border-radius: 4px;
  color: #fecaca;
  font-size: 11px;
  cursor: pointer;
}

.cancel-btn:hover {
  background: #991b1b;
}

.empty-state {
  text-align: center;
  padding: 20px;
  color: #64748b;
  font-size: 12px;
}

.place-order-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
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
  border-color: #3b82f6;
  background: rgba(59, 130, 246, 0.1);
  color: #93c5fd;
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
  border-color: #3b82f6;
}

.order-summary {
  background: #0f172a;
  border-radius: 6px;
  padding: 10px;
}

.summary-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #94a3b8;
  margin-bottom: 4px;
}

.summary-row:last-child {
  margin-bottom: 0;
}

.place-order-btn {
  padding: 10px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  border: none;
  border-radius: 6px;
  color: white;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.place-order-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.place-order-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.futures-subtabs {
  display: flex;
  gap: 4px;
  margin-bottom: 12px;
}

.subtab-btn {
  flex: 1;
  padding: 6px 8px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 4px;
  color: #94a3b8;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s;
}

.subtab-btn:hover {
  background: #334155;
}

.subtab-btn.active {
  background: #8b5cf6;
  border-color: #8b5cf6;
  color: white;
}

.futures-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 350px;
  overflow-y: auto;
}

.futures-card {
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 6px;
  padding: 10px;
}

.futures-card.open {
  border-left: 3px solid #8b5cf6;
}

.futures-card.active {
  border-left: 3px solid #3b82f6;
}

.futures-card.settled {
  border-left: 3px solid #10b981;
  opacity: 0.85;
}

.futures-card.liquidated {
  border-left: 3px solid #ef4444;
  opacity: 0.85;
}

.futures-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.header-right {
  display: flex;
  gap: 6px;
  align-items: center;
}

.futures-resource {
  color: #f1f5f9;
  font-size: 13px;
  font-weight: 600;
}

.futures-role {
  padding: 2px 6px;
  background: rgba(59, 130, 246, 0.2);
  color: #60a5fa;
  border-radius: 3px;
  font-size: 10px;
  font-weight: 500;
}

.futures-status {
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 10px;
  font-weight: 500;
}

.futures-card.open .futures-status {
  background: rgba(139, 92, 246, 0.2);
  color: #a78bfa;
}

.futures-card.active .futures-status {
  background: rgba(59, 130, 246, 0.2);
  color: #60a5fa;
}

.futures-card.settled .futures-status {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
}

.futures-card.liquidated .futures-status {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
}

.futures-details {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 6px;
  margin-bottom: 8px;
}

.futures-detail {
  display: flex;
  flex-direction: column;
}

.futures-detail .label {
  color: #64748b;
  font-size: 10px;
}

.futures-detail .value {
  color: #e2e8f0;
  font-size: 12px;
  font-weight: 500;
}

.futures-detail .value.pnl.profit {
  color: #10b981;
}

.futures-detail .value.pnl.loss {
  color: #ef4444;
}

.margin-status-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 8px;
  background: #0f172a;
  border-radius: 4px;
  margin-bottom: 8px;
}

.margin-label {
  color: #64748b;
  font-size: 11px;
}

.margin-indicator {
  padding: 2px 8px;
  border-radius: 3px;
  color: white;
  font-size: 10px;
  font-weight: 600;
}

.futures-actions {
  display: flex;
  gap: 6px;
  justify-content: flex-end;
}

.accept-btn {
  padding: 4px 10px;
  background: linear-gradient(135deg, #8b5cf6, #7c3aed);
  border: none;
  border-radius: 4px;
  color: white;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.accept-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(139, 92, 246, 0.4);
}

.accept-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.margin-btn {
  padding: 4px 10px;
  background: linear-gradient(135deg, #f59e0b, #d97706);
  border: none;
  border-radius: 4px;
  color: white;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.margin-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(245, 158, 11, 0.4);
}

.detail-btn {
  padding: 4px 10px;
  background: #334155;
  border: none;
  border-radius: 4px;
  color: #e2e8f0;
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s;
}

.detail-btn:hover {
  background: #475569;
}

.create-futures-form {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.futures-summary {
  background: #0f172a;
  border-radius: 6px;
  padding: 10px;
}

.futures-summary .summary-row {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #94a3b8;
  margin-bottom: 4px;
}

.futures-summary .summary-row:last-child {
  margin-bottom: 0;
}

.futures-summary .summary-row.hint {
  font-size: 11px;
  color: #64748b;
  font-style: italic;
  justify-content: center;
  padding-top: 4px;
  border-top: 1px solid #1e293b;
  margin-top: 6px;
}

.create-futures-btn {
  padding: 10px;
  background: linear-gradient(135deg, #8b5cf6, #7c3aed);
  border: none;
  border-radius: 6px;
  color: white;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.create-futures-btn:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.4);
}

.create-futures-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 20px;
  min-width: 300px;
  max-width: 90vw;
}

.modal-title {
  color: #f1f5f9;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
}

.modal-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 16px;
}

.modal-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.modal-btn.primary {
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  color: white;
}

.modal-btn.primary:hover {
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.4);
}

.modal-btn.secondary {
  background: #334155;
  color: #e2e8f0;
}

.modal-btn.secondary:hover {
  background: #475569;
}

.settlement-details {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: #94a3b8;
}

.detail-row span:last-child {
  color: #e2e8f0;
  font-weight: 500;
}

.detail-row.highlight {
  padding: 8px;
  background: #0f172a;
  border-radius: 4px;
  margin-top: 4px;
}

.detail-row.highlight span:last-child.profit {
  color: #10b981;
  font-size: 15px;
}

.detail-row.highlight span:last-child.loss {
  color: #ef4444;
  font-size: 15px;
}

.detail-row .danger {
  color: #ef4444;
}
</style>
