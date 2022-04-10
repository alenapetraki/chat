package service

import (
	"context"

	"github.com/alenapetraki/chat/auth"
	"github.com/alenapetraki/chat/chats"
	"github.com/alenapetraki/chat/commons"
	"github.com/pkg/errors" //todo: deprecated
)

type service struct {
	storage chats.Storage
}

func New(storage chats.Storage) chats.Chats {
	return &service{storage: storage}
}

func (s *service) CreateChat(ctx context.Context, chat *chats.Chat) (*chats.Chat, error) {
	if chat == nil {
		return nil, errors.New("type required")
	}
	if chat.Type == chats.GroupType || chat.Type == chats.ChannelType {
		if chat.Name == "" {
			return nil, errors.New("name required")
		}
		// todo: validate or normalize sting
	}

	if chat.Type == chats.DialogType {
		chat.Name = ""
		chat.Description = ""
		chat.AvatarURL = ""
	}

	chat.ID = commons.GenerateID()
	if err := s.storage.CreateChat(ctx, chat); err != nil {
		return nil, errors.Wrap(err, "failed to create chat")
	}

	if err := s.storage.AddMember(ctx, chat.ID, auth.GetUserID(ctx), chats.OwnerRole); err != nil {
		return nil, errors.Wrap(err, "failed to set membership")
	}

	return chat, nil
}

func (s *service) GetChat(ctx context.Context, chatID string) (*chats.Chat, error) {
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

func (s *service) UpdateChat(ctx context.Context, chat *chats.Chat) error {
	if chat == nil || chat.ID == "" {
		return errors.New("id required")
	}
	return s.storage.UpdateChat(ctx, chat)
}

func (s *service) AddMember(ctx context.Context, chatID, userID string) error {
	if userID == "" || chatID == "" {
		return errors.New("chat and user ids required")
	}

	// - проверить, что пользователь еще не в чате
	// - получить тип чата
	//   - диалог: можно добавить только еще одного владельца
	//   - группа или канал: добавляем простых членов

	if _, err := s.storage.GetRole(ctx, userID, chatID); err == nil {
		return errors.New("user already a member")
	}

	chat, err := s.storage.GetChat(ctx, chatID)
	if err != nil {
		return err
	}

	var role chats.Role
	switch chat.Type {
	case chats.DialogType:
		role = chats.OwnerRole
	case chats.GroupType, chats.ChannelType:
		role = chats.MemberRole
	}

	// todo: ограничения по вместимости чата

	return s.storage.AddMember(ctx, chatID, userID, role)
}

func (s *service) DeleteMember(ctx context.Context, chatID, userID string) error {
	if userID == "" || chatID == "" {
		return errors.New("chat and user ids required")
	}

	role, err := s.storage.GetRole(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if role == chats.OwnerRole {
		return errors.New("cannot delete owner")
	}

	return s.storage.DeleteMembers(ctx, chatID, userID)
}

func (s *service) GetRole(ctx context.Context, chatID, userID string) (chats.Role, error) {
	if userID == "" || chatID == "" {
		return 0, errors.New("chat and user ids required")
	}
	return s.storage.GetRole(ctx, chatID, userID)
}

// todo: default sort
//func (s *service) FindChatMembers(ctx context.Context, chatID string, options *commons.PaginationOptions) ([]*chats.ChatMember, error) {
//	return s.storage.FindChatMembers(ctx, chatID, options)
//}
