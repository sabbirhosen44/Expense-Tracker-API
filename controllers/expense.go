package controllers

import (
	"encoding/json"
	"expense-tracker/models"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

type ExpenseController struct {
	BaseController
}

type expenseInput struct {
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Note        string  `json:"note"`
	ExpenseDate string  `json:"expense_date"`
}

type expenseResponse struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Note        string  `json:"note"`
	ExpenseDate string  `json:"expense_date"`
}

func toExpenseResponse(e models.Expense) expenseResponse {
	return expenseResponse{
		ID:          e.ID,
		Title:       e.Title,
		Amount:      e.Amount,
		Category:    e.Category,
		Note:        e.Note,
		ExpenseDate: e.ExpenseDate,
	}
}

func isValidDate(d string) bool {
	_, err := time.Parse("2006-01-02", d)
	return err == nil
}

func validateExpenseInput(input expenseInput) string {
	if strings.TrimSpace(input.Title) == "" {
		return "Title is required"
	}
	if input.Amount <= 0 {
		return "Amount must be a positive number"
	}
	if strings.TrimSpace(input.ExpenseDate) == "" {
		return "Expense date is required"
	}
	if !isValidDate(input.ExpenseDate) {
		return "Expense date must be in YYYY-MM-DD format"
	}
	if strings.TrimSpace(input.Category) == "" {
		return "Category is required"
	}
	if !models.IsValidCategory(input.Category) {
		return "Invalid category"
	}
	return ""
}

func (c *ExpenseController) CreateExpense() {
	// Retrieve userID injected by the middleware
	userID := c.Ctx.Input.GetData("userID").(int)

	var input expenseInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.respondBadRequest("Invalid request body")
		return
	}

	if msg := validateExpenseInput(input); msg != "" {
		c.respondBadRequest(msg)
		return
	}

	expense := &models.Expense{
		UserID:      userID,
		Title:       strings.TrimSpace(input.Title),
		Amount:      input.Amount,
		Category:    input.Category,
		Note:        strings.TrimSpace(input.Note),
		ExpenseDate: input.ExpenseDate,
	}

	if err := models.CreateExpense(expense); err != nil {
		logs.Error("CreateExpense error:", err)
		c.respondInternalError("Failed to create expense")
		return
	}

	c.respondCreated("Expense created successfully", toExpenseResponse(*expense))
}

func (c *ExpenseController) ListExpenses() {
	userID := c.Ctx.Input.GetData("userID").(int)
	if userID == 0 {
		return
	}

	params := models.FilterExpensesParams{
		Category:  c.GetString("category"),
		DateFrom:  c.GetString("date_from"),
		DateTo:    c.GetString("date_to"),
		SortBy:    c.GetString("sort_by"),
		SortOrder: c.GetString("sort_order"),
	}

	// Validate sort_by value
	if params.SortBy != "" && params.SortBy != "amount" && params.SortBy != "expense_date" {
		c.respondBadRequest("sort_by must be 'amount' or 'expense_date'")
		return
	}
	// Validate sort_order value
	if params.SortOrder != "" && params.SortOrder != "asc" && params.SortOrder != "desc" {
		c.respondBadRequest("sort_order must be 'asc' or 'desc'")
		return
	}
	// Default sort order
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	// Validate date params
	if params.DateFrom != "" && !isValidDate(params.DateFrom) {
		c.respondBadRequest("date_from must be in YYYY-MM-DD format")
		return
	}
	if params.DateTo != "" && !isValidDate(params.DateTo) {
		c.respondBadRequest("date_to must be in YYYY-MM-DD format")
		return
	}

	// Pagination
	limit, _ := c.GetInt("limit", 10)
	offset, _ := c.GetInt("offset", 0)
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	params.Limit = limit
	params.Offset = offset

	expenses, err := models.FilterExpenses(userID, params)
	if err != nil {
		logs.Error("ListExpenses: error:", err)
		return
	}

	var result []expenseResponse
	for _, e := range expenses {
		result = append(result, toExpenseResponse(e))
	}
	if result == nil {
		result = []expenseResponse{}
	}

	c.respondOK("Expenses retrieved", result)
}

func (c *ExpenseController) GetExpense() {
	userID := c.Ctx.Input.GetData("userID").(int)
	if userID == 0 {
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.respondBadRequest("Invalid expense ID")
		return
	}

	expense, err := models.GetExpenseByID(id, userID)
	if err != nil {
		logs.Error("GetExpense: error:", err)
		c.respondInternalError("Failed to retrieve expense")
		return
	}
	if expense == nil {
		c.respondNotFound("Expense not found")
		return
	}

	c.respondOK("Expense retrieved", toExpenseResponse(*expense))
}

func (c *ExpenseController) UpdateExpense() {
	userID := c.Ctx.Input.GetData("userID").(int)
	if userID == 0 {
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.respondBadRequest("Invalid expense ID")
		return
	}

	// Verify ownership
	existing, err := models.GetExpenseByID(id, userID)
	if err != nil {
		logs.Error("UpdateExpense: error fetching expense:", err)
		c.respondInternalError("Failed to update expense")
		return
	}
	if existing == nil {
		c.respondNotFound("Expense not found")
		return
	}

	var input expenseInput
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		c.respondBadRequest("Invalid request body")
		return
	}

	if msg := validateExpenseInput(input); msg != "" {
		c.respondBadRequest(msg)
		return
	}

	updated := &models.Expense{
		ID:          id,
		UserID:      userID,
		Title:       strings.TrimSpace(input.Title),
		Amount:      input.Amount,
		Category:    input.Category,
		Note:        strings.TrimSpace(input.Note),
		ExpenseDate: input.ExpenseDate,
	}

	if err := models.UpdateExpense(updated); err != nil {
		logs.Error("UpdateExpense: error:", err)
		c.respondInternalError("Failed to update expense")
		return
	}

	logs.Info("UpdateExpense: updated id=", id, " for user=", userID)
	c.respondOK("Expense updated successfully", toExpenseResponse(*updated))
}

func (c *ExpenseController) DeleteExpense() {
	userID := c.Ctx.Input.GetData("userID").(int)
	if userID == 0 {
		return
	}

	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.respondBadRequest("Invalid expense ID")
		return
	}

	// Verify ownership
	existing, err := models.GetExpenseByID(id, userID)
	if err != nil {
		logs.Error("DeleteExpense: error fetching expense:", err)
		c.respondInternalError("Failed to delete expense")
		return
	}
	if existing == nil {
		c.respondNotFound("Expense not found")
		return
	}

	if err := models.DeleteExpense(id, userID); err != nil {
		logs.Error("DeleteExpense: error:", err)
		c.respondInternalError("Failed to delete expense")
		return
	}

	logs.Info("DeleteExpense: deleted id=", id, " for user=", userID)
	c.respondOK("Expense deleted successfully", nil)
}

func (c *ExpenseController) GetSummary() {
	userID := c.Ctx.Input.GetData("userID").(int)
	if userID == 0 {
		return
	}

	dateFrom := c.GetString("date_from")
	dateTo := c.GetString("date_to")

	// Only validate if they are provided
	if (dateFrom != "" && !isValidDate(dateFrom)) || (dateTo != "" && !isValidDate(dateTo)) {
		c.respondBadRequest("Invalid date format. Use YYYY-MM-DD")
		return
	}

	summary, err := models.GetExpenseSummary(userID, dateFrom, dateTo)
	if err != nil {
		c.respondInternalError("Failed to generate summary")
		return
	}

	c.respondOK("Summary generated", summary)
}
