package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alenapetraki/chat/account"
	"github.com/alenapetraki/chat/chats"
	"github.com/alenapetraki/chat/commons"
	"github.com/alenapetraki/chat/commons/postgres"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

type storage struct {
	cfg *Config
	db  *sql.DB
}

type Config struct {
	postgres.Config
}

func New(cfg *Config) chats.Storage {
	return &storage{cfg: cfg}
}

func (s *storage) Connect() error {

	db, err := postgres.Connect(&s.cfg.Config)
	if err != nil {
		return err
	}

	s.db = db

	if err := s.initChatTable(); err != nil {
		return err
	}
	if err := s.initMemberTable(); err != nil {
		return err
	}

	return nil
}

func (s *storage) Close() error {
	if s.db == nil {
		return nil
	}

	if err := s.db.Close(); err != nil {
		return errors.Wrap(err, "failed to close db connection")
	}

	return nil
}

func (s *storage) initChatTable() error {

	if _, err := s.db.Exec(`
create table if not exists chat (
	id text PRIMARY KEY,
	type integer NOT NULL,
	name text,
	description text,
	avatar_url text,
	deleted bool default false
)`); err != nil {
		return errors.Wrap(err, "failed to initialize 'chat' table")
	}

	return nil
}

func (s *storage) initMemberTable() error {

	if _, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS member ( 
	user_id text,
	chat_id text,
	role integer,
-- 	FOREIGN KEY (user_id) REFERENCES user (id) ON DELETE CASCADE,
-- 	FOREIGN KEY (chat_id) REFERENCES chat (id) ON DELETE CASCADE,
	primary key (user_id, chat_id)
)`); err != nil {
		return errors.Wrap(err, "failed to initialize 'chats' table")
	}

	return nil
}

func (s *storage) CreateChat(ctx context.Context, chat *chats.Chat) error {

	query := `
insert into "chat" (
	id, type, name, description, avatar_url
)
values (
	$1, $2, $3, $4, $5
)`
	_, err := s.db.ExecContext(
		ctx,
		query,
		chat.ID,
		chat.Type,
		chat.Name,
		chat.Description,
		chat.AvatarURL,
	)
	if err != nil {
		return errors.Wrap(err, "failed to insert 'chat' entity")
	}

	return nil
}

func (s *storage) UpdateChat(ctx context.Context, chat *chats.Chat) error {

	query := `
update "chat"
set 
    name=$1, 
    description=$2, 
    avatar_url=$3 
where 
    id=$4
` // todo: посмотреть style guide

	res, err := s.db.ExecContext(
		ctx,
		query,
		chat.Name,
		chat.Description,
		chat.AvatarURL,
		chat.ID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to update 'chat'")
	}

	if num, _ := res.RowsAffected(); num == 0 {
		return errors.New("not found")
	}

	return nil
}

func (s *storage) GetChat(ctx context.Context, chatID string) (*chats.Chat, error) {

	q := `
select 
  type, name, description, avatar_url
from "chat"
where id=$1`
	row := s.db.QueryRowContext(ctx, q, chatID)

	chat := chats.Chat{ID: chatID}

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

//func (s *storage) ForceDeleteChat(ctx context.Context, chatID string) error {
//
//	query := `delete from "chat" where id=$1`
//
//	res, err := s.db.ExecContext(ctx, query, chatID)
//	if err != nil {
//		return errors.Wrap(err, "failed to delete 'chat'")
//	}
//
//	if num, _ := res.RowsAffected(); num == 0 {
//		return errors.New("not found")
//	}
//
//	return nil
//}

func (s *storage) DeleteChat(ctx context.Context, chatID string, force ...bool) error {

	var query string
	if len(force) > 0 && force[0] {
		query = `delete from "chat" where id=$1`
	} else {
		query = `update "chat" set deleted=true where id=$1 and deleted=false`
	}

	res, err := s.db.ExecContext(ctx, query, chatID)
	if err != nil {
		return errors.Wrap(err, "failed to delete 'chat'")
	}

	if num, _ := res.RowsAffected(); num == 0 {
		return errors.New("not found")
	}

	return nil
}

func (s *storage) AddMember(ctx context.Context, chatID, userID string, role chats.Role) error {

	_, err := s.db.ExecContext(
		ctx, `insert into "member" (chat_id, user_id, role) values ($1, $2, $3)`,
		chatID, userID, role,
	)
	if err != nil {
		return errors.Wrap(err, "failed to insert 'member'")
	}

	return nil
}

func (s *storage) DeleteMembers(ctx context.Context, chatID string, userID ...string) error {

	q := `delete from "member" where chat_id=$1`
	args := []any{chatID}

	if len(userID) > 0 {
		q = fmt.Sprintf("%s and user_id in ($2)", q)
		args = append(args, pq.Array(userID))
	}

	_, err := s.db.ExecContext(ctx, q, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete 'member'")
	}

	return nil
}

func (s *storage) GetRole(ctx context.Context, chatID, userID string) (chats.Role, error) {

	row := s.db.QueryRowContext(ctx, `select role from "member" where chat_id=$1 and user_id=$2`, chatID, userID)

	var role chats.Role

	err := row.Scan(&role)
	if err != nil {
		return 0, err
	}

	return role, nil
}

func (s *storage) FindChatMembers(ctx context.Context, chatID string, options *commons.PaginationOptions) ([]*chats.ChatMember, error) {
	//	q := `
	//select user.id, user.username, user.password, user.email, user.full_name, user.status, member.role
	//from "member"
	//where chat_id=$1
	//join "user"
	//on member.user_id=user.id
	//`

	q := `select user_id, role from "member" where chat_id=$1 order by role desc, user_id asc`
	args := []any{chatID}

	if options != nil {
		if options.Limit != 0 {
			q += " limit $2 offset $3"
			args = append(args, options.Limit, options.Offset)
		}
	}

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*chats.ChatMember, 0)
	for rows.Next() {
		m := new(chats.ChatMember)
		m.Chat = &chats.Chat{ID: chatID}
		m.User = new(account.User)

		err := rows.Scan(
			&m.User.ID,
			&m.Role,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, m)
	}

	return res, nil

}
