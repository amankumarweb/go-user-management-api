package repository

import (
	"context"
	"errors"

	"github.com/ainyx/user-api/db/sqlc"
	"github.com/jackc/pgx/v5"
)

// ErrNotFound is returned when a database lookup yields no rows.
var ErrNotFound = errors.New("record not found")

// UserRepository wraps the SQLC-generated Queries and translates
// database-specific errors (e.g. pgx.ErrNoRows) into domain errors.
type UserRepository struct {
	queries *sqlc.Queries
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(queries *sqlc.Queries) *UserRepository {
	return &UserRepository{queries: queries}
}

// Create inserts a new user and returns the created record.
func (r *UserRepository) Create(ctx context.Context, arg sqlc.CreateUserParams) (sqlc.User, error) {
	return r.queries.CreateUser(ctx, arg)
}

// GetByID fetches a single user by primary key.
// Returns ErrNotFound when the id does not exist.
func (r *UserRepository) GetByID(ctx context.Context, id int32) (sqlc.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, ErrNotFound
		}
		return sqlc.User{}, err
	}
	return user, nil
}

// List returns a paginated slice of users.
func (r *UserRepository) List(ctx context.Context, arg sqlc.ListUsersParams) ([]sqlc.User, error) {
	return r.queries.ListUsers(ctx, arg)
}

// Update modifies an existing user and returns the updated record.
// Returns ErrNotFound when the id does not exist.
func (r *UserRepository) Update(ctx context.Context, arg sqlc.UpdateUserParams) (sqlc.User, error) {
	user, err := r.queries.UpdateUser(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, ErrNotFound
		}
		return sqlc.User{}, err
	}
	return user, nil
}

// Delete removes a user by id.
// Returns ErrNotFound when no row was affected.
func (r *UserRepository) Delete(ctx context.Context, id int32) error {
	rowsAffected, err := r.queries.DeleteUser(ctx, id)
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Count returns the total number of users in the table.
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountUsers(ctx)
}
