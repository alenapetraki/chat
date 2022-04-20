-- +goose Up
CREATE TYPE role AS ENUM (
    'member',
    'owner'
);

CREATE TYPE chat_type AS ENUM (
    'dialog',
    'group',
    'channel'
);

CREATE TABLE IF NOT EXISTS chat (
    id text PRIMARY KEY,
    type chat_type NOT NULL,
    name varchar(100),
    description text,
    avatar_url text,
    deleted_at timestamp
);

CREATE TABLE IF NOT EXISTS member (
    user_id text,
    chat_id text,
    role role,
    primary key (user_id, chat_id),
    deleted_at timestamp
);



-- +goose Down
DROP TABLE
    chat,
    member;

DROP TYPE
    role,
    chat_type;
