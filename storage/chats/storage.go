package chats

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/alenapetraki/chat/entities"
	"github.com/alenapetraki/chat/services/chats"
	"github.com/alenapetraki/chat/storage"
	"github.com/alenapetraki/chat/util"
	//"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type Storage struct {
	storage.DB
}

func New(db storage.DB) *Storage {
	return &Storage{DB: db}
}

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

func (s *Storage) CreateChat(ctx context.Context, chat *entities.Chat) error {
	const op = "Storage.CreateChat"

	_, err := psql.Insert("chat").
		Columns("id", "type", "name", "description", "avatar_url").
		Values(chat.ID, chat.Type, chat.Name, chat.Description, chat.AvatarURL).
		RunWith(s.DB).ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (s *Storage) UpdateChat(ctx context.Context, chat *entities.Chat) error {

	const op = "Storage.UpdateChat"

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
		RunWith(s.DB).
		ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, op)
	}

	if num, _ := res.RowsAffected(); num == 0 {
		return errors.Wrap(chats.ErrNotFound, op)
	}

	return nil
}

func (s *Storage) incrementChatMembersCount(ctx context.Context, chatID string, delta int) (int, error) {

	row := psql.Update("chat").
		Set("num_members", sq.Expr("num_members + ?", delta)).
		Where(
			sq.Eq{
				"id":         chatID,
				"deleted_at": nil,
			},
		).
		Suffix("RETURNING num_members").
		RunWith(s.DB).
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

	const op = "Storage.GetChat"

	row := psql.Select("type", "name", "num_members", "description", "avatar_url").
		From("chat").
		Where(
			sq.Eq{
				"id":         chatID,
				"deleted_at": nil,
			},
		).
		RunWith(s.DB).
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
			err = chats.ErrNotFound
		}
		return nil, errors.Wrap(err, op)
	}

	return &chat, nil
}

func (s *Storage) DeleteChat(ctx context.Context, chatID string, force ...bool) error {
	const op = "Storage.DeleteChat"

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

	res, err := s.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, op)
	}

	if num, _ := res.RowsAffected(); num == 0 {
		return errors.Wrap(chats.ErrNotFound, op)
	}

	return nil
}

func (s *Storage) SetMember(ctx context.Context, chatID, userID string, role entities.Role) error {
	const op = "Storage.SetMember"

	//return storage.RunTx(ctx, s.db,  func(tx *sql.Tx) error {
	_, err := psql.Insert("member").
		Columns("chat_id", "user_id", "role").
		Values(chatID, userID, role).
		Suffix("ON CONFLICT (user_id, chat_id) DO UPDATE SET user_id = ?", role).
		RunWith(s.DB).
		ExecContext(ctx)
	if err != nil {
		return errors.Wrap(err, op)
	}

	_, err = s.incrementChatMembersCount(ctx, chatID, 1)
	if err != nil {
		return errors.Wrap(err, op)
	}

	return nil
}

func (s *Storage) DeleteMembers(ctx context.Context, chatID string, userID ...string) (int, error) {

	const op = "Storage.DeleteMembers"

	eq := sq.Eq{
		"chat_id": chatID,
	}
	if len(userID) > 0 {
		eq["user_id"] = userID
	}

	res, err := psql.Delete("member").
		Where(eq).RunWith(s.DB).ExecContext(ctx)
	if err != nil {
		return 0, errors.Wrap(err, op)
	}

	var deleted int
	if res != nil {
		n, err := res.RowsAffected()
		if err != nil {
			return 0, errors.Wrap(err, op)
		}
		deleted = int(n)
	}

	if _, err = s.incrementChatMembersCount(ctx, chatID, -deleted); err != nil {
		return 0, errors.Wrap(err, op)
	}
	return deleted, nil
}

func (s *Storage) GetRole(ctx context.Context, chatID, userID string) (entities.Role, error) {

	const op = "Storage.GetRole"

	var role entities.Role

	err := psql.Select("role").
		From("member").
		Where(
			sq.Eq{
				"chat_id": chatID,
				"user_id": userID,
			},
		).
		RunWith(s.DB).
		QueryRowContext(ctx).
		Scan(&role)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = chats.ErrNotFound
		}
		return "", errors.Wrap(err, op)
	}

	return role, nil
}

func (s *Storage) FindChatMembers(ctx context.Context, chatID string, options *util.PaginationOptions) ([]*entities.ChatMember, error) {

	const op = "Storage.FindChatMembers"

	query := psql.Select("user_id", "role").
		From("member").
		Where(
			sq.Eq{"chat_id": chatID},
		).
		OrderBy("role", "user_id")

	if options != nil && options.Limit != 0 {
		query = query.Limit(uint64(options.Limit)).Offset(uint64(options.Offset))
	}

	rows, err := query.RunWith(s.DB).QueryContext(ctx)
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
			return nil, errors.Wrap(err, op)
		}

		res = append(res, m)
	}

	return res, nil

}
