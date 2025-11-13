package currency

import (
	"context"
	"testing"

	"gorm.io/gorm"
)

// TestDeductChipsWithTx_Success verifies transaction-aware deduction works
func TestDeductChipsWithTx_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	createTestUser(t, db, "user1", 1000)

	// Use explicit transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		return service.DeductChipsWithTx(ctx, tx, "user1", 200, TxTypeTournamentBuyIn, "test-ref", "Test deduction")
	})

	if err != nil {
		t.Fatalf("DeductChipsWithTx failed: %v", err)
	}

	balance := getBalance(t, db, "user1")
	if balance != 800 {
		t.Errorf("Expected balance 800, got %d", balance)
	}

	// Verify audit trail
	var txCount int64
	db.Model(&Transaction{}).Count(&txCount)
	if txCount != 1 {
		t.Errorf("Expected 1 transaction record, got %d", txCount)
	}
}

// TestDeductChipsWithTx_Rollback verifies transaction rollback
func TestDeductChipsWithTx_Rollback(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	createTestUser(t, db, "user1", 1000)

	// Use explicit transaction with forced rollback
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := service.DeductChipsWithTx(ctx, tx, "user1", 200, TxTypeTournamentBuyIn, "test-ref", "Test deduction"); err != nil {
			return err
		}
		// Force rollback by returning error
		return gorm.ErrInvalidTransaction
	})

	if err == nil {
		t.Fatal("Expected transaction to rollback, got nil error")
	}

	// Verify balance unchanged due to rollback
	balance := getBalance(t, db, "user1")
	if balance != 1000 {
		t.Errorf("Expected balance 1000 after rollback, got %d", balance)
	}

	// Verify no transaction records created
	var txCount int64
	db.Model(&Transaction{}).Count(&txCount)
	if txCount != 0 {
		t.Errorf("Expected 0 transaction records after rollback, got %d", txCount)
	}
}

// TestAddChipsWithTx_Success verifies transaction-aware addition works
func TestAddChipsWithTx_Success(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	createTestUser(t, db, "user1", 500)

	// Use explicit transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		return service.AddChipsWithTx(ctx, tx, "user1", 300, TxTypeTournamentPrize, "test-ref", "Test addition")
	})

	if err != nil {
		t.Fatalf("AddChipsWithTx failed: %v", err)
	}

	balance := getBalance(t, db, "user1")
	if balance != 800 {
		t.Errorf("Expected balance 800, got %d", balance)
	}

	// Verify audit trail
	var txCount int64
	db.Model(&Transaction{}).Count(&txCount)
	if txCount != 1 {
		t.Errorf("Expected 1 transaction record, got %d", txCount)
	}
}

// TestAddChipsWithTx_Rollback verifies transaction rollback
func TestAddChipsWithTx_Rollback(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	createTestUser(t, db, "user1", 500)

	// Use explicit transaction with forced rollback
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := service.AddChipsWithTx(ctx, tx, "user1", 300, TxTypeTournamentPrize, "test-ref", "Test addition"); err != nil {
			return err
		}
		// Force rollback by returning error
		return gorm.ErrInvalidTransaction
	})

	if err == nil {
		t.Fatal("Expected transaction to rollback, got nil error")
	}

	// Verify balance unchanged due to rollback
	balance := getBalance(t, db, "user1")
	if balance != 500 {
		t.Errorf("Expected balance 500 after rollback, got %d", balance)
	}

	// Verify no transaction records created
	var txCount int64
	db.Model(&Transaction{}).Count(&txCount)
	if txCount != 0 {
		t.Errorf("Expected 0 transaction records after rollback, got %d", txCount)
	}
}

// TestCoordinatedOperations_Atomic verifies multiple operations in one transaction
func TestCoordinatedOperations_Atomic(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	createTestUser(t, db, "user1", 1000)
	createTestUser(t, db, "user2", 500)

	// Perform coordinated operations in single transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		// Deduct from user1
		if err := service.DeductChipsWithTx(ctx, tx, "user1", 200, TxTypeCashGameBuyIn, "test-ref", "Buy in"); err != nil {
			return err
		}

		// Add to user2
		if err := service.AddChipsWithTx(ctx, tx, "user2", 200, TxTypeCashGameCashOut, "test-ref", "Cash out"); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Coordinated operations failed: %v", err)
	}

	// Verify both operations succeeded
	balance1 := getBalance(t, db, "user1")
	balance2 := getBalance(t, db, "user2")

	if balance1 != 800 {
		t.Errorf("Expected user1 balance 800, got %d", balance1)
	}

	if balance2 != 700 {
		t.Errorf("Expected user2 balance 700, got %d", balance2)
	}

	// Verify chips conserved
	totalChips := balance1 + balance2
	if totalChips != 1500 {
		t.Errorf("Chips not conserved! Expected 1500, got %d", totalChips)
	}

	// Verify both transaction records created
	var txCount int64
	db.Model(&Transaction{}).Count(&txCount)
	if txCount != 2 {
		t.Errorf("Expected 2 transaction records, got %d", txCount)
	}
}

// TestCoordinatedOperations_PartialFailure verifies rollback on partial failure
func TestCoordinatedOperations_PartialFailure(t *testing.T) {
	db := setupTestDB(t)
	service := NewService(db)
	ctx := context.Background()

	createTestUser(t, db, "user1", 1000)
	createTestUser(t, db, "user2", 500)

	initialBalance1 := getBalance(t, db, "user1")
	initialBalance2 := getBalance(t, db, "user2")

	// Perform coordinated operations with second operation failing
	err := db.Transaction(func(tx *gorm.DB) error {
		// First operation succeeds
		if err := service.DeductChipsWithTx(ctx, tx, "user1", 200, TxTypeCashGameBuyIn, "test-ref", "Buy in"); err != nil {
			return err
		}

		// Second operation fails (insufficient chips)
		if err := service.DeductChipsWithTx(ctx, tx, "user2", 1000, TxTypeCashGameBuyIn, "test-ref", "Buy in"); err != nil {
			return err // This will trigger rollback
		}

		return nil
	})

	if err == nil {
		t.Fatal("Expected transaction to fail, got nil error")
	}

	// Verify both balances unchanged (atomic rollback)
	balance1 := getBalance(t, db, "user1")
	balance2 := getBalance(t, db, "user2")

	if balance1 != initialBalance1 {
		t.Errorf("User1 balance changed after rollback! Expected %d, got %d", initialBalance1, balance1)
	}

	if balance2 != initialBalance2 {
		t.Errorf("User2 balance changed after rollback! Expected %d, got %d", initialBalance2, balance2)
	}

	// Verify no transaction records created
	var txCount int64
	db.Model(&Transaction{}).Count(&txCount)
	if txCount != 0 {
		t.Errorf("Expected 0 transaction records after rollback, got %d", txCount)
	}
}
