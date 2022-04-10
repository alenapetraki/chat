package auth

import "context"

type userIDKey struct{}

func WithUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

func GetUserID(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey{}).(string)
	return userID
}
