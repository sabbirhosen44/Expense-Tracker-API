package controllers

import (
	"encoding/json"
	"expense-tracker/models"
	"regexp"
	"strings"

	"github.com/beego/beego/v2/core/logs"
)

type AuthController struct {
	BaseController
}

type registerInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Regex For Email Validation
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Register Function
func (c *AuthController) Register() {
	var input registerInput

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logs.Warn("Register: failed to parse the request body: ", err)
		c.respondBadRequest("Invalid request body")
		return
	}

	// Sanitize The Input
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Password = strings.TrimSpace(input.Password)

	// Validation Check
	if input.Name == "" {
		c.respondBadRequest("Name is required")
		return
	}
	if input.Email == "" {
		c.respondBadRequest("Email is required")
		return
	}
	if !emailRegex.MatchString(input.Email) {
		c.respondBadRequest("Invalid email format")
		return
	}
	if input.Password == "" {
		c.respondBadRequest("Password is required")
		return
	}
	if len(input.Password) < 6 {
		c.respondBadRequest("Password must be at least 6 characters")
		return
	}

	// Check For Duplicate Email
	existing, err := models.GetUserByEmail(input.Email)
	if err != nil {
		logs.Error("Register: error checking email:", err)
		c.respondInternalError("Internal server error")
		return
	}
	if existing != nil {
		c.respondConflict("Email already exists")
		return
	}

	// Hashing Passwords

	hashedPassword, err := models.HashPassword(input.Password)
	if err != nil {
		c.respondInternalError("Error securing password")
		return
	}

	// New User Create
	user := &models.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: hashedPassword,
	}

	if err := models.CreateUser(user); err != nil {
		logs.Error("Register: error creating user:", err)
		c.respondInternalError("Failed to create user")
		return
	}

	logs.Info("Register: new user created, id=", user.ID)
	c.respondCreated("User registered successfully", nil)

}

func (c *AuthController) Login() {
	var input loginInput

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &input); err != nil {
		logs.Warn("Login: failed to parse request body:", err)
		c.respondBadRequest("Invalid request body")
		return
	}

	// Sanitize The Input
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.Password = strings.TrimSpace(input.Password)

	// Validation Check
	if input.Email == "" || input.Password == "" {
		c.respondBadRequest("Email and password are required")
		return
	}

	user, err := models.GetUserByEmail(input.Email)
	if err != nil {
		logs.Error("Login: error fetching user:", err)
		c.respondInternalError("Internal server error")
		return
	}
	if user == nil || !user.ValidatePassword(input.Password) {
		c.respondUnauthorized("Invalid email or password")
		return
	}

	// Login Response
	logs.Info("Login: user authenticated, id=", user.ID)
	c.respondOK("Login successful", map[string]interface{}{
		"user_id": user.ID,
		"name":    user.Name,
		"email":   user.Email,
	})
}
