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
    Datetime    string `json:"datetime"`    // RFC3339
	Title		string `json:"title"`
    Description string `json:"description"` // optional
}

func (s *Server) CreateAppointment(w http.ResponseWriter, r *http.Request) {
    userID, _ := GetUser(r.Context())
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
        ID:          uuid.New(),
        UserID:      nu,
        Datetime:    t,
		Title: 		 req.Title,
        Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
    })
    if err != nil {
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
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
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(toApptSliceDTO(items))
}

func (s *Server) AdminListAppointments(w http.ResponseWriter, r *http.Request) {
    items, err := s.queries.GetAllAppointments(r.Context())
    if err != nil {
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    // Use the admin-specific converter for rows with joined user info
    json.NewEncoder(w).Encode(toApptSliceDTOAdmin(items))
}

type updateStatusReq struct {
    Status string `json:"status"` // "accepted" | "rejected" | "pending"
}

func (s *Server) AdminUpdateStatus(w http.ResponseWriter, r *http.Request) {
    // Expected path: /api/admin/appointments/{id}/status
    parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
    // Validate the exact shape of the path
    if len(parts) != 5 || parts[0] != "api" || parts[1] != "admin" || parts[2] != "appointments" || parts[4] != "status" {
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

    if err := s.queries.UpdateAppointmentStatus(r.Context(), db.UpdateAppointmentStatusParams{
        ID:     uid,
        Status: req.Status,
    }); err != nil {
        http.Error(w, "db error", http.StatusInternalServerError)
        return
    }

	appt, err := s.queries.GetAppointmentsByID(r.Context(), uid)
	if err == nil {
		// notify the appointment's user if present
		if appt.UserID.Valid {
			s.publishToUser(r.Context(), appt.UserID.UUID.String(), Event{
				Type:			"appointment_status",
				Appointment: 	appt.ID.String(),
				Status: 		req.Status,
				Message:		"Your appointment status was updated",
			})
		}
	}

    w.WriteHeader(http.StatusNoContent)
}

func (s *Server) UserDeleteAppointment(w http.ResponseWriter, r *http.Request) {
	userID, _ := GetUser(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	// Expected path: /api/appointments/{id}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// Validate the exact shape of the path
	if len(parts) != 3 || parts[0] != "api" || parts[1] != "appointments" {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}

	idStr := parts[2]
	uid, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	nu, err := toNullUUID(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// Ensure the appointment belongs to the user
	appt, err := s.queries.GetAppointmentsByID(r.Context(), uid)
	if err != nil {
		http.Error(w, "appointment not found", http.StatusNotFound)
		return
	}
	if !appt.UserID.Valid || appt.UserID.UUID.String() != nu.UUID.String() {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if err := s.queries.DeleteAppointment(r.Context(), uid); err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	
}

// --- Helpers ---

func nullStr(ns sql.NullString) string {
    if ns.Valid {
        return ns.String
    }
    return ""
}

func nullUUID(u uuid.NullUUID) string {
    if u.Valid {
        return u.UUID.String()
    }
    return ""
}

// --- DTOs ---

type appointmentDTO struct {
    ID          string `json:"id"`
    UserID      string `json:"user_id"`
    Datetime    string `json:"datetime"`
	Title       string `json:"title"`
    Description string `json:"description"`
    Status      string `json:"status"`
    CreatedAt   string `json:"created_at"`
    UserName    string `json:"user_name"`
    UserEmail   string `json:"user_email"`
    UserPhone   string `json:"user_phone"`
}

// For user-specific queries (db.Appointment, no joined user info)
func toApptDTO(a db.Appointment) appointmentDTO {
    return appointmentDTO{
        ID:          a.ID.String(),
        UserID:      nullUUID(a.UserID),
        Datetime:    a.Datetime.Format(time.RFC3339),
		Title:       a.Title,
        Description: nullStr(a.Description),
        Status:      a.Status,
        CreatedAt:   a.CreatedAt.Format(time.RFC3339),
        // No joined fields here
        UserName:  "",
        UserEmail: "",
        UserPhone: "",
    }
}

func toApptSliceDTO(in []db.Appointment) []appointmentDTO {
    out := make([]appointmentDTO, 0, len(in))
    for _, a := range in {
        out = append(out, toApptDTO(a))
    }
    return out
}

// For admin list (db.GetAllAppointmentsRow, includes joined user info)
func toApptDTOAdmin(a db.GetAllAppointmentsRow) appointmentDTO {
    return appointmentDTO{
        ID:          a.ID.String(),
        UserID:      nullUUID(a.UserID),
        Datetime:    a.Datetime.Format(time.RFC3339),
		Title:       a.Title,
        Description: nullStr(a.Description),
        Status:      a.Status,
        CreatedAt:   a.CreatedAt.Format(time.RFC3339),
        UserName:    a.UserName,
        UserEmail:   a.UserEmail,
        UserPhone:   a.UserPhone,
    }
}

func toApptSliceDTOAdmin(in []db.GetAllAppointmentsRow) []appointmentDTO {
    out := make([]appointmentDTO, 0, len(in))
    for _, a := range in {
        out = append(out, toApptDTOAdmin(a))
    }
    return out
}

