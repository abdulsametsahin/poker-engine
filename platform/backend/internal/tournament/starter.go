package tournament

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"poker-platform/backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Starter manages tournament start conditions and initialization
type Starter struct {
	db              *gorm.DB
	service         *Service
	stopChan        chan struct{}
	onStartCallback func(tournamentID string) // Callback when tournament starts
}

// NewStarter creates a new tournament starter
func NewStarter(db *gorm.DB, service *Service) *Starter {
	return &Starter{
		db:              db,
		service:         service,
		stopChan:        make(chan struct{}),
		onStartCallback: nil,
	}
}

// SetOnStartCallback sets the callback function to be called when a tournament starts
func (s *Starter) SetOnStartCallback(callback func(tournamentID string)) {
	s.onStartCallback = callback
}

// Start begins monitoring tournaments for start conditions
func (s *Starter) Start() {
	log.Println("Tournament starter service started")
	ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkTournaments()
		case <-s.stopChan:
			log.Println("Tournament starter service stopped")
			return
		}
	}
}

// Stop stops the tournament starter service
func (s *Starter) Stop() {
	close(s.stopChan)
}

// checkTournaments checks all tournaments for start conditions
func (s *Starter) checkTournaments() {
	var tournaments []models.Tournament
	if err := s.db.Where("status = ?", "registering").Find(&tournaments).Error; err != nil {
		log.Printf("Error fetching tournaments: %v", err)
		return
	}

	now := time.Now()
	for _, tournament := range tournaments {
		if s.shouldStartTournament(tournament, now) {
			if err := s.StartTournament(tournament.ID); err != nil {
				log.Printf("Error starting tournament %s: %v", tournament.ID, err)
			} else {
				log.Printf("Tournament %s (%s) started successfully", tournament.ID, tournament.Name)
			}
		}
	}
}

// shouldStartTournament checks if a tournament should start
func (s *Starter) shouldStartTournament(tournament models.Tournament, now time.Time) bool {
	// Check if scheduled start time is reached
	if tournament.StartTime != nil && !tournament.StartTime.After(now) {
		if tournament.CurrentPlayers >= tournament.MinPlayers {
			return true
		}
		log.Printf("Tournament %s scheduled to start but only has %d/%d players",
			tournament.ID, tournament.CurrentPlayers, tournament.MinPlayers)
	}

	// Check if max players reached (immediate start)
	if tournament.CurrentPlayers >= tournament.MaxPlayers {
		return true
	}

	// Check if min players reached and auto-start delay expired
	if tournament.CurrentPlayers >= tournament.MinPlayers {
		// Find when min players was first reached
		// For now, we'll use a simpler approach: check if enough time has passed since creation
		// TODO: Track exact time when min_players was reached
		timeSinceCreation := now.Sub(tournament.CreatedAt)
		if timeSinceCreation.Seconds() >= float64(tournament.AutoStartDelay) {
			return true
		}
	}

	return false
}

// StartTournament starts a tournament
func (s *Starter) StartTournament(tournamentID string) error {
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
		return err
	}

	// Validate status
	if tournament.Status != "registering" {
		tx.Rollback()
		return ErrTournamentAlreadyStarted
	}

	// Validate player count
	if tournament.CurrentPlayers < tournament.MinPlayers {
		tx.Rollback()
		return ErrNotEnoughPlayers
	}

	// Update tournament status to 'starting'
	now := time.Now()
	if err := tx.Model(&tournament).Updates(map[string]interface{}{
		"status":           "starting",
		"started_at":       now,
		"level_started_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Get all registered players
	var players []models.TournamentPlayer
	if err := tx.Where("tournament_id = ?", tournamentID).Find(&players).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Assign players to tables
	tableAssignments, err := s.assignPlayersToTables(players, 8) // Max 8 players per table
	if err != nil {
		tx.Rollback()
		return err
	}

	// Parse tournament structure to get first blind level
	var structure models.TournamentStructure
	if err := json.Unmarshal([]byte(tournament.Structure), &structure); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to parse tournament structure: %w", err)
	}

	if len(structure.BlindLevels) == 0 {
		tx.Rollback()
		return ErrEmptyBlindStructure
	}

	firstLevel := structure.BlindLevels[0]

	// Create tables for each assignment
	for tableNum, assignment := range tableAssignments {
		tableName := fmt.Sprintf("%s - Table %d", tournament.Name, tableNum+1)
		tableNumber := tableNum + 1

		table := &models.Table{
			ID:           uuid.New().String(),
			TournamentID: &tournament.ID,
			TableNumber:  &tableNumber,
			Name:         tableName,
			GameType:     "tournament",
			Status:       "waiting",
			SmallBlind:   firstLevel.SmallBlind,
			BigBlind:     firstLevel.BigBlind,
			MaxPlayers:   8,
			MinBuyIn:     nil,
			MaxBuyIn:     nil,
			CreatedAt:    now,
		}

		if err := tx.Create(table).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Create table seats for assigned players
		for seatNum, playerID := range assignment {
			seat := &models.TableSeat{
				TableID:    table.ID,
				UserID:     playerID,
				SeatNumber: seatNum,
				Chips:      tournament.StartingChips,
				Status:     "active",
				JoinedAt:   now,
			}

			if err := tx.Create(seat).Error; err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// Update tournament status to 'in_progress'
	if err := tx.Model(&tournament).Update("status", "in_progress").Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	log.Printf("Tournament %s started with %d tables, %d players",
		tournamentID, len(tableAssignments), len(players))

	// Call the callback if set
	if s.onStartCallback != nil {
		s.onStartCallback(tournamentID)
	}

	return nil
}

// assignPlayersToTables assigns players to tables with randomized seating
// Returns a map of tableIndex -> []playerIDs (with seat positions as array indices)
func (s *Starter) assignPlayersToTables(players []models.TournamentPlayer, maxPlayersPerTable int) (map[int][]string, error) {
	if len(players) == 0 {
		return nil, fmt.Errorf("no players to assign")
	}

	// Shuffle players randomly
	shuffled := make([]models.TournamentPlayer, len(players))
	copy(shuffled, players)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Calculate table distribution
	distribution := DistributePlayersToTables(len(players), maxPlayersPerTable)

	assignments := make(map[int][]string)
	playerIndex := 0

	for tableIndex, playerCount := range distribution {
		tableAssignment := make([]string, playerCount)

		for seatNum := 0; seatNum < playerCount; seatNum++ {
			tableAssignment[seatNum] = shuffled[playerIndex].UserID
			playerIndex++
		}

		assignments[tableIndex] = tableAssignment
	}

	return assignments, nil
}

// ForceStartTournament manually starts a tournament (for testing/admin)
func (s *Starter) ForceStartTournament(tournamentID string) error {
	var tournament models.Tournament
	if err := s.db.Where("id = ?", tournamentID).First(&tournament).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrTournamentNotFound
		}
		return err
	}

	if tournament.Status != "registering" {
		return ErrTournamentAlreadyStarted
	}

	if tournament.CurrentPlayers < tournament.MinPlayers {
		return ErrNotEnoughPlayers
	}

	return s.StartTournament(tournamentID)
}
