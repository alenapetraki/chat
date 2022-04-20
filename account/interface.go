package account

import (
	"context"

	"github.com/alenapetraki/chat/models/entities"
)

// todo: WIP. непонятно пока, как будет работать

type Service interface {
	CreateUser(ctx context.Context, user *entities.User) (*entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error
	GetUser(ctx context.Context, userID string) (*entities.User, error)
	DeleteUser(ctx context.Context, userID string) error

	//Authenticate(ctx context.Context, creds *Credentials) (string, error)
	//Refresh(token string) (string, error)

	//FindUsers(ctx context.Context, filter *FindUsersFilter) ([]*User, error)
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
