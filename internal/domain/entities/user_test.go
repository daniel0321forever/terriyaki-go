package entities

import (
	"strings"
	"testing"
)

func TestNewUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		username    string
		email       string
		hashedPwd   string
		avatar      string
		wantErr     bool
		errContains string
	}{
		{
			name:      "creates user with normalized fields",
			username:  "  Alice  ",
			email:     "  ALICE@EXAMPLE.COM  ",
			hashedPwd: "hashed-secret",
			avatar:    "  https://example.com/a.png  ",
		},
		{
			name:        "rejects empty username",
			username:    "   ",
			email:       "alice@example.com",
			hashedPwd:   "hashed-secret",
			wantErr:     true,
			errContains: "username cannot be empty",
		},
		{
			name:        "rejects empty email",
			username:    "alice",
			email:       "   ",
			hashedPwd:   "hashed-secret",
			wantErr:     true,
			errContains: "email cannot be empty",
		},
		{
			name:        "rejects invalid email format",
			username:    "alice",
			email:       "alice.example.com",
			hashedPwd:   "hashed-secret",
			wantErr:     true,
			errContains: "invalid email format",
		},
		{
			name:        "rejects empty password",
			username:    "alice",
			email:       "alice@example.com",
			hashedPwd:   "   ",
			wantErr:     true,
			errContains: "password cannot be empty",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.username, tt.email, tt.hashedPwd, tt.avatar)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if user == nil {
				t.Fatalf("expected user, got nil")
			}
			if user.ID == "" {
				t.Fatalf("expected non-empty user ID")
			}
			if user.Username != strings.TrimSpace(tt.username) {
				t.Fatalf("expected normalized username %q, got %q", strings.TrimSpace(tt.username), user.Username)
			}
			if user.Email != strings.ToLower(strings.TrimSpace(tt.email)) {
				t.Fatalf("expected normalized email %q, got %q", strings.ToLower(strings.TrimSpace(tt.email)), user.Email)
			}
			if user.Avatar != strings.TrimSpace(tt.avatar) {
				t.Fatalf("expected normalized avatar %q, got %q", strings.TrimSpace(tt.avatar), user.Avatar)
			}
			if user.HashedPassword != tt.hashedPwd {
				t.Fatalf("expected hashed password to be preserved")
			}
		})
	}
}
