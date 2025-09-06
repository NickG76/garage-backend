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
	mux.Handle("DELETE /api/appointments/", s.AuthMiddleware(http.HandlerFunc(s.UserDeleteAppointment)))
	mux.Handle("PATCH /api/appointments/", s.AuthMiddleware(http.HandlerFunc(s.UserEditAppointment)))
	mux.Handle("PUT /api/appointments/", s.AuthMiddleware(http.HandlerFunc(s.UserEditAppointment)))

	// Admin routes
	adminList := s.AuthMiddleware(s.AdminOnly(http.HandlerFunc(s.AdminListAppointments)))
	adminUpdate := s.AuthMiddleware(s.AdminOnly(http.HandlerFunc(s.AdminUpdateStatus)))
	mux.Handle("GET /api/admin/appointments", adminList)
	mux.Handle("PATCH /api/admin/appointments/", adminUpdate)

	s.SetAdminAccountsFromEnv()
	// SSE events (JWT via query params)
	mux.HandleFunc("GET /api/events", s.Events)
	//
	// // --- Frontend routes fallback ---
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

	// return cors(mux)
	return mux
}


// // cors is a middleware that handles cross-origin requests.
// func cors(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// In production, you might want to restrict this to your frontend's domain,
// 		// e.g., "https://girlingdesign.co.uk"
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
// 		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS, DELETE, PUT")
//
// 		// Handle preflight requests
// 		if r.Method == "OPTIONS" {
// 			w.WriteHeader(http.StatusNoContent)
// 			return
// 		}
//
// 		next.ServeHTTP(w, r)
// 	})
// }
