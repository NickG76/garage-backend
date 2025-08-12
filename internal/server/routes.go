// internal/server/routes.go
package server

import (
    "net/http"
    "os"
    "path/filepath"

    "github.com/nickg76/garage-backend/internal/handlers"
)

func Routes(s *handlers.Server) http.Handler {
    mux := http.NewServeMux()

    // --- API routes ---
    mux.HandleFunc("POST /api/register", s.Register)
    mux.HandleFunc("POST /api/login", s.Login)
    mux.Handle("GET /api/me", s.AuthMiddleware(http.HandlerFunc(s.Me)))
    mux.Handle("GET /api/appointments", s.AuthMiddleware(http.HandlerFunc(s.GetMyAppointments)))
    mux.Handle("POST /api/appointments", s.AuthMiddleware(http.HandlerFunc(s.CreateAppointment)))

    // Admin routes
    adminList := s.AuthMiddleware(s.AdminOnly(http.HandlerFunc(s.AdminListAppointments)))
    adminUpdate := s.AuthMiddleware(s.AdminOnly(http.HandlerFunc(s.AdminUpdateStatus)))
    mux.Handle("GET /api/admin/appointments", adminList)
    mux.Handle("PATCH /api/admin/appointments/", adminUpdate)
    adminUpdateAdmins := s.AuthMiddleware(s.AdminOnly(http.HandlerFunc(s.UpdateAdminAccountsHandler)))
    mux.Handle("POST /api/admin/update-admins", adminUpdateAdmins)

    // --- Frontend routes fallback ---
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // Attempt to serve static file
        staticPath := filepath.Join("public", r.URL.Path)
        if info, err := os.Stat(staticPath); err == nil && !info.IsDir() {
            http.ServeFile(w, r, staticPath)
            return
        }

        // If not a file, fallback to React index.html
        http.ServeFile(w, r, "public/index.html")
    })

    return cors(mux)
}


// Minimal CORS for dev; adjust as needed or remove if same origin
func cors(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusNoContent)
            return
        }
        next.ServeHTTP(w, r)
    })
}

