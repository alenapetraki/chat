package postgres

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/alenapetraki/chat/users"
	"github.com/cockroachdb/errors"
	_ "github.com/lib/pq"
)

type storage struct {
	cfg *Config
	db  *sql.DB
}

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func New(cfg *Config) users.Storage {
	return &storage{cfg: cfg}
}

func (s *storage) Connect() error {

	if s.cfg.Host == "" || s.cfg.Port == "" || s.cfg.User == "" ||
		s.cfg.Password == "" || s.cfg.Database == "" {
		return errors.New("host:port, user:password and database parameters required")
	}

	db, err := sql.Open("postgres", fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		s.cfg.User, s.cfg.Password, s.cfg.Database, s.cfg.Host, s.cfg.Port))
	if err != nil {
		return errors.Wrap(err, "failed to open connection to postgres database")
	}

	if err = db.Ping(); err != nil {
		return errors.Wrap(err, "failed to ping")
	}

	s.db = db

	return s.initUserTable()
}

func (s *storage) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *storage) initUserTable() error {

	if _, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS "user" (
	id text PRIMARY KEY,
	username text UNIQUE NOT NULL,
	full_name text,
	status text,
	avatar_url text,
	deleted bool                  
)`); err != nil {
		return errors.Wrap(err, "failed to initialize 'user' table")
	}

	return nil
}

func (s *storage) CreateUser(ctx context.Context, user *users.User) error {

	query := `
INSERT INTO "user" (
	id, username, full_name, status, deleted
)
VALUES (
	$1, $2, $3, $4, false
)`
	_, err := s.db.Exec(
		query,
		user.ID,
		user.Username,
		user.FullName,
		user.Status,
	)
	if err != nil {
		return errors.Wrap(err, "failed to insert 'user'")
	}

	return nil
}

func (s *storage) UpdateUser(ctx context.Context, user *users.User) error {

	query := `UPDATE "user" SET full_name=$1, status=$2 WHERE id=$3 AND deleted=false` // todo: посмотреть style guide

	res, err := s.db.Exec(
		query,
		user.FullName,
		user.Status,
		user.ID,
	)
	if err != nil {
		return errors.Wrap(err, "failed to update 'user'")
	}

	if num, _ := res.RowsAffected(); num == 0 {
		return errors.New("not found")
	}

	return nil
}

func (s *storage) GetUser(ctx context.Context, userID string) (*users.User, error) {

	row := s.db.QueryRowContext(ctx, `SELECT username, full_name, status FROM "user" WHERE id=$1 AND deleted=false`, userID) // todo: лучше прописывать явно колонки или '*'

	user := &users.User{ID: userID}

	err := row.Scan(&user.Username, &user.FullName, &user.Status)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *storage) FindUsers(ctx context.Context, filter *users.FindUsersFilter) ([]*users.User, error) {

	//q := `SELECT id, username, full_name, status FROM "user"`
	//args := make([]any, 0)

	builder := sq.Select("id", "username", "full_name", "status").
		From("\"user\"").
		Where("deleted=false")

	if filter != nil {

		if filter.Username != "" {
			builder = builder.Where(sq.Like{"username": "?"}, filter.Username)

			//where = append(where, `username LIKE ?`)
			//args = append(args, filter.Username)
		}

		if len(filter.Sort) > 0 {
			builder = builder.OrderBy(filter.Sort...)
			//q += " ORDER BY " + strings.Join(filter.Sort, ",")
		}
		if filter.Offset != 0 {
			builder = builder.Offset(uint64(filter.Offset))
			//q = fmt.Sprintf(`%s OFFSET %d`, q, filter.Offset)
		}
		if filter.Limit != 0 {
			builder = builder.Limit(uint64(filter.Limit))
			//q = fmt.Sprintf(`%s LIMIT %d`, q, filter.Limit)
		}

	}

	str, args, err := builder.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(str, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]*users.User, 0)
	for rows.Next() {

		user := new(users.User)
		err = rows.Scan(&user.ID, &user.Username, &user.FullName, &user.Status)
		if err != nil {
			return nil, err
		}

		res = append(res, user)
	}
	return res, nil
}

func (s *storage) ForceDeleteUser(ctx context.Context, userID string) error {

	query := `DELETE FROM "user" WHERE id=$1`

	res, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete 'user'")
	}

	if num, _ := res.RowsAffected(); num == 0 {
		return errors.New("not found")
	}

	return nil
}

func (s *storage) DeleteUser(ctx context.Context, userID string) error {

	query := `UPDATE "user" SET deleted=true WHERE id=$1 AND deleted=false`

	res, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return errors.Wrap(err, "failed to delete 'user'")
	}

	if num, _ := res.RowsAffected(); num == 0 {
		return errors.New("not found")
	}

	return nil
}
