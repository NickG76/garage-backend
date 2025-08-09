// internal/server/routes.go
package server

import (
	"net/http"

	"github.com/nickg76/garage-backend/internal/handlers"
)

func Routes(s *handlers.Server) http.Handler {
	mux := http.NewServeMux()

	// auth
	mux.HandleFunc("POST /api/register", s.Register)
	mux.HandleFunc("POST /api/login", s.Login)

	// User-protected routes
	mux.Handle("GET /api/appointmments", s.AuthMiddleware(http.HandlerFunc(s.GetMyAppointments)))
	mux.Handle("POST /api/appointments", s.AuthMiddleware(http.HandlerFunc(s.CreateAppointment)))

	// Admin routes
	adminList := s.AuthMiddleware(s.AdminOnly(http.HandlerFunc(s.AdminListAppointments)))
	adminUpdate := s.AuthMiddleware(s.AdminOnly(http.HandlerFunc(s.AdminUpdateStatus)))
	mux.Handle("GET /api/admin/appointments", adminList)
	mux.Handle("PUT /api/admin/appointments/", adminUpdate)

	// Static files
	fs := http.FileServer(http.Dir("public"))
	mux.Handle("/", fs)

	return mux
} 


