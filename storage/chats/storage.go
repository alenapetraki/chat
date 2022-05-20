package chats

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/alenapetraki/chat/entities/entities"
	"github.com/alenapetraki/chat/services/chats"
	"github.com/alenapetraki/chat/storage"
	"github.com/alenapetraki/chat/util"
	//"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Storage struct {
	conn storage.Conn
	//db   *sql.DB
}

func New(db storage.Conn) *Storage {
	return &Storage{conn: db}
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (s *Storage) RunTx(ctx context.Context, f func(st chats.Storage) error) error {

	var err error

	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		p := recover()
		switch {
		case p != nil:
			_ = tx.Rollback()
			panic(p)
		case err != nil:
			_ = tx.Rollback()
		default:
			err = tx.Commit()
		}
	}()

	return f(New(storage.NewTransactioner(tx)))
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
		return err
	}

	if num, _ := res.RowsAffected(); num == 0 {
		return chats.ErrNotFound
	}

	return nil
}

func incrementChatMembersNum(ctx context.Context, runner sq.BaseRunner, chatID string, delta int) (int, error) {

	row := psql.Update("chat").
		Set("num_members", sq.Expr("num_members + ?", delta)).
		Where(
			sq.Eq{
				"id":         chatID,
				"deleted_at": nil,
			},
		).
		Suffix("RETURNING num_members").
		RunWith(runner).
		QueryRowContext(ctx)

	var num int
	if err := row.Scan(&num); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, chats.ErrNotFound
		}
		return 0, err
	}

	return num, nil
}

func (s *Storage) GetChat(ctx context.Context, chatID string) (*entities.Chat, error) {

	row := psql.Select("type", "name", "num_members", "description", "avatar_url").
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
		&chat.NumMembers,
		&chat.Description,
		&chat.AvatarURL,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, chats.ErrNotFound
		}
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
		return chats.ErrNotFound
	}

	return nil
}

func (s *Storage) SetMember(ctx context.Context, chatID, userID string, role entities.Role) error {

	//return storage.RunTx(ctx, s.db,  func(tx *sql.Tx) error {
	_, err := psql.Insert("member").
		Columns("chat_id", "user_id", "role").
		Values(chatID, userID, role).
		Suffix("ON CONFLICT (user_id, chat_id) DO UPDATE SET user_id = ?", role).
		RunWith(s.conn).
		ExecContext(ctx)
	if err != nil {
		return err
	}

	_, err = incrementChatMembersNum(ctx, s.conn, chatID, 1)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) DeleteMembers(ctx context.Context, chatID string, userID ...string) (int, error) {

	eq := sq.Eq{
		"chat_id": chatID,
	}
	if len(userID) > 0 {
		eq["user_id"] = userID
	}

	res, err := psql.Delete("member").
		Where(eq).RunWith(s.conn).ExecContext(ctx)

	var deleted int
	if res != nil {
		n, err := res.RowsAffected()
		if err != nil {
			return 0, errors.Wrap(err, "get number of deleted items")
		}
		deleted = int(n)
	}

	_, err = incrementChatMembersNum(ctx, s.conn, chatID, -deleted)
	return deleted, err
}

func (s *Storage) GetRole(ctx context.Context, chatID, userID string) (entities.Role, error) {

	var role entities.Role

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
		if errors.Is(err, sql.ErrNoRows) {
			return "", chats.ErrNotFound
		}
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
