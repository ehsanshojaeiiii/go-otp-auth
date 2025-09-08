package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ehsanshojaei/go-otp-auth/internal/config"
	"github.com/ehsanshojaei/go-otp-auth/internal/handler"
	"github.com/ehsanshojaei/go-otp-auth/internal/middleware"
	"github.com/ehsanshojaei/go-otp-auth/internal/model"
	"github.com/ehsanshojaei/go-otp-auth/internal/repository"
	"github.com/ehsanshojaei/go-otp-auth/internal/service"
	"github.com/ehsanshojaei/go-otp-auth/pkg/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// @title OTP Service API
// @version 1.0
// @description A service for OTP-based authentication and user management
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Redis
	redisClient := initRedis(cfg)

	// Initialize JWT manager
	jwtManager := jwt.NewJWTManager(cfg.JWT.SecretKey, cfg.JWT.ExpiryHours)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	otpRepo := repository.NewOTPRepository(redisClient)

	// Initialize services
	authService := service.NewAuthService(userRepo, otpRepo, jwtManager, cfg)
	userService := service.NewUserService(userRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Initialize Fiber app
	app := setupApp(authHandler, userHandler, authMiddleware, db, redisClient)

	// Start server with graceful shutdown
	go func() {
		log.Printf("Server starting on %s", cfg.ServerAddr())
		if err := app.Listen(cfg.ServerAddr()); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DatabaseDSN()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate
	if err := db.AutoMigrate(&model.User{}); err != nil {
		return nil, err
	}

	log.Println("Database connected and migrated successfully")
	return db, nil
}

func initRedis(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")
	return client
}

func setupApp(authHandler *handler.AuthHandler, userHandler *handler.UserHandler, authMiddleware *middleware.AuthMiddleware, db *gorm.DB, redisClient *redis.Client) *fiber.App {
	// Create Fiber app with custom configuration
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error":   "internal_server_error",
				"message": err.Error(),
			})
		},
		ServerHeader: "OTP-Service",
		AppName:      "OTP Service v1.0",
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(limiter.New(limiter.Config{
		Max:        100, // 100 requests per minute per IP
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests from this IP",
			})
		},
	}))
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${latency} - ${ip}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://127.0.0.1:3000",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
	}))

	// Health check endpoint with dependency checks
	app.Get("/health", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		status := fiber.Map{
			"status":  "healthy",
			"service": "OTP Service",
			"version": "1.0",
			"checks": fiber.Map{
				"database": "healthy",
				"redis":    "healthy",
			},
		}

		// Check database connection
		if sqlDB, err := db.DB(); err != nil || sqlDB.PingContext(ctx) != nil {
			status["status"] = "unhealthy"
			status["checks"].(fiber.Map)["database"] = "unhealthy"
		}

		// Check Redis connection
		if err := redisClient.Ping(ctx).Err(); err != nil {
			status["status"] = "unhealthy"
			status["checks"].(fiber.Map)["redis"] = "unhealthy"
		}

		statusCode := fiber.StatusOK
		if status["status"] == "unhealthy" {
			statusCode = fiber.StatusServiceUnavailable
		}

		return c.Status(statusCode).JSON(status)
	})

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)

	// API routes
	v1 := app.Group("/api/v1")

	// Auth routes (no authentication required)
	auth := v1.Group("/auth")
	auth.Post("/send-otp", authHandler.SendOTP)
	auth.Post("/verify-otp", authHandler.VerifyOTP)

	// User routes (authentication required)
	users := v1.Group("/users")
	users.Use(authMiddleware.RequireAuth())
	users.Get("/profile", userHandler.GetProfile)
	users.Get("/", userHandler.GetUsers)
	users.Get("/:id", userHandler.GetUser)

	return app
}
