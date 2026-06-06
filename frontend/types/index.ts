export type HexTerrain = 'shallow' | 'deep' | 'trench' | 'reef' | 'vent' | 'open_ocean'
export type ResourceType = 'oil' | 'gas' | 'manganese' | 'sulfide' | 'biomaterial'
export type FacilityType = 'drilling' | 'mine' | 'tidal' | 'farm' | 'port'
export type ShipType = 'explorer' | 'constructor' | 'transport' | 'escort'
export type GamePhase = 'production' | 'decision' | 'event' | 'settlement'
export type GameStatus = 'waiting' | 'playing' | 'finished'
export type TechnologyCategory = 'extraction' | 'ecology' | 'military'
export type RelationStatus = 'neutral' | 'nap' | 'alliance' | 'hostile'
export type TreatyType = 'nap' | 'alliance'
export type ProposalStatus = 'pending' | 'accepted' | 'rejected'

export interface Player {
  id: string
  game_id: string
  name: string
  color: string
  money: number
  reputation: number
  start_hex_q: number
  start_hex_r: number
  is_ai: boolean
  created_at: string
  updated_at: string
}

export interface Hex {
  q: number
  r: number
  terrain: HexTerrain
  resources: Record<ResourceType, number>
  discovered: boolean
  owner_id?: string
  facility?: Facility
  ecological_health: number
  pollution: number
  has_current: boolean
  current_dir: number
  is_eez: boolean
  eez_owner_id?: string
}

export interface Facility {
  id: string
  type: FacilityType
  level: number
  hex_q: number
  hex_r: number
  owner_id: string
  health: number
  max_health: number
  maintenance_cost: number
  build_turns_left: number
  is_active: boolean
  power_output: number
  power_consume: number
}

export interface Ship {
  id: string
  type: ShipType
  owner_id: string
  hex_q: number
  hex_r: number
  health: number
  max_health: number
  fuel: number
  max_fuel: number
  cargo: Record<ResourceType, number>
  cargo_capacity: number
  speed: number
  move_points: number
  attack: number
  defense: number
}

export interface Technology {
  id: string
  name: string
  category: TechnologyCategory
  cost: number
  turns: number
  description: string
  prerequisites: string[]
  effects: Record<string, number>
}

export interface PlayerTech {
  player_id: string
  tech_id: string
  researching: boolean
  turns_left: number
  completed: boolean
}

export interface Typhoon {
  id: string
  hex_q: number
  hex_r: number
  strength: number
  dir_q: number
  dir_r: number
  turns_left: number
}

export interface DiplomaticRelation {
  game_id: string
  player1_id: string
  player2_id: string
  status: RelationStatus
  has_nap: boolean
  has_alliance: boolean
  at_war: boolean
}

export interface DiplomaticProposal {
  id: string
  game_id: string
  from_player_id: string
  to_player_id: string
  treaty_type: TreatyType
  status: ProposalStatus
  created_at: number
}

export interface ReputationCooldown {
  player_id: string
  game_id: string
  turns_left: number
  reason: string
}

export interface BattleLog {
  id: string
  game_id: string
  turn: number
  attacker_id: string
  defender_id: string
  attacker_ship_id: string
  defender_ship_id: string
  hex_q: number
  hex_r: number
  attacker_damage: number
  defender_damage: number
  attacker_sunk: boolean
  defender_sunk: boolean
  timestamp: string
}

export interface GameLogEntry {
  id: string
  game_id: string
  turn: number
  message: string
  type: string
  player_id?: string
  timestamp: string
}

export interface Game {
  id: string
  name: string
  status: GameStatus
  current_turn: number
  max_turns: number
  phase: GamePhase
  map_radius: number
  players: Player[]
  current_player_index: number
  created_at: string
  updated_at: string
  win_condition: string
  winner_id?: string
}

export interface GameState {
  game: Game
  hexes: Record<string, Hex>
  players: Record<string, Player>
  ships: Ship[]
  facilities: Facility[]
  techs: PlayerTech[]
  relations: DiplomaticRelation[]
  proposals: DiplomaticProposal[]
  cooldowns: ReputationCooldown[]
  battle_logs: BattleLog[]
  game_logs: GameLogEntry[]
  typhoons: Typhoon[]
}

export const TERRAIN_COLORS: Record<HexTerrain, string> = {
  shallow: '#5DADE2',
  deep: '#2874A6',
  trench: '#1B4F72',
  reef: '#58D68D',
  vent: '#E67E22',
  open_ocean: '#3498DB'
}

export const TERRAIN_NAMES: Record<HexTerrain, string> = {
  shallow: '浅海',
  deep: '深海',
  trench: '海沟',
  reef: '珊瑚礁',
  vent: '热液喷口',
  open_ocean: '公海'
}

export const RESOURCE_NAMES: Record<ResourceType, string> = {
  oil: '石油',
  gas: '天然气',
  manganese: '锰结核',
  sulfide: '多金属硫化物',
  biomaterial: '生物原料'
}

export const FACILITY_NAMES: Record<FacilityType, string> = {
  drilling: '钻井平台',
  mine: '海底矿山',
  tidal: '潮汐发电站',
  farm: '养殖场',
  port: '港口'
}

export const SHIP_NAMES: Record<ShipType, string> = {
  explorer: '勘探船',
  constructor: '工程船',
  transport: '运输船',
  escort: '护卫舰'
}

export const PHASE_NAMES: Record<GamePhase, string> = {
  production: '生产阶段',
  decision: '决策阶段',
  event: '事件阶段',
  settlement: '结算阶段'
}

export const RELATION_STATUS_NAMES: Record<RelationStatus, string> = {
  neutral: '中立',
  nap: '互不侵犯',
  alliance: '军事同盟',
  hostile: '敌对'
}

export const RELATION_STATUS_COLORS: Record<RelationStatus, string> = {
  neutral: '#94a3b8',
  nap: '#22c55e',
  alliance: '#3b82f6',
  hostile: '#ef4444'
}

export const TREATY_TYPE_NAMES: Record<TreatyType, string> = {
  nap: '互不侵犯条约',
  alliance: '军事同盟'
}
