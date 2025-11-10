package tournament

import "poker-platform/backend/internal/models"

// Predefined Tournament Structures
var (
	// TurboStructure - Fast blind increases (5-minute levels)
	TurboStructure = models.TournamentStructure{
		Name:        "Turbo",
		Description: "Fast-paced tournament with 5-minute blind levels",
		BlindLevels: []models.BlindLevel{
			{Level: 1, SmallBlind: 10, BigBlind: 20, Ante: 0, Duration: 300},    // 5 min
			{Level: 2, SmallBlind: 15, BigBlind: 30, Ante: 0, Duration: 300},
			{Level: 3, SmallBlind: 25, BigBlind: 50, Ante: 0, Duration: 300},
			{Level: 4, SmallBlind: 50, BigBlind: 100, Ante: 10, Duration: 300},
			{Level: 5, SmallBlind: 75, BigBlind: 150, Ante: 15, Duration: 300},
			{Level: 6, SmallBlind: 100, BigBlind: 200, Ante: 20, Duration: 300},
			{Level: 7, SmallBlind: 150, BigBlind: 300, Ante: 30, Duration: 300},
			{Level: 8, SmallBlind: 200, BigBlind: 400, Ante: 40, Duration: 300},
			{Level: 9, SmallBlind: 300, BigBlind: 600, Ante: 60, Duration: 300},
			{Level: 10, SmallBlind: 400, BigBlind: 800, Ante: 80, Duration: 300},
			{Level: 11, SmallBlind: 600, BigBlind: 1200, Ante: 120, Duration: 300},
			{Level: 12, SmallBlind: 800, BigBlind: 1600, Ante: 160, Duration: 300},
			{Level: 13, SmallBlind: 1000, BigBlind: 2000, Ante: 200, Duration: 300},
			{Level: 14, SmallBlind: 1500, BigBlind: 3000, Ante: 300, Duration: 300},
			{Level: 15, SmallBlind: 2000, BigBlind: 4000, Ante: 400, Duration: 300},
		},
	}

	// StandardStructure - Regular blind increases (10-minute levels)
	StandardStructure = models.TournamentStructure{
		Name:        "Standard",
		Description: "Standard tournament with 10-minute blind levels",
		BlindLevels: []models.BlindLevel{
			{Level: 1, SmallBlind: 25, BigBlind: 50, Ante: 0, Duration: 600},    // 10 min
			{Level: 2, SmallBlind: 50, BigBlind: 100, Ante: 0, Duration: 600},
			{Level: 3, SmallBlind: 75, BigBlind: 150, Ante: 0, Duration: 600},
			{Level: 4, SmallBlind: 100, BigBlind: 200, Ante: 25, Duration: 600},
			{Level: 5, SmallBlind: 150, BigBlind: 300, Ante: 30, Duration: 600},
			{Level: 6, SmallBlind: 200, BigBlind: 400, Ante: 50, Duration: 600},
			{Level: 7, SmallBlind: 300, BigBlind: 600, Ante: 75, Duration: 600},
			{Level: 8, SmallBlind: 400, BigBlind: 800, Ante: 100, Duration: 600},
			{Level: 9, SmallBlind: 600, BigBlind: 1200, Ante: 150, Duration: 600},
			{Level: 10, SmallBlind: 800, BigBlind: 1600, Ante: 200, Duration: 600},
			{Level: 11, SmallBlind: 1000, BigBlind: 2000, Ante: 250, Duration: 600},
			{Level: 12, SmallBlind: 1500, BigBlind: 3000, Ante: 375, Duration: 600},
			{Level: 13, SmallBlind: 2000, BigBlind: 4000, Ante: 500, Duration: 600},
			{Level: 14, SmallBlind: 3000, BigBlind: 6000, Ante: 750, Duration: 600},
			{Level: 15, SmallBlind: 4000, BigBlind: 8000, Ante: 1000, Duration: 600},
			{Level: 16, SmallBlind: 6000, BigBlind: 12000, Ante: 1500, Duration: 600},
			{Level: 17, SmallBlind: 8000, BigBlind: 16000, Ante: 2000, Duration: 600},
			{Level: 18, SmallBlind: 10000, BigBlind: 20000, Ante: 2500, Duration: 600},
		},
	}

	// DeepStackStructure - Slow blind increases (15-minute levels)
	DeepStackStructure = models.TournamentStructure{
		Name:        "Deep Stack",
		Description: "Deep stack tournament with 15-minute blind levels",
		BlindLevels: []models.BlindLevel{
			{Level: 1, SmallBlind: 25, BigBlind: 50, Ante: 0, Duration: 900},    // 15 min
			{Level: 2, SmallBlind: 50, BigBlind: 100, Ante: 0, Duration: 900},
			{Level: 3, SmallBlind: 75, BigBlind: 150, Ante: 0, Duration: 900},
			{Level: 4, SmallBlind: 100, BigBlind: 200, Ante: 0, Duration: 900},
			{Level: 5, SmallBlind: 150, BigBlind: 300, Ante: 25, Duration: 900},
			{Level: 6, SmallBlind: 200, BigBlind: 400, Ante: 50, Duration: 900},
			{Level: 7, SmallBlind: 250, BigBlind: 500, Ante: 50, Duration: 900},
			{Level: 8, SmallBlind: 300, BigBlind: 600, Ante: 75, Duration: 900},
			{Level: 9, SmallBlind: 400, BigBlind: 800, Ante: 100, Duration: 900},
			{Level: 10, SmallBlind: 500, BigBlind: 1000, Ante: 100, Duration: 900},
			{Level: 11, SmallBlind: 600, BigBlind: 1200, Ante: 150, Duration: 900},
			{Level: 12, SmallBlind: 800, BigBlind: 1600, Ante: 200, Duration: 900},
			{Level: 13, SmallBlind: 1000, BigBlind: 2000, Ante: 250, Duration: 900},
			{Level: 14, SmallBlind: 1500, BigBlind: 3000, Ante: 300, Duration: 900},
			{Level: 15, SmallBlind: 2000, BigBlind: 4000, Ante: 500, Duration: 900},
			{Level: 16, SmallBlind: 3000, BigBlind: 6000, Ante: 600, Duration: 900},
			{Level: 17, SmallBlind: 4000, BigBlind: 8000, Ante: 1000, Duration: 900},
			{Level: 18, SmallBlind: 5000, BigBlind: 10000, Ante: 1000, Duration: 900},
			{Level: 19, SmallBlind: 6000, BigBlind: 12000, Ante: 1500, Duration: 900},
			{Level: 20, SmallBlind: 8000, BigBlind: 16000, Ante: 2000, Duration: 900},
		},
	}

	// HyperTurboStructure - Ultra-fast blind increases (3-minute levels)
	HyperTurboStructure = models.TournamentStructure{
		Name:        "Hyper Turbo",
		Description: "Lightning-fast tournament with 3-minute blind levels",
		BlindLevels: []models.BlindLevel{
			{Level: 1, SmallBlind: 10, BigBlind: 20, Ante: 0, Duration: 180},    // 3 min
			{Level: 2, SmallBlind: 15, BigBlind: 30, Ante: 0, Duration: 180},
			{Level: 3, SmallBlind: 25, BigBlind: 50, Ante: 5, Duration: 180},
			{Level: 4, SmallBlind: 50, BigBlind: 100, Ante: 10, Duration: 180},
			{Level: 5, SmallBlind: 75, BigBlind: 150, Ante: 15, Duration: 180},
			{Level: 6, SmallBlind: 100, BigBlind: 200, Ante: 25, Duration: 180},
			{Level: 7, SmallBlind: 150, BigBlind: 300, Ante: 40, Duration: 180},
			{Level: 8, SmallBlind: 200, BigBlind: 400, Ante: 50, Duration: 180},
			{Level: 9, SmallBlind: 300, BigBlind: 600, Ante: 75, Duration: 180},
			{Level: 10, SmallBlind: 500, BigBlind: 1000, Ante: 100, Duration: 180},
			{Level: 11, SmallBlind: 750, BigBlind: 1500, Ante: 150, Duration: 180},
			{Level: 12, SmallBlind: 1000, BigBlind: 2000, Ante: 250, Duration: 180},
		},
	}
)

// Predefined Prize Structures
var (
	// WinnerTakesAll - 100% to 1st place
	WinnerTakesAll = models.PrizeStructureConfig{
		Name:        "Winner Takes All",
		Description: "Single winner receives entire prize pool",
		Positions: []models.PrizePosition{
			{Position: 1, Percentage: 100.0},
		},
	}

	// Top3Payout - Standard 3-place payout
	Top3Payout = models.PrizeStructureConfig{
		Name:        "Top 3",
		Description: "Prize distribution for top 3 finishers",
		Positions: []models.PrizePosition{
			{Position: 1, Percentage: 50.0},
			{Position: 2, Percentage: 30.0},
			{Position: 3, Percentage: 20.0},
		},
	}

	// Top5Payout - 5-place payout for medium tournaments
	Top5Payout = models.PrizeStructureConfig{
		Name:        "Top 5",
		Description: "Prize distribution for top 5 finishers",
		Positions: []models.PrizePosition{
			{Position: 1, Percentage: 40.0},
			{Position: 2, Percentage: 25.0},
			{Position: 3, Percentage: 17.0},
			{Position: 4, Percentage: 11.0},
			{Position: 5, Percentage: 7.0},
		},
	}

	// Top10Payout - 10-place payout for larger tournaments
	Top10Payout = models.PrizeStructureConfig{
		Name:        "Top 10",
		Description: "Prize distribution for top 10 finishers",
		Positions: []models.PrizePosition{
			{Position: 1, Percentage: 30.0},
			{Position: 2, Percentage: 20.0},
			{Position: 3, Percentage: 13.0},
			{Position: 4, Percentage: 10.0},
			{Position: 5, Percentage: 8.0},
			{Position: 6, Percentage: 6.0},
			{Position: 7, Percentage: 5.0},
			{Position: 8, Percentage: 4.0},
			{Position: 9, Percentage: 2.5},
			{Position: 10, Percentage: 1.5},
		},
	}

	// Top10PercentPayout - WSOP-style payout for top 10% of field
	Top10PercentPayout = models.PrizeStructureConfig{
		Name:        "Top 10% (WSOP Style)",
		Description: "Pays top 10% of field with standard WSOP structure",
		Positions: []models.PrizePosition{
			{Position: 1, Percentage: 30.0},
			{Position: 2, Percentage: 18.0},
			{Position: 3, Percentage: 12.0},
			{Position: 4, Percentage: 9.0},
			{Position: 5, Percentage: 7.0},
			{Position: 6, Percentage: 5.5},
			{Position: 7, Percentage: 4.5},
			{Position: 8, Percentage: 3.5},
			{Position: 9, Percentage: 2.8},
			{Position: 10, Percentage: 2.2},
			// Remaining positions get proportional distribution of remaining 5.5%
		},
	}

	// HeadsUpPayout - 50/50 for heads-up
	HeadsUpPayout = models.PrizeStructureConfig{
		Name:        "Heads-Up (50/50)",
		Description: "Even split for heads-up tournament",
		Positions: []models.PrizePosition{
			{Position: 1, Percentage: 65.0},
			{Position: 2, Percentage: 35.0},
		},
	}
)

// StructurePresets maps preset names to structures
var StructurePresets = map[string]models.TournamentStructure{
	"turbo":       TurboStructure,
	"standard":    StandardStructure,
	"deep_stack":  DeepStackStructure,
	"hyper_turbo": HyperTurboStructure,
}

// PrizeStructurePresets maps preset names to prize structures
var PrizeStructurePresets = map[string]models.PrizeStructureConfig{
	"winner_takes_all": WinnerTakesAll,
	"top_3":            Top3Payout,
	"top_5":            Top5Payout,
	"top_10":           Top10Payout,
	"top_10_percent":   Top10PercentPayout,
	"heads_up":         HeadsUpPayout,
}

// GetStructurePreset retrieves a tournament structure by name
func GetStructurePreset(name string) (models.TournamentStructure, bool) {
	preset, exists := StructurePresets[name]
	return preset, exists
}

// GetPrizeStructurePreset retrieves a prize structure by name
func GetPrizeStructurePreset(name string) (models.PrizeStructureConfig, bool) {
	preset, exists := PrizeStructurePresets[name]
	return preset, exists
}

// GetDefaultStructure returns the default tournament structure
func GetDefaultStructure() models.TournamentStructure {
	return StandardStructure
}

// GetDefaultPrizeStructure returns the default prize structure
func GetDefaultPrizeStructure() models.PrizeStructureConfig {
	return Top3Payout
}

// ValidateStructure validates a tournament structure
func ValidateStructure(structure models.TournamentStructure) error {
	if len(structure.BlindLevels) == 0 {
		return ErrEmptyBlindStructure
	}

	for i, level := range structure.BlindLevels {
		if level.SmallBlind <= 0 || level.BigBlind <= 0 {
			return ErrInvalidBlindAmounts
		}
		if level.BigBlind <= level.SmallBlind {
			return ErrBigBlindTooSmall
		}
		if level.Duration <= 0 {
			return ErrInvalidLevelDuration
		}
		if level.Ante < 0 {
			return ErrNegativeAnte
		}
		if i > 0 && level.BigBlind <= structure.BlindLevels[i-1].BigBlind {
			return ErrBlindsNotIncreasing
		}
	}

	return nil
}

// ValidatePrizeStructure validates a prize structure
func ValidatePrizeStructure(structure models.PrizeStructureConfig) error {
	if len(structure.Positions) == 0 {
		return ErrEmptyPrizeStructure
	}

	totalPercentage := 0.0
	for i, pos := range structure.Positions {
		if pos.Position != i+1 {
			return ErrInvalidPrizePositions
		}
		if pos.Percentage <= 0 || pos.Percentage > 100 {
			return ErrInvalidPrizePercentage
		}
		totalPercentage += pos.Percentage
	}

	// Allow some tolerance for floating point arithmetic (99.9% - 100.1%)
	if totalPercentage < 99.9 || totalPercentage > 100.1 {
		return ErrPrizePercentageMismatch
	}

	return nil
}

// CalculatePrizeAmounts calculates actual prize amounts based on prize pool
func CalculatePrizeAmounts(prizePool int, structure models.PrizeStructureConfig) map[int]int {
	prizes := make(map[int]int)
	remainingPool := prizePool
	totalAllocated := 0

	for i, pos := range structure.Positions {
		if i < len(structure.Positions)-1 {
			// Calculate prize for this position
			amount := int(float64(prizePool) * pos.Percentage / 100.0)
			prizes[pos.Position] = amount
			totalAllocated += amount
		} else {
			// Give remaining pool to last position to handle rounding
			prizes[pos.Position] = remainingPool - totalAllocated
		}
	}

	return prizes
}
