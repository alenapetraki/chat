package entities

type Chat struct {
	ID          string
	Type        ChatType
	Name        string
	Description string
	AvatarURL   string
}

type ChatType string

const (
	DialogType  ChatType = "dialog"
	GroupType   ChatType = "group"
	ChannelType ChatType = "channel"
)

type Role string

const (
	RoleMember Role = "member"
	RoleOwner  Role = "owner"
)

type ChatMember struct {
	//Chat *Chat
	//User *account.User
	UserID string
	Role   string
}
