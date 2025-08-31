// internal/handlers/auth.go
package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"

    "github.com/google/uuid"
    "golang.org/x/crypto/bcrypt"

    "github.com/nickg76/garage-backend/internal/auth"
    "github.com/nickg76/garage-backend/internal/db"
)

type registerReq struct {
    Name     string `json:"name"`
    Email    string `json:"email"`
    Phone    string `json:"phone"`
    Password string `json:"password"`
}

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
    var req registerReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    if req.Email == "" || req.Password == "" || req.Name == "" {
        http.Error(w, "missing fields", http.StatusBadRequest)
        return
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "failed to hash", http.StatusInternalServerError)
        return
    }
    u, err := s.queries.CreateUser(r.Context(), db.CreateUserParams{
        ID:           uuid.New(),
        Name:         req.Name,
        Email:        req.Email,
        PasswordHash: string(hash),
        Phone:        req.Phone,
        IsAdmin:      sql.NullBool{Bool: false, Valid: true},
    })
    if err != nil {
        http.Error(w, "email may already exist", http.StatusConflict)
        return
    }
    json.NewEncoder(w).Encode(map[string]any{
        "id":    u.ID.String(),
        "name":  u.Name,
        "email": u.Email,
    })
}

type loginReq struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}


func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
    var req loginReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid body", http.StatusBadRequest)
        return
    }
    user, err := s.queries.GetUserByEmail(r.Context(), req.Email)
    if err != nil {
        http.Error(w, "invalid credentials", http.StatusUnauthorized)
        return
    }
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        http.Error(w, "invalid credentials", http.StatusUnauthorized)
        return
    }
    isAdmin := false
    if user.IsAdmin.Valid {
        isAdmin = user.IsAdmin.Bool
    }
    token, err := auth.GenerateJWT(user.ID.String(), isAdmin)
    if err != nil {
        http.Error(w, "token error", http.StatusInternalServerError)
        return
    }
	type u struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
		IsAdmin sql.NullBool `json:"is_admin"`
	}
	type userDetails struct {
		User u
	}

	returnedUser := u {
		Name: 	user.Name,
		Email:   user.Email,
		Phone:   user.Phone,
		IsAdmin: user.IsAdmin,
	}

	json.NewEncoder(w).Encode(map[string]any{"token": token, "user": returnedUser })

}

type meResp struct {
	ID 		string `json:"id"`
	Admin 	bool   `json:"admin"`
	Email  string `json:"email,omitempty"`
	Phone  string `json:"phone,omitempty"`
	Name   string `json:"name,omitempty"`
}

func (s *Server) Me(w http.ResponseWriter, r *http.Request) {
	userID, isAdmin := GetUser(r.Context())
	if userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userIDParsed, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user, err := s.queries.GetUserByID(r.Context(), userIDParsed)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	type u struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
		IsAdmin sql.NullBool `json:"is_admin"`
	}
	type userDetails struct {
		User u
	}

	returnedUser := u {
		Name: 	user.Name,
		Email:   user.Email,
		Phone:   user.Phone,
		IsAdmin: user.IsAdmin,
	}

	json.NewEncoder(w).Encode(meResp{
		ID:		userID,
		Admin:	isAdmin,
		Name:	returnedUser.Name,
		Email:	returnedUser.Email,
		Phone:	returnedUser.Phone,
	})
}
