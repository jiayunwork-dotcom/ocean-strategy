package game

import (
	"math"
	"math/rand"

	"ocean-strategy/internal/models"
)

const (
	hexWidth  = 1.0
	hexHeight = math.Sqrt(3) / 2
)

var axialDirections = [][2]int{
	{1, 0}, {1, -1}, {0, -1},
	{-1, 0}, {-1, 1}, {0, 1},
}

type MapGenerator struct {
	radius int
	rng    *rand.Rand
}

func NewMapGenerator(radius int, seed int64) *MapGenerator {
	return &MapGenerator{
		radius: radius,
		rng:    rand.New(rand.NewSource(seed)),
	}
}

func (mg *MapGenerator) Generate() map[string]*models.Hex {
	hexes := make(map[string]*models.Hex)

	for q := -mg.radius; q <= mg.radius; q++ {
		r1 := max(-mg.radius, -q-mg.radius)
		r2 := min(mg.radius, -q+mg.radius)
		for r := r1; r <= r2; r++ {
			hex := mg.generateHex(q, r)
			hexes[HexKey(q, r)] = hex
		}
	}

	mg.addCurrents(hexes)
	mg.distributeResources(hexes)

	return hexes
}

func (mg *MapGenerator) generateHex(q, r int) *models.Hex {
	terrain := mg.generateTerrain(q, r)
	ecoHealth := 100
	if terrain == models.TerrainReef {
		ecoHealth = 100
	}

	return &models.Hex{
		Q:               q,
		R:               r,
		Terrain:         terrain,
		Resources:       make(map[models.ResourceType]int),
		Discovered:      false,
		EcologicalHealth: ecoHealth,
		Pollution:       0,
		HasCurrent:      false,
		CurrentDir:      0,
		IsEEZ:           false,
	}
}

func (mg *MapGenerator) generateTerrain(q, r int) models.HexTerrain {
	distance := HexDistance(0, 0, q, r)
	randVal := mg.rng.Float64()

	switch {
	case distance <= 2:
		return models.TerrainShallow
	case distance <= 4:
		if randVal < 0.6 {
			return models.TerrainDeep
		} else if randVal < 0.75 {
			return models.TerrainShallow
		} else if randVal < 0.85 {
			return models.TerrainReef
		}
		return models.TerrainOpenOcean
	default:
		if randVal < 0.35 {
			return models.TerrainDeep
		} else if randVal < 0.5 {
			return models.TerrainTrench
		} else if randVal < 0.6 {
			return models.TerrainVent
		} else if randVal < 0.7 {
			return models.TerrainReef
		} else if randVal < 0.8 {
			return models.TerrainShallow
		}
		return models.TerrainOpenOcean
	}
}

func (mg *MapGenerator) addCurrents(hexes map[string]*models.Hex) {
	for _, hex := range hexes {
		if mg.rng.Float64() < 0.3 {
			hex.HasCurrent = true
			hex.CurrentDir = mg.rng.Intn(6)
		}
	}
}

func (mg *MapGenerator) distributeResources(hexes map[string]*models.Hex) {
	for _, hex := range hexes {
		switch hex.Terrain {
		case models.TerrainShallow:
			if mg.rng.Float64() < 0.4 {
				hex.Resources[models.ResourceOil] = mg.rng.Intn(200) + 50
				hex.Resources[models.ResourceGas] = mg.rng.Intn(150) + 30
			}
		case models.TerrainDeep:
			if mg.rng.Float64() < 0.6 {
				hex.Resources[models.ResourceManganese] = mg.rng.Intn(500) + 100
			}
			if mg.rng.Float64() < 0.2 {
				hex.Resources[models.ResourceOil] = mg.rng.Intn(300) + 100
			}
		case models.TerrainTrench:
			if mg.rng.Float64() < 0.8 {
				hex.Resources[models.ResourceSulfide] = mg.rng.Intn(800) + 200
			}
		case models.TerrainVent:
			hex.Resources[models.ResourceSulfide] = mg.rng.Intn(600) + 300
			hex.Resources[models.ResourceBiomaterial] = mg.rng.Intn(400) + 100
		case models.TerrainReef:
			hex.Resources[models.ResourceBiomaterial] = mg.rng.Intn(300) + 50
		}
	}
}

func HexKey(q, r int) string {
	return string(rune(q+100)) + string(rune(r+100))
}

func HexDistance(q1, r1, q2, r2 int) int {
	return (abs(q1-q2) + abs(q1+r1-q2-r2) + abs(r1-r2)) / 2
}

func HexNeighbors(q, r int) [][2]int {
	neighbors := make([][2]int, 6)
	for i, dir := range axialDirections {
		neighbors[i] = [2]int{q + dir[0], r + dir[1]}
	}
	return neighbors
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GetStartPositions(radius int, numPlayers int) [][2]int {
	positions := make([][2]int, numPlayers)
	angleStep := 2 * math.Pi / float64(numPlayers)
	startRadius := max(1, radius-2)

	for i := 0; i < numPlayers; i++ {
		angle := angleStep * float64(i)
		q := int(math.Round(float64(startRadius) * math.Cos(angle)))
		r := int(math.Round(float64(startRadius) * math.Sin(angle)))
		positions[i] = [2]int{q, r}
	}

	return positions
}
