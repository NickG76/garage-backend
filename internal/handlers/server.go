// internal/handlers/server.go
package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"

	"github.com/nickg76/garage-backend/internal/auth"
	"github.com/nickg76/garage-backend/internal/db"
)

type Server struct {
	db 		*sqlx.DB
	queries *db.Queries
}

func NewServer() *Server {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}
	conn := sqlx.MustConnect("postgres", dsn)
	return &Server{
		db:		 conn,
		queries: db.New(conn.DB),
	}
}

func (s *Server) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		token := strings.TrimSpace(h[len("Bearer "):])
		claims, err := auth.ParseJWT(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := r.Context()
		ctx = WithUser(ctx, claims.Sub, claims.Admin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, isAdmin := GetUser(r.Context())
		if !isAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func toNullUUID(idStr string) (uuid.NullUUID, error) {
	u, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.NullUUID{}, err
	}
	return uuid.NullUUID{UUID: u, Valid: true}, nil
}

func toNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
