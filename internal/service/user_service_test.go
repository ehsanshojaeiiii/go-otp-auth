package service

import (
	"testing"

	"github.com/ehsanshojaei/go-otp-auth/internal/model"
)

func createTestUserService() (UserService, *mockUserRepository) {
	userRepo := newMockUserRepository()
	userService := NewUserService(userRepo)
	return userService, userRepo
}

func TestUserService_GetUserByID(t *testing.T) {
	userService, userRepo := createTestUserService()

	// Create test user
	testUser := &model.User{
		PhoneNumber: "+1234567890",
	}
	userRepo.Create(testUser)

	tests := []struct {
		name    string
		userID  uint
		wantErr bool
		wantUser bool
	}{
		{
			name:    "Existing user",
			userID:  testUser.ID,
			wantErr: false,
			wantUser: true,
		},
		{
			name:    "Non-existing user",
			userID:  999,
			wantErr: true,
			wantUser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := userService.GetUserByID(tt.userID)
			
			if tt.wantErr {
				if err == nil {
					t.Error("GetUserByID() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GetUserByID() unexpected error = %v", err)
				return
			}

			if tt.wantUser {
				if user == nil {
					t.Error("GetUserByID() returned nil user")
					return
				}
				if user.ID != tt.userID {
					t.Errorf("GetUserByID() user ID = %v, want %v", user.ID, tt.userID)
				}
				if user.PhoneNumber != testUser.PhoneNumber {
					t.Errorf("GetUserByID() phone number = %v, want %v", user.PhoneNumber, testUser.PhoneNumber)
				}
			}
		})
	}
}

func TestUserService_GetUsers(t *testing.T) {
	userService, userRepo := createTestUserService()

	// Create test users
	users := []*model.User{
		{PhoneNumber: "+1234567890"},
		{PhoneNumber: "+1234567891"},
		{PhoneNumber: "+9876543210"},
	}

	for _, user := range users {
		userRepo.Create(user)
	}

	tests := []struct {
		name     string
		request  *model.GetUsersRequest
		wantErr  bool
		wantCount int
	}{
		{
			name: "Default pagination",
			request: &model.GetUsersRequest{
				Page:     1,
				PageSize: 10,
			},
			wantErr:  false,
			wantCount: 3,
		},
		{
			name: "Search by phone number",
			request: &model.GetUsersRequest{
				Page:        1,
				PageSize:    10,
				PhoneNumber: "+123456789",
			},
			wantErr:  false,
			wantCount: 2, // Should match first two users
		},
		{
			name: "Exact phone search",
			request: &model.GetUsersRequest{
				Page:        1,
				PageSize:    10,
				PhoneNumber: "+9876543210",
			},
			wantErr:  false,
			wantCount: 1,
		},
		{
			name: "No matches",
			request: &model.GetUsersRequest{
				Page:        1,
				PageSize:    10,
				PhoneNumber: "+5555555555",
			},
			wantErr:  false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.request.SetDefaults()
			
			result, err := userService.GetUsers(tt.request)
			
			if tt.wantErr {
				if err == nil {
					t.Error("GetUsers() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("GetUsers() unexpected error = %v", err)
				return
			}

			if result == nil {
				t.Error("GetUsers() returned nil result")
				return
			}

			if len(result.Users) != tt.wantCount {
				t.Errorf("GetUsers() user count = %v, want %v", len(result.Users), tt.wantCount)
			}

			if result.Total != int64(tt.wantCount) {
				t.Errorf("GetUsers() total = %v, want %v", result.Total, tt.wantCount)
			}

			if result.Page != tt.request.Page {
				t.Errorf("GetUsers() page = %v, want %v", result.Page, tt.request.Page)
			}

			if result.PageSize != tt.request.PageSize {
				t.Errorf("GetUsers() page size = %v, want %v", result.PageSize, tt.request.PageSize)
			}
		})
	}
}

func TestGetUsersRequest_SetDefaults(t *testing.T) {
	tests := []struct {
		name     string
		request  *model.GetUsersRequest
		wantPage int
		wantSize int
	}{
		{
			name:     "Zero values",
			request:  &model.GetUsersRequest{},
			wantPage: 1,
			wantSize: 10,
		},
		{
			name: "Custom values",
			request: &model.GetUsersRequest{
				Page:     2,
				PageSize: 20,
			},
			wantPage: 2,
			wantSize: 20,
		},
		{
			name: "Zero page only",
			request: &model.GetUsersRequest{
				Page:     0,
				PageSize: 5,
			},
			wantPage: 1,
			wantSize: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.request.SetDefaults()
			
			if tt.request.Page != tt.wantPage {
				t.Errorf("SetDefaults() page = %v, want %v", tt.request.Page, tt.wantPage)
			}
			
			if tt.request.PageSize != tt.wantSize {
				t.Errorf("SetDefaults() page size = %v, want %v", tt.request.PageSize, tt.wantSize)
			}
		})
	}
}
