package tournament

import "errors"

// Tournament errors
var (
	// Structure validation errors
	ErrEmptyBlindStructure     = errors.New("blind structure cannot be empty")
	ErrInvalidBlindAmounts     = errors.New("blind amounts must be positive")
	ErrBigBlindTooSmall        = errors.New("big blind must be greater than small blind")
	ErrInvalidLevelDuration    = errors.New("level duration must be positive")
	ErrNegativeAnte            = errors.New("ante cannot be negative")
	ErrBlindsNotIncreasing     = errors.New("blinds must increase with each level")

	// Prize structure validation errors
	ErrEmptyPrizeStructure      = errors.New("prize structure cannot be empty")
	ErrInvalidPrizePositions    = errors.New("prize positions must be sequential starting from 1")
	ErrInvalidPrizePercentage   = errors.New("prize percentage must be between 0 and 100")
	ErrPrizePercentageMismatch  = errors.New("prize percentages must sum to 100")

	// Tournament creation errors
	ErrInvalidTournamentName    = errors.New("tournament name is required")
	ErrInvalidBuyIn             = errors.New("buy-in must be non-negative")
	ErrInvalidStartingChips     = errors.New("starting chips must be at least 100")
	ErrInvalidMaxPlayers        = errors.New("max players must be between 2 and 1000")
	ErrInvalidMinPlayers        = errors.New("min players must be at least 2")
	ErrMinPlayersGreaterThanMax = errors.New("min players cannot exceed max players")
	ErrInvalidAutoStartDelay    = errors.New("auto start delay must be non-negative")
	ErrInvalidStartTime         = errors.New("start time cannot be in the past")
	ErrStructureNotFound        = errors.New("tournament structure preset not found")
	ErrPrizeStructureNotFound   = errors.New("prize structure preset not found")
	ErrInvalidStructure         = errors.New("invalid tournament structure")
	ErrInvalidPrizeStructure    = errors.New("invalid prize structure")

	// Tournament registration errors
	ErrTournamentNotFound         = errors.New("tournament not found")
	ErrTournamentNotRegistering   = errors.New("tournament is not accepting registrations")
	ErrTournamentFull             = errors.New("tournament is full")
	ErrAlreadyRegistered          = errors.New("already registered for this tournament")
	ErrInsufficientChips          = errors.New("insufficient chips for buy-in")
	ErrCannotUnregister           = errors.New("cannot unregister after tournament has started")
	ErrNotRegistered              = errors.New("not registered for this tournament")

	// Tournament start errors
	ErrNotEnoughPlayers           = errors.New("not enough players to start tournament")
	ErrTournamentAlreadyStarted   = errors.New("tournament has already started")
	ErrTournamentCancelled        = errors.New("tournament has been cancelled")
	ErrTournamentCompleted        = errors.New("tournament has already completed")

	// Tournament operation errors
	ErrNotTournamentCreator       = errors.New("only tournament creator can perform this action")
	ErrCannotCancelStarted        = errors.New("cannot cancel tournament that has already started")
	ErrInvalidBlindLevel          = errors.New("invalid blind level")
	ErrNoMoreBlindLevels          = errors.New("no more blind levels in structure")

	// Tournament code errors
	ErrInvalidTournamentCode      = errors.New("invalid tournament code")
	ErrTournamentCodeExists       = errors.New("tournament code already exists")
)
