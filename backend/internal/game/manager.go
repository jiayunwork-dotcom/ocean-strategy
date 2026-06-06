package game

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"ocean-strategy/internal/models"
)

type GameManager struct {
	games map[uuid.UUID]*GameInstance
	mu    sync.RWMutex
}

type GameInstance struct {
	ID        uuid.UUID
	Engine    *GameEngine
	Players   map[uuid.UUID]bool
	CreatedAt time.Time
	Seed      int64
}

var manager *GameManager
var managerOnce sync.Once

func GetGameManager() *GameManager {
	managerOnce.Do(func() {
		manager = &GameManager{
			games: make(map[uuid.UUID]*GameInstance),
		}
	})
	return manager
}

func (gm *GameManager) CreateGame(name string, maxTurns int, mapRadius int) (*models.Game, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	gameID := uuid.New()
	seed := time.Now().UnixNano()

	game := &models.Game{
		ID:           gameID,
		Name:         name,
		Status:       models.GameWaiting,
		CurrentTurn:  0,
		MaxTurns:     maxTurns,
		Phase:        models.PhaseProduction,
		MapRadius:    mapRadius,
		Players:      make([]models.Player, 0),
		CurrentPlayerIndex: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		WinCondition: "economic",
	}

	generator := NewMapGenerator(mapRadius, seed)
	hexes := generator.Generate()

	state := &models.GameState{
		Game:       game,
		Hexes:      hexes,
		Players:    make(map[uuid.UUID]*models.Player),
		Ships:      make([]*models.Ship, 0),
		Facilities: make([]*models.Facility, 0),
		Techs:      make([]*models.PlayerTech, 0),
		Relations:  make([]*models.DiplomaticRelation, 0),
		Typhoons:   make([]*models.Typhoon, 0),
	}

	engine := NewGameEngine(state, seed)

	gm.games[gameID] = &GameInstance{
		ID:        gameID,
		Engine:    engine,
		Players:   make(map[uuid.UUID]bool),
		CreatedAt: time.Now(),
		Seed:      seed,
	}

	return game, nil
}

func (gm *GameManager) AddPlayer(gameID uuid.UUID, playerName string, color string, isAI bool) (*models.Player, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	instance, ok := gm.games[gameID]
	if !ok {
		return nil, errors.New("game not found")
	}

	if instance.Engine.GetState().Game.Status != models.GameWaiting {
		return nil, errors.New("game already started")
	}

	if len(instance.Engine.GetState().Game.Players) >= 8 {
		return nil, errors.New("game is full")
	}

	playerID := uuid.New()
	startPositions := GetStartPositions(instance.Engine.GetState().Game.MapRadius, len(instance.Engine.GetState().Game.Players)+1)
	posIndex := len(instance.Engine.GetState().Game.Players)

	player := &models.Player{
		ID:          playerID,
		GameID:      gameID,
		Name:        playerName,
		Color:       color,
		Money:       15000,
		Reputation:  50,
		StartHexQ:   startPositions[posIndex][0],
		StartHexR:   startPositions[posIndex][1],
		IsAI:        isAI,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	instance.Engine.GetState().Players[playerID] = player
	instance.Engine.GetState().Game.Players = append(instance.Engine.GetState().Game.Players, *player)
	instance.Players[playerID] = true

	startHexKey := HexKey(player.StartHexQ, player.StartHexR)
	if hex, ok := instance.Engine.GetState().Hexes[startHexKey]; ok {
		hex.Discovered = true
		hex.OwnerID = &playerID
		hex.IsEEZ = true
		hex.EEZOwnerID = &playerID

		neighbors := HexNeighbors(player.StartHexQ, player.StartHexR)
		for _, pos := range neighbors {
			if nHex, ok := instance.Engine.GetState().Hexes[HexKey(pos[0], pos[1])]; ok {
				nHex.Discovered = true
				nHex.IsEEZ = true
				nHex.EEZOwnerID = &playerID
			}
		}
	}

	return player, nil
}

func (gm *GameManager) StartGame(gameID uuid.UUID) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	instance, ok := gm.games[gameID]
	if !ok {
		return errors.New("game not found")
	}

	if len(instance.Engine.GetState().Game.Players) < 2 {
		return errors.New("need at least 2 players")
	}

	instance.Engine.GetState().Game.Status = models.GamePlaying

	for _, player := range instance.Engine.GetState().Players {
		instance.Engine.BuildFacility(player.ID, player.StartHexQ, player.StartHexR, models.FacilityPort)
	}

	return nil
}

func (gm *GameManager) GetGame(gameID uuid.UUID) (*GameInstance, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	instance, ok := gm.games[gameID]
	if !ok {
		return nil, errors.New("game not found")
	}

	return instance, nil
}

func (gm *GameManager) ListGames() []*models.Game {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	games := make([]*models.Game, 0, len(gm.games))
	for _, instance := range gm.games {
		games = append(games, instance.Engine.GetState().Game)
	}

	return games
}

func (gm *GameManager) NextPhase(gameID uuid.UUID) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	instance, ok := gm.games[gameID]
	if !ok {
		return errors.New("game not found")
	}

	instance.Engine.NextPhase()
	instance.Engine.GetState().Game.UpdatedAt = time.Now()

	if instance.Engine.GetState().Game.CurrentTurn >= instance.Engine.GetState().Game.MaxTurns {
		instance.Engine.GetState().Game.Status = models.GameFinished
		gm.determineWinner(instance)
	}

	return nil
}

func (gm *GameManager) determineWinner(instance *GameInstance) {
	state := instance.Engine.GetState()
	var winner *models.Player
	highestScore := 0

	for _, player := range state.Players {
		score := player.Money
		if score > highestScore {
			highestScore = score
			winner = player
		}
	}

	if winner != nil {
		state.Game.WinnerID = &winner.ID
	}
}
