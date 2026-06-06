import type { Game, GameState, Player, Technology } from '~/types'

export function useGameApi() {
  const config = useRuntimeConfig()
  const baseUrl = config.public.apiBase

  const listGames = async (): Promise<Game[]> => {
    const res = await $fetch<Game[]>(`${baseUrl}/games`)
    return res
  }

  const createGame = async (name: string, maxTurns: number = 50, mapRadius: number = 6, winCondition: string = 'economic'): Promise<Game> => {
    const res = await $fetch<Game>(`${baseUrl}/games`, {
      method: 'POST',
      body: { name, max_turns: maxTurns, map_radius: mapRadius, win_condition: winCondition }
    })
    return res
  }

  const getGame = async (gameId: string): Promise<GameState> => {
    const res = await $fetch<GameState>(`${baseUrl}/games/${gameId}`)
    return res
  }

  const joinGame = async (gameId: string, playerName: string, color: string = '#3498db'): Promise<Player> => {
    const res = await $fetch<Player>(`${baseUrl}/games/${gameId}/join`, {
      method: 'POST',
      body: { player_name: playerName, color }
    })
    return res
  }

  const startGame = async (gameId: string): Promise<void> => {
    await $fetch(`${baseUrl}/games/${gameId}/start`, { method: 'POST' })
  }

  const nextPhase = async (gameId: string): Promise<GameState> => {
    const res = await $fetch<GameState>(`${baseUrl}/games/${gameId}/next-phase`, {
      method: 'POST'
    })
    return res
  }

  const buildFacility = async (gameId: string, playerId: string, q: number, r: number, type: string): Promise<void> => {
    await $fetch(`${baseUrl}/games/${gameId}/facilities`, {
      method: 'POST',
      body: { player_id: playerId, q, r, type }
    })
  }

  const buildShip = async (gameId: string, playerId: string, type: string, q: number, r: number): Promise<void> => {
    await $fetch(`${baseUrl}/games/${gameId}/ships`, {
      method: 'POST',
      body: { player_id: playerId, type, q, r }
    })
  }

  const moveShip = async (gameId: string, shipId: string, toQ: number, toR: number): Promise<void> => {
    await $fetch(`${baseUrl}/games/${gameId}/ships/move`, {
      method: 'POST',
      body: { ship_id: shipId, to_q: toQ, to_r: toR }
    })
  }

  const explore = async (gameId: string, shipId: string): Promise<void> => {
    await $fetch(`${baseUrl}/games/${gameId}/ships/explore`, {
      method: 'POST',
      body: { ship_id: shipId }
    })
  }

  const startResearch = async (gameId: string, playerId: string, techId: string): Promise<void> => {
    await $fetch(`${baseUrl}/games/${gameId}/research`, {
      method: 'POST',
      body: { player_id: playerId, tech_id: techId }
    })
  }

  const getTechnologies = async (): Promise<Technology[]> => {
    const res = await $fetch<Technology[]>(`${baseUrl}/technologies`)
    return res
  }

  return {
    listGames,
    createGame,
    getGame,
    joinGame,
    startGame,
    nextPhase,
    buildFacility,
    buildShip,
    moveShip,
    explore,
    startResearch,
    getTechnologies
  }
}
