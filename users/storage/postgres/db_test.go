package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"testing"

	"github.com/alenapetraki/chat/internal/util"
	"github.com/alenapetraki/chat/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	st := New(&Config{
		Host:     "localhost",
		Port:     "5435",
		User:     "chat_user",
		Password: "chat_password",
		Database: "chat",
	}).(*storage)

	require.NoError(t, st.Connect())
	defer st.Close()
	cleanUp(t, st.db)

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		{
			usr := &users.User{
				ID:       util.GenerateID(),
				Username: "test_usr_" + util.GenerateID(),
				FullName: "Test User 1",
				Status:   "texting",
			}
			require.NoError(t, st.CreateUser(ctx, usr))
		}
		{
			usr := &users.User{
				ID:       util.GenerateID(),
				Username: "test_usr_" + util.GenerateID(),
				FullName: "Test User 2",
				Status:   "also texting",
			}
			require.NoError(t, st.CreateUser(ctx, usr))
		}
	})
	t.Run("username must be unique", func(t *testing.T) {
		usr := &users.User{
			ID:       util.GenerateID(),
			Username: "test_usr_" + util.GenerateID(),
			FullName: "Test User 2",
			Status:   "also texting",
		}
		require.NoError(t, st.CreateUser(ctx, usr))

		err := st.CreateUser(ctx, usr)
		require.Error(t, err)
		assert.ErrorContains(t, err, "unique constraint")
	})

}

func TestRead(t *testing.T) {
	st := New(&Config{
		Host:     "localhost",
		Port:     "5435",
		User:     "chat_user",
		Password: "chat_password",
		Database: "chat",
	}).(*storage)

	require.NoError(t, st.Connect())
	defer st.Close()
	cleanUp(t, st.db)

	ctx := context.Background()

	userIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		usr := &users.User{
			ID:       util.GenerateID(),
			Username: "test_usr_" + strconv.Itoa(i),
			FullName: "Test User " + strconv.Itoa(i),
			Status:   "some",
		}
		require.NoError(t, st.CreateUser(ctx, usr))
		userIDs[i] = usr.ID
	}

	t.Run("get", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			usr, err := st.GetUser(ctx, userIDs[2])
			require.NoError(t, err)
			assert.Equal(t, userIDs[2], usr.ID)
			assert.Equal(t, "test_usr_2", usr.Username)
			assert.Equal(t, "Test User 2", usr.FullName)
			assert.Equal(t, "some", usr.Status)
		})
		t.Run("not found", func(t *testing.T) {
			usr, err := st.GetUser(ctx, "not_exist")
			require.Error(t, err)
			//assert.ErrorContains(t, err, "not found")
			assert.Nil(t, usr)
		})
	})

	t.Run("find", func(t *testing.T) {
		t.Run("all", func(t *testing.T) {
			res, err := st.FindUsers(ctx, nil)
			require.NoError(t, err)
			assert.Len(t, res, 5)
			for _, usr := range res {
				assert.NotEmpty(t, usr.ID)
				assert.NotEmpty(t, usr.Username)
				assert.NotEmpty(t, usr.FullName)
				assert.NotEmpty(t, usr.Status)
			}
		})
		t.Run("limit/offset", func(t *testing.T) {
			f := new(users.FindUsersFilter)
			f.Limit = 2
			f.Offset = 2
			f.Sort = []string{"username"}
			res, err := st.FindUsers(ctx, f)
			require.NoError(t, err)
			require.Len(t, res, 2)
			assert.Equal(t, userIDs[2], res[0].ID)
			assert.Equal(t, userIDs[3], res[1].ID)
		})
	})

}

func TestUpdate(t *testing.T) {
	st := New(&Config{
		Host:     "localhost",
		Port:     "5435",
		User:     "chat_user",
		Password: "chat_password",
		Database: "chat",
	}).(*storage)

	require.NoError(t, st.Connect())
	defer st.Close()
	cleanUp(t, st.db)

	ctx := context.Background()

	usr := &users.User{
		ID:       util.GenerateID(),
		Username: "test_usr",
		FullName: "Test User",
		Status:   "some",
	}
	require.NoError(t, st.CreateUser(ctx, usr))

	t.Run("success", func(t *testing.T) {
		err := st.UpdateUser(ctx, &users.User{
			ID:       usr.ID,
			Username: "another",
			FullName: "New test user",
			Status:   "ta-ta",
		})
		require.NoError(t, err)

		updated, err := st.GetUser(ctx, usr.ID)
		require.NoError(t, err)
		assert.Equal(t, "test_usr", updated.Username)
		assert.Equal(t, "New test user", updated.FullName)
		assert.Equal(t, "ta-ta", updated.Status)
	})
	t.Run("user not exist", func(t *testing.T) {
		err := st.UpdateUser(ctx, &users.User{
			ID:       "some",
			FullName: "another user name",
			Status:   "ta-ta",
		})
		require.Error(t, err)
		assert.ErrorContains(t, err, "not found")
	})

}

func TestDelete(t *testing.T) {
	st := New(&Config{
		Host:     "localhost",
		Port:     "5435",
		User:     "chat_user",
		Password: "chat_password",
		Database: "chat",
	}).(*storage)

	require.NoError(t, st.Connect())
	defer st.Close()
	cleanUp(t, st.db)

	ctx := context.Background()

	usr := &users.User{
		ID:       util.GenerateID(),
		Username: "test_usr",
		FullName: "Test User",
		Status:   "some",
	}
	require.NoError(t, st.CreateUser(ctx, usr))

	t.Run("success", func(t *testing.T) {
		err := st.DeleteUser(ctx, usr.ID)
		require.NoError(t, err)

		_, err = st.GetUser(ctx, usr.ID)
		require.Error(t, err)
		//assert.ErrorContains(t, err, "cc")
	})
	t.Run("user not exist", func(t *testing.T) {
		err := st.DeleteUser(ctx, usr.ID)
		require.Error(t, err)
		assert.ErrorContains(t, err, "not found")
	})

}

func cleanUp(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`DELETE FROM "user"`)
	require.NoError(t, err)
}
