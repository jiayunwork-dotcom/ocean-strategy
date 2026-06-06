package handlers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"ocean-strategy/internal/game"
	"ocean-strategy/internal/models"
	ws "ocean-strategy/internal/websocket"
)

type CreateGameRequest struct {
	Name         string `json:"name"`
	MaxTurns     int    `json:"max_turns"`
	MapRadius    int    `json:"map_radius"`
	WinCondition string `json:"win_condition"`
}

type JoinGameRequest struct {
	PlayerName string `json:"player_name"`
	Color      string `json:"color"`
}

type BuildFacilityRequest struct {
	PlayerID uuid.UUID          `json:"player_id"`
	Q        int                `json:"q"`
	R        int                `json:"r"`
	Type     models.FacilityType `json:"type"`
}

type BuildShipRequest struct {
	PlayerID uuid.UUID      `json:"player_id"`
	Type     models.ShipType `json:"type"`
	Q        int            `json:"q"`
	R        int            `json:"r"`
}

type MoveShipRequest struct {
	ShipID uuid.UUID `json:"ship_id"`
	ToQ    int       `json:"to_q"`
	ToR    int       `json:"to_r"`
}

type ExploreRequest struct {
	ShipID uuid.UUID `json:"ship_id"`
}

type ResearchRequest struct {
	PlayerID uuid.UUID `json:"player_id"`
	TechID   string    `json:"tech_id"`
}

type ProposeTreatyRequest struct {
	FromPlayerID uuid.UUID         `json:"from_player_id"`
	ToPlayerID   uuid.UUID         `json:"to_player_id"`
	TreatyType   models.TreatyType `json:"treaty_type"`
}

type RespondProposalRequest struct {
	ProposalID uuid.UUID `json:"proposal_id"`
	PlayerID   uuid.UUID `json:"player_id"`
	Accept     bool      `json:"accept"`
}

type BreakTreatyRequest struct {
	PlayerID      uuid.UUID `json:"player_id"`
	OtherPlayerID uuid.UUID `json:"other_player_id"`
}

type PlaceOrderRequest struct {
	PlayerID uuid.UUID          `json:"player_id"`
	OrderType models.OrderType  `json:"order_type"`
	Resource  models.ResourceType `json:"resource"`
	Quantity  int                `json:"quantity"`
	Price     int                `json:"price"`
}

type CancelOrderRequest struct {
	PlayerID uuid.UUID `json:"player_id"`
	OrderID  uuid.UUID `json:"order_id"`
}

type CreateAuctionRequest struct {
	PlayerID      uuid.UUID           `json:"player_id"`
	ItemType      models.AuctionItemType `json:"item_type"`
	ItemID        string              `json:"item_id"`
	StartingPrice int                 `json:"starting_price"`
}

type PlaceBidRequest struct {
	PlayerID  uuid.UUID `json:"player_id"`
	AuctionID uuid.UUID `json:"auction_id"`
	Amount    int       `json:"amount"`
}

type CreateFuturesContractRequest struct {
	PlayerID      uuid.UUID          `json:"player_id"`
	Resource      models.ResourceType `json:"resource"`
	Quantity      int                `json:"quantity"`
	ContractPrice int                `json:"contract_price"`
	DeliveryTurn  int                `json:"delivery_turn"`
}

type AcceptFuturesContractRequest struct {
	PlayerID   uuid.UUID `json:"player_id"`
	ContractID uuid.UUID `json:"contract_id"`
}

type CancelFuturesContractRequest struct {
	PlayerID   uuid.UUID `json:"player_id"`
	ContractID uuid.UUID `json:"contract_id"`
}

type AddFuturesMarginRequest struct {
	PlayerID   uuid.UUID `json:"player_id"`
	ContractID uuid.UUID `json:"contract_id"`
	Amount     int       `json:"amount"`
}

func CreateGame(c *fiber.Ctx) error {
	var req CreateGameRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if req.MaxTurns == 0 {
		req.MaxTurns = 50
	}
	if req.MapRadius == 0 {
		req.MapRadius = 6
	}

	gm := game.GetGameManager()
	gameModel, err := gm.CreateGame(req.Name, req.MaxTurns, req.MapRadius, req.WinCondition)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(http.StatusCreated).JSON(gameModel)
}

func ListGames(c *fiber.Ctx) error {
	gm := game.GetGameManager()
	games := gm.ListGames()
	return c.JSON(games)
}

func GetGame(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(instance.Engine.GetState())
}

func JoinGame(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req JoinGameRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if req.Color == "" {
		req.Color = "#3498db"
	}

	gm := game.GetGameManager()
	player, err := gm.AddPlayer(gameID, req.PlayerName, req.Color, false)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "player_joined", player)

	return c.JSON(player)
}

func StartGame(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	gm := game.GetGameManager()
	if err := gm.StartGame(gameID); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	instance, _ := gm.GetGame(gameID)
	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "game_started", instance.Engine.GetState())

	return c.JSON(fiber.Map{"status": "started"})
}

func NextPhase(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	gm := game.GetGameManager()
	if err := gm.NextPhase(gameID); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	instance, _ := gm.GetGame(gameID)
	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "phase_changed", instance.Engine.GetState())

	return c.JSON(instance.Engine.GetState())
}

func BuildFacility(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req BuildFacilityRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	if !instance.Engine.BuildFacility(req.PlayerID, req.Q, req.R, req.Type) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "could not build facility"})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "facility_built", fiber.Map{
		"player_id": req.PlayerID,
		"q":         req.Q,
		"r":         req.R,
		"type":      req.Type,
	})

	return c.JSON(fiber.Map{"status": "success"})
}

func BuildShip(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req BuildShipRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	if !instance.Engine.BuildShip(req.PlayerID, req.Type, req.Q, req.R) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "could not build ship"})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "ship_built", fiber.Map{
		"player_id": req.PlayerID,
		"type":      req.Type,
		"q":         req.Q,
		"r":         req.R,
	})

	return c.JSON(fiber.Map{"status": "success"})
}

func MoveShip(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req MoveShipRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	success, battleLog := instance.Engine.MoveShip(req.ShipID, req.ToQ, req.ToR)
	if !success {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "could not move ship"})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "ship_moved", fiber.Map{
		"ship_id": req.ShipID,
		"to_q":    req.ToQ,
		"to_r":    req.ToR,
	})

	if battleLog != nil {
		hub.BroadcastToGame(gameID, "battle_occurred", battleLog)
	}

	return c.JSON(fiber.Map{"status": "success", "battle_log": battleLog})
}

func Explore(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req ExploreRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	if !instance.Engine.Explore(req.ShipID) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "could not explore"})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "explored", fiber.Map{
		"ship_id": req.ShipID,
	})

	return c.JSON(fiber.Map{"status": "success"})
}

func StartResearch(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req ResearchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	if !instance.Engine.StartResearch(req.PlayerID, req.TechID) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "could not start research"})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "research_started", fiber.Map{
		"player_id": req.PlayerID,
		"tech_id":   req.TechID,
	})

	return c.JSON(fiber.Map{"status": "success"})
}

func GetTechnologies(c *fiber.Ctx) error {
	return c.JSON(game.Technologies)
}

func ProposeTreaty(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req ProposeTreatyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	success, msg := instance.Engine.ProposeTreaty(req.FromPlayerID, req.ToPlayerID, req.TreatyType)
	if !success {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "treaty_proposed", fiber.Map{
		"from_player_id": req.FromPlayerID,
		"to_player_id":   req.ToPlayerID,
		"treaty_type":    req.TreatyType,
	})

	return c.JSON(fiber.Map{"status": "success"})
}

func RespondToProposal(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req RespondProposalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	success, msg := instance.Engine.RespondToProposal(req.ProposalID, req.PlayerID, req.Accept)
	if !success {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "proposal_responded", fiber.Map{
		"proposal_id": req.ProposalID,
		"player_id":   req.PlayerID,
		"accept":      req.Accept,
	})

	return c.JSON(fiber.Map{"status": "success"})
}

func BreakTreaty(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req BreakTreatyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	success, msg := instance.Engine.BreakTreaty(req.PlayerID, req.OtherPlayerID)
	if !success {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "treaty_broken", fiber.Map{
		"player_id":       req.PlayerID,
		"other_player_id": req.OtherPlayerID,
	})

	return c.JSON(fiber.Map{"status": "success"})
}

func PlaceOrder(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req PlaceOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	success, msg := instance.Engine.PlaceOrder(req.PlayerID, req.OrderType, req.Resource, req.Quantity, req.Price)
	if !success {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "order_placed", fiber.Map{
		"player_id": req.PlayerID,
		"order_type": req.OrderType,
		"resource":  req.Resource,
	})

	return c.JSON(fiber.Map{"status": "success"})
}

func CancelOrder(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req CancelOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	success, msg := instance.Engine.CancelOrder(req.PlayerID, req.OrderID)
	if !success {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "order_cancelled", fiber.Map{
		"player_id": req.PlayerID,
		"order_id":  req.OrderID,
	})

	return c.JSON(fiber.Map{"status": "success"})
}

func GetMarketData(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	playerIDStr := c.Query("player_id")
	var playerID uuid.UUID
	if playerIDStr != "" {
		playerID, _ = uuid.Parse(playerIDStr)
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	state := instance.Engine.GetState()

	var visibleOrders []*models.MarketOrder
	if playerID != uuid.Nil {
		visibleOrders = instance.Engine.GetVisibleOrders(playerID)
	} else {
		visibleOrders = state.MarketOrders
	}

	var visibleAuctions []*models.Auction
	if playerID != uuid.Nil {
		visibleAuctions = instance.Engine.GetVisibleAuctions(playerID)
	} else {
		visibleAuctions = state.Auctions
	}

	return c.JSON(fiber.Map{
		"current_prices": state.CurrentPrices,
		"price_history":  state.PriceHistory,
		"orders":         visibleOrders,
		"auctions":       visibleAuctions,
		"resource_stats": state.ResourceStats,
	})
}

func CreateAuction(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req CreateAuctionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	success, msg, auction := instance.Engine.CreateAuction(req.PlayerID, req.ItemType, req.ItemID, req.StartingPrice)
	if !success {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "auction_created", auction)

	return c.JSON(fiber.Map{"status": "success", "auction": auction})
}

func PlaceBid(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req PlaceBidRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	success, msg := instance.Engine.PlaceBid(req.PlayerID, req.AuctionID, req.Amount)
	if !success {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "bid_placed", fiber.Map{
		"auction_id": req.AuctionID,
		"player_id":  req.PlayerID,
		"amount":     req.Amount,
	})

	return c.JSON(fiber.Map{"status": "success"})
}

func CreateFuturesContract(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req CreateFuturesContractRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	success, msg, contract := instance.Engine.CreateFuturesContract(req.PlayerID, req.Resource, req.Quantity, req.ContractPrice, req.DeliveryTurn)
	if !success {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "futures_created", fiber.Map{
		"contract_id": contract.ID,
		"player_id":   req.PlayerID,
		"resource":    req.Resource,
	})

	return c.JSON(fiber.Map{"status": "success", "contract": contract})
}

func AcceptFuturesContract(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req AcceptFuturesContractRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	success, msg := instance.Engine.AcceptFuturesContract(req.PlayerID, req.ContractID)
	if !success {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "futures_accepted", fiber.Map{
		"contract_id": req.ContractID,
		"player_id":   req.PlayerID,
	})

	return c.JSON(fiber.Map{"status": "success"})
}

func CancelFuturesContract(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req CancelFuturesContractRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	success, msg := instance.Engine.CancelFuturesContract(req.PlayerID, req.ContractID)
	if !success {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "futures_cancelled", fiber.Map{
		"contract_id": req.ContractID,
		"player_id":   req.PlayerID,
	})

	return c.JSON(fiber.Map{"status": "success"})
}

func AddFuturesMargin(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	var req AddFuturesMarginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	success, msg := instance.Engine.AddFuturesMargin(req.PlayerID, req.ContractID, req.Amount)
	if !success {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": msg})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "futures_margin_added", fiber.Map{
		"contract_id": req.ContractID,
		"player_id":   req.PlayerID,
		"amount":      req.Amount,
	})

	return c.JSON(fiber.Map{"status": "success"})
}

func GetFuturesData(c *fiber.Ctx) error {
	gameID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid game id"})
	}

	playerIDStr := c.Query("player_id")
	var playerID uuid.UUID
	if playerIDStr != "" {
		playerID, _ = uuid.Parse(playerIDStr)
	}

	gm := game.GetGameManager()
	instance, err := gm.GetGame(gameID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	state := instance.Engine.GetState()

	var visibleContracts []*models.FuturesContract
	if playerID != uuid.Nil {
		visibleContracts = instance.Engine.GetVisibleFuturesContracts(playerID)
	} else {
		visibleContracts = state.FuturesContracts
	}

	contractsWithPnL := make([]fiber.Map, 0, len(visibleContracts))
	for _, contract := range visibleContracts {
		contractMap := fiber.Map{
			"id":               contract.ID,
			"game_id":          contract.GameID,
			"creator_id":       contract.CreatorID,
			"accepter_id":      contract.AccepterID,
			"resource":         contract.Resource,
			"quantity":         contract.Quantity,
			"contract_price":   contract.ContractPrice,
			"delivery_turn":    contract.DeliveryTurn,
			"creator_margin":   contract.CreatorMargin,
			"accepter_margin":  contract.AccepterMargin,
			"initial_margin":   contract.InitialMargin,
			"status":           contract.Status,
			"margin_call_turn": contract.MarginCallTurn,
			"margin_call_party": contract.MarginCallParty,
			"created_turn":     contract.CreatedTurn,
			"settled_turn":     contract.SettledTurn,
			"settlement_price": contract.SettlementPrice,
			"creator_pnl":      contract.CreatorPnL,
			"accepter_pnl":     contract.AccepterPnL,
			"created_at":       contract.CreatedAt,
		}
		if playerID != uuid.Nil {
			contractMap["floating_pnl"] = instance.Engine.GetFloatingPnL(contract, playerID)
			contractMap["margin_status"] = instance.Engine.GetMarginStatus(contract, playerID)
		}
		contractsWithPnL = append(contractsWithPnL, contractMap)
	}

	return c.JSON(fiber.Map{
		"contracts":           contractsWithPnL,
		"settlements":         state.FuturesSettlements,
		"manipulation_penalties": state.ManipulationPenalties,
	})
}
