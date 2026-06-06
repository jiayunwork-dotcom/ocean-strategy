package game

import (
	"math/rand"
	"strconv"
	"time"

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
					if ge.state.ResourceStats != nil && ge.state.ResourceStats[resource] != nil {
						ge.state.ResourceStats[resource].TotalMined += produced
					}
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
					if ge.state.ResourceStats != nil && ge.state.ResourceStats[resource] != nil {
						ge.state.ResourceStats[resource].TotalMined += produced
					}
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
			if ge.state.ResourceStats != nil && ge.state.ResourceStats[models.ResourceBiomaterial] != nil {
				ge.state.ResourceStats[models.ResourceBiomaterial].TotalMined += produced
			}
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
	ge.MatchOrders()
	ge.CalculateResourceStats()
	ge.UpdatePrices()
	ge.ProcessAuctions()
	ge.destroyDestroyedShips()
	ge.repairFacilities()
	ge.progressFacilityConstruction()
	ge.progressResearch()
	ge.resetShipMovePoints()
	ge.ProcessCooldowns()
}

func (ge *GameEngine) progressFacilityConstruction() {
	for _, facility := range ge.state.Facilities {
		if facility.BuildTurnsLeft > 0 {
			facility.BuildTurnsLeft--
			if facility.BuildTurnsLeft <= 0 {
				facility.IsActive = true
			}
		}
	}
}

func (ge *GameEngine) progressResearch() {
	for _, pt := range ge.state.Techs {
		if pt.Researching && !pt.Completed {
			pt.TurnsLeft--
			if pt.TurnsLeft <= 0 {
				pt.Completed = true
				pt.Researching = false
			}
		}
	}
}

func (ge *GameEngine) resetShipMovePoints() {
	for _, ship := range ge.state.Ships {
		ship.MovePoints = ship.Speed
	}
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

func (ge *GameEngine) MoveShip(shipID uuid.UUID, toQ, toR int) (bool, *models.BattleLog) {
	var ship *models.Ship
	for _, s := range ge.state.Ships {
		if s.ID == shipID {
			ship = s
			break
		}
	}
	if ship == nil {
		return false, nil
	}

	distance := HexDistance(ship.HexQ, ship.HexR, toQ, toR)
	if distance > ship.MovePoints {
		return false, nil
	}

	fuelCost := ge.calculateFuelCost(ship.HexQ, ship.HexR, toQ, toR)
	if ship.Fuel < fuelCost {
		return false, nil
	}

	ship.HexQ = toQ
	ship.HexR = toR
	ship.MovePoints -= distance
	ship.Fuel -= fuelCost

	var battleLog *models.BattleLog
	if ship.Type == models.ShipEscort {
		if log, triggered := ge.CheckAndTriggerBattle(shipID, toQ, toR); triggered {
			battleLog = log
		}
	}

	return true, battleLog
}

func (ge *GameEngine) calculateFuelCost(fromQ, fromR, toQ, toR int) int {
	fromHexKey := HexKey(fromQ, fromR)
	fromHex, ok := ge.state.Hexes[fromHexKey]
	if !ok || !fromHex.HasCurrent {
		return HexDistance(fromQ, fromR, toQ, toR) * 2
	}

	dq := toQ - fromQ
	dr := toR - fromR

	currentDir := axialDirections[fromHex.CurrentDir]
	cdq, cdr := currentDir[0], currentDir[1]

	dotProduct := dq*cdq + dr*cdr + (dq+dr)*(cdq+cdr)

	if dotProduct > 0 {
		return HexDistance(fromQ, fromR, toQ, toR) * 1
	} else if dotProduct < 0 {
		return HexDistance(fromQ, fromR, toQ, toR) * 4
	}

	return HexDistance(fromQ, fromR, toQ, toR) * 2
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

	player := ge.state.Players[ship.OwnerID]

	playerAdded := false
	if player != nil {
		found := false
		for _, hk := range player.DiscoveredHexes {
			if hk == hexKey {
				found = true
				break
			}
		}
		if !found {
			player.DiscoveredHexes = append(player.DiscoveredHexes, hexKey)
			playerAdded = true
		}
	}

	if playerAdded {
		for _, rel := range ge.state.Relations {
			if rel.Status == models.RelationAlliance {
				var allyID uuid.UUID
				if rel.Player1ID == ship.OwnerID {
					allyID = rel.Player2ID
				} else if rel.Player2ID == ship.OwnerID {
					allyID = rel.Player1ID
				} else {
					continue
				}

				ally := ge.state.Players[allyID]
				if ally != nil {
					found := false
					for _, hk := range ally.DiscoveredHexes {
						if hk == hexKey {
							found = true
							break
						}
					}
					if !found {
						ally.DiscoveredHexes = append(ally.DiscoveredHexes, hexKey)
					}
				}
			}
		}
	}

	if !hex.Discovered {
		hex.Discovered = true
		return true
	}

	return playerAdded
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

func (ge *GameEngine) InitDiplomaticRelations() {
	playerIDs := make([]uuid.UUID, 0, len(ge.state.Players))
	for id := range ge.state.Players {
		playerIDs = append(playerIDs, id)
	}

	for i := 0; i < len(playerIDs); i++ {
		for j := i + 1; j < len(playerIDs); j++ {
			relation := &models.DiplomaticRelation{
				GameID:      ge.state.Game.ID,
				Player1ID:   playerIDs[i],
				Player2ID:   playerIDs[j],
				Status:      models.RelationNeutral,
				HasNAP:      false,
				HasAlliance: false,
				AtWar:       false,
			}
			ge.state.Relations = append(ge.state.Relations, relation)
		}
	}
}

func (ge *GameEngine) GetRelation(player1ID, player2ID uuid.UUID) *models.DiplomaticRelation {
	for _, rel := range ge.state.Relations {
		if (rel.Player1ID == player1ID && rel.Player2ID == player2ID) ||
			(rel.Player1ID == player2ID && rel.Player2ID == player1ID) {
			return rel
		}
	}
	return nil
}

func (ge *GameEngine) HasCooldown(playerID uuid.UUID) bool {
	for _, cd := range ge.state.Cooldowns {
		if cd.PlayerID == playerID && cd.TurnsLeft > 0 {
			return true
		}
	}
	return false
}

func (ge *GameEngine) ProposeTreaty(fromPlayerID, toPlayerID uuid.UUID, treatyType models.TreatyType) (bool, string) {
	if fromPlayerID == toPlayerID {
		return false, "不能向自己提议条约"
	}

	fromPlayer := ge.state.Players[fromPlayerID]
	toPlayer := ge.state.Players[toPlayerID]
	if fromPlayer == nil || toPlayer == nil {
		return false, "玩家不存在"
	}

	if ge.HasCooldown(fromPlayerID) {
		return false, "该玩家正处于声誉冷却期，无法发起条约提议"
	}

	relation := ge.GetRelation(fromPlayerID, toPlayerID)
	if relation == nil {
		return false, "外交关系不存在"
	}

	for _, p := range ge.state.Proposals {
		if p.FromPlayerID == fromPlayerID && p.ToPlayerID == toPlayerID && p.Status == "pending" {
			return false, "已有待处理的提议"
		}
		if p.FromPlayerID == toPlayerID && p.ToPlayerID == fromPlayerID && p.Status == "pending" {
			return false, "对方已向你发起提议，请先回应"
		}
	}

	switch treatyType {
	case models.TreatyNAP:
		if relation.Status != models.RelationNeutral {
			return false, "只能向中立关系的玩家提议互不侵犯条约"
		}
	case models.TreatyAlliance:
		if relation.Status != models.RelationNAP {
			return false, "只能向互不侵犯关系的玩家提议军事同盟"
		}
	default:
		return false, "未知条约类型"
	}

	proposal := &models.DiplomaticProposal{
		ID:           uuid.New(),
		GameID:       ge.state.Game.ID,
		FromPlayerID: fromPlayerID,
		ToPlayerID:   toPlayerID,
		TreatyType:   treatyType,
		Status:       "pending",
		CreatedAt:    ge.state.Game.CurrentTurn,
	}
	ge.state.Proposals = append(ge.state.Proposals, proposal)

	ge.addGameLog("diplomacy", fromPlayer.Name+" 向 "+toPlayer.Name+" 发起了"+getTreatyName(treatyType)+"提议", &fromPlayerID)

	return true, ""
}

func getTreatyName(treatyType models.TreatyType) string {
	switch treatyType {
	case models.TreatyNAP:
		return "互不侵犯条约"
	case models.TreatyAlliance:
		return "军事同盟"
	default:
		return "条约"
	}
}

func (ge *GameEngine) RespondToProposal(proposalID uuid.UUID, acceptorPlayerID uuid.UUID, accept bool) (bool, string) {
	var proposal *models.DiplomaticProposal
	for _, p := range ge.state.Proposals {
		if p.ID == proposalID {
			proposal = p
			break
		}
	}

	if proposal == nil {
		return false, "提议不存在"
	}

	if proposal.ToPlayerID != acceptorPlayerID {
		return false, "你不是该提议的接收方"
	}

	if proposal.Status != "pending" {
		return false, "该提议已被处理"
	}

	fromPlayer := ge.state.Players[proposal.FromPlayerID]
	toPlayer := ge.state.Players[proposal.ToPlayerID]

	if accept {
		if ge.HasCooldown(proposal.FromPlayerID) {
			proposal.Status = "rejected"
			return false, "对方正处于声誉冷却期，自动拒绝"
		}

		relation := ge.GetRelation(proposal.FromPlayerID, proposal.ToPlayerID)
		if relation != nil {
			switch proposal.TreatyType {
			case models.TreatyNAP:
				relation.Status = models.RelationNAP
				relation.HasNAP = true
			case models.TreatyAlliance:
				relation.Status = models.RelationAlliance
				relation.HasAlliance = true
				ge.shareAllianceVision(proposal.FromPlayerID, proposal.ToPlayerID)
			}
		}
		proposal.Status = "accepted"
		ge.addGameLog("diplomacy", toPlayer.Name+" 接受了 "+fromPlayer.Name+" 的"+getTreatyName(proposal.TreatyType), &acceptorPlayerID)
	} else {
		proposal.Status = "rejected"
		ge.addGameLog("diplomacy", toPlayer.Name+" 拒绝了 "+fromPlayer.Name+" 的"+getTreatyName(proposal.TreatyType), &acceptorPlayerID)
	}

	return true, ""
}

func (ge *GameEngine) shareAllianceVision(player1ID, player2ID uuid.UUID) {
	player1 := ge.state.Players[player1ID]
	player2 := ge.state.Players[player2ID]

	if player1 == nil || player2 == nil {
		return
	}

	combinedDiscovered := make(map[string]bool)
	for _, key := range player1.DiscoveredHexes {
		combinedDiscovered[key] = true
	}
	for _, key := range player2.DiscoveredHexes {
		combinedDiscovered[key] = true
	}

	allDiscovered := make([]string, 0, len(combinedDiscovered))
	for key := range combinedDiscovered {
		allDiscovered = append(allDiscovered, key)
		if hex, ok := ge.state.Hexes[key]; ok {
			hex.Discovered = true
		}
	}

	player1.DiscoveredHexes = allDiscovered
	player2.DiscoveredHexes = allDiscovered
}

func (ge *GameEngine) BreakTreaty(playerID, otherPlayerID uuid.UUID) (bool, string) {
	relation := ge.GetRelation(playerID, otherPlayerID)
	if relation == nil {
		return false, "外交关系不存在"
	}

	if relation.Status == models.RelationNeutral || relation.Status == models.RelationHostile {
		return false, "当前没有可撕毁的条约"
	}

	player := ge.state.Players[playerID]
	otherPlayer := ge.state.Players[otherPlayerID]

	relation.Status = models.RelationHostile
	relation.HasNAP = false
	relation.HasAlliance = false
	relation.AtWar = true

	player.Reputation = max(0, player.Reputation-20)

	cooldown := &models.ReputationCooldown{
		PlayerID:  playerID,
		GameID:    ge.state.Game.ID,
		TurnsLeft: 3,
		Reason:    "撕毁条约",
	}
	ge.state.Cooldowns = append(ge.state.Cooldowns, cooldown)

	ge.addGameLog("diplomacy", player.Name+" 撕毁了与 "+otherPlayer.Name+" 的条约，双方进入敌对状态", &playerID)

	return true, ""
}

func (ge *GameEngine) ProcessCooldowns() {
	remaining := make([]*models.ReputationCooldown, 0)
	for _, cd := range ge.state.Cooldowns {
		cd.TurnsLeft--
		if cd.TurnsLeft > 0 {
			remaining = append(remaining, cd)
		}
	}
	ge.state.Cooldowns = remaining
}

func (ge *GameEngine) CheckAndTriggerBattle(shipID uuid.UUID, toQ, toR int) (*models.BattleLog, bool) {
	var movingShip *models.Ship
	for _, s := range ge.state.Ships {
		if s.ID == shipID {
			movingShip = s
			break
		}
	}
	if movingShip == nil || movingShip.Type != models.ShipEscort {
		return nil, false
	}

	hexKey := HexKey(toQ, toR)
	hex, ok := ge.state.Hexes[hexKey]
	if !ok || hex.OwnerID == nil {
		return nil, false
	}

	ownerID := *hex.OwnerID
	if ownerID == movingShip.OwnerID {
		return nil, false
	}

	relation := ge.GetRelation(movingShip.OwnerID, ownerID)
	if relation == nil || relation.Status != models.RelationHostile {
		return nil, false
	}

	var defenderShip *models.Ship
	for _, s := range ge.state.Ships {
		if s.OwnerID == ownerID && s.HexQ == toQ && s.HexR == toR && s.Type == models.ShipEscort {
			defenderShip = s
			break
		}
	}

	if defenderShip == nil {
		return nil, false
	}

	return ge.resolveBattle(movingShip, defenderShip, toQ, toR), true
}

func (ge *GameEngine) resolveBattle(attacker, defender *models.Ship, q, r int) *models.BattleLog {
	attackerDamage := max(1, attacker.Attack-defender.Defense/2)
	defenderDamage := max(1, defender.Attack-attacker.Defense/2)

	defender.Health -= attackerDamage
	attacker.Health -= defenderDamage

	attackerSunk := attacker.Health <= 0
	defenderSunk := defender.Health <= 0

	battleLog := &models.BattleLog{
		ID:             uuid.New(),
		GameID:         ge.state.Game.ID,
		Turn:           ge.state.Game.CurrentTurn,
		AttackerID:     attacker.OwnerID,
		DefenderID:     defender.OwnerID,
		AttackerShipID: attacker.ID,
		DefenderShipID: defender.ID,
		HexQ:           q,
		HexR:           r,
		AttackerDamage: attackerDamage,
		DefenderDamage: defenderDamage,
		AttackerSunk:   attackerSunk,
		DefenderSunk:   defenderSunk,
		Timestamp:      time.Now(),
	}
	ge.state.BattleLogs = append(ge.state.BattleLogs, battleLog)

	attackerPlayer := ge.state.Players[attacker.OwnerID]
	defenderPlayer := ge.state.Players[defender.OwnerID]

	logMsg := "战斗：" + attackerPlayer.Name + " 的护卫舰与 " + defenderPlayer.Name + " 的护卫舰在海域(" + strconv.Itoa(q) + "," + strconv.Itoa(r) + ")交战"
	if attackerSunk {
		logMsg += "，攻击方舰船被击沉"
	}
	if defenderSunk {
		logMsg += "，防守方舰船被击沉"
	}
	ge.addGameLog("battle", logMsg, nil)

	return battleLog
}

func (ge *GameEngine) addGameLog(logType, message string, playerID *uuid.UUID) {
	entry := &models.GameLogEntry{
		ID:        uuid.New(),
		GameID:    ge.state.Game.ID,
		Turn:      ge.state.Game.CurrentTurn,
		Message:   message,
		Type:      logType,
		PlayerID:  playerID,
		Timestamp: time.Now(),
	}
	ge.state.GameLogs = append(ge.state.GameLogs, entry)
}

func (ge *GameEngine) GetPendingProposalsForPlayer(playerID uuid.UUID) []*models.DiplomaticProposal {
	result := make([]*models.DiplomaticProposal, 0)
	for _, p := range ge.state.Proposals {
		if p.ToPlayerID == playerID && p.Status == "pending" {
			result = append(result, p)
		}
	}
	return result
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
