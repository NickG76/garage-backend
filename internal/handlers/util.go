// internal/handlers/util.go
package handlers

import (
	"context"

	"github.com/nickg76/garage-backend/internal/auth"
)

func WithUser(ctx context.Context, userID string, isAdmin bool) context.Context {
	ctx = context.WithValue(ctx, auth.UserIDKey, userID)
	ctx = context.WithValue(ctx, auth.AdminKey, isAdmin)
	return ctx
}

func GetUser(ctx context.Context) (string, bool) {
	id, _ := ctx.Value(auth.UserIDKey).(string)
	admin, _ := ctx.Value(auth.AdminKey).(bool)
	return id, admin
}
