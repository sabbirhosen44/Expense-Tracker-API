package models

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	beego "github.com/beego/beego/v2/server/web"
)

var mu sync.Mutex

type User struct {
	ID        int
	Name      string
	Email     string
	Password  string
	CreatedAt string
}

// usersFilePath returns the path to the CSV file
func usersFilePath() string {
	path, _ := beego.AppConfig.String("users_csv")
	if path == "" {
		path = "data/users.csv"
	}
	return path
}

// ensureUsersFile creates the CSV file with headers if it doesn't exist
func ensureUsersFile() error {
	path := usersFilePath()
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
		if err := w.Write([]string{"id", "name", "email", "password", "created_at"}); err != nil {
			return err
		}
		w.Flush()
	}
	return nil
}

// HashPassword generates a salted SHA-256 hash
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	rand.Read(salt)

	hash := sha256.New()
	hash.Write(salt)
	hash.Write([]byte(password))

	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash.Sum(nil)), nil
}

// ValidatePassword checks a plain password against the stored salted hash
func (u *User) ValidatePassword(password string) bool {
	parts := strings.Split(u.Password, ":")
	if len(parts) != 2 {
		return false
	}

	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}

	hash := sha256.New()
	hash.Write(salt)
	hash.Write([]byte(password))

	return hex.EncodeToString(hash.Sum(nil)) == parts[1]
}

// GetAllUsers returns all users from the CSV
func GetAllUsers() ([]User, error) {
	mu.Lock()
	defer mu.Unlock()

	if err := ensureUsersFile(); err != nil {
		return nil, err
	}
	f, err := os.Open(usersFilePath())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, err
	}

	var users []User
	for i, record := range records {
		if i == 0 {
			continue
		}
		id, _ := strconv.Atoi(record[0])
		users = append(users, User{
			ID:        id,
			Name:      record[1],
			Email:     record[2],
			Password:  record[3],
			CreatedAt: record[4],
		})
	}
	return users, nil
}

// GetUserByEmail searches for a user record by email
func GetUserByEmail(email string) (*User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if strings.TrimSpace(u.Email) == email {
			copy := u
			return &copy, nil
		}
	}

	return nil, nil
}

// GetUserByID finds a user by ID. Returns nil if not found.
func GetUserByID(id int) (*User, error) {
	users, err := GetAllUsers()
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if u.ID == id {
			copy := u
			return &copy, nil
		}
	}
	return nil, nil
}

// CreateUser appends a new user to the CSV file
func CreateUser(user *User) error {
	mu.Lock()
	defer mu.Unlock()

	if err := ensureUsersFile(); err != nil {
		return err
	}

	user.ID = GetNextUserIDInternal()
	user.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	f, err := os.OpenFile(usersFilePath(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	err = w.Write([]string{
		strconv.Itoa(user.ID),
		user.Name,
		user.Email,
		user.Password,
		user.CreatedAt,
	})
	if err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

// GetNextUserIDInternal calculates the next available ID (assumes caller holds the lock)
func GetNextUserIDInternal() int {
	f, err := os.Open(usersFilePath())
	if err != nil {
		return 1
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil || len(records) <= 1 {
		return 1
	}

	max := 0
	for i, record := range records {
		if i == 0 {
			continue
		}
		if len(record) > 0 {
			id, _ := strconv.Atoi(strings.TrimSpace(record[0]))
			if id > max {
				max = id
			}
		}
	}
	return max + 1
}

// GetNextUserID is a thread-safe wrapper to get the next ID
func GetNextUserID() int {
	mu.Lock()
	defer mu.Unlock()
	return GetNextUserIDInternal()
}

// ErrUserNotFound is returned when a user lookup fails.
var ErrUserNotFound = errors.New("user not found")
