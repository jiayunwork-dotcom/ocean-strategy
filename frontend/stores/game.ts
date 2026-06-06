import type { GameState, Player, Hex, Ship, Facility, PlayerTech, Typhoon, DiplomaticRelation, DiplomaticProposal, GameLogEntry } from '~/types'
import { hexKey } from '~/utils/hex'

export const useGameStore = defineStore('game', () => {
  const gameState = ref<GameState | null>(null)
  const currentPlayerId = ref<string | null>(null)
  const selectedHex = ref<{ q: number; r: number } | null>(null)
  const selectedShip = ref<Ship | null>(null)
  const selectedFacility = ref<Facility | null>(null)
  const activeTab = ref<string>('info')
  const showProposalModal = ref(false)
  const pendingProposals = ref<DiplomaticProposal[]>([])

  const currentPlayer = computed(() => {
    if (!gameState.value || !currentPlayerId.value) return null
    return gameState.value.players[currentPlayerId.value] || null
  })

  const playerShips = computed(() => {
    if (!gameState.value || !currentPlayerId.value) return []
    return gameState.value.ships.filter(s => s.owner_id === currentPlayerId.value)
  })

  const playerFacilities = computed(() => {
    if (!gameState.value || !currentPlayerId.value) return []
    return gameState.value.facilities.filter(f => f.owner_id === currentPlayerId.value)
  })

  const playerTechs = computed(() => {
    if (!gameState.value || !currentPlayerId.value) return []
    return gameState.value.techs.filter(t => t.player_id === currentPlayerId.value)
  })

  const otherPlayers = computed(() => {
    if (!gameState.value || !currentPlayerId.value) return []
    return Object.values(gameState.value.players).filter(p => p.id !== currentPlayerId.value)
  })

  const relations = computed(() => {
    if (!gameState.value) return []
    return gameState.value.relations
  })

  const proposals = computed(() => {
    if (!gameState.value) return []
    return gameState.value.proposals
  })

  const gameLogs = computed(() => {
    if (!gameState.value) return []
    return [...gameState.value.game_logs].reverse()
  })

  const myPendingProposals = computed(() => {
    if (!gameState.value || !currentPlayerId.value) return []
    return gameState.value.proposals.filter(
      p => p.to_player_id === currentPlayerId.value && p.status === 'pending'
    )
  })

  function getRelationWithPlayer(otherPlayerId: string): DiplomaticRelation | null {
    if (!gameState.value || !currentPlayerId.value) return null
    return relations.value.find(
      r =>
        (r.player1_id === currentPlayerId.value && r.player2_id === otherPlayerId) ||
        (r.player1_id === otherPlayerId && r.player2_id === currentPlayerId.value)
    ) || null
  }

  function hasCooldown(playerId: string): boolean {
    if (!gameState.value) return false
    return gameState.value.cooldowns.some(cd => cd.player_id === playerId && cd.turns_left > 0)
  }

  function getCooldown(playerId: string): number {
    if (!gameState.value) return 0
    const cd = gameState.value.cooldowns.find(c => c.player_id === playerId)
    return cd ? cd.turns_left : 0
  }

  function setGameState(state: GameState) {
    gameState.value = state
    if (currentPlayerId.value) {
      pendingProposals.value = state.proposals.filter(
        p => p.to_player_id === currentPlayerId.value && p.status === 'pending'
      )
      if (pendingProposals.value.length > 0) {
        showProposalModal.value = true
      }
    }
  }

  function setCurrentPlayer(playerId: string) {
    currentPlayerId.value = playerId
  }

  function selectHex(q: number, r: number) {
    selectedHex.value = { q, r }
    selectedShip.value = null
    selectedFacility.value = null
  }

  function selectShip(ship: Ship) {
    selectedShip.value = ship
    selectedHex.value = { q: ship.hex_q, r: ship.hex_r }
    selectedFacility.value = null
  }

  function selectFacility(facility: Facility) {
    selectedFacility.value = facility
    selectedHex.value = { q: facility.hex_q, r: facility.hex_r }
    selectedShip.value = null
  }

  function clearSelection() {
    selectedHex.value = null
    selectedShip.value = null
    selectedFacility.value = null
  }

  function getHexAt(q: number, r: number): Hex | null {
    if (!gameState.value) return null
    const key = hexKey(q, r)
    return gameState.value.hexes[key] || null
  }

  function setActiveTab(tab: string) {
    activeTab.value = tab
  }

  function closeProposalModal() {
    showProposalModal.value = false
  }

  function openProposalModal() {
    showProposalModal.value = true
  }

  return {
    gameState,
    currentPlayerId,
    selectedHex,
    selectedShip,
    selectedFacility,
    activeTab,
    showProposalModal,
    pendingProposals,
    currentPlayer,
    playerShips,
    playerFacilities,
    playerTechs,
    otherPlayers,
    relations,
    proposals,
    gameLogs,
    myPendingProposals,
    setGameState,
    setCurrentPlayer,
    selectHex,
    selectShip,
    selectFacility,
    clearSelection,
    getHexAt,
    getRelationWithPlayer,
    hasCooldown,
    getCooldown,
    setActiveTab,
    closeProposalModal,
    openProposalModal
  }
})
