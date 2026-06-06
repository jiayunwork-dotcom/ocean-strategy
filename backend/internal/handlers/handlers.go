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
	Name      string `json:"name"`
	MaxTurns  int    `json:"max_turns"`
	MapRadius int    `json:"map_radius"`
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
	gameModel, err := gm.CreateGame(req.Name, req.MaxTurns, req.MapRadius)
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

	if !instance.Engine.MoveShip(req.ShipID, req.ToQ, req.ToR) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "could not move ship"})
	}

	hub := ws.GetHub()
	hub.BroadcastToGame(gameID, "ship_moved", fiber.Map{
		"ship_id": req.ShipID,
		"to_q":    req.ToQ,
		"to_r":    req.ToR,
	})

	return c.JSON(fiber.Map{"status": "success"})
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
