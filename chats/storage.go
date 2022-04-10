package chats

import (
	"context"

	"github.com/alenapetraki/chat/commons"
)

type Storage interface {
	Connect() error
	Close() error

	/*
		BeginTr()
		CommitTR()
	*/

	CreateChat(ctx context.Context, chat *Chat) error
	UpdateChat(ctx context.Context, chat *Chat) error
	GetChat(ctx context.Context, chatID string) (*Chat, error)
	DeleteChat(ctx context.Context, chatID string, force ...bool) error
	//FindChats(ctx context.Context, filter *FindChatsFilter, option *chat.PaginationOptions) ([]*chat.Chat, int, error)

	AddMember(ctx context.Context, chatID, userID string, role Role) error
	DeleteMembers(ctx context.Context, chatID string, userID ...string) error
	GetRole(ctx context.Context, chatID, userID string) (Role, error)

	// FindChatMembers возникает зависимость от хранилища другого вообще сервиса. Возвращать просто userID?
	FindChatMembers(ctx context.Context, chatID string, options *commons.PaginationOptions) ([]*ChatMember, error)
}
