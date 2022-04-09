package users

import (
	"context"
)

type Users interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, userID string) error
	//FindUsers(ctx context.Context, filter *FindUsersFilter) ([]*User, error)
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Status   string `json:"status"`
	//IsVerified bool   `json:"isverified" sql:"isverified"`
}

//type FindUsersFilter struct {
//	chat.PaginationOptions
//	chat.SortOptions
//	Username string `json:"username,omitempty"`
//}

type Authentication interface {
	Authenticate(ctx context.Context, creds *Credentials) (string, error)
	Refresh(token string) (string, error)
}

type Credentials struct {
	Username string `json:"username" sql:"username"`
	Password string `json:"password" sql:"password"`
	//Email     string `json:"email" sql:"email"`
	//TokenHash string `json:"token_hash" sql:"token_hash"`
}
