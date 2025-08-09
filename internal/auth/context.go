// internal/auth/context.go
package auth

type contextKey string

const (
	UserIDKey	contextKey = "user_id"
	AdminKey	contextKey = "is_admin"
)
