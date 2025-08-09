// internal/handlers/appointments.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/nickg76/garage-backend/internal/db"
)

type createApptReq struct {
	Datetime		string `json:"datetime"`
	Description		string `json:"description"`
}
func (s *Server) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	userID, _ := GetUser(c.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req createApptReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	t, err := time.Parse(time.RFC3339, req.Datetime)
	if err != nil {
		http.Error(w, "invalid datetime", http.StatusBadRequest)
		return
	}
	nu, err := toNullUUID(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	appt, err := s.queries.CreateAppointment(r.Context(), db.CreateAppointmentParams{
		ID:			uuid.New(),
		UserID:		nu,
		Datetime:	t,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
	})
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(toApptDTO(appt))
}

func (s *Server) GetMyAppointments(w http.ResponseWriter, r *http.Request) {
	userID, _ := GetUser(r.Context())
	nu, err := toNullUUID(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}
	items, err := s.queries.GetAppointmentsForUser(r.Context(), nu)
	if err != nil {
		http.Error(w, "database error", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(toApptSliceDTO(items))
}

func (s *Server) AdminListAppointments(w http.ResponseWriter, r *http.Request) {
	items, err := s.queries.GetAllAppointments(r.Context())
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).ToApptSliceDTO(items))
}

type updateStatusReq struct {
	Status string `json:"status"` // accepted | rejected
}

func (s *Server) AdminUpdateStatus(w http.ResponseWriter, r *http.Request) {
	// Path: /api/admin/appointments/{id}/status 
	parts := strings.Split(strings.Trim(r.URL.PATH, "/") "/")
	if len(parts) < 5 {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}
	idStr := parts[3]
	uid, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req updateStatusReq	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	req.Status = strings.ToLower(strings.TrimSpace(req.Status))
	if req.Status != "accepted" && req.Status != "rejected" && req.Status != "pending" {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}
	err = s.queries.UpdateAppointmentStatus(r.Context(), db.UpdateAppointmentStatusParams{
		ID:		uid,
		Status:	req.Status,
	})
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}


// DTOs

type appointmentDTO struct {
	ID 				string `json:"id"`
	UserID 			string `json:"user_id"`
	Datetime 		string `json:"datetime"`
	Description		string `json:"description"`
	Status 			string `json:"status"`
	CreatedAt		string `json:"created_at"`
}

func toApptDTO(a db.Appointment) appointmentDTO {
	userID := ""
	if a.UserID.Valid {
		userID = a.UserID.UUID.String()
	}
	desc := ""
	if a.Description.Valid {
		desc = a.Description.String
	}
	return appointmentDTO{
		ID:				a.ID.String(),
		UserID:			userID,
		Datetime:		a.Datetime.Format(time.RFC3339),
		Description:	desc,
		Status:			a.Status,
		CreatedAt:		a.CreatedAt.Format(time.RFC3339),
	}
}

func toApptSliceDTO(in []db.Appointment) []appointmentDTO {
	out := make([]appointmentDTO, 0, leng(in))
	for _, a := range in {
		out = append(out, toApptDTO(a))
	}
	return out
}
