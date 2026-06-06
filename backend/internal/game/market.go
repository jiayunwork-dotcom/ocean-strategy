package game

import (
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"ocean-strategy/internal/models"
)

const (
	DefaultFeeRate         = 0.05
	AllianceFeeRate        = 0.025
	MinReputationForBuying = 30
	AuctionDuration        = 3
	MinBidIncrement        = 0.10
)

func (ge *GameEngine) InitMarket() {
	ge.state.CurrentPrices = map[models.ResourceType]int{
		models.ResourceOil:         50,
		models.ResourceGas:         40,
		models.ResourceManganese:   30,
		models.ResourceSulfide:     80,
		models.ResourceBiomaterial: 100,
	}

	ge.state.ResourceStats = make(map[models.ResourceType]*models.ResourceStats)
	resources := []models.ResourceType{
		models.ResourceOil,
		models.ResourceGas,
		models.ResourceManganese,
		models.ResourceSulfide,
		models.ResourceBiomaterial,
	}
	for _, r := range resources {
		ge.state.ResourceStats[r] = &models.ResourceStats{
			Resource:    r,
			TotalMined:  0,
			TotalUsed:   0,
			TotalTraded: 0,
			Reserve:     0,
		}
	}

	ge.state.MarketOrders = make([]*models.MarketOrder, 0)
	ge.state.TradeRecords = make([]*models.TradeRecord, 0)
	ge.state.PriceHistory = make([]*models.PriceHistoryEntry, 0)
	ge.state.Auctions = make([]*models.Auction, 0)
	ge.state.AuctionBids = make([]*models.AuctionBid, 0)
	ge.state.FrozenShips = make(map[uuid.UUID]uuid.UUID)
	ge.state.FrozenTechs = make(map[string]uuid.UUID)

	for _, r := range resources {
		ge.state.PriceHistory = append(ge.state.PriceHistory, &models.PriceHistoryEntry{
			Resource: r,
			Turn:     0,
			Price:    ge.state.CurrentPrices[r],
			Volume:   0,
		})
	}
}

func (ge *GameEngine) ResetResourceStats() {
	if ge.state.ResourceStats == nil {
		return
	}
	for _, stats := range ge.state.ResourceStats {
		stats.TotalMined = 0
		stats.TotalUsed = 0
		stats.TotalTraded = 0
	}
}

func (ge *GameEngine) PlaceOrder(playerID uuid.UUID, orderType models.OrderType, resource models.ResourceType, quantity int, price int) (bool, string) {
	player := ge.state.Players[playerID]
	if player == nil {
		return false, "玩家不存在"
	}

	if ge.state.Game.Phase != models.PhaseDecision {
		return false, "只能在决策阶段挂单"
	}

	if quantity <= 0 || price <= 0 {
		return false, "数量和价格必须大于0"
	}

	if player.Reputation < MinReputationForBuying && orderType == models.OrderTypeBuy {
		return false, "声誉低于30，无法挂买单"
	}

	if orderType == models.OrderTypeSell {
		hasResource := false
		for _, ship := range ge.state.Ships {
			if ship.OwnerID == playerID && ship.Cargo[resource] >= quantity {
				hasResource = true
				break
			}
		}
		if !hasResource {
			return false, "资源不足"
		}
	} else {
		totalCost := quantity * price
		fee := int(float64(totalCost) * DefaultFeeRate)
		if player.Money < totalCost+fee {
			return false, "金币不足（需支付货款+5%手续费）"
		}
	}

	order := &models.MarketOrder{
		ID:           uuid.New(),
		GameID:       ge.state.Game.ID,
		PlayerID:     playerID,
		OrderType:    orderType,
		Resource:     resource,
		Quantity:     quantity,
		RemainingQty: quantity,
		Price:        price,
		Status:       models.OrderStatusActive,
		CreatedTurn:  ge.state.Game.CurrentTurn,
		CreatedAt:    time.Now(),
	}

	ge.state.MarketOrders = append(ge.state.MarketOrders, order)

	orderTypeStr := "买单"
	if orderType == models.OrderTypeSell {
		orderTypeStr = "卖单"
	}
	ge.addGameLog("market", player.Name+" 挂出"+orderTypeStr+"："+strResource(resource)+" x"+
		strconv.Itoa(quantity)+" @ "+strconv.Itoa(price)+"金币", &playerID)

	return true, ""
}

func (ge *GameEngine) CancelOrder(playerID uuid.UUID, orderID uuid.UUID) (bool, string) {
	var order *models.MarketOrder
	for _, o := range ge.state.MarketOrders {
		if o.ID == orderID {
			order = o
			break
		}
	}
	if order == nil {
		return false, "订单不存在"
	}
	if order.PlayerID != playerID {
		return false, "不是你的订单"
	}
	if order.Status != models.OrderStatusActive && order.Status != models.OrderStatusPartial {
		return false, "订单无法取消"
	}

	order.Status = models.OrderStatusCancelled
	player := ge.state.Players[playerID]
	ge.addGameLog("market", player.Name+" 取消了订单", &playerID)

	return true, ""
}

func (ge *GameEngine) MatchOrders() {
	resources := []models.ResourceType{
		models.ResourceOil,
		models.ResourceGas,
		models.ResourceManganese,
		models.ResourceSulfide,
		models.ResourceBiomaterial,
	}

	for _, resource := range resources {
		ge.matchOrdersForResource(resource)
	}
}

func (ge *GameEngine) matchOrdersForResource(resource models.ResourceType) {
	buyOrders := make([]*models.MarketOrder, 0)
	sellOrders := make([]*models.MarketOrder, 0)

	for _, order := range ge.state.MarketOrders {
		if order.Resource != resource {
			continue
		}
		if order.Status != models.OrderStatusActive && order.Status != models.OrderStatusPartial {
			continue
		}
		if order.OrderType == models.OrderTypeBuy {
			buyOrders = append(buyOrders, order)
		} else {
			sellOrders = append(sellOrders, order)
		}
	}

	sort.Slice(buyOrders, func(i, j int) bool {
		if buyOrders[i].Price != buyOrders[j].Price {
			return buyOrders[i].Price > buyOrders[j].Price
		}
		return buyOrders[i].CreatedAt.Before(buyOrders[j].CreatedAt)
	})

	sort.Slice(sellOrders, func(i, j int) bool {
		if sellOrders[i].Price != sellOrders[j].Price {
			return sellOrders[i].Price < sellOrders[j].Price
		}
		return sellOrders[i].CreatedAt.Before(sellOrders[j].CreatedAt)
	})

	for _, buyOrder := range buyOrders {
		if buyOrder.Status == models.OrderStatusCancelled || buyOrder.Status == models.OrderStatusFilled {
			continue
		}
		for _, sellOrder := range sellOrders {
			if sellOrder.Status == models.OrderStatusCancelled || sellOrder.Status == models.OrderStatusFilled {
				continue
			}
			if buyOrder.PlayerID == sellOrder.PlayerID {
				continue
			}
			if ge.isHostile(buyOrder.PlayerID, sellOrder.PlayerID) {
				continue
			}
			if buyOrder.Price < sellOrder.Price {
				break
			}

			tradeQty := min(buyOrder.RemainingQty, sellOrder.RemainingQty)
			if tradeQty <= 0 {
				continue
			}

			ge.executeTrade(buyOrder, sellOrder, tradeQty)

			if buyOrder.RemainingQty == 0 {
				buyOrder.Status = models.OrderStatusFilled
			} else {
				buyOrder.Status = models.OrderStatusPartial
			}
			if sellOrder.RemainingQty == 0 {
				sellOrder.Status = models.OrderStatusFilled
			} else {
				sellOrder.Status = models.OrderStatusPartial
			}

			if buyOrder.RemainingQty == 0 {
				break
			}
		}
	}
}

func (ge *GameEngine) executeTrade(buyOrder, sellOrder *models.MarketOrder, quantity int) {
	tradePrice := sellOrder.Price
	buyer := ge.state.Players[buyOrder.PlayerID]
	seller := ge.state.Players[sellOrder.PlayerID]

	isAlliance := ge.isAlliance(buyOrder.PlayerID, sellOrder.PlayerID)
	feeRate := DefaultFeeRate
	if isAlliance {
		feeRate = AllianceFeeRate
	}

	totalAmount := quantity * tradePrice
	fee := int(float64(totalAmount) * feeRate)

	buyOrder.RemainingQty -= quantity
	sellOrder.RemainingQty -= quantity

	buyer.Money -= (totalAmount + fee)
	seller.Money += (totalAmount - fee)

	ge.transferResource(sellOrder.PlayerID, buyOrder.PlayerID, buyOrder.Resource, quantity)

	trade := &models.TradeRecord{
		ID:          uuid.New(),
		GameID:      ge.state.Game.ID,
		BuyOrderID:  buyOrder.ID,
		SellOrderID: sellOrder.ID,
		BuyerID:     buyOrder.PlayerID,
		SellerID:    sellOrder.PlayerID,
		Resource:    buyOrder.Resource,
		Quantity:    quantity,
		Price:       tradePrice,
		Fee:         fee,
		Turn:        ge.state.Game.CurrentTurn,
		Timestamp:   time.Now(),
	}
	ge.state.TradeRecords = append(ge.state.TradeRecords, trade)

	if ge.state.ResourceStats[buyOrder.Resource] != nil {
		ge.state.ResourceStats[buyOrder.Resource].TotalTraded += quantity
	}

	ge.addGameLog("market", buyer.Name+" 和 "+seller.Name+" 成交："+
		strResource(buyOrder.Resource)+" x"+strconv.Itoa(quantity)+" @ "+strconv.Itoa(tradePrice)+"金币", nil)
}

func (ge *GameEngine) transferResource(fromID, toID uuid.UUID, resource models.ResourceType, quantity int) {
	remaining := quantity
	for _, ship := range ge.state.Ships {
		if remaining <= 0 {
			break
		}
		if ship.OwnerID == fromID && ship.Cargo[resource] > 0 {
			take := min(remaining, ship.Cargo[resource])
			ship.Cargo[resource] -= take
			remaining -= take
		}
	}

	remaining = quantity
	for _, ship := range ge.state.Ships {
		if remaining <= 0 {
			break
		}
		if ship.OwnerID == toID {
			space := ship.CargoCapacity
			for _, qty := range ship.Cargo {
				space -= qty
			}
			if space > 0 {
				give := min(remaining, space)
				ship.Cargo[resource] += give
				remaining -= give
			}
		}
	}
}

func (ge *GameEngine) CalculateResourceStats() {
	for resource, stats := range ge.state.ResourceStats {
		totalReserve := 0
		for _, hex := range ge.state.Hexes {
			if hex.Resources != nil {
				totalReserve += hex.Resources[resource]
			}
		}
		for _, ship := range ge.state.Ships {
			if ship.Cargo != nil {
				totalReserve += ship.Cargo[resource]
			}
		}
		stats.Reserve = totalReserve
	}
}

func (ge *GameEngine) UpdatePrices() {
	turn := ge.state.Game.CurrentTurn

	for resource, price := range ge.state.CurrentPrices {
		stats := ge.state.ResourceStats[resource]
		if stats == nil {
			continue
		}

		supply := stats.TotalMined
		demand := stats.TotalUsed + stats.TotalTraded

		baseVolatility := 0.10
		if stats.Reserve < 1000 {
			baseVolatility = 0.20
		} else if stats.Reserve < 5000 {
			baseVolatility = 0.15
		}

		priceChange := 0.0
		if supply > 0 || demand > 0 {
			if demand == 0 && supply > 0 {
				priceChange = -baseVolatility * 0.5
			} else if supply == 0 && demand > 0 {
				priceChange = baseVolatility * 0.5
			} else {
				ratio := float64(demand) / float64(supply)
				priceChange = (ratio - 1.0) * baseVolatility
			}
		}

		newPrice := int(float64(price) * (1.0 + priceChange))
		minPrice := 10
		maxPrice := price * 5
		if newPrice < minPrice {
			newPrice = minPrice
		}
		if newPrice > maxPrice {
			newPrice = maxPrice
		}

		ge.state.CurrentPrices[resource] = newPrice

		volume := 0
		for _, trade := range ge.state.TradeRecords {
			if trade.Resource == resource && trade.Turn == turn {
				volume += trade.Quantity
			}
		}

		ge.state.PriceHistory = append(ge.state.PriceHistory, &models.PriceHistoryEntry{
			Resource: resource,
			Turn:     turn,
			Price:    newPrice,
			Volume:   volume,
		})
	}

	ge.TrimPriceHistory(20)
}

func (ge *GameEngine) TrimPriceHistory(maxTurns int) {
	turnThreshold := ge.state.Game.CurrentTurn - maxTurns
	filtered := make([]*models.PriceHistoryEntry, 0)
	for _, entry := range ge.state.PriceHistory {
		if entry.Turn >= turnThreshold {
			filtered = append(filtered, entry)
		}
	}
	ge.state.PriceHistory = filtered
}

func (ge *GameEngine) GetVisibleOrders(playerID uuid.UUID) []*models.MarketOrder {
	visible := make([]*models.MarketOrder, 0)
	for _, order := range ge.state.MarketOrders {
		if order.Status != models.OrderStatusActive && order.Status != models.OrderStatusPartial {
			continue
		}
		if order.PlayerID == playerID {
			visible = append(visible, order)
			continue
		}
		if ge.isHostile(playerID, order.PlayerID) {
			continue
		}
		visible = append(visible, order)
	}
	return visible
}

func (ge *GameEngine) isHostile(p1, p2 uuid.UUID) bool {
	rel := ge.GetRelation(p1, p2)
	if rel == nil {
		return false
	}
	return rel.Status == models.RelationHostile
}

func (ge *GameEngine) isAlliance(p1, p2 uuid.UUID) bool {
	rel := ge.GetRelation(p1, p2)
	if rel == nil {
		return false
	}
	return rel.Status == models.RelationAlliance
}

func (ge *GameEngine) CreateAuction(playerID uuid.UUID, itemType models.AuctionItemType, itemID string, startingPrice int) (bool, string, *models.Auction) {
	player := ge.state.Players[playerID]
	if player == nil {
		return false, "玩家不存在", nil
	}

	if ge.state.Game.Phase != models.PhaseDecision {
		return false, "只能在决策阶段发起拍卖", nil
	}

	if startingPrice <= 0 {
		return false, "起拍价必须大于0", nil
	}

	itemName := ""
	switch itemType {
	case models.AuctionItemTech:
		found := false
		for _, pt := range ge.state.Techs {
			if pt.PlayerID == playerID && pt.TechID == itemID && pt.Completed {
				found = true
				tech := GetTechnology(itemID)
				if tech != nil {
					itemName = tech.Name
				}
				break
			}
		}
		if !found {
			return false, "未拥有该科技或研发未完成", nil
		}
		if _, frozen := ge.state.FrozenTechs[playerID.String()+":"+itemID]; frozen {
			return false, "该科技正在拍卖中", nil
		}

	case models.AuctionItemShip:
		shipUUID, err := uuid.Parse(itemID)
		if err != nil {
			return false, "无效的船只ID", nil
		}
		found := false
		for _, ship := range ge.state.Ships {
			if ship.ID == shipUUID && ship.OwnerID == playerID {
				found = true
				itemName = getShipTypeName(ship.Type)
				break
			}
		}
		if !found {
			return false, "未拥有该船只", nil
		}
		if _, frozen := ge.state.FrozenShips[shipUUID]; frozen {
			return false, "该船只正在拍卖中", nil
		}

	default:
		return false, "不支持的拍卖物品类型", nil
	}

	auction := &models.Auction{
		ID:            uuid.New(),
		GameID:        ge.state.Game.ID,
		SellerID:      playerID,
		Item: models.AuctionItem{
			ItemType: itemType,
			ItemID:   itemID,
			ItemName: itemName,
		},
		StartingPrice: startingPrice,
		CurrentBid:    startingPrice,
		CurrentBidder: nil,
		StartTurn:     ge.state.Game.CurrentTurn,
		Duration:      AuctionDuration,
		Status:        models.AuctionStatusActive,
		CreatedAt:     time.Now(),
	}

	ge.state.Auctions = append(ge.state.Auctions, auction)

	switch itemType {
	case models.AuctionItemTech:
		ge.state.FrozenTechs[playerID.String()+":"+itemID] = auction.ID
	case models.AuctionItemShip:
		shipUUID, _ := uuid.Parse(itemID)
		ge.state.FrozenShips[shipUUID] = auction.ID
	}

	ge.addGameLog("auction", player.Name+" 发起拍卖："+itemName+"，起拍价 "+strconv.Itoa(startingPrice)+"金币", &playerID)

	return true, "", auction
}

func (ge *GameEngine) PlaceBid(playerID uuid.UUID, auctionID uuid.UUID, amount int) (bool, string) {
	player := ge.state.Players[playerID]
	if player == nil {
		return false, "玩家不存在"
	}

	if ge.state.Game.Phase != models.PhaseDecision {
		return false, "只能在决策阶段出价"
	}

	var auction *models.Auction
	for _, a := range ge.state.Auctions {
		if a.ID == auctionID {
			auction = a
			break
		}
	}
	if auction == nil {
		return false, "拍卖不存在"
	}
	if auction.Status != models.AuctionStatusActive {
		return false, "拍卖已结束"
	}
	if auction.SellerID == playerID {
		return false, "不能给自己的拍卖出价"
	}

	minBid := int(float64(auction.CurrentBid) * (1.0 + MinBidIncrement))
	if auction.CurrentBidder != nil && amount < minBid {
		return false, "出价必须至少高于当前最高价10%，最低 "+strconv.Itoa(minBid)+"金币"
	}
	if auction.CurrentBidder == nil && amount < auction.StartingPrice {
		return false, "出价不能低于起拍价"
	}
	if player.Money < amount {
		return false, "金币不足"
	}

	bid := &models.AuctionBid{
		ID:        uuid.New(),
		AuctionID: auctionID,
		PlayerID:  playerID,
		Amount:    amount,
		Turn:      ge.state.Game.CurrentTurn,
		Timestamp: time.Now(),
	}
	ge.state.AuctionBids = append(ge.state.AuctionBids, bid)

	auction.CurrentBid = amount
	auction.CurrentBidder = &playerID

	ge.addGameLog("auction", player.Name+" 对 "+auction.Item.ItemName+" 出价 "+strconv.Itoa(amount)+"金币", &playerID)

	return true, ""
}

func (ge *GameEngine) ProcessAuctions() {
	turn := ge.state.Game.CurrentTurn

	for _, auction := range ge.state.Auctions {
		if auction.Status != models.AuctionStatusActive {
			continue
		}

		elapsed := turn - auction.StartTurn
		if elapsed >= auction.Duration {
			ge.finalizeAuction(auction)
		}
	}
}

func (ge *GameEngine) finalizeAuction(auction *models.Auction) {
	seller := ge.state.Players[auction.SellerID]
	if seller == nil {
		return
	}

	if auction.CurrentBidder != nil {
		buyer := ge.state.Players[*auction.CurrentBidder]
		if buyer != nil && buyer.Money >= auction.CurrentBid {
			buyer.Money -= auction.CurrentBid
			seller.Money += auction.CurrentBid

			ge.transferAuctionItem(auction)

			auction.Status = models.AuctionStatusFinished

			ge.addGameLog("auction", buyer.Name+" 以 "+strconv.Itoa(auction.CurrentBid)+
				"金币 拍得 "+auction.Item.ItemName, nil)
		} else {
			auction.Status = models.AuctionStatusExpired
			ge.addGameLog("auction", "拍卖 "+auction.Item.ItemName+" 因买家金币不足流拍", nil)
		}
	} else {
		auction.Status = models.AuctionStatusExpired
		ge.addGameLog("auction", "拍卖 "+auction.Item.ItemName+" 无人出价，流拍", &auction.SellerID)
	}

	ge.unfreezeAuctionItem(auction)
}

func (ge *GameEngine) transferAuctionItem(auction *models.Auction) {
	if auction.CurrentBidder == nil {
		return
	}
	buyerID := *auction.CurrentBidder

	switch auction.Item.ItemType {
	case models.AuctionItemTech:
		tech := GetTechnology(auction.Item.ItemID)
		if tech != nil {
			// 移除卖家的科技
			newTechs := make([]*models.PlayerTech, 0)
			for _, pt := range ge.state.Techs {
				if !(pt.PlayerID == auction.SellerID && pt.TechID == tech.ID) {
					newTechs = append(newTechs, pt)
				}
			}
			ge.state.Techs = newTechs

			// 给买家添加科技
			ge.state.Techs = append(ge.state.Techs, &models.PlayerTech{
				PlayerID:    buyerID,
				TechID:      tech.ID,
				Researching: false,
				TurnsLeft:   0,
				Completed:   true,
			})
		}

	case models.AuctionItemShip:
		shipID, _ := uuid.Parse(auction.Item.ItemID)
		for _, ship := range ge.state.Ships {
			if ship.ID == shipID {
				ship.OwnerID = buyerID
				break
			}
		}
	}
}

func (ge *GameEngine) unfreezeAuctionItem(auction *models.Auction) {
	switch auction.Item.ItemType {
	case models.AuctionItemTech:
		key := auction.SellerID.String() + ":" + auction.Item.ItemID
		delete(ge.state.FrozenTechs, key)
	case models.AuctionItemShip:
		shipID, _ := uuid.Parse(auction.Item.ItemID)
		delete(ge.state.FrozenShips, shipID)
	}
}

func (ge *GameEngine) GetVisibleAuctions(playerID uuid.UUID) []*models.Auction {
	visible := make([]*models.Auction, 0)
	for _, auction := range ge.state.Auctions {
		if ge.isHostile(playerID, auction.SellerID) {
			continue
		}
		visible = append(visible, auction)
	}
	return visible
}

func getShipTypeName(st models.ShipType) string {
	names := map[models.ShipType]string{
		models.ShipExplorer:    "勘探船",
		models.ShipConstructor: "工程船",
		models.ShipTransport:   "运输船",
		models.ShipEscort:      "护卫舰",
	}
	return names[st]
}

func strResource(r models.ResourceType) string {
	names := map[models.ResourceType]string{
		models.ResourceOil:         "石油",
		models.ResourceGas:         "天然气",
		models.ResourceManganese:   "锰结核",
		models.ResourceSulfide:     "多金属硫化物",
		models.ResourceBiomaterial: "生物原料",
	}
	return names[r]
}
