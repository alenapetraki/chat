package users

import (
	"context"

	"github.com/alenapetraki/chat"
)

type Service interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, userID string) (*User, error)
	FindUsers(ctx context.Context, filter *FindUsersFilter) ([]*User, error)
	DeleteUser(ctx context.Context, userID string) error
}

type User struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	FullName string `json:"full_name,omitempty"`
	Status   string `json:"status,omitempty"`
}

type FindUsersFilter struct {
	chat.PaginationOptions
	chat.SortOptions
	Username string `json:"username,omitempty"`
}
