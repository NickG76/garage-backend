-- name: CreateUser :one
INSERT INTO users (id, name, email, password_hash, phone, is_admin)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateAppointment :one
INSERT INTO appointments (id, user_id, datetime, description)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAppointmentsForUser :many
SELECT * FROM appointments WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetAllAppointments :many
SELECT * FROM appointments ORDER BY created_at DESC;

-- name: UpdateAppointmentStatus :exec
UPDATE appointments SET status = $2 WHERE id = $1;
