// internal/handlers/admins.go
package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/nickg76/garage-backend/internal/db"
)

// SetAdminAccountsFromEnv updates admins based on env vars
func (s *Server) SetAdminAccountsFromEnv() {
	adminsToAdd := strings.Split(os.Getenv("ADMIN_EMAILS"), ",")
	adminsToRemove := strings.Split(os.Getenv("ADMIN_REMOVE_EMAILS"), ",")

	for _, email := range adminsToAdd {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}
		updateAdmin(s, email, true)
	}

	for _, email := range adminsToRemove {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}
		updateAdmin(s, email, false)
	}
}

func updateAdmin(s *Server, email string, isAdmin bool) {
	log.Printf("Setting %s admin=%v", email, isAdmin)
	err := s.queries.SetAdmin(context.Background(), db.SetAdminParams{
		Email:   email,
		IsAdmin: sql.NullBool{Bool: isAdmin, Valid: true},
	})
	if err != nil {
		log.Printf("❌ Failed to update %s: %v", email, err)
	} else {
		log.Printf("✅ Updated %s successfully", email)
	}
}

// UpdateAdminAccountsHandler triggers SetAdminAccountsFromEnv via HTTP
func (s *Server) UpdateAdminAccountsHandler(w http.ResponseWriter, r *http.Request) {
	// Optional: disable this route in production unless DEV_MODE=true
	if os.Getenv("DEV_MODE") != "true" {
		http.Error(w, "not available in production", http.StatusForbidden)
		return
	}

	// Check for extra shared secret
	secret := r.URL.Query().Get("secret")
	if secret != os.Getenv("ADMIN_UPDATE_SECRET") {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Perform update
	s.SetAdminAccountsFromEnv()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Admin accounts updated successfully"))
}

