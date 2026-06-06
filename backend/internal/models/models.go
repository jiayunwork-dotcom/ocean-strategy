package models

import (
	"time"

	"github.com/google/uuid"
)

type Player struct {
	ID              uuid.UUID   `json:"id" gorm:"type:uuid;primaryKey"`
	GameID          uuid.UUID   `json:"game_id" gorm:"type:uuid;index"`
	Name            string      `json:"name"`
	Color           string      `json:"color"`
	Money           int         `json:"money"`
	Reputation      int         `json:"reputation"`
	StartHexQ       int         `json:"start_hex_q"`
	StartHexR       int         `json:"start_hex_r"`
	IsAI            bool        `json:"is_ai"`
	DiscoveredHexes []string    `json:"discovered_hexes"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type ResourceType string

const (
	ResourceOil        ResourceType = "oil"
	ResourceGas        ResourceType = "gas"
	ResourceManganese  ResourceType = "manganese"
	ResourceSulfide    ResourceType = "sulfide"
	ResourceBiomaterial ResourceType = "biomaterial"
)

type HexTerrain string

const (
	TerrainShallow    HexTerrain = "shallow"
	TerrainDeep       HexTerrain = "deep"
	TerrainTrench     HexTerrain = "trench"
	TerrainReef       HexTerrain = "reef"
	TerrainVent       HexTerrain = "vent"
	TerrainOpenOcean  HexTerrain = "open_ocean"
)

type Hex struct {
	Q               int               `json:"q"`
	R               int               `json:"r"`
	Terrain         HexTerrain        `json:"terrain"`
	Resources       map[ResourceType]int `json:"resources"`
	Discovered      bool              `json:"discovered"`
	OwnerID         *uuid.UUID        `json:"owner_id,omitempty"`
	Facility        *Facility         `json:"facility,omitempty"`
	EcologicalHealth int              `json:"ecological_health"`
	Pollution       int               `json:"pollution"`
	HasCurrent      bool              `json:"has_current"`
	CurrentDir      int               `json:"current_dir"`
	IsEEZ           bool              `json:"is_eez"`
	EEZOwnerID      *uuid.UUID        `json:"eez_owner_id,omitempty"`
}

type FacilityType string

const (
	FacilityDrilling   FacilityType = "drilling"
	FacilityMine       FacilityType = "mine"
	FacilityTidal      FacilityType = "tidal"
	FacilityFarm       FacilityType = "farm"
	FacilityPort       FacilityType = "port"
)

type Facility struct {
	ID           uuid.UUID    `json:"id"`
	Type         FacilityType `json:"type"`
	Level        int          `json:"level"`
	HexQ         int          `json:"hex_q"`
	HexR         int          `json:"hex_r"`
	OwnerID      uuid.UUID    `json:"owner_id"`
	Health       int          `json:"health"`
	MaxHealth    int          `json:"max_health"`
	MaintenanceCost int       `json:"maintenance_cost"`
	BuildTurnsLeft int         `json:"build_turns_left"`
	IsActive     bool         `json:"is_active"`
	PowerOutput  int          `json:"power_output"`
	PowerConsume int          `json:"power_consume"`
}

type ShipType string

const (
	ShipExplorer  ShipType = "explorer"
	ShipConstructor ShipType = "constructor"
	ShipTransport ShipType = "transport"
	ShipEscort    ShipType = "escort"
)

type Ship struct {
	ID           uuid.UUID `json:"id"`
	Type         ShipType  `json:"type"`
	OwnerID      uuid.UUID `json:"owner_id"`
	HexQ         int       `json:"hex_q"`
	HexR         int       `json:"hex_r"`
	Health       int       `json:"health"`
	MaxHealth    int       `json:"max_health"`
	Fuel         int       `json:"fuel"`
	MaxFuel      int       `json:"max_fuel"`
	Cargo        map[ResourceType]int `json:"cargo"`
	CargoCapacity int      `json:"cargo_capacity"`
	Speed        int       `json:"speed"`
	MovePoints   int       `json:"move_points"`
	Attack       int       `json:"attack"`
	Defense      int       `json:"defense"`
}

type TechnologyCategory string

const (
	TechExtraction TechnologyCategory = "extraction"
	TechEcology    TechnologyCategory = "ecology"
	TechMilitary   TechnologyCategory = "military"
)

type Technology struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Category     TechnologyCategory `json:"category"`
	Cost         int                `json:"cost"`
	Turns        int                `json:"turns"`
	Description  string             `json:"description"`
	Prerequisites []string          `json:"prerequisites"`
	Effects      map[string]int     `json:"effects"`
}

type PlayerTech struct {
	PlayerID    uuid.UUID `json:"player_id"`
	TechID      string    `json:"tech_id"`
	Researching bool      `json:"researching"`
	TurnsLeft   int       `json:"turns_left"`
	Completed   bool      `json:"completed"`
}

type RelationStatus string

const (
	RelationNeutral   RelationStatus = "neutral"
	RelationNAP       RelationStatus = "nap"
	RelationAlliance  RelationStatus = "alliance"
	RelationHostile   RelationStatus = "hostile"
)

type TreatyType string

const (
	TreatyNAP      TreatyType = "nap"
	TreatyAlliance TreatyType = "alliance"
)

type DiplomaticProposal struct {
	ID         uuid.UUID    `json:"id"`
	GameID     uuid.UUID    `json:"game_id"`
	FromPlayerID uuid.UUID  `json:"from_player_id"`
	ToPlayerID uuid.UUID    `json:"to_player_id"`
	TreatyType TreatyType   `json:"treaty_type"`
	Status     string       `json:"status"`
	CreatedAt  int          `json:"created_at"`
}

type DiplomaticRelation struct {
	GameID        uuid.UUID      `json:"game_id"`
	Player1ID     uuid.UUID      `json:"player1_id"`
	Player2ID     uuid.UUID      `json:"player2_id"`
	Status        RelationStatus `json:"status"`
	HasNAP        bool           `json:"has_nap"`
	HasAlliance   bool           `json:"has_alliance"`
	AtWar         bool           `json:"at_war"`
}

type ReputationCooldown struct {
	PlayerID    uuid.UUID `json:"player_id"`
	GameID      uuid.UUID `json:"game_id"`
	TurnsLeft   int       `json:"turns_left"`
	Reason      string    `json:"reason"`
}

type BattleLog struct {
	ID             uuid.UUID `json:"id"`
	GameID         uuid.UUID `json:"game_id"`
	Turn           int       `json:"turn"`
	AttackerID     uuid.UUID `json:"attacker_id"`
	DefenderID     uuid.UUID `json:"defender_id"`
	AttackerShipID uuid.UUID `json:"attacker_ship_id"`
	DefenderShipID uuid.UUID `json:"defender_ship_id"`
	HexQ           int       `json:"hex_q"`
	HexR           int       `json:"hex_r"`
	AttackerDamage int       `json:"attacker_damage"`
	DefenderDamage int       `json:"defender_damage"`
	AttackerSunk   bool      `json:"attacker_sunk"`
	DefenderSunk   bool      `json:"defender_sunk"`
	Timestamp      time.Time `json:"timestamp"`
}

type GameLogEntry struct {
	ID        uuid.UUID `json:"id"`
	GameID    uuid.UUID `json:"game_id"`
	Turn      int       `json:"turn"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	PlayerID  *uuid.UUID `json:"player_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type GamePhase string

const (
	PhaseProduction GamePhase = "production"
	PhaseDecision   GamePhase = "decision"
	PhaseEvent      GamePhase = "event"
	PhaseSettlement GamePhase = "settlement"
)

type GameStatus string

const (
	GameWaiting  GameStatus = "waiting"
	GamePlaying  GameStatus = "playing"
	GameFinished GameStatus = "finished"
)

type Game struct {
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	Name         string     `json:"name"`
	Status       GameStatus `json:"status"`
	CurrentTurn  int        `json:"current_turn"`
	MaxTurns     int        `json:"max_turns"`
	Phase        GamePhase  `json:"phase"`
	MapRadius    int        `json:"map_radius"`
	Players      []Player   `json:"players" gorm:"foreignKey:GameID"`
	CurrentPlayerIndex int   `json:"current_player_index"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	WinCondition string     `json:"win_condition"`
	WinnerID     *uuid.UUID `json:"winner_id,omitempty"`
}

type GameState struct {
	Game           *Game                     `json:"game"`
	Hexes          map[string]*Hex           `json:"hexes"`
	Players        map[uuid.UUID]*Player     `json:"players"`
	Ships          []*Ship                   `json:"ships"`
	Facilities     []*Facility               `json:"facilities"`
	Techs          []*PlayerTech             `json:"techs"`
	Relations      []*DiplomaticRelation     `json:"relations"`
	Proposals      []*DiplomaticProposal     `json:"proposals"`
	Cooldowns      []*ReputationCooldown     `json:"cooldowns"`
	BattleLogs     []*BattleLog              `json:"battle_logs"`
	GameLogs       []*GameLogEntry           `json:"game_logs"`
	Typhoons       []*Typhoon                `json:"typhoons"`
	MarketOrders   []*MarketOrder            `json:"market_orders"`
	TradeRecords   []*TradeRecord            `json:"trade_records"`
	PriceHistory   []*PriceHistoryEntry      `json:"price_history"`
	ResourceStats  map[ResourceType]*ResourceStats `json:"resource_stats"`
	Auctions       []*Auction                `json:"auctions"`
	AuctionBids    []*AuctionBid             `json:"auction_bids"`
	CurrentPrices        map[ResourceType]int              `json:"current_prices"`
	FrozenShips          map[uuid.UUID]uuid.UUID           `json:"frozen_ships"`
	FrozenTechs          map[string]uuid.UUID              `json:"frozen_techs"`
	FuturesContracts     []*FuturesContract                `json:"futures_contracts"`
	FuturesSettlements   []*FuturesSettlement              `json:"futures_settlements"`
	ManipulationPenalties []*MarketManipulationPenalty    `json:"manipulation_penalties"`
}

type Typhoon struct {
	ID         uuid.UUID `json:"id"`
	HexQ       int       `json:"hex_q"`
	HexR       int       `json:"hex_r"`
	Strength   int       `json:"strength"`
	DirQ       int       `json:"dir_q"`
	DirR       int       `json:"dir_r"`
	TurnsLeft  int       `json:"turns_left"`
}

type MarketPrice struct {
	Resource ResourceType `json:"resource"`
	Price    int          `json:"price"`
	Demand   int          `json:"demand"`
}

type OrderType string

const (
	OrderTypeBuy  OrderType = "buy"
	OrderTypeSell OrderType = "sell"
)

type OrderStatus string

const (
	OrderStatusActive   OrderStatus = "active"
	OrderStatusPartial  OrderStatus = "partial"
	OrderStatusFilled   OrderStatus = "filled"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type MarketOrder struct {
	ID            uuid.UUID    `json:"id"`
	GameID        uuid.UUID    `json:"game_id"`
	PlayerID      uuid.UUID    `json:"player_id"`
	OrderType     OrderType    `json:"order_type"`
	Resource      ResourceType `json:"resource"`
	Quantity      int          `json:"quantity"`
	RemainingQty  int          `json:"remaining_qty"`
	Price         int          `json:"price"`
	Status        OrderStatus  `json:"status"`
	CreatedTurn   int          `json:"created_turn"`
	CreatedAt     time.Time    `json:"created_at"`
}

type TradeRecord struct {
	ID           uuid.UUID    `json:"id"`
	GameID       uuid.UUID    `json:"game_id"`
	BuyOrderID   uuid.UUID    `json:"buy_order_id"`
	SellOrderID  uuid.UUID    `json:"sell_order_id"`
	BuyerID      uuid.UUID    `json:"buyer_id"`
	SellerID     uuid.UUID    `json:"seller_id"`
	Resource     ResourceType `json:"resource"`
	Quantity     int          `json:"quantity"`
	Price        int          `json:"price"`
	Fee          int          `json:"fee"`
	Turn         int          `json:"turn"`
	Timestamp    time.Time    `json:"timestamp"`
}

type PriceHistoryEntry struct {
	Resource ResourceType `json:"resource"`
	Turn     int          `json:"turn"`
	Price    int          `json:"price"`
	Volume   int          `json:"volume"`
}

type ResourceStats struct {
	Resource    ResourceType `json:"resource"`
	TotalMined  int          `json:"total_mined"`
	TotalUsed   int          `json:"total_used"`
	TotalTraded int          `json:"total_traded"`
	Reserve     int          `json:"reserve"`
}

type AuctionItemType string

const (
	AuctionItemTech     AuctionItemType = "tech"
	AuctionItemShip     AuctionItemType = "ship"
	AuctionItemBlueprint AuctionItemType = "blueprint"
)

type AuctionStatus string

const (
	AuctionStatusActive   AuctionStatus = "active"
	AuctionStatusFinished AuctionStatus = "finished"
	AuctionStatusExpired  AuctionStatus = "expired"
)

type AuctionItem struct {
	ItemType   AuctionItemType `json:"item_type"`
	ItemID     string          `json:"item_id"`
	ItemName   string          `json:"item_name"`
	ItemData   interface{}     `json:"item_data,omitempty"`
}

type Auction struct {
	ID            uuid.UUID     `json:"id"`
	GameID        uuid.UUID     `json:"game_id"`
	SellerID      uuid.UUID     `json:"seller_id"`
	Item          AuctionItem   `json:"item"`
	StartingPrice int           `json:"starting_price"`
	CurrentBid    int           `json:"current_bid"`
	CurrentBidder *uuid.UUID    `json:"current_bidder,omitempty"`
	StartTurn     int           `json:"start_turn"`
	Duration      int           `json:"duration"`
	Status        AuctionStatus `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
}

type AuctionBid struct {
	ID         uuid.UUID `json:"id"`
	AuctionID  uuid.UUID `json:"auction_id"`
	PlayerID   uuid.UUID `json:"player_id"`
	Amount     int       `json:"amount"`
	Turn       int       `json:"turn"`
	Timestamp  time.Time `json:"timestamp"`
}

type FuturesContractStatus string

const (
	FuturesStatusOpen       FuturesContractStatus = "open"
	FuturesStatusActive     FuturesContractStatus = "active"
	FuturesStatusSettled    FuturesContractStatus = "settled"
	FuturesStatusLiquidated FuturesContractStatus = "liquidated"
	FuturesStatusCancelled  FuturesContractStatus = "cancelled"
)

type FuturesContract struct {
	ID              uuid.UUID            `json:"id"`
	GameID          uuid.UUID            `json:"game_id"`
	CreatorID       uuid.UUID            `json:"creator_id"`
	AccepterID      *uuid.UUID           `json:"accepter_id,omitempty"`
	Resource        ResourceType         `json:"resource"`
	Quantity        int                  `json:"quantity"`
	ContractPrice   int                  `json:"contract_price"`
	DeliveryTurn    int                  `json:"delivery_turn"`
	CreatorMargin   int                  `json:"creator_margin"`
	AccepterMargin  int                  `json:"accepter_margin"`
	InitialMargin   int                  `json:"initial_margin"`
	Status          FuturesContractStatus `json:"status"`
	MarginCallTurn  *int                 `json:"margin_call_turn,omitempty"`
	MarginCallParty *uuid.UUID           `json:"margin_call_party,omitempty"`
	CreatedTurn     int                  `json:"created_turn"`
	SettledTurn     *int                 `json:"settled_turn,omitempty"`
	SettlementPrice *int                 `json:"settlement_price,omitempty"`
	CreatorPnL      *int                 `json:"creator_pnl,omitempty"`
	AccepterPnL     *int                 `json:"accepter_pnl,omitempty"`
	CreatedAt       time.Time            `json:"created_at"`
}

type FuturesSettlement struct {
	ContractID      uuid.UUID `json:"contract_id"`
	Resource        ResourceType `json:"resource"`
	Quantity        int       `json:"quantity"`
	ContractPrice   int       `json:"contract_price"`
	SettlementPrice int       `json:"settlement_price"`
	CreatorPnL      int       `json:"creator_pnl"`
	AccepterPnL     int       `json:"accepter_pnl"`
	IsLiquidated    bool      `json:"is_liquidated"`
	Turn            int       `json:"turn"`
	Timestamp       time.Time `json:"timestamp"`
}

type MarketManipulationPenalty struct {
	PlayerID      uuid.UUID    `json:"player_id"`
	GameID        uuid.UUID    `json:"game_id"`
	Resource      ResourceType `json:"resource"`
	TurnsLeft     int          `json:"turns_left"`
	FeeMultiplier float64      `json:"fee_multiplier"`
	CreatedTurn   int          `json:"created_turn"`
}
