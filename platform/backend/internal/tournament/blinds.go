package tournament

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"poker-platform/backend/internal/models"

	"gorm.io/gorm"
)

// BlindManager manages blind level increases for tournaments
type BlindManager struct {
	db                   *gorm.DB
	stopChan             chan struct{}
	onBlindIncreaseCallback func(tournamentID string, newLevel models.BlindLevel) // Callback when blinds increase
}

// NewBlindManager creates a new blind manager
func NewBlindManager(db *gorm.DB) *BlindManager {
	return &BlindManager{
		db:                      db,
		stopChan:                make(chan struct{}),
		onBlindIncreaseCallback: nil,
	}
}

// SetOnBlindIncreaseCallback sets the callback function to be called when blinds increase
func (bm *BlindManager) SetOnBlindIncreaseCallback(callback func(tournamentID string, newLevel models.BlindLevel)) {
	bm.onBlindIncreaseCallback = callback
}

// Start begins monitoring tournaments for blind level increases
func (bm *BlindManager) Start() {
	log.Println("Blind manager service started")
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bm.checkBlinds()
		case <-bm.stopChan:
			log.Println("Blind manager service stopped")
			return
		}
	}
}

// Stop stops the blind manager service
func (bm *BlindManager) Stop() {
	close(bm.stopChan)
}

// checkBlinds checks all active tournaments for blind increases
func (bm *BlindManager) checkBlinds() {
	var tournaments []models.Tournament
	if err := bm.db.Where("status = ?", "in_progress").Find(&tournaments).Error; err != nil {
		log.Printf("Error fetching active tournaments: %v", err)
		return
	}

	now := time.Now()
	for _, tournament := range tournaments {
		if bm.shouldIncreaseBlinds(tournament, now) {
			if err := bm.IncreaseBlinds(tournament.ID); err != nil {
				log.Printf("Error increasing blinds for tournament %s: %v", tournament.ID, err)
			} else {
				log.Printf("Tournament %s: Blinds increased to level %d", tournament.ID, tournament.CurrentLevel+1)
			}
		}
	}
}

// shouldIncreaseBlinds checks if blinds should be increased for a tournament
func (bm *BlindManager) shouldIncreaseBlinds(tournament models.Tournament, now time.Time) bool {
	// Tournament must be in progress
	if tournament.Status != "in_progress" {
		return false
	}

	// Must have a level start time
	if tournament.LevelStartedAt == nil {
		return false
	}

	// Parse tournament structure
	var structure models.TournamentStructure
	if err := json.Unmarshal([]byte(tournament.Structure), &structure); err != nil {
		log.Printf("Error parsing tournament structure: %v", err)
		return false
	}

	// Check if we have more levels
	if tournament.CurrentLevel >= len(structure.BlindLevels) {
		// We're at the final level, no more increases
		return false
	}

	// Get current level configuration
	currentLevelIndex := tournament.CurrentLevel - 1 // CurrentLevel is 1-indexed
	if currentLevelIndex < 0 || currentLevelIndex >= len(structure.BlindLevels) {
		return false
	}

	currentLevelConfig := structure.BlindLevels[currentLevelIndex]

	// Check if enough time has passed
	timeSinceLevelStart := now.Sub(*tournament.LevelStartedAt)
	if timeSinceLevelStart.Seconds() >= float64(currentLevelConfig.Duration) {
		return true
	}

	return false
}

// IncreaseBlinds increases the blind level for a tournament
func (bm *BlindManager) IncreaseBlinds(tournamentID string) error {
	// Start transaction
	tx := bm.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get tournament with lock
	var tournament models.Tournament
	if err := tx.Clauses().Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Validate status
	if tournament.Status != "in_progress" {
		tx.Rollback()
		return fmt.Errorf("tournament is not in progress")
	}

	// Parse tournament structure
	var structure models.TournamentStructure
	if err := json.Unmarshal([]byte(tournament.Structure), &structure); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to parse tournament structure: %w", err)
	}

	// Check if we can increase
	if tournament.CurrentLevel >= len(structure.BlindLevels) {
		tx.Rollback()
		return ErrNoMoreBlindLevels
	}

	// Get next level
	newLevel := tournament.CurrentLevel + 1
	newLevelIndex := newLevel - 1
	newLevelConfig := structure.BlindLevels[newLevelIndex]

	// Update tournament
	now := time.Now()
	if err := tx.Model(&tournament).Updates(map[string]interface{}{
		"current_level":     newLevel,
		"level_started_at":  now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Get all tables for this tournament
	var tables []models.Table
	if err := tx.Where("tournament_id = ? AND status != ?", tournamentID, "completed").Find(&tables).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update all tournament tables
	for _, table := range tables {
		if err := tx.Model(&table).Updates(map[string]interface{}{
			"small_blind": newLevelConfig.SmallBlind,
			"big_blind":   newLevelConfig.BigBlind,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	log.Printf("Tournament %s: Increased to level %d (SB: %d, BB: %d, Ante: %d)",
		tournamentID, newLevel, newLevelConfig.SmallBlind, newLevelConfig.BigBlind, newLevelConfig.Ante)

	// Call the callback if set
	if bm.onBlindIncreaseCallback != nil {
		bm.onBlindIncreaseCallback(tournamentID, newLevelConfig)
	}

	return nil
}

// GetCurrentBlindLevel returns the current blind level configuration for a tournament
func (bm *BlindManager) GetCurrentBlindLevel(tournamentID string) (*models.BlindLevel, error) {
	var tournament models.Tournament
	if err := bm.db.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		return nil, err
	}

	// Parse tournament structure
	var structure models.TournamentStructure
	if err := json.Unmarshal([]byte(tournament.Structure), &structure); err != nil {
		return nil, fmt.Errorf("failed to parse tournament structure: %w", err)
	}

	// Get current level
	levelIndex := tournament.CurrentLevel - 1
	if levelIndex < 0 || levelIndex >= len(structure.BlindLevels) {
		return nil, ErrInvalidBlindLevel
	}

	return &structure.BlindLevels[levelIndex], nil
}

// GetNextBlindLevel returns the next blind level configuration for a tournament
func (bm *BlindManager) GetNextBlindLevel(tournamentID string) (*models.BlindLevel, error) {
	var tournament models.Tournament
	if err := bm.db.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		return nil, err
	}

	// Parse tournament structure
	var structure models.TournamentStructure
	if err := json.Unmarshal([]byte(tournament.Structure), &structure); err != nil {
		return nil, fmt.Errorf("failed to parse tournament structure: %w", err)
	}

	// Get next level
	nextLevelIndex := tournament.CurrentLevel // CurrentLevel is 1-indexed, so CurrentLevel is the next index
	if nextLevelIndex >= len(structure.BlindLevels) {
		return nil, ErrNoMoreBlindLevels
	}

	return &structure.BlindLevels[nextLevelIndex], nil
}

// GetTimeUntilNextLevel returns the time remaining until the next blind level
func (bm *BlindManager) GetTimeUntilNextLevel(tournamentID string) (time.Duration, error) {
	var tournament models.Tournament
	if err := bm.db.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		return 0, err
	}

	if tournament.Status != "in_progress" {
		return 0, fmt.Errorf("tournament is not in progress")
	}

	if tournament.LevelStartedAt == nil {
		return 0, fmt.Errorf("level start time not set")
	}

	// Parse tournament structure
	var structure models.TournamentStructure
	if err := json.Unmarshal([]byte(tournament.Structure), &structure); err != nil {
		return 0, fmt.Errorf("failed to parse tournament structure: %w", err)
	}

	// Get current level
	levelIndex := tournament.CurrentLevel - 1
	if levelIndex < 0 || levelIndex >= len(structure.BlindLevels) {
		return 0, ErrInvalidBlindLevel
	}

	currentLevel := structure.BlindLevels[levelIndex]

	// Calculate time until next level
	levelDuration := time.Duration(currentLevel.Duration) * time.Second
	elapsed := time.Since(*tournament.LevelStartedAt)
	remaining := levelDuration - elapsed

	if remaining < 0 {
		return 0, nil // Overdue for increase
	}

	return remaining, nil
}
