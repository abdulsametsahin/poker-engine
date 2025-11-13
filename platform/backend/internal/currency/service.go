package currency

import (
	"context"
	"fmt"

	"poker-platform/backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Service handles all currency operations
type Service struct {
	db *gorm.DB
}

// NewService creates a new currency service
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// GetBalance retrieves the current chip balance for a user
func (s *Service) GetBalance(ctx context.Context, userID string) (int, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Select("chips").First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, ErrUserNotFound
		}
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}
	return user.Chips, nil
}

// HasSufficientBalance checks if a user has enough chips
func (s *Service) HasSufficientBalance(ctx context.Context, userID string, amount int) (bool, error) {
	balance, err := s.GetBalance(ctx, userID)
	if err != nil {
		return false, err
	}
	return balance >= amount, nil
}

// ValidateAmount checks if a transaction amount is valid
func (s *Service) ValidateAmount(amount int) error {
	if amount < 0 {
		return ErrNegativeAmount
	}
	if amount < MinimumTransaction {
		return ErrInvalidAmount
	}
	if amount > MaximumTransaction {
		return ErrExceedsMaximum
	}
	return nil
}

// deductChipsInTx removes chips from a user's balance within an existing transaction
// Internal function - use DeductChips for standalone operations
func (s *Service) deductChipsInTx(ctx context.Context, tx *gorm.DB, userID string, amount int, txType TransactionType, refID string, description string) error {
	// Get current balance with row lock
	var user models.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to lock user record: %w", err)
	}

	// Check sufficient balance
	if user.Chips < amount {
		return ErrInsufficientChips
	}

	balanceBefore := user.Chips
	balanceAfter := balanceBefore - amount

	// Update balance
	if err := tx.Model(&user).Update("chips", balanceAfter).Error; err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Create audit record
	transaction := Transaction{
		ID:              uuid.New().String(),
		UserID:          userID,
		Amount:          -amount, // Negative for deduction
		BalanceBefore:   balanceBefore,
		BalanceAfter:    balanceAfter,
		TransactionType: txType,
		ReferenceID:     &refID,
		Description:     description,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		return fmt.Errorf("failed to create transaction record: %w", err)
	}

	return nil
}

// DeductChips removes chips from a user's balance with validation and audit trail
func (s *Service) DeductChips(ctx context.Context, userID string, amount int, txType TransactionType, refID string, description string) error {
	if err := s.ValidateAmount(amount); err != nil {
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.deductChipsInTx(ctx, tx, userID, amount, txType, refID, description)
	})
}

// addChipsInTx adds chips to a user's balance within an existing transaction
// Internal function - use AddChips for standalone operations
func (s *Service) addChipsInTx(ctx context.Context, tx *gorm.DB, userID string, amount int, txType TransactionType, refID string, description string) error {
	// Get current balance with row lock
	var user models.User
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to lock user record: %w", err)
	}

	balanceBefore := user.Chips
	balanceAfter := balanceBefore + amount

	// Update balance
	if err := tx.Model(&user).Update("chips", balanceAfter).Error; err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	// Create audit record
	transaction := Transaction{
		ID:              uuid.New().String(),
		UserID:          userID,
		Amount:          amount, // Positive for addition
		BalanceBefore:   balanceBefore,
		BalanceAfter:    balanceAfter,
		TransactionType: txType,
		ReferenceID:     &refID,
		Description:     description,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		return fmt.Errorf("failed to create transaction record: %w", err)
	}

	return nil
}

// AddChips adds chips to a user's balance with audit trail
func (s *Service) AddChips(ctx context.Context, userID string, amount int, txType TransactionType, refID string, description string) error {
	if err := s.ValidateAmount(amount); err != nil {
		return err
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.addChipsInTx(ctx, tx, userID, amount, txType, refID, description)
	})
}

// TransferChips transfers chips from one user to another atomically
// CRITICAL: Uses a single transaction to ensure atomicity - if either operation fails,
// both are rolled back, preventing money loss or duplication
func (s *Service) TransferChips(ctx context.Context, fromUserID, toUserID string, amount int, txType TransactionType, refID string, description string) error {
	if err := s.ValidateAmount(amount); err != nil {
		return err
	}

	// Single atomic transaction for both operations
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Deduct from sender (uses same transaction tx)
		if err := s.deductChipsInTx(ctx, tx, fromUserID, amount, txType, refID, description); err != nil {
			return fmt.Errorf("failed to deduct from sender: %w", err)
		}

		// Add to receiver (uses same transaction tx)
		if err := s.addChipsInTx(ctx, tx, toUserID, amount, txType, refID, description); err != nil {
			return fmt.Errorf("failed to add to receiver: %w", err)
		}

		return nil
	})
}

// GetTransactionHistory retrieves transaction history for a user
func (s *Service) GetTransactionHistory(ctx context.Context, userID string, limit int) ([]Transaction, error) {
	var transactions []Transaction
	query := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&transactions).Error; err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}

	return transactions, nil
}

// CalculatePrizeAmount calculates prize amount using basis points (integer math)
// basisPoints: 5000 = 50.00%, 10000 = 100.00%
func CalculatePrizeAmount(prizePool int, basisPoints int) int {
	return (prizePool * basisPoints) / BasisPointsTotal
}

// DistributePrizesExact distributes prizes ensuring exact total equals prize pool
// Returns slice of prize amounts, giving any remainder to first place
func DistributePrizesExact(prizePool int, basisPointsSlice []int) []int {
	if len(basisPointsSlice) == 0 {
		return []int{}
	}

	prizes := make([]int, len(basisPointsSlice))
	totalDistributed := 0

	// Calculate each prize
	for i, basisPoints := range basisPointsSlice {
		prizes[i] = CalculatePrizeAmount(prizePool, basisPoints)
		totalDistributed += prizes[i]
	}

	// Give any remainder to first place (due to integer division)
	remainder := prizePool - totalDistributed
	prizes[0] += remainder

	return prizes
}

// ValidateBasisPoints ensures basis points are valid and sum to 100%
func ValidateBasisPoints(basisPoints []int) error {
	if len(basisPoints) == 0 {
		return fmt.Errorf("no basis points provided")
	}

	total := 0
	for i, bp := range basisPoints {
		if bp <= 0 {
			return fmt.Errorf("basis points at position %d must be positive", i)
		}
		if bp > BasisPointsMax {
			return fmt.Errorf("basis points at position %d exceeds maximum", i)
		}
		total += bp
	}

	if total != BasisPointsTotal {
		return fmt.Errorf("basis points sum to %d, expected %d (100%%)", total, BasisPointsTotal)
	}

	return nil
}
