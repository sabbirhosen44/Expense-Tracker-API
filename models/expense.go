package models

import (
	"encoding/csv"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/server/web"
)

// AllowedCategories defines the valid expense categories.
var AllowedCategories = []string{
	"Food", "Transport", "Housing", "Entertainment",
	"Shopping", "Healthcare", "Education", "Utilities", "Other",
}

// Expense represents a single expense record.
type Expense struct {
	ID          int
	UserID      int
	Title       string
	Amount      float64
	Category    string
	Note        string
	ExpenseDate string
	CreatedAt   string
}

// CategorySummary holds the total and count for one category.
type CategorySummary struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
	Count    int     `json:"count"`
}

// ExpenseSummary is the response payload for the summary endpoint.
type ExpenseSummary struct {
	DateFrom    string            `json:"date_from,omitempty"`
	DateTo      string            `json:"date_to,omitempty"`
	TotalAmount float64           `json:"total_amount"`
	TotalCount  int               `json:"total_count"`
	ByCategory  []CategorySummary `json:"by_category"`
}

// FilterExpensesParams holds optional filter and sort parameters.
type FilterExpensesParams struct {
	Category  string
	DateFrom  string
	DateTo    string
	SortBy    string
	SortOrder string
	Limit     int
	Offset    int
}

// writeAllExpenses rewrites the entire CSV with the provided slice.
func writeAllExpenses(expenses []Expense) error {
	f, err := os.OpenFile(expensesFilePath(), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write([]string{"id", "user_id", "title", "amount", "category", "note", "expense_date", "created_at"}); err != nil {
		return err
	}
	for _, e := range expenses {
		if err := w.Write(expenseToRecord(e)); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// expenseToRecord converts an Expense to a CSV string slice.
func expenseToRecord(e Expense) []string {
	return []string{
		strconv.Itoa(e.ID),
		strconv.Itoa(e.UserID),
		e.Title,
		strconv.FormatFloat(e.Amount, 'f', 2, 64),
		e.Category,
		e.Note,
		e.ExpenseDate,
		e.CreatedAt,
	}
}

// IsValidCategory returns true if the category string is in AllowedCategories.
func IsValidCategory(category string) bool {
	for _, c := range AllowedCategories {
		if c == category {
			return true
		}
	}
	return false
}

func expensesFilePath() string {
	path, _ := beego.AppConfig.String("expenses_csv")
	if path == "" {
		path = "data/expenses.csv"
	}
	return path
}

// ensureExpensesFile creates the CSV file with headers if it does not exist.
func ensureExpensesFile() error {
	path := expensesFilePath()
	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		w := csv.NewWriter(f)
		if err := w.Write([]string{"id", "user_id", "title", "amount", "category", "note", "expense_date", "created_at"}); err != nil {
			return err
		}
		w.Flush()
	}
	return nil
}

// readAllExpenses reads every expense row from CSV.
func readAllExpenses() ([]Expense, error) {
	if err := ensureExpensesFile(); err != nil {
		return nil, err
	}
	f, err := os.Open(expensesFilePath())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	var expenses []Expense
	for i, r := range records {
		if i == 0 || len(r) < 8 {
			continue
		}

		id, _ := strconv.Atoi(strings.TrimSpace(r[0]))
		userID, _ := strconv.Atoi(strings.TrimSpace(r[1]))

		// Trim space before parsing to avoid 0 values
		amount, err := strconv.ParseFloat(strings.TrimSpace(r[3]), 64)
		if err != nil {
			continue
		}

		expenses = append(expenses, Expense{
			ID:          id,
			UserID:      userID,
			Title:       strings.TrimSpace(r[2]),
			Amount:      amount,
			Category:    strings.TrimSpace(r[4]),
			Note:        strings.TrimSpace(r[5]),
			ExpenseDate: strings.TrimSpace(r[6]),
			CreatedAt:   strings.TrimSpace(r[7]),
		})
	}
	return expenses, nil
}

// GetExpensesByUserID returns all expenses belonging to the given user.
func GetExpensesByUserID(userID int) ([]Expense, error) {
	all, err := readAllExpenses()
	if err != nil {
		return nil, err
	}
	var result []Expense
	for _, e := range all {
		if e.UserID == userID {
			result = append(result, e)
		}
	}
	return result, nil
}

// GetExpenseByID returns a single expense belonging to the user, or nil.
func GetExpenseByID(id int, userID int) (*Expense, error) {
	all, err := readAllExpenses()
	if err != nil {
		return nil, err
	}
	for _, e := range all {
		if e.ID == id && e.UserID == userID {
			copy := e
			return &copy, nil
		}
	}
	return nil, nil
}

// CreateExpense appends a new expense to the CSV.
func CreateExpense(expense *Expense) error {
	if err := ensureExpensesFile(); err != nil {
		return err
	}
	expense.ID = GetNextExpenseID()
	expense.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	f, err := os.OpenFile(expensesFilePath(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(expenseToRecord(*expense)); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

// UpdateExpense rewrites the CSV replacing the matching expense.
func UpdateExpense(expense *Expense) error {
	all, err := readAllExpenses()
	if err != nil {
		return err
	}
	found := false
	for i, e := range all {
		if e.ID == expense.ID && e.UserID == expense.UserID {
			expense.CreatedAt = e.CreatedAt // preserve original timestamp
			all[i] = *expense
			found = true
			break
		}
	}
	if !found {
		return ErrUserNotFound
	}
	return writeAllExpenses(all)
}

// DeleteExpense rewrites the CSV omitting the row with the given id/userID.
func DeleteExpense(id int, userID int) error {
	all, err := readAllExpenses()
	if err != nil {
		return err
	}
	var filtered []Expense
	found := false
	for _, e := range all {
		if e.ID == id && e.UserID == userID {
			found = true
			continue
		}
		filtered = append(filtered, e)
	}
	if !found {
		return ErrUserNotFound
	}
	return writeAllExpenses(filtered)
}

// GetNextExpenseID returns the next available expense ID.
func GetNextExpenseID() int {
	all, err := readAllExpenses()
	if err != nil || len(all) == 0 {
		return 1
	}
	max := 0
	for _, e := range all {
		if e.ID > max {
			max = e.ID
		}
	}
	return max + 1
}

// FilterExpenses applies category/date filters, sorting, and pagination to user expenses.
func FilterExpenses(userID int, params FilterExpensesParams) ([]Expense, error) {
	expenses, err := GetExpensesByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Data Normalization: Trim all strings in retrieved expenses
	for i := range expenses {
		expenses[i].Title = strings.TrimSpace(expenses[i].Title)
		expenses[i].Category = strings.TrimSpace(expenses[i].Category)
		expenses[i].Note = strings.TrimSpace(expenses[i].Note)
		expenses[i].ExpenseDate = strings.TrimSpace(expenses[i].ExpenseDate)
	}

	// Filter by Category
	if cat := strings.TrimSpace(params.Category); cat != "" {
		var filtered []Expense
		for _, e := range expenses {
			if strings.EqualFold(e.Category, cat) {
				filtered = append(filtered, e)
			}
		}
		expenses = filtered
	}

	// Filter by Date Range (YYYY-MM-DD comparison)
	if from := strings.TrimSpace(params.DateFrom); from != "" {
		var filtered []Expense
		for _, e := range expenses {
			if e.ExpenseDate >= from {
				filtered = append(filtered, e)
			}
		}
		expenses = filtered
	}
	if to := strings.TrimSpace(params.DateTo); to != "" {
		var filtered []Expense
		for _, e := range expenses {
			if e.ExpenseDate <= to {
				filtered = append(filtered, e)
			}
		}
		expenses = filtered
	}

	// Sorting
	if params.SortBy != "" {
		sort.Slice(expenses, func(i, j int) bool {
			var less bool
			switch params.SortBy {
			case "amount":
				less = expenses[i].Amount < expenses[j].Amount
			case "expense_date":
				less = expenses[i].ExpenseDate < expenses[j].ExpenseDate
			default:
				less = expenses[i].ID < expenses[j].ID
			}
			if params.SortOrder == "asc" {
				return less
			}
			return !less
		})
	} else {
		// Default: newest first
		sort.Slice(expenses, func(i, j int) bool {
			return expenses[i].ExpenseDate > expenses[j].ExpenseDate
		})
	}

	// Pagination
	total := len(expenses)
	if params.Offset >= total {
		return []Expense{}, nil
	}
	expenses = expenses[params.Offset:]
	if params.Limit > 0 && params.Limit < len(expenses) {
		expenses = expenses[:params.Limit]
	}

	return expenses, nil
}

// GetExpenseSummary computes totals grouped by category for a date range.
func GetExpenseSummary(userID int, dateFrom, dateTo string) (*ExpenseSummary, error) {
	expenses, err := GetExpensesByUserID(userID)
	if err != nil {
		return nil, err
	}

	categoryTotals := make(map[string]float64)
	categoryCounts := make(map[string]int)
	var totalAmount float64
	var totalCount int

	for _, e := range expenses {
		// Date filtering logic
		if dateFrom != "" && e.ExpenseDate < dateFrom {
			continue
		}
		if dateTo != "" && e.ExpenseDate > dateTo {
			continue
		}

		categoryTotals[e.Category] += e.Amount
		categoryCounts[e.Category]++
		totalAmount += e.Amount
		totalCount++
	}

	var byCategory []CategorySummary
	for cat, total := range categoryTotals {
		byCategory = append(byCategory, CategorySummary{
			Category: cat,
			Total:    total,
			Count:    categoryCounts[cat],
		})
	}

	// Sort by total amount descending
	sort.Slice(byCategory, func(i, j int) bool {
		return byCategory[i].Total > byCategory[j].Total
	})

	return &ExpenseSummary{
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		TotalAmount: totalAmount,
		TotalCount:  totalCount,
		ByCategory:  byCategory,
	}, nil
}
