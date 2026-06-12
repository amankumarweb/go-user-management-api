package service

import (
	"context"
	"errors"
	"time"

	"github.com/ainyx/user-api/db/sqlc"
	"github.com/ainyx/user-api/internal/models"
	"github.com/ainyx/user-api/internal/repository"
	"github.com/go-playground/validator/v10"
)

// ErrDOBInFuture is returned when the supplied date of birth is in the future.
var ErrDOBInFuture = errors.New("date of birth cannot be in the future")

// UserService contains all business logic for user operations.
type UserService struct {
	repo     *repository.UserRepository
	validate *validator.Validate
}

// NewUserService creates a UserService with the given repository.
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo:     repo,
		validate: validator.New(),
	}
}

// CreateUser validates input, persists the user, and returns a response DTO.
func (s *UserService) CreateUser(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return models.UserResponse{}, err
	}

	dob, err := time.Parse("2006-01-02", req.DOB)
	if err != nil {
		return models.UserResponse{}, err
	}

	if dob.After(time.Now()) {
		return models.UserResponse{}, ErrDOBInFuture
	}

	user, err := s.repo.Create(ctx, sqlc.CreateUserParams{
		Name: req.Name,
		Dob:  dob,
	})
	if err != nil {
		return models.UserResponse{}, err
	}

	return models.UserResponse{
		ID:   user.ID,
		Name: user.Name,
		DOB:  user.Dob.Format("2006-01-02"),
	}, nil
}

// GetUser fetches a single user and calculates their age dynamically.
func (s *UserService) GetUser(ctx context.Context, id int32) (models.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.UserResponse{}, err
	}

	return models.UserResponse{
		ID:   user.ID,
		Name: user.Name,
		DOB:  user.Dob.Format("2006-01-02"),
		Age:  models.CalculateAge(user.Dob),
	}, nil
}

// ListUsers returns a paginated list of users with their ages.
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) (models.PaginatedResponse, error) {
	offset := (page - 1) * pageSize

	users, err := s.repo.List(ctx, sqlc.ListUsersParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return models.PaginatedResponse{}, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return models.PaginatedResponse{}, err
	}

	userResponses := make([]models.UserResponse, 0, len(users))
	for _, u := range users {
		userResponses = append(userResponses, models.UserResponse{
			ID:   u.ID,
			Name: u.Name,
			DOB:  u.Dob.Format("2006-01-02"),
			Age:  models.CalculateAge(u.Dob),
		})
	}

	return models.PaginatedResponse{
		Users:    userResponses,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// UpdateUser validates input, updates the user, and returns the updated record.
func (s *UserService) UpdateUser(ctx context.Context, id int32, req models.UpdateUserRequest) (models.UserResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return models.UserResponse{}, err
	}

	dob, err := time.Parse("2006-01-02", req.DOB)
	if err != nil {
		return models.UserResponse{}, err
	}

	if dob.After(time.Now()) {
		return models.UserResponse{}, ErrDOBInFuture
	}

	user, err := s.repo.Update(ctx, sqlc.UpdateUserParams{
		ID:   id,
		Name: req.Name,
		Dob:  dob,
	})
	if err != nil {
		return models.UserResponse{}, err
	}

	return models.UserResponse{
		ID:   user.ID,
		Name: user.Name,
		DOB:  user.Dob.Format("2006-01-02"),
	}, nil
}

// DeleteUser removes a user by id.
func (s *UserService) DeleteUser(ctx context.Context, id int32) error {
	return s.repo.Delete(ctx, id)
}
