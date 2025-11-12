package tournament

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"poker-platform/backend/internal/currency"
	"poker-platform/backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service handles tournament operations
type Service struct {
	db              *gorm.DB
	currencyService *currency.Service
}

// NewService creates a new tournament service
func NewService(db *gorm.DB, currencyService *currency.Service) *Service {
	return &Service{
		db:              db,
		currencyService: currencyService,
	}
}

// CreateTournament creates a new tournament
func (s *Service) CreateTournament(req models.CreateTournamentRequest, creatorID string) (*models.Tournament, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Get or validate structure
	var structure models.TournamentStructure
	if req.StructurePreset != "" {
		preset, exists := GetStructurePreset(req.StructurePreset)
		if !exists {
			return nil, ErrStructureNotFound
		}
		structure = preset
	} else if req.CustomStructure != nil {
		if err := ValidateStructure(*req.CustomStructure); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidStructure, err)
		}
		structure = *req.CustomStructure
	} else {
		structure = GetDefaultStructure()
	}

	// Get or validate prize structure
	var prizeStructure models.PrizeStructureConfig
	if req.PrizeStructurePreset != "" {
		preset, exists := GetPrizeStructurePreset(req.PrizeStructurePreset)
		if !exists {
			return nil, ErrPrizeStructureNotFound
		}
		prizeStructure = preset
	} else if req.CustomPrizeStructure != nil {
		if err := ValidatePrizeStructure(*req.CustomPrizeStructure); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidPrizeStructure, err)
		}
		prizeStructure = *req.CustomPrizeStructure
	} else {
		prizeStructure = GetDefaultPrizeStructure()
	}

	// Generate unique tournament code
	var tournamentCode string
	var err error
	for i := 0; i < 10; i++ { // Try up to 10 times
		tournamentCode, err = GenerateTournamentCode()
		if err != nil {
			return nil, err
		}

		// Check if code already exists
		var existing models.Tournament
		result := s.db.Where("tournament_code = ?", tournamentCode).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			break // Code is unique
		}
		if i == 9 {
			return nil, ErrTournamentCodeExists
		}
	}

	// Serialize structures to JSON
	structureJSON, err := json.Marshal(structure)
	if err != nil {
		return nil, err
	}

	prizeStructureJSON, err := json.Marshal(prizeStructure)
	if err != nil {
		return nil, err
	}

	// Set default auto start delay if not provided
	autoStartDelay := req.AutoStartDelay
	if autoStartDelay == 0 {
		autoStartDelay = 300 // 5 minutes default
	}

	// Create tournament
	tournament := &models.Tournament{
		ID:                   uuid.New().String(),
		TournamentCode:       tournamentCode,
		Name:                 req.Name,
		CreatorID:            &creatorID,
		Status:               "registering",
		BuyIn:                req.BuyIn,
		StartingChips:        req.StartingChips,
		MaxPlayers:           req.MaxPlayers,
		MinPlayers:           req.MinPlayers,
		CurrentPlayers:       0,
		PrizePool:            0,
		Structure:            string(structureJSON),
		PrizeStructure:       string(prizeStructureJSON),
		StartTime:            req.StartTime,
		RegistrationClosesAt: nil, // Can be set later
		AutoStartDelay:       autoStartDelay,
		CurrentLevel:         1,
		LevelStartedAt:       nil,
		CreatedAt:            time.Now(),
	}

	if err := s.db.Create(tournament).Error; err != nil {
		return nil, err
	}

	return tournament, nil
}

// RegisterPlayer registers a player for a tournament
func (s *Service) RegisterPlayer(tournamentID, userID string) error {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get tournament with lock
	var tournament models.Tournament
	if err := tx.Clauses().Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return ErrTournamentNotFound
		}
		return err
	}

	// Validate tournament status
	if tournament.Status != "registering" {
		tx.Rollback()
		return ErrTournamentNotRegistering
	}

	// Check if tournament is full
	if tournament.CurrentPlayers >= tournament.MaxPlayers {
		tx.Rollback()
		return ErrTournamentFull
	}

	// Check if player is already registered
	var existing models.TournamentPlayer
	result := tx.Where("tournament_id = ? AND user_id = ?", tournamentID, userID).First(&existing)
	if result.Error == nil {
		tx.Rollback()
		return ErrAlreadyRegistered
	}

	// Deduct buy-in from user using currency service (with validation and audit trail)
	ctx := context.Background()
	description := fmt.Sprintf("Buy-in for tournament: %s", tournament.Name)
	if err := s.currencyService.DeductChips(
		ctx,
		userID,
		tournament.BuyIn,
		currency.TxTypeTournamentBuyIn,
		tournamentID,
		description,
	); err != nil {
		tx.Rollback()
		if err == currency.ErrInsufficientChips {
			return ErrInsufficientChips
		}
		return fmt.Errorf("failed to deduct buy-in: %w", err)
	}

	// Create tournament player entry
	tournamentPlayer := &models.TournamentPlayer{
		TournamentID: tournamentID,
		UserID:       userID,
		Position:     nil,
		Chips:        &tournament.StartingChips,
		PrizeAmount:  0,
		RegisteredAt: time.Now(),
	}

	if err := tx.Create(tournamentPlayer).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update tournament player count and prize pool
	newPlayerCount := tournament.CurrentPlayers + 1
	newPrizePool := tournament.PrizePool + tournament.BuyIn

	updates := map[string]interface{}{
		"current_players": newPlayerCount,
		"prize_pool":      newPrizePool,
	}

	// If we just reached min_players and don't have a scheduled start time,
	// set registration_completed_at for auto-start countdown
	if newPlayerCount == tournament.MinPlayers && tournament.StartTime == nil && tournament.RegistrationCompletedAt == nil {
		now := time.Now()
		updates["registration_completed_at"] = now
	}

	if err := tx.Model(&tournament).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// UnregisterPlayer removes a player from a tournament
func (s *Service) UnregisterPlayer(tournamentID, userID string) error {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get tournament
	var tournament models.Tournament
	if err := tx.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return ErrTournamentNotFound
		}
		return err
	}

	// Check if tournament has started
	if tournament.Status != "registering" {
		tx.Rollback()
		return ErrCannotUnregister
	}

	// Get tournament player
	var tournamentPlayer models.TournamentPlayer
	if err := tx.Where("tournament_id = ? AND user_id = ?", tournamentID, userID).First(&tournamentPlayer).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return ErrNotRegistered
		}
		return err
	}

	// Refund buy-in to user using currency service (with audit trail)
	ctx := context.Background()
	description := fmt.Sprintf("Refund for tournament: %s", tournament.Name)
	if err := s.currencyService.AddChips(
		ctx,
		userID,
		tournament.BuyIn,
		currency.TxTypeTournamentRefund,
		tournamentID,
		description,
	); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to refund buy-in: %w", err)
	}

	// Delete tournament player entry
	if err := tx.Delete(&tournamentPlayer).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update tournament player count and prize pool
	newPlayerCount := tournament.CurrentPlayers - 1
	newPrizePool := tournament.PrizePool - tournament.BuyIn

	updates := map[string]interface{}{
		"current_players": newPlayerCount,
		"prize_pool":      newPrizePool,
	}

	// If we drop below min_players, clear the registration_completed_at timestamp
	if newPlayerCount < tournament.MinPlayers && tournament.RegistrationCompletedAt != nil {
		updates["registration_completed_at"] = nil
	}

	if err := tx.Model(&tournament).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// CancelTournament cancels a tournament and refunds all players
func (s *Service) CancelTournament(tournamentID, userID string) error {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get tournament
	var tournament models.Tournament
	if err := tx.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return ErrTournamentNotFound
		}
		return err
	}

	// Check if user is creator
	if tournament.CreatorID == nil || *tournament.CreatorID != userID {
		tx.Rollback()
		return ErrNotTournamentCreator
	}

	// Check if tournament has already started
	if tournament.Status != "registering" {
		tx.Rollback()
		return ErrCannotCancelStarted
	}

	// Get all registered players
	var players []models.TournamentPlayer
	if err := tx.Where("tournament_id = ?", tournamentID).Find(&players).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Refund all players using currency service (with audit trail)
	ctx := context.Background()
	for _, player := range players {
		description := fmt.Sprintf("Refund from cancelled tournament: %s", tournament.Name)
		if err := s.currencyService.AddChips(
			ctx,
			player.UserID,
			tournament.BuyIn,
			currency.TxTypeTournamentRefund,
			tournamentID,
			description,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to refund player %s: %w", player.UserID, err)
		}
	}

	// Update tournament status
	if err := tx.Model(&tournament).Updates(map[string]interface{}{
		"status":          "cancelled",
		"current_players": 0,
		"prize_pool":      0,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

// GetTournament retrieves a tournament by ID
func (s *Service) GetTournament(tournamentID string) (*models.Tournament, error) {
	var tournament models.Tournament
	if err := s.db.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTournamentNotFound
		}
		return nil, err
	}
	return &tournament, nil
}

// GetTournamentByCode retrieves a tournament by code
func (s *Service) GetTournamentByCode(code string) (*models.Tournament, error) {
	normalizedCode := NormalizeTournamentCode(code)
	var tournament models.Tournament
	if err := s.db.Where("tournament_code = ?", normalizedCode).First(&tournament).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTournamentNotFound
		}
		return nil, err
	}
	return &tournament, nil
}

// ListTournaments retrieves tournaments with optional filters
func (s *Service) ListTournaments(status string, limit, offset int) ([]models.Tournament, error) {
	query := s.db.Model(&models.Tournament{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var tournaments []models.Tournament
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tournaments).Error; err != nil {
		return nil, err
	}

	return tournaments, nil
}

// GetTournamentPlayers retrieves all players registered for a tournament
func (s *Service) GetTournamentPlayers(tournamentID string) ([]models.TournamentPlayer, error) {
	var players []models.TournamentPlayer
	if err := s.db.Where("tournament_id = ?", tournamentID).Order("registered_at ASC").Find(&players).Error; err != nil {
		return nil, err
	}
	return players, nil
}

// IsPlayerRegistered checks if a player is registered for a tournament
func (s *Service) IsPlayerRegistered(tournamentID, userID string) (bool, error) {
	var count int64
	if err := s.db.Model(&models.TournamentPlayer{}).Where("tournament_id = ? AND user_id = ?", tournamentID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// validateCreateRequest validates tournament creation request
func (s *Service) validateCreateRequest(req models.CreateTournamentRequest) error {
	if req.Name == "" {
		return ErrInvalidTournamentName
	}
	if req.BuyIn < 0 {
		return ErrInvalidBuyIn
	}
	if req.StartingChips < 100 {
		return ErrInvalidStartingChips
	}
	if req.MaxPlayers < 2 || req.MaxPlayers > 1000 {
		return ErrInvalidMaxPlayers
	}
	if req.MinPlayers < 2 {
		return ErrInvalidMinPlayers
	}
	if req.MinPlayers > req.MaxPlayers {
		return ErrMinPlayersGreaterThanMax
	}
	if req.AutoStartDelay < 0 {
		return ErrInvalidAutoStartDelay
	}
	if req.StartTime != nil && req.StartTime.Before(time.Now()) {
		return ErrInvalidStartTime
	}

	return nil
}

// PauseTournament pauses a tournament and all its tables
func (s *Service) PauseTournament(tournamentID string, pausedBy string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get tournament
	var tournament models.Tournament
	if err := tx.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Validate status
	if tournament.Status != "in_progress" {
		tx.Rollback()
		return fmt.Errorf("can only pause in-progress tournament, current: %s", tournament.Status)
	}

	// Validate only creator can pause
	if tournament.CreatorID == nil || *tournament.CreatorID != pausedBy {
		tx.Rollback()
		return fmt.Errorf("only tournament creator can pause")
	}

	// Update tournament status
	now := time.Now()
	if err := tx.Model(&tournament).Updates(map[string]interface{}{
		"status":    "paused",
		"paused_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Update all tournament tables in DB
	if err := tx.Model(&models.Table{}).
		Where("tournament_id = ? AND status = ?", tournamentID, "playing").
		Update("status", "paused").Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// ResumeTournament resumes a paused tournament
func (s *Service) ResumeTournament(tournamentID string, resumedBy string) error {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var tournament models.Tournament
	if err := tx.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		tx.Rollback()
		return err
	}

	if tournament.Status != "paused" {
		tx.Rollback()
		return fmt.Errorf("tournament not paused, current: %s", tournament.Status)
	}

	if tournament.CreatorID == nil || *tournament.CreatorID != resumedBy {
		tx.Rollback()
		return fmt.Errorf("only tournament creator can resume")
	}

	// Calculate pause duration
	pauseDuration := 0
	if tournament.PausedAt != nil {
		pauseDuration = int(time.Since(*tournament.PausedAt).Seconds())
	}

	// Update tournament
	now := time.Now()
	updates := map[string]interface{}{
		"status":                 "in_progress",
		"resumed_at":            now,
		"total_paused_duration": tournament.TotalPausedDuration + pauseDuration,
	}

	// Adjust level_started_at to account for pause
	if tournament.LevelStartedAt != nil {
		adjustedLevelStart := tournament.LevelStartedAt.Add(time.Duration(pauseDuration) * time.Second)
		updates["level_started_at"] = adjustedLevelStart
	}

	if err := tx.Model(&tournament).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Resume all tournament tables
	if err := tx.Model(&models.Table{}).
		Where("tournament_id = ? AND status = ?", tournamentID, "paused").
		Update("status", "playing").Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
