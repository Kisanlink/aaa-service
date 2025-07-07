package models

import (
	"testing"
	"time"

	"github.com/Kisanlink/kisanlink-db/pkg/core/hash"
)

func TestNewUser(t *testing.T) {
	for _, tt := range NewUserTests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser(tt.name, tt.email)

			if user == nil {
				t.Fatal("NewUser returned nil")
			}

			if user.Name != tt.name {
				t.Errorf("Expected name %s, got %s", tt.name, user.Name)
			}

			if user.Email != tt.email {
				t.Errorf("Expected email %s, got %s", tt.email, user.Email)
			}

			if user.GetTableIdentifier() != "USER" {
				t.Errorf("Expected table identifier USER, got %s", user.GetTableIdentifier())
			}

			if user.GetTableSize() != hash.Medium {
				t.Errorf("Expected table size Medium, got %s", user.GetTableSize())
			}

			// Verify ID starts with USER
			if len(user.ID) < 4 || user.ID[:4] != "USER" {
				t.Errorf("Expected ID to start with USER, got %s", user.ID)
			}

			// Verify timestamps are set
			if user.CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}

			if user.UpdatedAt.IsZero() {
				t.Error("UpdatedAt should not be zero")
			}
		})
	}
}

func TestUserBeforeCreate(t *testing.T) {
	for _, tt := range UserBeforeCreateTests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser(tt.name, tt.email)
			err := user.BeforeCreate()

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tt.shouldError {
				// Verify timestamps are updated
				if user.CreatedAt.IsZero() {
					t.Error("CreatedAt should not be zero after BeforeCreate")
				}

				if user.UpdatedAt.IsZero() {
					t.Error("UpdatedAt should not be zero after BeforeCreate")
				}
			}
		})
	}
}

func TestUserBeforeUpdate(t *testing.T) {
	for _, tt := range UserBeforeUpdateTests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser(tt.name, tt.email)
			originalUpdatedAt := user.UpdatedAt

			// Sleep to ensure time difference
			time.Sleep(time.Millisecond)

			err := user.BeforeUpdate()

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tt.shouldError {
				// Verify UpdatedAt is updated
				if user.UpdatedAt.Equal(originalUpdatedAt) {
					t.Error("UpdatedAt should be updated after BeforeUpdate")
				}
			}
		})
	}
}

func TestUserValidation(t *testing.T) {
	for _, tt := range UserValidationTests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser(tt.name, tt.email)
			user.Phone = tt.phone
			user.Status = tt.status

			err := user.Validate()

			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestUserInterfaceCompliance(t *testing.T) {
	user := NewUser("Test User", "test@example.com")

	// Test ModelInterface compliance
	if user.GetID() != user.ID {
		t.Errorf("GetID() returned %s, expected %s", user.GetID(), user.ID)
	}

	if user.GetCreatedAt() != user.CreatedAt {
		t.Errorf("GetCreatedAt() returned %v, expected %v", user.GetCreatedAt(), user.CreatedAt)
	}

	if user.GetUpdatedAt() != user.UpdatedAt {
		t.Errorf("GetUpdatedAt() returned %v, expected %v", user.GetUpdatedAt(), user.UpdatedAt)
	}

	// Test setters
	newID := "USER123456789"
	user.SetID(newID)
	if user.GetID() != newID {
		t.Errorf("SetID() failed, got %s, expected %s", user.GetID(), newID)
	}

	newTime := time.Now().Add(time.Hour)
	user.SetCreatedAt(newTime)
	if user.GetCreatedAt() != newTime {
		t.Errorf("SetCreatedAt() failed, got %v, expected %v", user.GetCreatedAt(), newTime)
	}

	user.SetUpdatedAt(newTime)
	if user.GetUpdatedAt() != newTime {
		t.Errorf("SetUpdatedAt() failed, got %v, expected %v", user.GetUpdatedAt(), newTime)
	}
}
