package chats

import (
	"context"
	"database/sql"
	"strconv"
	"testing"

	"github.com/alenapetraki/chat/models/entities"
	"github.com/alenapetraki/chat/services/chats"
	"github.com/alenapetraki/chat/storage"
	"github.com/alenapetraki/chat/util/id"
	"github.com/stretchr/testify/suite"
)

type testSuite struct {
	suite.Suite
	db *sql.DB
	st *Storage
}

func TestStorage(t *testing.T) {
	db, err := storage.Connect("postgres", &storage.Config{
		Host:     "localhost",
		Port:     "5435",
		User:     "chat_user",
		Password: "chat_password",
		Database: "chat",
	})
	if err != nil {
		panic(err)
	}
	suite.Run(t, &testSuite{db: db})
	db.Close()
}

func (t *testSuite) SetupTest() {

	t.st = New(t.db)

	t.db.Exec(`truncate user`)
	t.db.Exec(`truncate member`)
	t.db.Exec(`truncate chat`)
}

func (t *testSuite) TearDownTest() {
}

func (t *testSuite) TestChats_Create() {

	ctx := context.Background()

	{
		chat := &entities.Chat{
			ID:          id.MustNewULID(),
			Type:        entities.ChannelType,
			Name:        "channel one",
			Description: "just a channel",
			AvatarURL:   "https://test.some",
		}
		t.Require().NoError(t.st.CreateChat(ctx, chat))
	}
	{
		chat := &entities.Chat{
			ID:          id.MustNewULID(),
			Type:        entities.GroupType,
			Name:        "group one",
			Description: "just a group",
			AvatarURL:   "https://test.some",
		}
		t.Require().NoError(t.st.CreateChat(ctx, chat))
	}

}

func (t *testSuite) TestGetChat() {

	ctx := context.Background()

	chatIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		chat := &entities.Chat{
			ID:          id.MustNewULID(),
			Type:        entities.DialogType,
			Name:        "chat " + strconv.Itoa(i),
			Description: "just a chat",
			AvatarURL:   "https://test.some" + strconv.Itoa(i),
		}
		if i%3 == 0 {
			chat.Type = entities.ChannelType
		}
		t.Require().NoError(t.st.CreateChat(ctx, chat))
		chatIDs[i] = chat.ID
	}

	{
		chat, err := t.st.GetChat(ctx, chatIDs[2])
		t.Require().NoError(err)
		t.Assert().Equal(chatIDs[2], chat.ID)
		t.Assert().Equal(entities.DialogType, chat.Type)
		t.Assert().Equal("chat 2", chat.Name)
		t.Assert().Equal("just a chat", chat.Description)
		t.Assert().Equal("https://test.some2", chat.AvatarURL)
	}
	{
		chat, err := t.st.GetChat(ctx, chatIDs[3])
		t.Require().NoError(err)
		t.Assert().Equal(chatIDs[3], chat.ID)
		t.Assert().Equal(entities.ChannelType, chat.Type)
		t.Assert().Equal("chat 3", chat.Name)
		t.Assert().Equal("just a chat", chat.Description)
		t.Assert().Equal("https://test.some3", chat.AvatarURL)
	}
}

func (t *testSuite) TestGetChat_NotFound() {

	ctx := context.Background()

	chatIDs := make([]string, 5)
	for i := 0; i < 3; i++ {
		chat := &entities.Chat{
			ID:   id.MustNewULID(),
			Type: entities.DialogType,
			Name: "chat " + strconv.Itoa(i),
		}
		t.Require().NoError(t.st.CreateChat(ctx, chat))
		chatIDs[i] = chat.ID
	}

	chat, err := t.st.GetChat(ctx, "not_exist")
	t.Require().Error(err)
	t.Assert().ErrorIs(err, chats.ErrNotFound)
	t.Assert().Nil(chat)
}

func (t *testSuite) TestUpdateChat_NotFound() {

	ctx := context.Background()

	chat := &entities.Chat{
		ID:   id.MustNewULID(),
		Type: entities.DialogType,
		Name: "chat",
	}
	t.Require().NoError(t.st.CreateChat(ctx, chat))

	err := t.st.UpdateChat(ctx, &entities.Chat{
		ID:   "not_exist",
		Type: "channel",
		Name: "new name",
	})
	t.Require().Error(err)
	t.Assert().ErrorIs(err, chats.ErrNotFound)
}

func (t *testSuite) TestMembers() {

	ctx := context.Background()

	userIDs := []string{"user_1", "user_2", "user_3"}
	chatIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		chat := &entities.Chat{
			ID:   id.MustNewULID(),
			Type: entities.GroupType,
			Name: "chat " + strconv.Itoa(i),
		}
		t.Require().NoError(t.st.CreateChat(ctx, chat))
		chatIDs[i] = chat.ID

		for j := 0; j < i+1; j++ {
			role := entities.RoleMember
			if j == 0 {
				role = entities.RoleOwner
			}
			t.Require().NoError(t.st.SetMember(ctx, chat.ID, userIDs[j], role))
		}
	}

	t.Run("get role", func() {
		r, err := t.st.GetRole(ctx, chatIDs[0], userIDs[0])
		t.Require().NoError(err)
		t.Assert().Equal(entities.RoleOwner, r)
	})

	t.Run("delete members", func() {

		ms, err := t.st.FindChatMembers(ctx, chatIDs[2], nil)
		t.Require().NoError(err)
		t.Assert().Len(ms, 3)

		n, err := t.st.DeleteMembers(ctx, chatIDs[2])
		t.Require().NoError(err)
		t.Assert().Equal(3, n)

		ms, err = t.st.FindChatMembers(ctx, chatIDs[2], nil)
		t.Require().NoError(err)
		t.Assert().Len(ms, 0, "Должны быть удалены все участники чата")
	})
}

func (t *testSuite) TestTX() {

	ctx := context.Background()

	tx, err := t.st.BeginTx(ctx)
	t.Require().NoError(err)

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
	t.Require().NoError(err)

}
