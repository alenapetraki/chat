package service

import (
	"context"

	"github.com/alenapetraki/chat/auth"
	"github.com/alenapetraki/chat/entities"
	"github.com/alenapetraki/chat/services/chats"
	"github.com/alenapetraki/chat/storage"
	chatsstorage "github.com/alenapetraki/chat/storage/chats"
	"github.com/alenapetraki/chat/util/id"
	"github.com/pkg/errors" //todo: deprecated. choose another package
)

type service struct {
	storage chats.Storage
}

func New(storage chats.Storage) *service {
	return &service{storage: storage}
}

func (s *service) CreateChat(ctx context.Context, chat *entities.Chat) (*entities.Chat, error) {
	const op = "ChatService.CreateChat"

	if chat.Type == entities.GroupType || chat.Type == entities.ChannelType {
		if chat.Name == "" {
			return nil, errors.Wrap(errors.New("name required"), op)
		}
	}

	if chat.Type == entities.DialogType {
		chat.Name = ""
		chat.Description = ""
		chat.AvatarURL = ""
	}

	if chat.Type != entities.DialogType && chat.Type != entities.GroupType && chat.Type != entities.ChannelType {
		return nil, errors.Wrap(errors.New("unknown type"), op)
	}

	var err error
	chat.ID, err = id.NewULID()
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if err := s.storage.RunTx(func(tx *storage.Transaction) error {

		st := chatsstorage.New(tx)

		if err := st.CreateChat(ctx, chat); err != nil {
			return errors.Wrap(err, op)
		}
		if err := st.SetMember(ctx, chat.ID, auth.GetUserID(ctx), entities.RoleOwner); err != nil {
			return errors.Wrap(err, op)
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, op)
	}

	return chat, nil
}

func (s *service) GetChat(ctx context.Context, chatID string) (*entities.Chat, error) {
	const op = "ChatService.GetChat"

	chat, err := s.storage.GetChat(ctx, chatID)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	return chat, nil
}

func (s *service) DeleteChat(ctx context.Context, chatID string) error {
	const op = "ChatService.DeleteChat"
	if _, err := s.storage.DeleteMembers(ctx, chatID); err != nil {
		return errors.Wrap(err, op)
	}
	if err := s.storage.DeleteChat(ctx, chatID); err != nil {
		return errors.Wrap(err, op)
	}
	return nil
}

func (s *service) UpdateChat(ctx context.Context, chat *entities.Chat) error {
	const op = "ChatService.UpdateChat"
	return errors.Wrap(s.storage.UpdateChat(ctx, chat), op)
}

func (s *service) SetMember(ctx context.Context, chatID, userID string, role entities.Role) error {
	const op = "ChatService.SetMember"

	// - получить тип чата
	//   - диалог: можно добавить только еще одного владельца

	chat, err := s.storage.GetChat(ctx, chatID)
	if err != nil {
		return errors.Wrap(err, op)
	}

	if chat.Type == entities.DialogType {

		role, err = s.storage.GetRole(ctx, chatID, userID)
		if err != nil && !errors.Is(err, chats.ErrNotFound) {
			return err
		}

		if chat.NumMembers == 2 && role == "" {
			return errors.Wrap(chats.ErrMaxMembersNumExceeded, op)
		}

		role = entities.RoleOwner
	}

	if chat.Type == entities.GroupType && chat.NumMembers >= chats.MaxGroupMembersAllowed-1 {
		return errors.Wrap(chats.ErrMaxMembersNumExceeded, op)
	}

	return s.storage.SetMember(ctx, chatID, userID, role)
}

func (s *service) DeleteMember(ctx context.Context, chatID, userID string) error {
	const op = "ChatService.DeleteMember"

	role, err := s.storage.GetRole(ctx, chatID, userID)
	if err != nil {
		return errors.Wrap(err, op)
	}
	if role == entities.RoleOwner {
		// найти других владельцев, нельзя удалить только если владелец один
		return errors.Wrap(errors.New("cannot delete owner"), op)
	}

	_, err = s.storage.DeleteMembers(ctx, chatID, userID)
	return errors.Wrap(err, op)
}

func (s *service) GetRole(ctx context.Context, chatID, userID string) (entities.Role, error) {
	const op = "ChatService.GetRole"
	role, err := s.storage.GetRole(ctx, chatID, userID)
	if err != nil {
		return "", errors.Wrap(err, op)
	}
	return role, nil
}

//func (s *service) FindChatMembers(ctx context.Context, chatID string, options *commons.PaginationOptions) ([]*ChatMember, error) {
//	return s.storage.FindChatMembers(ctx, chatID, options)
//}
