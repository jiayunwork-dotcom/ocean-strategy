package game

import (
	"math/rand"

	"github.com/google/uuid"
	"ocean-strategy/internal/models"
)

type GameEngine struct {
	state *models.GameState
	rng   *rand.Rand
}

func NewGameEngine(state *models.GameState, seed int64) *GameEngine {
	return &GameEngine{
		state: state,
		rng:   rand.New(rand.NewSource(seed)),
	}
}

func (ge *GameEngine) GetState() *models.GameState {
	return ge.state
}

func (ge *GameEngine) NextPhase() {
	switch ge.state.Game.Phase {
	case models.PhaseProduction:
		ge.runProductionPhase()
		ge.state.Game.Phase = models.PhaseDecision
	case models.PhaseDecision:
		ge.state.Game.Phase = models.PhaseEvent
	case models.PhaseEvent:
		ge.runEventPhase()
		ge.state.Game.Phase = models.PhaseSettlement
	case models.PhaseSettlement:
		ge.runSettlementPhase()
		ge.state.Game.CurrentTurn++
		ge.state.Game.Phase = models.PhaseProduction
	}
}

func (ge *GameEngine) runProductionPhase() {
	for _, facility := range ge.state.Facilities {
		if !facility.IsActive || facility.BuildTurnsLeft > 0 {
			continue
		}

		hexKey := HexKey(facility.HexQ, facility.HexR)
		hex := ge.state.Hexes[hexKey]
		player := ge.state.Players[facility.OwnerID]

		ge.produceFromFacility(facility, hex, player)
	}

	for _, player := range ge.state.Players {
		player.Money += 1000
	}
}

func (ge *GameEngine) produceFromFacility(facility *models.Facility, hex *models.Hex, player *models.Player) {
	productionMultiplier := float64(facility.Level) * (float64(facility.Health) / float64(facility.MaxHealth))

	switch facility.Type {
	case models.FacilityDrilling:
		for resource, amount := range hex.Resources {
			if resource == models.ResourceOil || resource == models.ResourceGas {
				produced := int(float64(amount) * 0.05 * productionMultiplier)
				if produced > 0 && hex.Resources[resource] > 0 {
					player.Money += produced * ge.getMarketPrice(resource)
					hex.Resources[resource] -= produced
					hex.Pollution = min(100, hex.Pollution+2)
				}
			}
		}

	case models.FacilityMine:
		for resource, amount := range hex.Resources {
			if resource == models.ResourceManganese || resource == models.ResourceSulfide {
				produced := int(float64(amount) * 0.03 * productionMultiplier)
				if produced > 0 && hex.Resources[resource] > 0 {
					player.Money += produced * ge.getMarketPrice(resource)
					hex.Resources[resource] -= produced
				}
			}
		}
		facility.Health = max(0, facility.Health-1)

	case models.FacilityTidal:
		if hex.HasCurrent {
			player.Money += int(100 * productionMultiplier)
		} else {
			player.Money += int(50 * productionMultiplier)
		}

	case models.FacilityFarm:
		if hex.EcologicalHealth > 30 {
			produced := int(50 * productionMultiplier * (float64(hex.EcologicalHealth) / 100))
			player.Money += produced * ge.getMarketPrice(models.ResourceBiomaterial)
			hex.EcologicalHealth = max(0, hex.EcologicalHealth-1)
		}
	}

	player.Money -= facility.MaintenanceCost
}

func (ge *GameEngine) getMarketPrice(resource models.ResourceType) int {
	prices := map[models.ResourceType]int{
		models.ResourceOil:         50,
		models.ResourceGas:         40,
		models.ResourceManganese:   30,
		models.ResourceSulfide:     80,
		models.ResourceBiomaterial: 100,
	}
	return prices[resource]
}

func (ge *GameEngine) runEventPhase() {
	ge.moveTyphoons()
	ge.generateTyphoons()
	ge.applyTyphoonDamage()
	ge.updateCurrents()
	ge.spreadPollution()
	ge.updateEcology()
}

func (ge *GameEngine) generateTyphoons() {
	if ge.rng.Float64() < 0.15 {
		q := ge.rng.Intn(ge.state.Game.MapRadius*2) - ge.state.Game.MapRadius
		r := ge.rng.Intn(ge.state.Game.MapRadius*2) - ge.state.Game.MapRadius
		if HexDistance(0, 0, q, r) <= ge.state.Game.MapRadius {
			typhoon := &models.Typhoon{
				ID:        uuid.New(),
				HexQ:      q,
				HexR:      r,
				Strength:  ge.rng.Intn(5) + 3,
				DirQ:      ge.rng.Intn(3) - 1,
				DirR:      ge.rng.Intn(3) - 1,
				TurnsLeft: ge.rng.Intn(4) + 3,
			}
			ge.state.Typhoons = append(ge.state.Typhoons, typhoon)
		}
	}
}

func (ge *GameEngine) moveTyphoons() {
	remaining := make([]*models.Typhoon, 0)
	for _, t := range ge.state.Typhoons {
		t.HexQ += t.DirQ
		t.HexR += t.DirR
		t.TurnsLeft--
		if t.TurnsLeft > 0 && HexDistance(0, 0, t.HexQ, t.HexR) <= ge.state.Game.MapRadius+2 {
			remaining = append(remaining, t)
		}
	}
	ge.state.Typhoons = remaining
}

func (ge *GameEngine) applyTyphoonDamage() {
	for _, t := range ge.state.Typhoons {
		neighbors := HexNeighbors(t.HexQ, t.HexR)
		affectedHexes := append([][2]int{{t.HexQ, t.HexR}}, neighbors...)

		for _, pos := range affectedHexes {
			hexKey := HexKey(pos[0], pos[1])
			if hex, ok := ge.state.Hexes[hexKey]; ok {
				if hex.Facility != nil && ge.rng.Float64() < float64(t.Strength)*0.1 {
					hex.Facility.Health = max(0, hex.Facility.Health-t.Strength*10)
				}
			}
		}

		for _, ship := range ge.state.Ships {
			if HexDistance(ship.HexQ, ship.HexR, t.HexQ, t.HexR) <= 1 {
				if ge.rng.Float64() < float64(t.Strength)*0.15 {
					ship.Health = max(0, ship.Health-t.Strength*15)
				}
			}
		}
	}
}

func (ge *GameEngine) updateCurrents() {
	if ge.state.Game.CurrentTurn%5 == 0 {
		for _, hex := range ge.state.Hexes {
			if hex.HasCurrent {
				hex.CurrentDir = (hex.CurrentDir + ge.rng.Intn(3) - 1 + 6) % 6
			}
		}
	}
}

func (ge *GameEngine) spreadPollution() {
	for _, hex := range ge.state.Hexes {
		if hex.Pollution > 0 {
			neighbors := HexNeighbors(hex.Q, hex.R)
			spreadAmount := hex.Pollution / 10
			for _, pos := range neighbors {
				if nHex, ok := ge.state.Hexes[HexKey(pos[0], pos[1])]; ok {
					nHex.Pollution = min(100, nHex.Pollution+spreadAmount/6)
				}
			}
		}
	}
}

func (ge *GameEngine) updateEcology() {
	for _, hex := range ge.state.Hexes {
		if hex.Terrain == models.TerrainReef {
			hex.EcologicalHealth = max(0, hex.EcologicalHealth-hex.Pollution/20)
		}
	}
}

func (ge *GameEngine) runSettlementPhase() {
	ge.destroyDestroyedShips()
	ge.repairFacilities()
}

func (ge *GameEngine) destroyDestroyedShips() {
	alive := make([]*models.Ship, 0)
	for _, ship := range ge.state.Ships {
		if ship.Health > 0 {
			alive = append(alive, ship)
		}
	}
	ge.state.Ships = alive
}

func (ge *GameEngine) repairFacilities() {
	for _, facility := range ge.state.Facilities {
		if facility.Health < facility.MaxHealth {
			player := ge.state.Players[facility.OwnerID]
			repairCost := (facility.MaxHealth - facility.Health) * 2
			if player.Money >= repairCost {
				player.Money -= repairCost
				facility.Health = min(facility.MaxHealth, facility.Health+10)
			}
		}
	}
}

func (ge *GameEngine) BuildFacility(playerID uuid.UUID, q, r int, facilityType models.FacilityType) bool {
	hexKey := HexKey(q, r)
	hex, ok := ge.state.Hexes[hexKey]
	if !ok || hex.Facility != nil {
		return false
	}

	player := ge.state.Players[playerID]
	cost := ge.getFacilityCost(facilityType)
	if player.Money < cost {
		return false
	}

	if !ge.canBuildFacility(hex, facilityType) {
		return false
	}

	player.Money -= cost

	facility := &models.Facility{
		ID:            uuid.New(),
		Type:          facilityType,
		Level:         1,
		HexQ:          q,
		HexR:          r,
		OwnerID:       playerID,
		Health:        100,
		MaxHealth:     100,
		MaintenanceCost: cost / 20,
		BuildTurnsLeft: ge.getBuildTurns(facilityType),
		IsActive:      false,
	}

	hex.Facility = facility
	hex.OwnerID = &playerID
	ge.state.Facilities = append(ge.state.Facilities, facility)

	return true
}

func (ge *GameEngine) getFacilityCost(ft models.FacilityType) int {
	costs := map[models.FacilityType]int{
		models.FacilityDrilling: 5000,
		models.FacilityMine:     8000,
		models.FacilityTidal:    3000,
		models.FacilityFarm:     4000,
		models.FacilityPort:     10000,
	}
	return costs[ft]
}

func (ge *GameEngine) getBuildTurns(ft models.FacilityType) int {
	turns := map[models.FacilityType]int{
		models.FacilityDrilling: 3,
		models.FacilityMine:     5,
		models.FacilityTidal:    2,
		models.FacilityFarm:     3,
		models.FacilityPort:     6,
	}
	return turns[ft]
}

func (ge *GameEngine) canBuildFacility(hex *models.Hex, ft models.FacilityType) bool {
	switch ft {
	case models.FacilityDrilling:
		return hex.Terrain == models.TerrainShallow || hex.Terrain == models.TerrainDeep
	case models.FacilityMine:
		return hex.Terrain == models.TerrainDeep || hex.Terrain == models.TerrainTrench || hex.Terrain == models.TerrainVent
	case models.FacilityTidal:
		return hex.HasCurrent
	case models.FacilityFarm:
		return hex.Terrain == models.TerrainReef
	case models.FacilityPort:
		return hex.Terrain == models.TerrainShallow
	}
	return false
}

func (ge *GameEngine) BuildShip(playerID uuid.UUID, shipType models.ShipType, q, r int) bool {
	hexKey := HexKey(q, r)
	hex, ok := ge.state.Hexes[hexKey]
	if !ok || hex.Facility == nil || hex.Facility.Type != models.FacilityPort {
		return false
	}

	player := ge.state.Players[playerID]
	cost := ge.getShipCost(shipType)
	if player.Money < cost {
		return false
	}

	player.Money -= cost

	ship := &models.Ship{
		ID:            uuid.New(),
		Type:          shipType,
		OwnerID:       playerID,
		HexQ:          q,
		HexR:          r,
		Health:        ge.getShipMaxHealth(shipType),
		MaxHealth:     ge.getShipMaxHealth(shipType),
		Fuel:          ge.getShipMaxFuel(shipType),
		MaxFuel:       ge.getShipMaxFuel(shipType),
		Cargo:         make(map[models.ResourceType]int),
		CargoCapacity: ge.getShipCargoCapacity(shipType),
		Speed:         ge.getShipSpeed(shipType),
		MovePoints:    ge.getShipSpeed(shipType),
		Attack:        ge.getShipAttack(shipType),
		Defense:       ge.getShipDefense(shipType),
	}

	ge.state.Ships = append(ge.state.Ships, ship)
	return true
}

func (ge *GameEngine) getShipCost(st models.ShipType) int {
	costs := map[models.ShipType]int{
		models.ShipExplorer:    2000,
		models.ShipConstructor: 3000,
		models.ShipTransport:   2500,
		models.ShipEscort:      4000,
	}
	return costs[st]
}

func (ge *GameEngine) getShipMaxHealth(st models.ShipType) int {
	hp := map[models.ShipType]int{
		models.ShipExplorer:    50,
		models.ShipConstructor: 80,
		models.ShipTransport:   60,
		models.ShipEscort:      100,
	}
	return hp[st]
}

func (ge *GameEngine) getShipMaxFuel(st models.ShipType) int {
	fuel := map[models.ShipType]int{
		models.ShipExplorer:    100,
		models.ShipConstructor: 80,
		models.ShipTransport:   120,
		models.ShipEscort:      90,
	}
	return fuel[st]
}

func (ge *GameEngine) getShipCargoCapacity(st models.ShipType) int {
	cargo := map[models.ShipType]int{
		models.ShipExplorer:    20,
		models.ShipConstructor: 100,
		models.ShipTransport:   200,
		models.ShipEscort:      30,
	}
	return cargo[st]
}

func (ge *GameEngine) getShipSpeed(st models.ShipType) int {
	speed := map[models.ShipType]int{
		models.ShipExplorer:    4,
		models.ShipConstructor: 2,
		models.ShipTransport:   3,
		models.ShipEscort:      4,
	}
	return speed[st]
}

func (ge *GameEngine) getShipAttack(st models.ShipType) int {
	attack := map[models.ShipType]int{
		models.ShipExplorer:    5,
		models.ShipConstructor: 10,
		models.ShipTransport:   5,
		models.ShipEscort:      25,
	}
	return attack[st]
}

func (ge *GameEngine) getShipDefense(st models.ShipType) int {
	defense := map[models.ShipType]int{
		models.ShipExplorer:    5,
		models.ShipConstructor: 15,
		models.ShipTransport:   10,
		models.ShipEscort:      20,
	}
	return defense[st]
}

func (ge *GameEngine) MoveShip(shipID uuid.UUID, toQ, toR int) bool {
	var ship *models.Ship
	for _, s := range ge.state.Ships {
		if s.ID == shipID {
			ship = s
			break
		}
	}
	if ship == nil {
		return false
	}

	distance := HexDistance(ship.HexQ, ship.HexR, toQ, toR)
	if distance > ship.MovePoints {
		return false
	}

	if ship.Fuel < distance*2 {
		return false
	}

	ship.HexQ = toQ
	ship.HexR = toR
	ship.MovePoints -= distance
	ship.Fuel -= distance * 2

	return true
}

func (ge *GameEngine) Explore(shipID uuid.UUID) bool {
	var ship *models.Ship
	for _, s := range ge.state.Ships {
		if s.ID == shipID {
			ship = s
			break
		}
	}
	if ship == nil || ship.Type != models.ShipExplorer {
		return false
	}

	hexKey := HexKey(ship.HexQ, ship.HexR)
	hex := ge.state.Hexes[hexKey]

	if !hex.Discovered {
		hex.Discovered = true
		return true
	}

	return false
}

func (ge *GameEngine) StartResearch(playerID uuid.UUID, techID string) bool {
	for _, pt := range ge.state.Techs {
		if pt.PlayerID == playerID && pt.TechID == techID && !pt.Completed {
			return false
		}
	}

	tech := GetTechnology(techID)
	if tech == nil {
		return false
	}

	player := ge.state.Players[playerID]
	if player.Money < tech.Cost {
		return false
	}

	for _, prereq := range tech.Prerequisites {
		found := false
		for _, pt := range ge.state.Techs {
			if pt.PlayerID == playerID && pt.TechID == prereq && pt.Completed {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	player.Money -= tech.Cost

	ge.state.Techs = append(ge.state.Techs, &models.PlayerTech{
		PlayerID:    playerID,
		TechID:      techID,
		Researching: true,
		TurnsLeft:   tech.Turns,
		Completed:   false,
	})

	return true
}

func GetTechnology(id string) *models.Technology {
	for _, t := range Technologies {
		if t.ID == id {
			return t
		}
	}
	return nil
}

var Technologies = []*models.Technology{
	{
		ID:           "deep_drilling_1",
		Name:         "深海钻探 I",
		Category:     models.TechExtraction,
		Cost:         5000,
		Turns:        4,
		Description:  "允许在深海区域建造钻井平台",
		Prerequisites: []string{},
		Effects:      map[string]int{"deep_drilling": 1},
	},
	{
		ID:           "deep_drilling_2",
		Name:         "深海钻探 II",
		Category:     models.TechExtraction,
		Cost:         10000,
		Turns:        6,
		Description:  "提升深海开采效率20%",
		Prerequisites: []string{"deep_drilling_1"},
		Effects:      map[string]int{"extraction_efficiency": 20},
	},
	{
		ID:           "trench_mining",
		Name:         "海沟采矿技术",
		Category:     models.TechExtraction,
		Cost:         15000,
		Turns:        8,
		Description:  "允许在海沟区域建造矿山",
		Prerequisites: []string{"deep_drilling_1"},
		Effects:      map[string]int{"trench_mining": 1},
	},
	{
		ID:           "eco_friendly_1",
		Name:         "环保技术 I",
		Category:     models.TechEcology,
		Cost:         4000,
		Turns:        3,
		Description:  "降低污染扩散速度50%",
		Prerequisites: []string{},
		Effects:      map[string]int{"pollution_reduction": 50},
	},
	{
		ID:           "eco_friendly_2",
		Name:         "环保技术 II",
		Category:     models.TechEcology,
		Cost:         8000,
		Turns:        5,
		Description:  "珊瑚礁生态恢复,提升国际声誉",
		Prerequisites: []string{"eco_friendly_1"},
		Effects:      map[string]int{"reputation_bonus": 10},
	},
	{
		ID:           "sustainable_farm",
		Name:         "可持续养殖",
		Category:     models.TechEcology,
		Cost:         6000,
		Turns:        4,
		Description:  "养殖场产出提升30%,生态破坏减少",
		Prerequisites: []string{"eco_friendly_1"},
		Effects:      map[string]int{"farm_efficiency": 30},
	},
	{
		ID:           "ship_armor",
		Name:         "舰船装甲",
		Category:     models.TechMilitary,
		Cost:         6000,
		Turns:        4,
		Description:  "所有舰船生命值+20%",
		Prerequisites: []string{},
		Effects:      map[string]int{"ship_health": 20},
	},
	{
		ID:           "ship_weapons",
		Name:         "舰载武器",
		Category:     models.TechMilitary,
		Cost:         8000,
		Turns:        5,
		Description:  "护卫舰攻击力+30%",
		Prerequisites: []string{"ship_armor"},
		Effects:      map[string]int{"ship_attack": 30},
	},
	{
		ID:           "typhoon_resist",
		Name:         "抗台风技术",
		Category:     models.TechMilitary,
		Cost:         5000,
		Turns:        3,
		Description:  "降低台风对舰船的伤害50%",
		Prerequisites: []string{},
		Effects:      map[string]int{"typhoon_resist": 50},
	},
}
