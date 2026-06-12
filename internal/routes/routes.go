package routes

import (
	"github.com/ainyx/user-api/internal/handler"
	"github.com/gofiber/fiber/v2"
)

// Setup registers all user routes on the Fiber app.
func Setup(app *fiber.App, userHandler *handler.UserHandler) {
	users := app.Group("/users")

	users.Post("/", userHandler.CreateUser)
	users.Get("/", userHandler.ListUsers)
	users.Get("/:id", userHandler.GetUser)
	users.Put("/:id", userHandler.UpdateUser)
	users.Delete("/:id", userHandler.DeleteUser)
}
