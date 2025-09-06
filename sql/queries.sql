-- name: CreateUser :one
INSERT INTO users (id, name, email, password_hash, phone, is_admin)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateAppointment :one
INSERT INTO appointments (id, user_id, datetime, title, description)
VALUES ($1, $2, $3, $5, $4)
RETURNING *;

-- name: DeleteAppointment :exec
DELETE FROM appointments WHERE id = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetAppointmentsForUser :many
SELECT * FROM appointments WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpdateAppointmentStatus :exec
UPDATE appointments SET status = $2 WHERE id = $1;

-- name: UserUpdateAppointment :exec
UPDATE appointments SET datetime = $2, title = $3, description = $4 WHERE user_id = $5 AND id = $1;

-- name: SetAdmin :exec
UPDATE users SET is_admin = $2 WHERE email = $1;

-- name: GetAllAppointments :many
SELECT
  a.id,
  a.user_id,
  a.datetime,
  a.title,
  a.description,
  a.status,
  a.created_at,
  u.name AS user_name,
  u.email AS user_email,
  u.phone AS user_phone
FROM appointments a
JOIN users u ON a.user_id = u.id
ORDER BY a.created_at DESC;

-- name: GetAppointmentsByID :one
SELECT id, user_id, datetime, title, description, status, created_at
FROM appointments
WHERE id = $1;
