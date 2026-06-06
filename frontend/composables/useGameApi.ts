import type { Game, GameState, Player, Technology, MarketData, Auction, OrderType, ResourceType, AuctionItemType } from '~/types'

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

  const proposeTreaty = async (gameId: string, fromPlayerId: string, toPlayerId: string, treatyType: string): Promise<void> => {
    await $fetch(`${baseUrl}/games/${gameId}/diplomacy/propose`, {
      method: 'POST',
      body: { from_player_id: fromPlayerId, to_player_id: toPlayerId, treaty_type: treatyType }
    })
  }

  const respondToProposal = async (gameId: string, proposalId: string, playerId: string, accept: boolean): Promise<void> => {
    await $fetch(`${baseUrl}/games/${gameId}/diplomacy/respond`, {
      method: 'POST',
      body: { proposal_id: proposalId, player_id: playerId, accept }
    })
  }

  const breakTreaty = async (gameId: string, playerId: string, otherPlayerId: string): Promise<void> => {
    await $fetch(`${baseUrl}/games/${gameId}/diplomacy/break`, {
      method: 'POST',
      body: { player_id: playerId, other_player_id: otherPlayerId }
    })
  }

  const getMarketData = async (gameId: string, playerId?: string): Promise<MarketData> => {
    const query = playerId ? { player_id: playerId } : {}
    const res = await $fetch<MarketData>(`${baseUrl}/games/${gameId}/market`, { query })
    return res
  }

  const placeOrder = async (gameId: string, playerId: string, orderType: OrderType, resource: ResourceType, quantity: number, price: number): Promise<void> => {
    await $fetch(`${baseUrl}/games/${gameId}/market/orders`, {
      method: 'POST',
      body: { player_id: playerId, order_type: orderType, resource, quantity, price }
    })
  }

  const cancelOrder = async (gameId: string, playerId: string, orderId: string): Promise<void> => {
    await $fetch(`${baseUrl}/games/${gameId}/market/orders/cancel`, {
      method: 'POST',
      body: { player_id: playerId, order_id: orderId }
    })
  }

  const createAuction = async (gameId: string, playerId: string, itemType: AuctionItemType, itemId: string, startingPrice: number): Promise<Auction> => {
    const res = await $fetch<any>(`${baseUrl}/games/${gameId}/auctions`, {
      method: 'POST',
      body: { player_id: playerId, item_type: itemType, item_id: itemId, starting_price: startingPrice }
    })
    return res.auction
  }

  const placeBid = async (gameId: string, playerId: string, auctionId: string, amount: number): Promise<void> => {
    await $fetch(`${baseUrl}/games/${gameId}/auctions/bid`, {
      method: 'POST',
      body: { player_id: playerId, auction_id: auctionId, amount }
    })
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
    getTechnologies,
    proposeTreaty,
    respondToProposal,
    breakTreaty,
    getMarketData,
    placeOrder,
    cancelOrder,
    createAuction,
    placeBid
  }
}
