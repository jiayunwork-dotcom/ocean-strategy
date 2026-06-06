import type { GameState, Player, Hex, Ship, Facility, PlayerTech, Typhoon } from '~/types'
import { hexKey } from '~/utils/hex'

export const useGameStore = defineStore('game', () => {
  const gameState = ref<GameState | null>(null)
  const currentPlayerId = ref<string | null>(null)
  const selectedHex = ref<{ q: number; r: number } | null>(null)
  const selectedShip = ref<Ship | null>(null)
  const selectedFacility = ref<Facility | null>(null)

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

  function setGameState(state: GameState) {
    gameState.value = state
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

  return {
    gameState,
    currentPlayerId,
    selectedHex,
    selectedShip,
    selectedFacility,
    currentPlayer,
    playerShips,
    playerFacilities,
    playerTechs,
    setGameState,
    setCurrentPlayer,
    selectHex,
    selectShip,
    selectFacility,
    clearSelection,
    getHexAt
  }
})
