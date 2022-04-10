package commons

import "github.com/rs/xid"

func GenerateID() string {
	return xid.New().String()
}
