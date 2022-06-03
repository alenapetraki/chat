package chats

import "github.com/pkg/errors"

var (
	ErrNotFound              = errors.New("not found")
	ErrMaxMembersNumExceeded = errors.New("max number of members is exceeded")
)
