package models

import (
	"os"
	"testing"
)

// ─── Helper ──────────────────────────────────────────────────────────────────

func setupExpenseCSV(t *testing.T) func() {
	t.Helper()
	original, _ := os.Getwd()

	dir := t.TempDir()
	os.MkdirAll(dir+"/data", 0755)
	os.Chdir(dir)

	// users.csv needed by any indirect user lookups
	uf, _ := os.Create(dir + "/data/users.csv")
	uf.WriteString("id,name,email,password,created_at\n")
	uf.Close()

	ef, _ := os.Create(dir + "/data/expenses.csv")
	ef.WriteString("id,user_id,title,amount,category,note,expense_date,created_at\n")
	ef.Close()

	return func() { os.Chdir(original) }
}

func sampleExpense(userID int) *Expense {
	return &Expense{
		UserID:      userID,
		Title:       "Lunch",
		Amount:      12.50,
		Category:    "Food",
		Note:        "test note",
		ExpenseDate: "2024-06-01",
	}
}

// ─── IsValidCategory ─────────────────────────────────────────────────────────

func TestIsValidCategory(t *testing.T) {
	tests := []struct {
		name     string
		category string
		want     bool
	}{
		{
			name:     "valid Food",
			category: "Food",
			want:     true,
		},
		{
			name:     "valid Transport",
			category: "Transport",
			want:     true,
		},
		{
			name:     "valid Entertainment",
			category: "Entertainment",
			want:     true,
		},
		{
			name:     "invalid Gambling",
			category: "Gambling",
			want:     false,
		},
		{
			name:     "empty string",
			category: "",
			want:     false,
		},
		{
			name:     "case sensitive lowercase",
			category: "food",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidCategory(tt.category)
			if got != tt.want {
				t.Errorf("IsValidCategory(%q) = %v, want %v", tt.category, got, tt.want)
			}
		})
	}
}

// ─── CreateExpense ────────────────────────────────────────────────────────────

func TestCreateExpense(t *testing.T) {
	tests := []struct {
		name             string
		expenseFn        func() *Expense
		wantIDPositive   bool
		wantCreatedAtSet bool
	}{
		{
			name: "assigns positive ID",
			expenseFn: func() *Expense {
				return sampleExpense(1)
			},
			wantIDPositive:   true,
			wantCreatedAtSet: true,
		},
		{
			name: "sets created at",
			expenseFn: func() *Expense {
				return sampleExpense(1)
			},
			wantIDPositive:   true,
			wantCreatedAtSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseCSV(t)
			defer cleanup()

			e := tt.expenseFn()
			if err := CreateExpense(e); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantIDPositive && e.ID <= 0 {
				t.Errorf("expected ID > 0, got %d", e.ID)
			}

			if tt.wantCreatedAtSet && e.CreatedAt == "" {
				t.Error("expected CreatedAt to be set")
			}
		})
	}
}

func TestCreateExpense_IDsIncrement(t *testing.T) {
	cleanup := setupExpenseCSV(t)
	defer cleanup()

	e1 := sampleExpense(1)
	e2 := sampleExpense(1)
	e3 := sampleExpense(1)
	CreateExpense(e1)
	CreateExpense(e2)
	CreateExpense(e3)

	if !(e1.ID < e2.ID && e2.ID < e3.ID) {
		t.Errorf("expected incrementing IDs, got %d %d %d", e1.ID, e2.ID, e3.ID)
	}
}

// ─── GetExpensesByUserID ──────────────────────────────────────────────────────

func TestGetExpensesByUserID(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T)
		userID        int
		expectedCount int
	}{
		{
			name: "returns only owner expenses",
			setup: func(t *testing.T) {
				CreateExpense(sampleExpense(1))
				CreateExpense(sampleExpense(1))
				CreateExpense(sampleExpense(2))
			},
			userID:        1,
			expectedCount: 2,
		},
		{
			name: "no expenses returns empty",
			setup: func(t *testing.T) {
				// no setup
			},
			userID:        99,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseCSV(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t)
			}

			expenses, err := GetExpensesByUserID(tt.userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(expenses) != tt.expectedCount {
				t.Errorf("expected %d expenses, got %d", tt.expectedCount, len(expenses))
			}
		})
	}
}

// ─── GetExpenseByID ───────────────────────────────────────────────────────────

func TestGetExpenseByID(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) *Expense
		userID      int
		expenseIDFn func(e *Expense) int
		wantNil     bool
	}{
		{
			name: "found returns expense",
			setup: func(t *testing.T) *Expense {
				e := sampleExpense(1)
				CreateExpense(e)
				return e
			},
			userID: 1,
			expenseIDFn: func(e *Expense) int {
				return e.ID
			},
			wantNil: false,
		},
		{
			name: "wrong user returns nil",
			setup: func(t *testing.T) *Expense {
				e := sampleExpense(1)
				CreateExpense(e)
				return e
			},
			userID: 2,
			expenseIDFn: func(e *Expense) int {
				return e.ID
			},
			wantNil: true,
		},
		{
			name: "not found returns nil",
			setup: func(t *testing.T) *Expense {
				return nil
			},
			userID: 1,
			expenseIDFn: func(e *Expense) int {
				return 9999
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseCSV(t)
			defer cleanup()

			e := tt.setup(t)
			expenseID := tt.expenseIDFn(e)

			found, err := GetExpenseByID(expenseID, tt.userID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil && found != nil {
				t.Errorf("expected nil, got %v", found)
			}
			if !tt.wantNil && found == nil {
				t.Fatal("expected expense, got nil")
			}
			if !tt.wantNil && found.ID != expenseID {
				t.Errorf("expected ID %d, got %d", expenseID, found.ID)
			}
		})
	}
}

// ─── UpdateExpense ────────────────────────────────────────────────────────────

func TestUpdateExpense(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) *Expense
		modifyFn func(e *Expense)
		userID   int
		wantErr  bool
	}{
		{
			name: "updates fields",
			setup: func(t *testing.T) *Expense {
				e := sampleExpense(1)
				CreateExpense(e)
				return e
			},
			modifyFn: func(e *Expense) {
				e.Title = "Dinner"
				e.Amount = 25.00
			},
			userID:  1,
			wantErr: false,
		},
		{
			name: "preserves created at",
			setup: func(t *testing.T) *Expense {
				e := sampleExpense(1)
				CreateExpense(e)
				return e
			},
			modifyFn: func(e *Expense) {
				e.Title = "Updated"
			},
			userID:  1,
			wantErr: false,
		},
		{
			name: "wrong user returns error",
			setup: func(t *testing.T) *Expense {
				e := sampleExpense(1)
				CreateExpense(e)
				return e
			},
			modifyFn: func(e *Expense) {
				e.UserID = 2
			},
			userID:  1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseCSV(t)
			defer cleanup()

			e := tt.setup(t)
			originalCreatedAt := e.CreatedAt
			tt.modifyFn(e)

			err := UpdateExpense(e)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateExpense error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.name == "preserves created at" {
				updated, _ := GetExpenseByID(e.ID, tt.userID)
				if updated.CreatedAt != originalCreatedAt {
					t.Errorf("expected CreatedAt to be preserved, got %s", updated.CreatedAt)
				}
			}
		})
	}
}

// ─── DeleteExpense ────────────────────────────────────────────────────────────

func TestDeleteExpense(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) *Expense
		userID  int
		wantErr bool
	}{
		{
			name: "removes expense",
			setup: func(t *testing.T) *Expense {
				e := sampleExpense(1)
				CreateExpense(e)
				return e
			},
			userID:  1,
			wantErr: false,
		},
		{
			name: "wrong user returns error",
			setup: func(t *testing.T) *Expense {
				e := sampleExpense(1)
				CreateExpense(e)
				return e
			},
			userID:  2,
			wantErr: true,
		},
		{
			name: "non-existent ID returns error",
			setup: func(t *testing.T) *Expense {
				return &Expense{ID: 9999}
			},
			userID:  1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseCSV(t)
			defer cleanup()

			e := tt.setup(t)
			err := DeleteExpense(e.ID, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteExpense error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && tt.name == "removes expense" {
				found, _ := GetExpenseByID(e.ID, tt.userID)
				if found != nil {
					t.Error("expected expense to be deleted")
				}
			}
		})
	}
}

// ─── GetNextExpenseID ─────────────────────────────────────────────────────────

func TestGetNextExpenseID(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		expectedNext int
	}{
		{
			name:         "empty file returns one",
			expectedNext: 1,
		},
		{
			name: "after insert increments",
			setup: func(t *testing.T) {
				CreateExpense(sampleExpense(1))
			},
			expectedNext: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseCSV(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t)
			}

			id := GetNextExpenseID()
			if id != tt.expectedNext {
				t.Errorf("expected %d, got %d", tt.expectedNext, id)
			}
		})
	}
}

// ─── FilterExpenses ───────────────────────────────────────────────────────────

func TestFilterExpenses(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T)
		params        FilterExpensesParams
		expectedCount int
		checkOrder    bool
		checkTitle    string
	}{
		{
			name: "by category",
			setup: func(t *testing.T) {
				CreateExpense(sampleExpense(1))
				CreateExpense(&Expense{
					UserID:      1,
					Title:       "Bus",
					Amount:      5,
					Category:    "Transport",
					ExpenseDate: "2024-06-01",
				})
			},
			params:        FilterExpensesParams{Category: "Food"},
			expectedCount: 1,
			checkTitle:    "Lunch",
		},
		{
			name: "by date from",
			setup: func(t *testing.T) {
				CreateExpense(&Expense{
					UserID:      1,
					Title:       "Old",
					Amount:      5,
					Category:    "Food",
					ExpenseDate: "2024-01-01",
				})
				CreateExpense(&Expense{
					UserID:      1,
					Title:       "New",
					Amount:      5,
					Category:    "Food",
					ExpenseDate: "2024-06-01",
				})
			},
			params:        FilterExpensesParams{DateFrom: "2024-03-01"},
			expectedCount: 1,
			checkTitle:    "New",
		},
		{
			name: "by date to",
			setup: func(t *testing.T) {
				CreateExpense(&Expense{
					UserID:      1,
					Title:       "Old",
					Amount:      5,
					Category:    "Food",
					ExpenseDate: "2024-01-01",
				})
				CreateExpense(&Expense{
					UserID:      1,
					Title:       "New",
					Amount:      5,
					Category:    "Food",
					ExpenseDate: "2024-06-01",
				})
			},
			params:        FilterExpensesParams{DateTo: "2024-03-01"},
			expectedCount: 1,
			checkTitle:    "Old",
		},
		{
			name: "sort by amount asc",
			setup: func(t *testing.T) {
				CreateExpense(&Expense{
					UserID:      1,
					Title:       "Cheap",
					Amount:      5,
					Category:    "Food",
					ExpenseDate: "2024-06-01",
				})
				CreateExpense(&Expense{
					UserID:      1,
					Title:       "Expensive",
					Amount:      100,
					Category:    "Food",
					ExpenseDate: "2024-06-02",
				})
			},
			params:        FilterExpensesParams{SortBy: "amount", SortOrder: "asc"},
			expectedCount: 2,
			checkOrder:    true,
		},
		{
			name: "sort by amount desc",
			setup: func(t *testing.T) {
				CreateExpense(&Expense{
					UserID:      1,
					Title:       "Cheap",
					Amount:      5,
					Category:    "Food",
					ExpenseDate: "2024-06-01",
				})
				CreateExpense(&Expense{
					UserID:      1,
					Title:       "Expensive",
					Amount:      100,
					Category:    "Food",
					ExpenseDate: "2024-06-02",
				})
			},
			params:        FilterExpensesParams{SortBy: "amount", SortOrder: "desc"},
			expectedCount: 2,
			checkOrder:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseCSV(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t)
			}

			results, err := FilterExpenses(1, tt.params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(results) != tt.expectedCount {
				t.Errorf("expected %d results, got %d", tt.expectedCount, len(results))
			}

			if tt.checkTitle != "" && len(results) > 0 {
				if results[0].Title != tt.checkTitle {
					t.Errorf("expected title %q, got %q", tt.checkTitle, results[0].Title)
				}
			}

			if tt.checkOrder && len(results) >= 2 {
				switch tt.name {
				case "sort by amount asc":
					if results[0].Amount > results[1].Amount {
						t.Error("expected ascending sort by amount")
					}
				case "sort by amount desc":
					if results[0].Amount < results[1].Amount {
						t.Error("expected descending sort by amount")
					}
				}
			}
		})
	}
}

// ─── GetExpenseSummary ────────────────────────────────────────────────────────

func TestGetExpenseSummary(t *testing.T) {
	tests := []struct {
		name               string
		setup              func(t *testing.T)
		dateFrom           string
		dateTo             string
		expectedTotal      float64
		expectedCount      int
		expectedCategories int
	}{
		{
			name: "total amount",
			setup: func(t *testing.T) {
				CreateExpense(&Expense{UserID: 1, Title: "A", Amount: 10, Category: "Food", ExpenseDate: "2024-06-01"})
				CreateExpense(&Expense{UserID: 1, Title: "B", Amount: 20, Category: "Food", ExpenseDate: "2024-06-02"})
			},
			expectedTotal:      30,
			expectedCount:      2,
			expectedCategories: 1,
		},
		{
			name: "groups by category",
			setup: func(t *testing.T) {
				CreateExpense(&Expense{UserID: 1, Title: "A", Amount: 10, Category: "Food", ExpenseDate: "2024-06-01"})
				CreateExpense(&Expense{UserID: 1, Title: "B", Amount: 50, Category: "Transport", ExpenseDate: "2024-06-01"})
			},
			expectedTotal:      60,
			expectedCount:      2,
			expectedCategories: 2,
		},
		{
			name: "with date range",
			setup: func(t *testing.T) {
				CreateExpense(&Expense{UserID: 1, Title: "In", Amount: 10, Category: "Food", ExpenseDate: "2024-06-15"})
				CreateExpense(&Expense{UserID: 1, Title: "Out", Amount: 99, Category: "Food", ExpenseDate: "2024-01-01"})
			},
			dateFrom:           "2024-06-01",
			dateTo:             "2024-06-30",
			expectedTotal:      10,
			expectedCount:      1,
			expectedCategories: 1,
		},
		{
			name:               "no expenses returns zero",
			setup:              func(t *testing.T) {},
			expectedTotal:      0,
			expectedCount:      0,
			expectedCategories: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseCSV(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t)
			}

			summary, err := GetExpenseSummary(1, tt.dateFrom, tt.dateTo)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if summary.TotalAmount != tt.expectedTotal {
				t.Errorf("expected total %f, got %f", tt.expectedTotal, summary.TotalAmount)
			}
			if summary.TotalCount != tt.expectedCount {
				t.Errorf("expected count %d, got %d", tt.expectedCount, summary.TotalCount)
			}
			if len(summary.ByCategory) != tt.expectedCategories {
				t.Errorf("expected %d categories, got %d", tt.expectedCategories, len(summary.ByCategory))
			}
		})
	}
}
