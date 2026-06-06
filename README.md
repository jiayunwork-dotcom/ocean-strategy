# 🌊 深海策略 - Ocean Strategy

多人回合制深海资源开采策略游戏

## 项目概述

4-8个玩家在随机生成的六角格海洋地图上争夺海底资源，通过建设开采设施、组建船队、外交博弈争夺海域控制权。

## 技术栈

### 后端
- Go 1.21 + Fiber 框架
- WebSocket 实时同步
- PostgreSQL 16 数据持久化
- Redis 7 会话管理

### 前端
- Nuxt 3 + Vue 3
- TypeScript
- Tailwind CSS
- Pinia 状态管理

### 部署
- Docker + Docker Compose
- Nginx 静态资源服务

## 项目结构

```
ocean-strategy/
├── backend/                 # Go 后端服务
│   ├── cmd/server/          # 服务入口
│   ├── internal/
│   │   ├── game/            # 游戏核心逻辑
│   │   ├── models/          # 数据模型
│   │   ├── handlers/        # HTTP/WebSocket 处理器
│   │   ├── websocket/       # WebSocket Hub
│   │   ├── database/        # PostgreSQL 连接
│   │   └── cache/           # Redis 连接
│   └── Dockerfile
├── frontend/                # Nuxt 3 前端
│   ├── components/          # Vue 组件
│   ├── pages/               # 页面路由
│   ├── stores/              # Pinia 状态
│   ├── composables/       # 组合式函数
│   ├── types/               # TypeScript 类型
│   └── Dockerfile
└── docker-compose.yml       # 容器编排
```

## 游戏特性

### 🗺️ 地图系统
- 六角格随机生成地图
- 6种海域类型：浅海、深海、海沟、珊瑚礁、热液喷口、公海
- 5种资源：石油、天然气、锰结核、多金属硫化物、生物制药原料

### 🏗️ 设施系统
- 钻井平台：浅海开采石油天然气
- 海底矿山：深海开采锰结核和硫化物
- 潮汐发电站：利用洋流产生电力
- 养殖场：珊瑚礁培育海洋生物
- 港口：船只建造基地

### 🚢 船队系统
- 勘探船：探测未开发海域
- 工程船：建设远海建设设施
- 运输船：运送物资
- 护卫舰：保护运输线路

### 🌪️ 环境系统
- 台风：随机生成并移动，损坏设施和船只
- 洋流：影响船只移动成本
- 生态系统：珊瑚礁健康值影响产出
- 污染：开采设施向周围扩散污染

### 🔬 科技系统
三大科技方向，共9种科技：
- 开采科技：深海钻探、海沟采矿
- 环保科技：污染控制、可持续养殖
- 军事科技：舰船装甲、舰载武器、抗台风

## 快速开始

### 使用 Docker Compose 启动

```bash
# 克隆项目
cd ocean-strategy

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps
```

启动后访问：
- 前端: http://localhost:3000
- 后端 API: http://localhost:8080
- WebSocket: ws://localhost:8080/ws

### 本地开发

#### 后端开发

```bash
cd backend
go mod download
go run cmd/server/main.go
```

#### 前端开发

```bash
cd frontend
npm install
npm run dev
```

## API 接口

### 游戏管理
- `GET /api/v1/games` - 获取游戏列表
- `POST /api/v1/games` - 创建游戏
- `GET /api/v1/games/:id` - 获取游戏状态
- `POST /api/v1/games/:id/join` - 加入游戏
- `POST /api/v1/games/:id/start` - 开始游戏
- `POST /api/v1/games/:id/next-phase` - 下一阶段

### 游戏操作
- `POST /api/v1/games/:id/facilities` - 建造设施
- `POST /api/v1/games/:id/ships` - 建造船只
- `POST /api/v1/games/:id/ships/move` - 移动船只
- `POST /api/v1/games/:id/ships/explore` - 勘探海域
- `POST /api/v1/games/:id/research` - 研发科技

### 其他
- `GET /api/v1/technologies` - 获取科技列表
- `GET /health` - 健康检查

## 游戏流程

每个回合分为4个阶段：
1. **生产阶段**：所有设施自动产出资源
2. **决策阶段**：玩家下达建设、派遣、研发、外交指令
3. **事件阶段**：台风移动、环境变化生效
4. **结算阶段**：计算收入支出、维护成本

## 胜利条件

- 经济总量第一
- 控制海域面积最大
- 科技全部研发完成
- 外交手段让多数玩家成为附庸
- 回合数上限超时按综合评分判定

## 开发说明

### 后端代码规范
- 遵循 Go 官方代码规范
- 使用 GORM 进行数据库操作
- WebSocket 使用 Fiber websocket 中间件

### 前端代码规范
- 使用 Composition API + TypeScript
- 组件命名使用 PascalCase
- 状态管理使用 Pinia
- 样式使用 Tailwind CSS

## License

MIT
