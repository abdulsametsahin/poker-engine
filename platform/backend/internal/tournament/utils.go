package tournament

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	// TournamentCodeLength is the length of generated tournament codes
	TournamentCodeLength = 8
	// TournamentCodeChars are the characters used in tournament codes (excluding ambiguous chars)
	TournamentCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // Removed I, O, 0, 1 to avoid confusion
)

// GenerateTournamentCode generates a unique random tournament code
func GenerateTournamentCode() (string, error) {
	code := make([]byte, TournamentCodeLength)
	charsLen := big.NewInt(int64(len(TournamentCodeChars)))

	for i := 0; i < TournamentCodeLength; i++ {
		num, err := rand.Int(rand.Reader, charsLen)
		if err != nil {
			return "", err
		}
		code[i] = TournamentCodeChars[num.Int64()]
	}

	return string(code), nil
}

// ValidateTournamentCode checks if a tournament code is valid format
func ValidateTournamentCode(code string) bool {
	if len(code) != TournamentCodeLength {
		return false
	}

	code = strings.ToUpper(code)
	for _, char := range code {
		if !strings.ContainsRune(TournamentCodeChars, char) {
			return false
		}
	}

	return true
}

// NormalizeTournamentCode normalizes a tournament code to uppercase
func NormalizeTournamentCode(code string) string {
	return strings.ToUpper(code)
}

// CalculatePrizePool calculates total prize pool based on buy-in and players
func CalculatePrizePool(buyIn int, playerCount int) int {
	return buyIn * playerCount
}

// CalculateTablesNeeded calculates how many tables are needed for player count
func CalculateTablesNeeded(playerCount int, maxPlayersPerTable int) int {
	if playerCount == 0 {
		return 0
	}
	tables := playerCount / maxPlayersPerTable
	if playerCount%maxPlayersPerTable != 0 {
		tables++
	}
	return tables
}

// DistributePlayersToTables distributes players across tables as evenly as possible
func DistributePlayersToTables(playerCount int, maxPlayersPerTable int) []int {
	if playerCount == 0 {
		return []int{}
	}

	tablesNeeded := CalculateTablesNeeded(playerCount, maxPlayersPerTable)
	distribution := make([]int, tablesNeeded)

	// Distribute players evenly
	basePlayersPerTable := playerCount / tablesNeeded
	remainder := playerCount % tablesNeeded

	for i := 0; i < tablesNeeded; i++ {
		distribution[i] = basePlayersPerTable
		if i < remainder {
			distribution[i]++
		}
	}

	return distribution
}

// ShouldConsolidateTables determines if tables should be consolidated
func ShouldConsolidateTables(tableCounts []int, maxPlayersPerTable int) bool {
	if len(tableCounts) <= 1 {
		return false
	}

	totalPlayers := 0
	for _, count := range tableCounts {
		totalPlayers += count
	}

	// If all players can fit on fewer tables, consolidate
	minTablesNeeded := CalculateTablesNeeded(totalPlayers, maxPlayersPerTable)
	return minTablesNeeded < len(tableCounts)
}

// CalculateTableBalance checks if tables need balancing
// Returns true if any two tables have a difference > 2 players
func CalculateTableBalance(tableCounts []int) bool {
	if len(tableCounts) <= 1 {
		return true // Single table is always balanced
	}

	minPlayers := tableCounts[0]
	maxPlayers := tableCounts[0]

	for _, count := range tableCounts {
		if count < minPlayers {
			minPlayers = count
		}
		if count > maxPlayers {
			maxPlayers = count
		}
	}

	// Tables are unbalanced if difference is > 2
	return (maxPlayers - minPlayers) <= 2
}

// FindTableToMove finds which table to move a player from for balancing
// Returns table index with most players
func FindTableToMove(tableCounts []int) int {
	maxPlayers := 0
	tableIndex := 0

	for i, count := range tableCounts {
		if count > maxPlayers {
			maxPlayers = count
			tableIndex = i
		}
	}

	return tableIndex
}

// FindTableToReceive finds which table should receive a player for balancing
// Returns table index with fewest players
func FindTableToReceive(tableCounts []int) int {
	minPlayers := tableCounts[0]
	tableIndex := 0

	for i, count := range tableCounts {
		if count < minPlayers {
			minPlayers = count
			tableIndex = i
		}
	}

	return tableIndex
}

// CalculateAverageStack calculates the average chip stack in a tournament
func CalculateAverageStack(totalChips int, remainingPlayers int) int {
	if remainingPlayers == 0 {
		return 0
	}
	return totalChips / remainingPlayers
}

// IsFinalTable checks if only one table remains
func IsFinalTable(tableCount int) bool {
	return tableCount == 1
}

// IsOnTheBubble checks if we're one elimination away from prizes
// playerCount is current players remaining
// prizePositions is number of positions that pay
func IsOnTheBubble(playerCount int, prizePositions int) bool {
	return playerCount == prizePositions+1
}

// IsInTheMoney checks if a position receives a prize
func IsInTheMoney(position int, prizePositions int) bool {
	return position <= prizePositions
}
