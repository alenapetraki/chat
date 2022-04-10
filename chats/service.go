package chats

import (
	"context"

	"github.com/alenapetraki/chat/account"
)

type Chats interface {
	//CreateDialog(ctx context.Context, userID string) (*Chat, error)
	//DeleteDialog(ctx context.Context, dialogID string) error
	//
	//CreateGroup(ctx context.Context, group *Chat) (*Chat, error)
	//UpdateGroup(ctx context.Context, group *Chat) error
	//GetGroup(ctx context.Context, groupID string) (*Chat, error)
	//DeleteGroup(ctx context.Context, groupID string) error
	//
	//JoinGroup(ctx context.Context, groupID string) error
	//LeaveGroup(ctx context.Context, groupID string) error
	//ListGroupMembers(ctx context.Context, groupID string, options *chat.PaginationOptions) ([]*ChatMember, error)
	//
	//CreateChannel(ctx context.Context, channel *Chat) (*Chat, error)
	//UpdateChannel(ctx context.Context, channel *Chat) error
	//SubscribeChannel(ctx context.Context, channelID string) error
	//UnsubscribeChannel(ctx context.Context, channelID string) error
	//GetChannel(ctx context.Context, channelID string) (*Chat, error)

	CreateChat(ctx context.Context, chat *Chat) (*Chat, error)
	UpdateChat(ctx context.Context, chat *Chat) error
	GetChat(ctx context.Context, chatID string) (*Chat, error)
	DeleteChat(ctx context.Context, chatID string) error

	AddMember(ctx context.Context, chatID, userID string) error
	DeleteMember(ctx context.Context, chatID, userID string) error
	GetRole(ctx context.Context, chatID, userID string) (Role, error)
	//FindChatMembers(ctx context.Context, chatID string, options *commons.PaginationOptions) ([]*ChatMember, error)
}

type Chat struct {
	ID   string
	Type ChatType
	Name string
	//Title       string
	Description string
	AvatarURL   string
}

type ChatType uint8

const (
	DialogType ChatType = iota + 1
	GroupType
	ChannelType
)

type Role uint8

//func (r *Role) Scan(src any) error {
//	switch v := src.(type) {
//	case *int64:
//		fmt.Println(v, r)
//	default:
//		fmt.Println(v, r)
//
//	}
//	return nil
//}

const (
	MemberRole Role = iota + 1
	OwnerRole
)

type ChatMember struct {
	Chat *Chat
	User *account.User
	Role Role
}
