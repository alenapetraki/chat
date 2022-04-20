package storage

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/alenapetraki/chat/chats"
	"github.com/alenapetraki/chat/models/entities"
	"github.com/alenapetraki/chat/util"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Storage struct {
	conn Conn
}

func New(db Conn) *Storage {
	return &Storage{conn: db}
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// tx - адаптер для использования как execer
type transactioner struct {
	Tx
}

func (t *transactioner) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, nil
}

func (s *Storage) BeginTx(ctx context.Context) (chats.Storage, error) {
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "get tx")
	}
	return New(&transactioner{tx}), nil
}

func (s *Storage) EndTx(f func() error) error {
	tx := s.conn.(*transactioner)

	err := f()
	if err != nil {
		rErr := tx.Rollback()
		_ = rErr // тоже вернуть
		return err
	}

	return errors.Wrap(tx.Commit(), "commit tx")
}

func (s *Storage) CreateChat(ctx context.Context, chat *entities.Chat) error {

	_, err := psql.Insert("chat").
		Columns("id", "type", "name", "description", "avatar_url").
		Values(chat.ID, chat.Type, chat.Name, chat.Description, chat.AvatarURL).
		RunWith(s.conn).ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to insert 'chat' entity")
	}

	return nil
}

func (s *Storage) UpdateChat(ctx context.Context, chat *entities.Chat) error {

	res, err := psql.Update("chat").
		Set("name", chat.Name).
		Set("description", chat.Description).
		Set("avatar_url", chat.AvatarURL).
		Where(
			sq.Eq{
				"id":         chat.ID,
				"deleted_at": nil,
			},
		).
		RunWith(s.conn).
		ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to update 'chat'")
	}

	if num, _ := res.RowsAffected(); num == 0 {
		return errors.New("not found")
	}

	return nil
}

func (s *Storage) GetChat(ctx context.Context, chatID string) (*entities.Chat, error) {

	row := psql.Select("type", "name", "description", "avatar_url").
		From("chat").
		Where(
			sq.Eq{
				"id":         chatID,
				"deleted_at": nil,
			},
		).
		RunWith(s.conn).
		QueryRowContext(ctx)

	chat := entities.Chat{ID: chatID}

	err := row.Scan(
		&chat.Type,
		&chat.Name,
		&chat.Description,
		&chat.AvatarURL,
	)
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (s *Storage) DeleteChat(ctx context.Context, chatID string, force ...bool) error {

	var (
		query string
		args  []any
	)
	if len(force) > 0 && force[0] {
		query, args = psql.Delete("chat").Where(sq.Eq{"id": chatID}).MustSql()
	} else {
		query, args = psql.Update("chat").
			Where(
				sq.Eq{
					"id":         chatID,
					"deleted_at": nil,
				},
			).
			MustSql()
	}

	res, err := s.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete 'chat'")
	}

	if num, _ := res.RowsAffected(); num == 0 {
		return errors.New("not found")
	}

	return nil
}

func (s *Storage) SetMember(ctx context.Context, chatID, userID string, role string) error {

	_, err := psql.Insert("member").
		Columns("chat_id", "user_id", "role").
		Values(chatID, userID, role).
		Suffix("ON CONFLICT (user_id, chat_id) DO UPDATE SET user_id = ?", role).
		RunWith(s.conn).
		ExecContext(ctx)

	return err
}

func (s *Storage) DeleteMembers(ctx context.Context, chatID string, userID ...string) error {

	eq := sq.Eq{
		"chat_id": chatID,
	}
	if len(userID) > 0 {
		eq["user_id"] = userID
	}

	_, err := psql.Delete("member").
		Where(eq).RunWith(s.conn).ExecContext(ctx)

	return err
}

func (s *Storage) GetRole(ctx context.Context, chatID, userID string) (string, error) {

	var role string

	err := psql.Select("role").
		From("member").
		Where(
			sq.Eq{
				"chat_id": chatID,
				"user_id": userID,
			},
		).
		RunWith(s.conn).
		QueryRowContext(ctx).
		Scan(&role)

	if err != nil {
		return "", err
	}

	return role, nil
}

func (s *Storage) FindChatMembers(ctx context.Context, chatID string, options *util.PaginationOptions) ([]*entities.ChatMember, error) {

	query := psql.Select("user_id", "role").
		From("member").
		Where(
			sq.Eq{"chat_id": chatID},
		).
		OrderBy("role", "user_id")

	//Join("user on member.user_id=user.id"). // todo по готовности users
	//OrderBy("role", "user.name")

	if options != nil && options.Limit != 0 {
		query = query.Limit(uint64(options.Limit)).Offset(uint64(options.Offset))
	}

	rows, err := query.RunWith(s.conn).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*entities.ChatMember, 0)
	for rows.Next() {
		m := new(entities.ChatMember)
		//m.Chat = &chats.Chat{ID: chatID}
		//m.User = new(account.User)

		err := rows.Scan(
			&m.UserID,
			&m.Role,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, m)
	}

	return res, nil

}
