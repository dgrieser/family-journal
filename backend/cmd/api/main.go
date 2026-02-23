package main

import (
	"crypto/sha256"
	"encoding/base64"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/mysql/v2"
	"github.com/joho/godotenv"
	"github.com/user/family-journal/internal/handlers"
	"github.com/user/family-journal/internal/middleware"
	"github.com/user/family-journal/internal/models"
	"github.com/user/family-journal/internal/repository"
	"github.com/user/family-journal/internal/services"
)

func main() {
	godotenv.Load()

	db, err := repository.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto Migration
	if os.Getenv("AUTO_MIGRATE") == "true" {
		log.Println("Running database migrations...")
		err = db.AutoMigrate(
			&models.User{},
			&models.Post{},
			&models.Comment{},
			&models.Hashtag{},
			&models.Attachment{},
		)
		if err != nil {
			log.Fatal("Failed to migrate database:", err)
		}
	} else {
		log.Println("Skipping database migrations (set AUTO_MIGRATE=true to enable)")
	}

	dbPortStr := os.Getenv("DB_PORT")
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		if dbPortStr != "" {
			log.Printf("Warning: Invalid DB_PORT '%s', using default 3306", dbPortStr)
		}
		dbPort = 3306
	}

	// Session storage
	storage := mysql.New(mysql.Config{
		Host:       os.Getenv("DB_HOST"),
		Port:       dbPort,
		Username:   os.Getenv("DB_USER"),
		Password:   os.Getenv("DB_PASSWORD"),
		Database:   os.Getenv("DB_NAME"),
		Table:      "sessions",
		GCInterval: 10 * time.Second,
	})

	isSecure := os.Getenv("COOKIE_SECURE") == "true"
	store := session.New(session.Config{
		Storage:        storage,
		CookieHTTPOnly: true,
		CookieSecure:   isSecure,
		CookieSameSite: "Lax",
		Expiration:     24 * time.Hour,
	})

	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB limit
	})

	app.Use(logger.New())
	allowOrigins := os.Getenv("CORS_ALLOW_ORIGIN")
	if allowOrigins == "" {
		allowOrigins = "http://localhost:3000"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowCredentials: true,
	}))

	// Encrypt Cookies
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		log.Fatal("SESSION_SECRET environment variable is required")
	}
	if len(secret) < 32 {
		log.Fatal("SESSION_SECRET environment variable must be at least 32 characters long")
	}
	secretHash := sha256.Sum256([]byte(secret))
	cookieKey := base64.StdEncoding.EncodeToString(secretHash[:])
	app.Use(encryptcookie.New(encryptcookie.Config{
		Key:    cookieKey,
		Except: []string{"csrf_"}, // Allow frontend to read CSRF token
	}))

	// CSRF Protection
	app.Use(csrf.New(csrf.Config{
		KeyLookup:      "header:X-Csrf-Token",
		CookieName:     "csrf_",
		CookieHTTPOnly: false,
		CookieSameSite: "Lax",
		CookieSecure:   isSecure,
		Expiration:     1 * time.Hour,
		ContextKey:     "csrf",
	}))

	// Repositories
	userRepo := repository.NewUserRepository(db)
	personRepo := repository.NewPersonRepository(db)
	postRepo := repository.NewPostRepository(db)

	// Services
	authService := services.NewAuthService(userRepo)
	postService := services.NewPostService(postRepo, personRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService, store)
	personHandler := handlers.NewPersonHandler(personRepo)
	postHandler := handlers.NewPostHandler(postService)
	adminHandler := handlers.NewAdminHandler(userRepo)

	api := app.Group("/api/v1")

	// Auth rate limiting
	authLimiter := limiter.New(limiter.Config{
		Max:        20,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})

	// Public auth routes
	api.Post("/auth/register", authLimiter, authHandler.Register)
	api.Post("/auth/login", authLimiter, authHandler.Login)
	api.Post("/auth/logout", authHandler.Logout)

	// Protected routes
	protected := api.Use(middleware.AuthRequired(store, authService))
	protected.Get("/auth/profile", authHandler.Me)
	protected.Put("/auth/profile", authHandler.UpdateProfile)

	// Persons
	protected.Get("/persons", personHandler.GetAll)
	protected.Post("/persons", personHandler.Create)
	protected.Put("/persons/:id", personHandler.Update)
	protected.Delete("/persons/:id", personHandler.Delete)

	// Posts
	protected.Get("/posts", postHandler.GetPosts)
	protected.Get("/posts/:id", postHandler.GetPost)
	protected.Post("/posts", postHandler.Create)
	protected.Put("/posts/:id", postHandler.Update)
	protected.Delete("/posts/:id", postHandler.Delete)

	// Comments
	protected.Post("/posts/:id/comments", postHandler.AddComment)
	protected.Delete("/comments/:id", postHandler.DeleteComment)

	// Attachments
	protected.Get("/attachments/:id/download", postHandler.DownloadAttachment)

	// Hashtags
	protected.Get("/hashtags", postHandler.GetHashtags)

	// Admin routes
	admin := protected.Use(middleware.AdminRequired())
	admin.Get("/admin/users", adminHandler.GetAllUsers)
	admin.Put("/admin/users/:id/role", adminHandler.UpdateUserRole)
	admin.Put("/admin/users/:id/active", adminHandler.ToggleUserActive)

	// Note: Uploads are served via protected /attachments/:id/download endpoint

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Listen from a different goroutine
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Panic(err)
		}
	}()

	// Create channel for idle connections.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c // This blocks the main thread until an interrupt is received
	log.Println("Gracefully shutting down...")
	_ = app.Shutdown()

	log.Println("Fiber was successful shutdown.")
}
