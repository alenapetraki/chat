package service

import (
	"context"

	"github.com/alenapetraki/chat/auth"
	"github.com/alenapetraki/chat/chats"
	"github.com/alenapetraki/chat/models/entities"
	"github.com/alenapetraki/chat/util/id"
	"github.com/pkg/errors" //todo: deprecated
)

type service struct {
	storage chats.Storage
}

func New(storage chats.Storage) *service {
	return &service{storage: storage}
}

func (s *service) CreateChat(ctx context.Context, chat *entities.Chat) (*entities.Chat, error) {
	if chat.Type == entities.GroupType || chat.Type == entities.ChannelType {
		if chat.Name == "" {
			return nil, errors.New("name required")
		}
	}

	if chat.Type == entities.DialogType {
		chat.Name = ""
		chat.Description = ""
		chat.AvatarURL = ""
	}

	var err error
	chat.ID, err = id.NewULID()
	if err != nil {
		return nil, err
	}

	stTx, err := s.storage.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	if err := stTx.EndTx(func() error {
		if err := stTx.CreateChat(ctx, chat); err != nil {
			return errors.Wrap(err, "failed to create chat")
		}
		if err := stTx.SetMember(ctx, chat.ID, auth.GetUserID(ctx), entities.RoleOwner); err != nil {
			return errors.Wrap(err, "failed to set membership")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return chat, nil
}

func (s *service) GetChat(ctx context.Context, chatID string) (*entities.Chat, error) {
	if chatID == "" {
		return nil, errors.New("chat identifier required")
	}

	chat, err := s.storage.GetChat(ctx, chatID)
	if err != nil {
		return nil, errors.New("get chat")
	}
	return chat, nil
}

func (s *service) DeleteChat(ctx context.Context, chatID string) error {
	if chatID == "" {
		return errors.New("chat identifier required")
	}
	if err := s.storage.DeleteChat(ctx, chatID); err != nil {
		return err
	}

	return s.storage.DeleteMembers(ctx, chatID)
}

func (s *service) UpdateChat(ctx context.Context, chat *entities.Chat) error {
	if chat == nil || chat.ID == "" {
		return errors.New("id required")
	}
	return s.storage.UpdateChat(ctx, chat)
}

func (s *service) SetMember(ctx context.Context, chatID, userID, role string) error {
	if userID == "" || chatID == "" {
		return errors.New("chat and user ids required")
	}

	// - получить тип чата
	//   - диалог: можно добавить только еще одного владельца

	chat, err := s.storage.GetChat(ctx, chatID)
	if err != nil {
		return err
	}

	if chat.Type == entities.DialogType {
		members, err := s.storage.FindChatMembers(ctx, chatID, nil)
		if err != nil {
			return errors.Wrap(err, "check members")
		}

		isMember := false
		//isMemberf := func() bool {
		for _, m := range members {
			if m.UserID == userID {
				isMember = true
				break
			}
		}
		//}()

		if len(members) == 2 && !isMember {
			return errors.New("max two members for dialog are allowed")
		}
		role = entities.RoleOwner
	}

	// todo: ограничения по вместимости чата - запрос на Count к БД?

	return s.storage.SetMember(ctx, chatID, userID, role)
}

func (s *service) DeleteMember(ctx context.Context, chatID, userID string) error {
	if userID == "" || chatID == "" {
		return errors.New("chat and user ids required")
	}

	role, err := s.storage.GetRole(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if role == entities.RoleOwner {
		return errors.New("cannot delete owner")
	}

	return s.storage.DeleteMembers(ctx, chatID, userID)
}

func (s *service) GetRole(ctx context.Context, chatID, userID string) (string, error) {
	if userID == "" || chatID == "" {
		return "", errors.New("chat and user ids required")
	}
	return s.storage.GetRole(ctx, chatID, userID)
}

// todo: default sort
//func (s *service) FindChatMembers(ctx context.Context, chatID string, options *commons.PaginationOptions) ([]*ChatMember, error) {
//	return s.storage.FindChatMembers(ctx, chatID, options)
//}
