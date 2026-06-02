package models

import (
	"os"
	"testing"
)

// ─── Test Helper ─────────────────────────────────────────────────────────────

// setupTempCSV redirects the model to a fresh temp directory so every test
// starts with an empty users.csv and never touches real data.
func setupTempCSV(t *testing.T) func() {
	t.Helper()
	original, _ := os.Getwd()

	dir := t.TempDir()
	os.MkdirAll(dir+"/data", 0755)
	os.Chdir(dir)

	// Write just the header row so ensureUsersFile skips re-creating the file
	f, _ := os.Create(dir + "/data/users.csv")
	f.WriteString("id,name,email,password,created_at\n")
	f.Close()

	return func() { os.Chdir(original) }
}

// splitColon splits on the first colon only (avoids importing "strings")
func splitColon(s string) []string {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}

// ─── HashPassword ─────────────────────────────────────────────────────────────

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		check    func(hash string) bool
	}{
		{
			name:     "returns non-empty",
			password: "secret123",
			check: func(hash string) bool {
				return hash != ""
			},
		},
		{
			name:     "has salt:hash format",
			password: "mypassword",
			check: func(hash string) bool {
				parts := splitColon(hash)
				return len(parts) == 2 && parts[0] != "" && parts[1] != ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.check(hash) {
				t.Errorf("check failed for hash: %s", hash)
			}
		})
	}
}

func TestHashPassword_SameInputGivesDifferentHashes(t *testing.T) {
	h1, _ := HashPassword("same")
	h2, _ := HashPassword("same")
	if h1 == h2 {
		t.Error("expected different hashes (random salt), got identical results")
	}
}

// ─── ValidatePassword ─────────────────────────────────────────────────────────

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name          string
		setupHash     func() string
		inputPassword string
		want          bool
	}{
		{
			name: "correct password returns true",
			setupHash: func() string {
				hash, _ := HashPassword("correcthorse")
				return hash
			},
			inputPassword: "correcthorse",
			want:          true,
		},
		{
			name: "wrong password returns false",
			setupHash: func() string {
				hash, _ := HashPassword("correcthorse")
				return hash
			},
			inputPassword: "wrongpassword",
			want:          false,
		},
		{
			name: "empty input returns false",
			setupHash: func() string {
				hash, _ := HashPassword("secret")
				return hash
			},
			inputPassword: "",
			want:          false,
		},
		{
			name: "missing colon returns false",
			setupHash: func() string {
				return "nocolonatall"
			},
			inputPassword: "anything",
			want:          false,
		},
		{
			name: "invalid hex salt returns false",
			setupHash: func() string {
				return "ZZZINVALIDHEX:abc123def456"
			},
			inputPassword: "anything",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := tt.setupHash()
			u := &User{Password: hash}
			got := u.ValidatePassword(tt.inputPassword)
			if got != tt.want {
				t.Errorf("ValidatePassword(%q) = %v, want %v", tt.inputPassword, got, tt.want)
			}
		})
	}
}

// ─── GetAllUsers ──────────────────────────────────────────────────────────────

func TestGetAllUsers(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T)
		expectedCount int
		checkData     bool
	}{
		{
			name:          "empty file returns empty slice",
			setup:         func(t *testing.T) {},
			expectedCount: 0,
		},
		{
			name: "returns inserted users",
			setup: func(t *testing.T) {
				hash, _ := HashPassword("pass")
				CreateUser(&User{Name: "Alice", Email: "alice@test.com", Password: hash})
			},
			expectedCount: 1,
			checkData:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTempCSV(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t)
			}

			users, err := GetAllUsers()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(users) != tt.expectedCount {
				t.Errorf("expected %d users, got %d", tt.expectedCount, len(users))
			}

			if tt.checkData && len(users) > 0 {
				if users[0].Name != "Alice" || users[0].Email != "alice@test.com" {
					t.Errorf("unexpected user data: %+v", users[0])
				}
			}
		})
	}
}

// ─── CreateUser ───────────────────────────────────────────────────────────────

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name             string
		userFn           func() *User
		wantIDPositive   bool
		wantCreatedAtSet bool
	}{
		{
			name: "assigns positive ID",
			userFn: func() *User {
				hash, _ := HashPassword("pass")
				return &User{Name: "Bob", Email: "bob@test.com", Password: hash}
			},
			wantIDPositive:   true,
			wantCreatedAtSet: true,
		},
		{
			name: "sets created at",
			userFn: func() *User {
				hash, _ := HashPassword("pass")
				return &User{Name: "Carol", Email: "carol@test.com", Password: hash}
			},
			wantIDPositive:   true,
			wantCreatedAtSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTempCSV(t)
			defer cleanup()

			u := tt.userFn()
			if err := CreateUser(u); err != nil {
				t.Fatalf("CreateUser error: %v", err)
			}

			if tt.wantIDPositive && u.ID <= 0 {
				t.Errorf("expected ID > 0, got %d", u.ID)
			}

			if tt.wantCreatedAtSet && u.CreatedAt == "" {
				t.Error("expected CreatedAt to be set")
			}
		})
	}
}

func TestCreateUser_IDsIncrement(t *testing.T) {
	cleanup := setupTempCSV(t)
	defer cleanup()

	hash, _ := HashPassword("pass")
	u1 := &User{Name: "U1", Email: "u1@test.com", Password: hash}
	u2 := &User{Name: "U2", Email: "u2@test.com", Password: hash}
	u3 := &User{Name: "U3", Email: "u3@test.com", Password: hash}
	CreateUser(u1)
	CreateUser(u2)
	CreateUser(u3)

	if !(u1.ID < u2.ID && u2.ID < u3.ID) {
		t.Errorf("expected incrementing IDs, got %d %d %d", u1.ID, u2.ID, u3.ID)
	}
}

// ─── GetUserByEmail ───────────────────────────────────────────────────────────

func TestGetUserByEmail(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T)
		emailLookup string
		wantNil     bool
		checkEmail  string
	}{
		{
			name: "existing email returns user",
			setup: func(t *testing.T) {
				hash, _ := HashPassword("pass")
				CreateUser(&User{Name: "Dave", Email: "dave@test.com", Password: hash})
			},
			emailLookup: "dave@test.com",
			wantNil:     false,
			checkEmail:  "dave@test.com",
		},
		{
			name:        "missing email returns nil",
			emailLookup: "ghost@test.com",
			wantNil:     true,
		},
		{
			name: "case sensitive",
			setup: func(t *testing.T) {
				hash, _ := HashPassword("pass")
				CreateUser(&User{Name: "Eve", Email: "eve@test.com", Password: hash})
			},
			emailLookup: "EVE@TEST.COM",
			wantNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTempCSV(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t)
			}

			user, err := GetUserByEmail(tt.emailLookup)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil && user != nil {
				t.Errorf("expected nil, got %v", user)
			}
			if !tt.wantNil && user == nil {
				t.Fatal("expected user, got nil")
			}
			if !tt.wantNil && tt.checkEmail != "" && user.Email != tt.checkEmail {
				t.Errorf("expected %s, got %s", tt.checkEmail, user.Email)
			}
		})
	}
}

// ─── GetUserByID ──────────────────────────────────────────────────────────────

func TestGetUserByID(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) int
		idToLookup func(int) int
		wantNil    bool
	}{
		{
			name: "existing ID returns user",
			setup: func(t *testing.T) int {
				hash, _ := HashPassword("pass")
				u := &User{Name: "Frank", Email: "frank@test.com", Password: hash}
				CreateUser(u)
				return u.ID
			},
			idToLookup: func(id int) int {
				return id
			},
			wantNil: false,
		},
		{
			name: "missing ID returns nil",
			setup: func(t *testing.T) int {
				return 9999
			},
			idToLookup: func(id int) int {
				return id
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTempCSV(t)
			defer cleanup()

			id := tt.setup(t)
			lookupID := tt.idToLookup(id)

			found, err := GetUserByID(lookupID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantNil && found != nil {
				t.Errorf("expected nil, got %v", found)
			}
			if !tt.wantNil && found == nil {
				t.Fatal("expected user, got nil")
			}
			if !tt.wantNil && found.ID != lookupID {
				t.Errorf("expected ID %d, got %d", lookupID, found.ID)
			}
		})
	}
}

// ─── GetNextUserID ────────────────────────────────────────────────────────────

func TestGetNextUserID(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T)
		expectedNext int
	}{
		{
			name:         "empty file returns one",
			setup:        func(t *testing.T) {},
			expectedNext: 1,
		},
		{
			name: "after one insert returns two",
			setup: func(t *testing.T) {
				hash, _ := HashPassword("pass")
				CreateUser(&User{Name: "Grace", Email: "grace@test.com", Password: hash})
			},
			expectedNext: 2,
		},
		{
			name: "after three inserts returns four",
			setup: func(t *testing.T) {
				hash, _ := HashPassword("pass")
				for i, name := range []string{"H1", "H2", "H3"} {
					CreateUser(&User{Name: name, Email: name + "@test.com", Password: hash})
					_ = i
				}
			},
			expectedNext: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTempCSV(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t)
			}

			id := GetNextUserID()
			if id != tt.expectedNext {
				t.Errorf("expected %d, got %d", tt.expectedNext, id)
			}
		})
	}
}
