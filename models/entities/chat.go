package entities

type Chat struct {
	ID          string
	Type        string
	Name        string
	Description string
	AvatarURL   string
}

const (
	DialogType  = "dialog"
	GroupType   = "group"
	ChannelType = "channel"

	RoleMember = "member"
	RoleOwner  = "owner"
)

type ChatMember struct {
	//Chat *Chat
	//User *account.User
	UserID string
	Role   string
}
