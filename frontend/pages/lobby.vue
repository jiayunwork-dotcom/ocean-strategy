<template>
  <div class="lobby-page">
    <div class="lobby-header">
      <button class="back-btn" @click="navigateTo('/')">← 返回首页</button>
      <h1 class="lobby-title">🎮 游戏大厅</h1>
      <button class="create-btn" @click="showCreateDialog = true">+ 创建游戏</button>
    </div>
    
    <div class="lobby-content">
      <div v-if="games.length === 0" class="empty-state">
        <span class="empty-icon">🌊</span>
        <p class="empty-text">暂无可用游戏</p>
        <p class="empty-desc">创建一个新游戏开始深海冒险吧！</p>
        <button class="btn primary" @click="showCreateDialog = true">创建游戏</button>
      </div>
      
      <div v-else class="games-grid">
        <div v-for="game in games" :key="game.id" class="game-card">
          <div class="game-card-header">
            <h3 class="game-name">{{ game.name }}</h3>
            <span class="game-status" :class="game.status">{{ getStatusText(game.status) }}</span>
          </div>
          <div class="game-info">
            <div class="info-item">
              <span class="info-label">玩家</span>
              <span class="info-value">{{ game.players?.length || 0 }} / 8</span>
            </div>
            <div class="info-item">
              <span class="info-label">回合</span>
              <span class="info-value">{{ game.current_turn }} / {{ game.max_turns }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">地图</span>
              <span class="info-value">半径 {{ game.map_radius }}</span>
            </div>
          </div>
          <div class="game-actions">
            <button
              v-if="game.status === 'waiting'"
              class="btn join-btn"
              @click="joinGame(game.id)"
            >
              加入游戏
            </button>
            <button
              v-else
              class="btn view-btn"
              @click="viewGame(game.id)"
            >
              查看游戏
            </button>
          </div>
        </div>
      </div>
    </div>
    
    <Teleport to="body">
      <div v-if="showCreateDialog" class="modal-overlay" @click.self="showCreateDialog = false">
        <div class="modal-content">
          <h2 class="modal-title">创建新游戏</h2>
          <div class="form-group">
            <label>游戏名称</label>
            <input v-model="newGameName" type="text" placeholder="输入游戏名称" class="form-input" />
          </div>
          <div class="form-group">
            <label>最大回合数</label>
            <input v-model.number="maxTurns" type="number" min="10" max="200" class="form-input" />
          </div>
          <div class="form-group">
            <label>地图大小（半径）</label>
            <input v-model.number="mapRadius" type="number" min="4" max="10" class="form-input" />
          </div>
          <div class="form-group">
            <label>胜利条件</label>
            <select v-model="winCondition" class="form-input">
              <option value="economic">💰 经济总量第一</option>
              <option value="territory">🗺️ 控制海域面积最大</option>
              <option value="technology">🔬 科技全部研发完成</option>
              <option value="diplomatic">🤝 外交声誉最高</option>
              <option value="comprehensive">📊 综合评分</option>
            </select>
          </div>
          <div class="form-actions">
            <button class="btn secondary" @click="showCreateDialog = false">取消</button>
            <button class="btn primary" @click="createGame">创建游戏</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import type { Game } from '~/types'

const gameApi = useGameApi()
const router = useRouter()

const games = ref<Game[]>([])
const showCreateDialog = ref(false)
const newGameName = ref('我的深海游戏')
const maxTurns = ref(50)
const mapRadius = ref(6)
const winCondition = ref('economic')

onMounted(async () => {
  await loadGames()
})

async function loadGames() {
  try {
    games.value = await gameApi.listGames()
  } catch (e) {
    console.error('Load games failed:', e)
  }
}

function getStatusText(status: string): string {
  const texts: Record<string, string> = {
    waiting: '等待中',
    playing: '进行中',
    finished: '已结束'
  }
  return texts[status] || status
}

async function joinGame(gameId: string) {
  try {
    const color = '#' + Math.floor(Math.random()*16777215).toString(16).padStart(6, '0')
    const player = await gameApi.joinGame(gameId, '玩家', color)
    
    localStorage.setItem('playerId', player.id)
    localStorage.setItem('gameId', gameId)
    
    router.push(`/game/${gameId}`)
  } catch (e) {
    console.error('Join game failed:', e)
  }
}

function viewGame(gameId: string) {
  router.push(`/game/${gameId}`)
}

async function createGame() {
  try {
    const game = await gameApi.createGame(newGameName.value, maxTurns.value, mapRadius.value, winCondition.value)
    
    const color = '#' + Math.floor(Math.random()*16777215).toString(16).padStart(6, '0')
    const player = await gameApi.joinGame(game.id, '玩家1', color)
    
    localStorage.setItem('playerId', player.id)
    localStorage.setItem('gameId', game.id)
    
    showCreateDialog.value = false
    router.push(`/game/${game.id}`)
  } catch (e) {
    console.error('Create game failed:', e)
  }
}
</script>

<style scoped>
.lobby-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #0a1628 0%, #1a3a5c 100%);
  padding: 20px;
}

.lobby-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  max-width: 1200px;
  margin: 0 auto 32px;
}

.back-btn {
  padding: 10px 20px;
  background: rgba(71, 85, 105, 0.6);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
}

.back-btn:hover {
  background: rgba(71, 85, 105, 0.9);
}

.lobby-title {
  font-size: 32px;
  font-weight: bold;
  color: white;
}

.create-btn {
  padding: 10px 20px;
  background: #22c55e;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
}

.create-btn:hover {
  background: #16a34a;
}

.lobby-content {
  max-width: 1200px;
  margin: 0 auto;
}

.empty-state {
  text-align: center;
  padding: 80px 20px;
}

.empty-icon {
  font-size: 64px;
  display: block;
  margin-bottom: 16px;
}

.empty-text {
  font-size: 24px;
  color: white;
  margin-bottom: 8px;
}

.empty-desc {
  color: #94a3b8;
  margin-bottom: 24px;
}

.games-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.game-card {
  background: rgba(15, 23, 42, 0.9);
  border-radius: 12px;
  padding: 20px;
  border: 1px solid rgba(59, 130, 246, 0.2);
  transition: all 0.3s;
}

.game-card:hover {
  transform: translateY(-4px);
  border-color: rgba(59, 130, 246, 0.5);
}

.game-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.game-name {
  font-size: 18px;
  font-weight: 600;
  color: white;
}

.game-status {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;
}

.game-status.waiting {
  background: rgba(234, 179, 8, 0.2);
  color: #eab308;
}

.game-status.playing {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
}

.game-status.finished {
  background: rgba(107, 114, 128, 0.2);
  color: #9ca3af;
}

.game-info {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
  margin-bottom: 16px;
}

.info-item {
  text-align: center;
}

.info-label {
  display: block;
  font-size: 12px;
  color: #64748b;
  margin-bottom: 4px;
}

.info-value {
  display: block;
  font-size: 14px;
  font-weight: 600;
  color: white;
}

.game-actions {
  display: flex;
  gap: 8px;
}

.btn {
  flex: 1;
  padding: 10px;
  border: none;
  border-radius: 8px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn.primary {
  background: #3b82f6;
  color: white;
}

.join-btn {
  background: #22c55e;
  color: white;
}

.join-btn:hover {
  background: #16a34a;
}

.view-btn {
  background: rgba(59, 130, 246, 0.2);
  color: #60a5fa;
}

.view-btn:hover {
  background: rgba(59, 130, 246, 0.4);
}

.btn.secondary {
  background: rgba(71, 85, 105, 0.8);
  color: white;
}

.btn.secondary:hover {
  background: rgba(71, 85, 105, 1);
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: #1e293b;
  padding: 32px;
  border-radius: 16px;
  min-width: 400px;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.modal-title {
  font-size: 24px;
  font-weight: bold;
  color: white;
  margin-bottom: 24px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  color: #cbd5e1;
  margin-bottom: 8px;
  font-size: 14px;
}

.form-input {
  width: 100%;
  padding: 12px 16px;
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 8px;
  color: white;
  font-size: 14px;
}

.form-input:focus {
  outline: none;
  border-color: #3b82f6;
}

.form-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  margin-top: 24px;
}
</style>
