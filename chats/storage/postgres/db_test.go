package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"testing"

	"github.com/alenapetraki/chat/chats"
	"github.com/alenapetraki/chat/commons"
	"github.com/alenapetraki/chat/commons/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cfg = postgres.Config{
	Host:     "localhost",
	Port:     "5435",
	User:     "chat_user",
	Password: "chat_password",
	Database: "chat",
}

func TestChats_Create(t *testing.T) {
	st := New(&Config{cfg}).(*storage)

	require.NoError(t, st.Connect())
	defer st.Close()
	cleanUp(t, st.db)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		{
			chat := &chats.Chat{
				ID:          commons.GenerateID(),
				Type:        chats.ChannelType,
				Name:        "channel one",
				Description: "just a channel",
				AvatarURL:   "https://test.some",
			}
			require.NoError(t, st.CreateChat(ctx, chat))
		}
		{
			chat := &chats.Chat{
				ID:          commons.GenerateID(),
				Type:        chats.GroupType,
				Name:        "group one",
				Description: "just a group",
				AvatarURL:   "https://test.some",
			}
			require.NoError(t, st.CreateChat(ctx, chat))
		}
	})

}

func TestChats_Read(t *testing.T) {
	st := New(&Config{cfg}).(*storage)

	require.NoError(t, st.Connect())
	defer st.Close()
	cleanUp(t, st.db)

	ctx := context.Background()

	chatIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		chat := &chats.Chat{
			ID:          commons.GenerateID(),
			Type:        chats.ChatType((i+1)%3 + 1),
			Name:        "chat " + strconv.Itoa(i),
			Description: "just a chat",
			AvatarURL:   "https://test.some",
		}
		require.NoError(t, st.CreateChat(ctx, chat))
		chatIDs[i] = chat.ID
	}

	t.Run("get", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			chat, err := st.GetChat(ctx, chatIDs[2])
			require.NoError(t, err)
			assert.Equal(t, chatIDs[2], chat.ID)
			assert.Equal(t, chats.DialogType, chat.Type)
			assert.Equal(t, "chat 2", chat.Name)
			assert.Equal(t, "just a chat", chat.Description)
			assert.Equal(t, "https://test.some", chat.AvatarURL)
		})
		t.Run("not found", func(t *testing.T) {
			usr, err := st.GetChat(ctx, "not_exist")
			require.Error(t, err)
			assert.Nil(t, usr)
		})
	})

}

func TestMembers(t *testing.T) {
	st := New(&Config{cfg}).(*storage)

	require.NoError(t, st.Connect())
	defer st.Close()
	cleanUp(t, st.db)

	ctx := context.Background()

	userIDs := []string{"user_1", "user_2", "user_3"}
	chatIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		chat := &chats.Chat{
			ID:   commons.GenerateID(),
			Type: chats.GroupType,
			Name: "chat " + strconv.Itoa(i),
		}
		require.NoError(t, st.CreateChat(ctx, chat))
		chatIDs[i] = chat.ID

		for j := 0; j < i+1; j++ {
			role := chats.MemberRole
			if j == 0 {
				role = chats.OwnerRole
			}
			require.NoError(t, st.AddMember(ctx, chat.ID, userIDs[j], role))
		}
	}

	t.Run("get role", func(t *testing.T) {
		r, err := st.GetRole(ctx, chatIDs[0], userIDs[0])
		require.NoError(t, err)
		assert.Equal(t, chats.OwnerRole, r)
	})

	t.Run("delete members", func(t *testing.T) {

		ms, err := st.FindChatMembers(ctx, chatIDs[2], nil)
		require.NoError(t, err)
		assert.Len(t, ms, 3)

		require.NoError(t, st.DeleteMembers(ctx, chatIDs[2]))

		ms, err = st.FindChatMembers(ctx, chatIDs[2], nil)
		require.NoError(t, err)
		assert.Len(t, ms, 0, "Должны быть удалены все участники чата")
	})
}

func cleanUp(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`delete from "user"`)
	require.NoError(t, err)
	_, err = db.Exec(`delete from "member"`)
	require.NoError(t, err)
	_, err = db.Exec(`delete from "chat"`)
	require.NoError(t, err)
}
