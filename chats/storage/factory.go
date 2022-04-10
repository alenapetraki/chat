package storage

import (
	"github.com/alenapetraki/chat/chats"
	"github.com/alenapetraki/chat/chats/storage/postgres"
	"github.com/pkg/errors"
)

func New(typ string, config any) (chats.Storage, error) {

	switch typ {
	case "postgres":
		pConfig, ok := config.(*postgres.Config)
		if !ok {
			return nil, errors.New("invalid config")
		}
		return postgres.New(pConfig), nil
	default:
		return nil, errors.Errorf("unknown storage type: '%s'", typ)
	}
}
