package service

import (
	"context"

	"github.com/alenapetraki/chat/internal/util"
	"github.com/alenapetraki/chat/users"
	"github.com/cockroachdb/errors"
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	storage users.Storage
}

func New(storage users.Storage) users.Users {
	return &service{storage: storage}
}

func (s *service) CreateUser(ctx context.Context, user *users.User) (*users.User, error) {

	if user == nil || user.Username == "" || user.Password == "" {
		return nil, errors.New("username and password required")
	}

	user.ID = util.GenerateID()

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(hashedPass)

	if err := s.storage.CreateUser(ctx, user); err != nil {
		return nil, errors.Wrap(err, "create user")
	}

	return user, nil
}

func (s *service) UpdateUser(ctx context.Context, user *users.User) error {
	if user == nil || user.ID == "" {
		return errors.New("user identifier required")
	}

	// supposed that username cannot be changed
	return s.storage.UpdateUser(ctx, user)
}

func (s *service) GetUser(ctx context.Context, userID string) (*users.User, error) {
	if userID == "" {
		return nil, errors.New("user identifier required")
	}
	return s.storage.GetUser(ctx, userID)
}

//func (s *service) FindUsers(ctx context.Context, filter *account.FindUsersFilter) ([]*account.User, error) {
//	return s.storage.FindUsers(ctx, filter)
//}

func (s *service) DeleteUser(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("user identifier required")
	}
	return s.storage.DeleteUser(ctx, userID)
}
