package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ainyx/user-api/config"
	"github.com/ainyx/user-api/db/sqlc"
	"github.com/ainyx/user-api/internal/handler"
	"github.com/ainyx/user-api/internal/logger"
	"github.com/ainyx/user-api/internal/middleware"
	"github.com/ainyx/user-api/internal/repository"
	"github.com/ainyx/user-api/internal/routes"
	"github.com/ainyx/user-api/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Load .env file if present (local dev convenience; ignored in Docker).
	_ = godotenv.Load()

	// Load configuration from environment variables.
	cfg := config.Load()

	// Initialize structured logger.
	logger.Init()
	defer logger.Sync()

	// Connect to PostgreSQL.
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL())
	if err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Log.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Log.Info("Connected to database successfully")

	// Auto-apply migrations on startup.
	if err := runMigrations(ctx, pool); err != nil {
		logger.Log.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Wire up the dependency chain:
	//   SQLC Queries → Repository → Service → Handler
	queries := sqlc.New(pool)
	userRepo := repository.NewUserRepository(queries)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// Create Fiber app.
	app := fiber.New(fiber.Config{
		AppName: "User API v1.0.0",
	})

	// Global middleware.
	app.Use(cors.New())
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogger())

	// Health check endpoint.
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Register API routes.
	routes.Setup(app, userHandler)

	// Graceful shutdown on SIGINT / SIGTERM.
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Log.Info("Shutting down server...")
		_ = app.Shutdown()
	}()

	// Start listening.
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	logger.Log.Info("Starting server", zap.String("address", addr))
	if err := app.Listen(addr); err != nil {
		logger.Log.Fatal("Server failed to start", zap.Error(err))
	}
}

// runMigrations applies the users table schema idempotently.
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migration := `
	CREATE TABLE IF NOT EXISTS users (
		id   SERIAL PRIMARY KEY,
		name TEXT   NOT NULL,
		dob  DATE   NOT NULL
	);`

	if _, err := pool.Exec(ctx, migration); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	logger.Log.Info("Database migrations applied successfully")
	return nil
}
