package service

import (
	"context"

	"github.com/alenapetraki/chat/auth"
	"github.com/alenapetraki/chat/models/entities"
	"github.com/alenapetraki/chat/services/chats"
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

	if chat.Type != entities.DialogType && chat.Type != entities.GroupType && chat.Type != entities.ChannelType {
		return nil, errors.New("unknown type")
	}

	var err error
	chat.ID, err = id.NewULID()
	if err != nil {
		return nil, err
	}

	if err := s.storage.RunTx(ctx, func(st chats.Storage) error {
		if err := st.CreateChat(ctx, chat); err != nil {
			return errors.Wrap(err, "failed to create chat")
		}
		if err := st.SetMember(ctx, chat.ID, auth.GetUserID(ctx), entities.RoleOwner); err != nil {
			return errors.Wrap(err, "failed to set membership")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return chat, nil
}

func (s *service) GetChat(ctx context.Context, chatID string) (*entities.Chat, error) {
	chat, err := s.storage.GetChat(ctx, chatID)
	if err != nil {
		return nil, errors.New("get chat")
	}
	return chat, nil
}

func (s *service) DeleteChat(ctx context.Context, chatID string) error {
	if err := s.storage.DeleteChat(ctx, chatID); err != nil {
		return err
	}
	_, err := s.storage.DeleteMembers(ctx, chatID)
	return err
}

func (s *service) UpdateChat(ctx context.Context, chat *entities.Chat) error {
	return s.storage.UpdateChat(ctx, chat)
}

func (s *service) SetMember(ctx context.Context, chatID, userID string, role entities.Role) error {
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

		role, err = s.storage.GetRole(ctx, chatID, userID)
		if err != nil && !errors.Is(err, chats.ErrNotFound) {
			return err
		}

		if chat.NumMembers == 2 && role == "" {
			return chats.ErrMaxMembersNumExceeded
		}

		role = entities.RoleOwner
	}

	// непревышение количества участников нужно контролировать в транзакции, причем с уровнями изоляции разбираться?
	// предлагаю оставить такое решение - да, могут возникнуть ситуации, когда 1002 участника. Да и ок, нет?
	if chat.Type == entities.GroupType && chat.NumMembers >= chats.MaxGroupMembersAllowed-1 {
		return chats.ErrMaxMembersNumExceeded
	}

	return s.storage.SetMember(ctx, chatID, userID, role)
}

func (s *service) DeleteMember(ctx context.Context, chatID, userID string) error {

	role, err := s.storage.GetRole(ctx, chatID, userID)
	if err != nil {
		return err
	}
	if role == entities.RoleOwner {
		// найти других владельцев, нельзя удалить только если владелец один
		return errors.New("cannot delete owner")
	}

	_, err = s.storage.DeleteMembers(ctx, chatID, userID)
	return err
}

func (s *service) GetRole(ctx context.Context, chatID, userID string) (entities.Role, error) {
	return s.storage.GetRole(ctx, chatID, userID)
}

//func (s *service) FindChatMembers(ctx context.Context, chatID string, options *commons.PaginationOptions) ([]*ChatMember, error) {
//	return s.storage.FindChatMembers(ctx, chatID, options)
//}
