package handler

import (
	"errors"
	"strconv"

	"github.com/ainyx/user-api/internal/logger"
	"github.com/ainyx/user-api/internal/models"
	"github.com/ainyx/user-api/internal/repository"
	"github.com/ainyx/user-api/internal/service"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// UserHandler holds the HTTP handlers for user CRUD operations.
type UserHandler struct {
	service *service.UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{service: svc}
}

// CreateUser handles POST /users.
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Log.Warn("Failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	resp, err := h.service.CreateUser(c.Context(), req)
	if err != nil {
		return handleServiceError(c, "create user", err)
	}

	logger.Log.Info("User created",
		zap.Int32("id", resp.ID),
		zap.String("name", resp.Name),
	)
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// GetUser handles GET /users/:id.
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	resp, err := h.service.GetUser(c.Context(), id)
	if err != nil {
		return handleServiceError(c, "get user", err)
	}

	return c.JSON(resp)
}

// UpdateUser handles PUT /users/:id.
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Log.Warn("Failed to parse request body", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	resp, err := h.service.UpdateUser(c.Context(), id, req)
	if err != nil {
		return handleServiceError(c, "update user", err)
	}

	logger.Log.Info("User updated", zap.Int32("id", id))
	return c.JSON(resp)
}

// DeleteUser handles DELETE /users/:id.
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	if err := h.service.DeleteUser(c.Context(), id); err != nil {
		return handleServiceError(c, "delete user", err)
	}

	logger.Log.Info("User deleted", zap.Int32("id", id))
	return c.SendStatus(fiber.StatusNoContent)
}

// ListUsers handles GET /users with optional pagination query params.
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	resp, err := h.service.ListUsers(c.Context(), page, pageSize)
	if err != nil {
		logger.Log.Error("Failed to list users", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list users",
		})
	}

	return c.JSON(resp)
}

// --- helpers ---

// parseID extracts and validates the :id path parameter.
func parseID(c *fiber.Ctx) (int32, error) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return 0, err
	}
	return int32(id), nil
}

// handleServiceError maps domain/service errors to the correct HTTP status.
func handleServiceError(c *fiber.Ctx, action string, err error) error {
	// Validation errors → 400
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		logger.Log.Warn("Validation failed", zap.String("action", action), zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": ve.Error(),
		})
	}

	// DOB in the future → 400
	if errors.Is(err, service.ErrDOBInFuture) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Not found → 404
	if errors.Is(err, repository.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Everything else → 500
	logger.Log.Error("Failed to "+action, zap.Error(err))
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "Failed to " + action,
	})
}
