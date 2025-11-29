package rdb

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// User represents a user in the system
type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash string
	Email        string
	CreatedAt    time.Time
	ModifiedAt   time.Time
	DeletedAt    sql.NullTime
}

// Role represents a role that can be assigned to users
type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
	ModifiedAt  time.Time
	DeletedAt   sql.NullTime
}

// UserRole represents the binding between a user and a role
type UserRole struct {
	UserID    uuid.UUID
	RoleID    uuid.UUID
	CreatedAt time.Time
}

// --- User methods ---

// GetAll retrieves all active users
func (u *User) GetAll(ctx context.Context) ([]User, error) {
	rows, err := db.Query(ctx, `
		SELECT id, username, password_hash, email, created_at, modified_at 
		FROM bind_dns.users 
		WHERE deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.CreatedAt, &user.ModifiedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// FindByID retrieves a user by ID
func (u *User) FindByID(ctx context.Context) error {
	row := db.QueryRow(ctx, `
		SELECT id, username, password_hash, email, created_at, modified_at, deleted_at 
		FROM bind_dns.users 
		WHERE id = $1
	`, u.ID)
	return row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.CreatedAt, &u.ModifiedAt, &u.DeletedAt)
}

// FindByUsername retrieves a user by username
func (u *User) FindByUsername(ctx context.Context) error {
	row := db.QueryRow(ctx, `
		SELECT id, username, password_hash, email, created_at, modified_at, deleted_at 
		FROM bind_dns.users 
		WHERE username = $1 AND deleted_at IS NULL
	`, u.Username)
	return row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.CreatedAt, &u.ModifiedAt, &u.DeletedAt)
}

// Create inserts a new user into the database
func (u *User) Create(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	u.ID = uuid.New()
	now := time.Now()
	u.CreatedAt = now
	u.ModifiedAt = now

	result, err := tx.Exec(ctx, `
		INSERT INTO bind_dns.users (id, username, password_hash, email, created_at, modified_at) 
		VALUES ($1, $2, $3, $4, $5, $6)
	`, u.ID, u.Username, u.PasswordHash, u.Email, u.CreatedAt, u.ModifiedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotCreated
	}
	return tx.Commit(ctx)
}

// Update modifies an existing user
func (u *User) Update(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	u.ModifiedAt = time.Now()

	result, err := tx.Exec(ctx, `
		UPDATE bind_dns.users 
		SET username = $1, password_hash = $2, email = $3, modified_at = $4 
		WHERE id = $5 AND deleted_at IS NULL
	`, u.Username, u.PasswordHash, u.Email, u.ModifiedAt, u.ID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}

// Delete soft-deletes a user
func (u *User) Delete(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `
		UPDATE bind_dns.users 
		SET deleted_at = NOW() 
		WHERE id = $1 AND deleted_at IS NULL
	`, u.ID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}

// GetRoles retrieves all roles assigned to a user
func (u *User) GetRoles(ctx context.Context) ([]Role, error) {
	rows, err := db.Query(ctx, `
		SELECT r.id, r.name, r.description, r.created_at, r.modified_at 
		FROM bind_dns.roles r
		INNER JOIN bind_dns.user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.deleted_at IS NULL
	`, u.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.ModifiedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// --- Role methods ---

// GetAll retrieves all active roles
func (r *Role) GetAll(ctx context.Context) ([]Role, error) {
	rows, err := db.Query(ctx, `
		SELECT id, name, description, created_at, modified_at 
		FROM bind_dns.roles 
		WHERE deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.ModifiedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// FindByID retrieves a role by ID
func (r *Role) FindByID(ctx context.Context) error {
	row := db.QueryRow(ctx, `
		SELECT id, name, description, created_at, modified_at, deleted_at 
		FROM bind_dns.roles 
		WHERE id = $1
	`, r.ID)
	return row.Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt, &r.ModifiedAt, &r.DeletedAt)
}

// FindByName retrieves a role by name
func (r *Role) FindByName(ctx context.Context) error {
	row := db.QueryRow(ctx, `
		SELECT id, name, description, created_at, modified_at, deleted_at 
		FROM bind_dns.roles 
		WHERE name = $1 AND deleted_at IS NULL
	`, r.Name)
	return row.Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt, &r.ModifiedAt, &r.DeletedAt)
}

// Create inserts a new role into the database
func (r *Role) Create(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	r.ID = uuid.New()
	now := time.Now()
	r.CreatedAt = now
	r.ModifiedAt = now

	result, err := tx.Exec(ctx, `
		INSERT INTO bind_dns.roles (id, name, description, created_at, modified_at) 
		VALUES ($1, $2, $3, $4, $5)
	`, r.ID, r.Name, r.Description, r.CreatedAt, r.ModifiedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotCreated
	}
	return tx.Commit(ctx)
}

// Update modifies an existing role
func (r *Role) Update(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	r.ModifiedAt = time.Now()

	result, err := tx.Exec(ctx, `
		UPDATE bind_dns.roles 
		SET name = $1, description = $2, modified_at = $3 
		WHERE id = $4 AND deleted_at IS NULL
	`, r.Name, r.Description, r.ModifiedAt, r.ID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}

// Delete soft-deletes a role
func (r *Role) Delete(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `
		UPDATE bind_dns.roles 
		SET deleted_at = NOW() 
		WHERE id = $1 AND deleted_at IS NULL
	`, r.ID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}

// --- UserRole methods ---

// AssignRole assigns a role to a user
func (ur *UserRole) Create(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	ur.CreatedAt = time.Now()

	result, err := tx.Exec(ctx, `
		INSERT INTO bind_dns.user_roles (user_id, role_id, created_at) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`, ur.UserID, ur.RoleID, ur.CreatedAt)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotCreated
	}
	return tx.Commit(ctx)
}

// RevokeRole removes a role from a user
func (ur *UserRole) Delete(ctx context.Context) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx, `
		DELETE FROM bind_dns.user_roles 
		WHERE user_id = $1 AND role_id = $2
	`, ur.UserID, ur.RoleID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return tx.Commit(ctx)
}

// GetUsersByRole retrieves all users with a specific role
func (r *Role) GetUsers(ctx context.Context) ([]User, error) {
	rows, err := db.Query(ctx, `
		SELECT u.id, u.username, u.password_hash, u.email, u.created_at, u.modified_at 
		FROM bind_dns.users u
		INNER JOIN bind_dns.user_roles ur ON u.id = ur.user_id
		WHERE ur.role_id = $1 AND u.deleted_at IS NULL
	`, r.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.CreatedAt, &user.ModifiedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}
