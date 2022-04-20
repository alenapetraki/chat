package storage

import (
	"context"
	"database/sql"
	"strconv"
	"testing"

	"github.com/alenapetraki/chat/models/entities"
	"github.com/alenapetraki/chat/util/id"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cfg = &Config{
	Host:     "localhost",
	Port:     "5435",
	User:     "chat_user",
	Password: "chat_password",
	Database: "chat",
}

func TestChats_Create(t *testing.T) {

	db, err := Connect("postgres", cfg)
	require.NoError(t, err)
	defer db.Close()
	cleanUp(t, db)

	st := New(db)

	ctx := context.Background()

	{
		chat := &entities.Chat{
			ID:          id.MustNewULID(),
			Type:        entities.ChannelType,
			Name:        "channel one",
			Description: "just a channel",
			AvatarURL:   "https://test.some",
		}
		require.NoError(t, st.CreateChat(ctx, chat))
	}
	{
		chat := &entities.Chat{
			ID:          id.MustNewULID(),
			Type:        entities.GroupType,
			Name:        "group one",
			Description: "just a group",
			AvatarURL:   "https://test.some",
		}
		require.NoError(t, st.CreateChat(ctx, chat))
	}

}

func TestChats_Read(t *testing.T) {

	db, err := Connect("postgres", cfg)
	require.NoError(t, err)
	cleanUp(t, db)

	st := New(db)

	ctx := context.Background()

	chatIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		chat := &entities.Chat{
			ID:          id.MustNewULID(),
			Type:        entities.DialogType,
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
			assert.Equal(t, entities.DialogType, chat.Type)
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

	db, err := Connect("postgres", cfg)
	require.NoError(t, err)
	cleanUp(t, db)

	st := New(db)

	ctx := context.Background()

	userIDs := []string{"user_1", "user_2", "user_3"}
	chatIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		chat := &entities.Chat{
			ID:   id.MustNewULID(),
			Type: entities.GroupType,
			Name: "chat " + strconv.Itoa(i),
		}
		require.NoError(t, st.CreateChat(ctx, chat))
		chatIDs[i] = chat.ID

		for j := 0; j < i+1; j++ {
			role := entities.RoleMember
			if j == 0 {
				role = entities.RoleOwner
			}
			require.NoError(t, st.SetMember(ctx, chat.ID, userIDs[j], role))
		}
	}

	t.Run("get role", func(t *testing.T) {
		r, err := st.GetRole(ctx, chatIDs[0], userIDs[0])
		require.NoError(t, err)
		assert.Equal(t, entities.RoleOwner, r)
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

func TestTX(t *testing.T) {

	db, err := Connect("postgres", cfg)
	require.NoError(t, err)
	cleanUp(t, db)

	st := New(db)

	ctx := context.Background()

	tx, err := st.BeginTx(ctx)
	require.NoError(t, err)

	err = tx.EndTx(func() error {
		chat := &entities.Chat{
			ID:   id.MustNewULID(),
			Type: entities.DialogType,
			Name: "chat foo",
		}
		err := tx.CreateChat(ctx, chat)
		if err != nil {
			return err
		}
		err = tx.SetMember(ctx, chat.ID, "user42", entities.RoleOwner)
		if err != nil {
			return err
		}
		return nil
	})
	require.NoError(t, err)

}

func cleanUp(t *testing.T, db *sql.DB) {
	db.Exec(`truncate user`)
	db.Exec(`truncate member`)
	db.Exec(`truncate chat`)
}
