package users

import "context"

type Storage interface {
	Connect() error
	Close() error

	/*
		BeginTr()
		CommitTR()
	*/

	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, userID string) (*User, error)
	//FindUsers(ctx context.Context, filter *FindUsersFilter) ([]*User, error)
	DeleteUser(ctx context.Context, userID string) error
}
