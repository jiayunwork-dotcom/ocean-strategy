<template>
  <div class="home-page">
    <div class="hero-section">
      <div class="hero-content">
        <h1 class="hero-title">🌊 深海策略</h1>
        <p class="hero-subtitle">多人回合制深海资源开采策略游戏</p>
        <p class="hero-desc">
          在广袤的海洋地图上，与其他玩家争夺海底资源。<br/>
          建设开采设施、组建船队、外交博弈，成为海洋霸主！
        </p>
        <div class="hero-actions">
          <button class="btn primary" @click="showCreateDialog = true">
            🎮 创建游戏
          </button>
          <button class="btn secondary" @click="navigateTo('/lobby')">
            📋 游戏大厅
          </button>
        </div>
      </div>
    </div>
    
    <div class="features-section">
      <div class="features-grid">
        <div class="feature-card">
          <span class="feature-icon">🗺️</span>
          <h3 class="feature-title">随机生成地图</h3>
          <p class="feature-desc">六角格构成的海洋地图，浅海、深海、海沟、珊瑚礁、热液喷口等多种地形</p>
        </div>
        <div class="feature-card">
          <span class="feature-icon">🏗️</span>
          <h3 class="feature-title">设施建造</h3>
          <p class="feature-desc">钻井平台、海底矿山、潮汐发电站、养殖场等多种设施类型</p>
        </div>
        <div class="feature-card">
          <span class="feature-icon">🚢</span>
          <h3 class="feature-title">船队系统</h3>
          <p class="feature-desc">勘探船、工程船、运输船、护卫舰，组建你的海上力量</p>
        </div>
        <div class="feature-card">
          <span class="feature-icon">🌪️</span>
          <h3 class="feature-title">环境系统</h3>
          <p class="feature-desc">台风、洋流、生态系统，真实的海洋环境变化</p>
        </div>
        <div class="feature-card">
          <span class="feature-icon">🤝</span>
          <h3 class="feature-title">外交博弈</h3>
          <p class="feature-desc">签订条约、组建同盟、资源共享，或用武力强占</p>
        </div>
        <div class="feature-card">
          <span class="feature-icon">🔬</span>
          <h3 class="feature-title">科技研发</h3>
          <p class="feature-desc">三大科技方向，解锁更深水域和更高效的技术</p>
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
const gameApi = useGameApi()
const router = useRouter()

const showCreateDialog = ref(false)
const newGameName = ref('我的深海游戏')
const maxTurns = ref(50)
const mapRadius = ref(6)
const winCondition = ref('economic')

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
.home-page {
  min-height: 100vh;
  background: linear-gradient(180deg, #0a1628 0%, #1a3a5c 50%, #0f2942 100%);
}

.hero-section {
  padding: 80px 20px;
  text-align: center;
}

.hero-content {
  max-width: 800px;
  margin: 0 auto;
}

.hero-title {
  font-size: 56px;
  font-weight: bold;
  color: white;
  margin-bottom: 16px;
  text-shadow: 0 4px 20px rgba(59, 130, 246, 0.5);
}

.hero-subtitle {
  font-size: 24px;
  color: #60a5fa;
  margin-bottom: 24px;
}

.hero-desc {
  font-size: 16px;
  color: #94a3b8;
  line-height: 1.8;
  margin-bottom: 40px;
}

.hero-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
}

.btn {
  padding: 14px 32px;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  border: none;
  transition: all 0.3s;
}

.btn.primary {
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  color: white;
  box-shadow: 0 4px 20px rgba(59, 130, 246, 0.4);
}

.btn.primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 30px rgba(59, 130, 246, 0.6);
}

.btn.secondary {
  background: rgba(71, 85, 105, 0.8);
  color: white;
}

.btn.secondary:hover {
  background: rgba(71, 85, 105, 1);
}

.features-section {
  padding: 60px 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 24px;
}

.feature-card {
  background: rgba(15, 23, 42, 0.8);
  padding: 32px 24px;
  border-radius: 16px;
  border: 1px solid rgba(59, 130, 246, 0.2);
  transition: all 0.3s;
}

.feature-card:hover {
  transform: translateY(-4px);
  border-color: rgba(59, 130, 246, 0.5);
  box-shadow: 0 8px 30px rgba(59, 130, 246, 0.2);
}

.feature-icon {
  font-size: 40px;
  margin-bottom: 16px;
}

.feature-title {
  font-size: 20px;
  font-weight: 600;
  color: white;
  margin-bottom: 12px;
}

.feature-desc {
  color: #94a3b8;
  font-size: 14px;
  line-height: 1.6;
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
