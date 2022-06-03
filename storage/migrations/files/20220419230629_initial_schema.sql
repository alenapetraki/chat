-- +goose Up

CREATE TABLE IF NOT EXISTS chat (
    id text PRIMARY KEY,
    type text NOT NULL,
    name text,
    num_members int default 0,
    description text,
    avatar_url text,
    deleted_at timestamp
);

CREATE TABLE IF NOT EXISTS member (
    user_id text,
    chat_id text,
    role text,
    primary key (user_id, chat_id),
    deleted_at timestamp
);



-- +goose Down
DROP TABLE
    chat,
    member;
