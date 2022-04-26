package chats

import (
	"context"

	"github.com/alenapetraki/chat/models/entities"
	"github.com/alenapetraki/chat/util"
)

type Chats interface {
	CreateChat(ctx context.Context, chat *entities.Chat) (*entities.Chat, error)
	UpdateChat(ctx context.Context, chat *entities.Chat) error
	GetChat(ctx context.Context, chatID string) (*entities.Chat, error)
	DeleteChat(ctx context.Context, chatID string) error

	SetMember(ctx context.Context, chatID, userID, role string) error
	DeleteMember(ctx context.Context, chatID, userID string) error
	GetRole(ctx context.Context, chatID, userID string) (string, error)
	//FindChatMembers(ctx context.Context, chatID string, options *commons.PaginationOptions) ([]*ChatMember, error)
}

type Storage interface {
	Tx

	CreateChat(ctx context.Context, chat *entities.Chat) error
	UpdateChat(ctx context.Context, chat *entities.Chat) error
	GetChat(ctx context.Context, chatID string) (*entities.Chat, error)
	DeleteChat(ctx context.Context, chatID string, force ...bool) error
	//FindChats(ctx context.Context, filter *FindChatsFilter, option *chats.Chat.PaginationOptions) ([]*chats.Chat.Chat, int, error)

	SetMember(ctx context.Context, chatID, userID string, role string) error
	DeleteMembers(ctx context.Context, chatID string, userID ...string) error
	GetRole(ctx context.Context, chatID, userID string) (string, error)
	FindChatMembers(ctx context.Context, chatID string, options *util.PaginationOptions) ([]*entities.ChatMember, error)
}

type Tx interface {
	BeginTx(ctx context.Context) (Storage, error)
	EndTx(func() error) error
}

//type StorageTx interface {
//	Storage
//	EndTx(func() error) error
//}