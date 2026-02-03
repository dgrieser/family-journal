package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
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
	log.Println("Running database migrations...")
	err = db.AutoMigrate(
		&models.User{},
		&models.Person{},
		&models.Post{},
		&models.Comment{},
		&models.Hashtag{},
		&models.Attachment{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	if dbPort == 0 {
		dbPort = 3306
	}

	// Session storage
	storage := mysql.New(mysql.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     dbPort,
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
		Table:    "sessions",
		GCInterval: 10 * time.Second,
	})

	store := session.New(session.Config{
		Storage:        storage,
		CookieHTTPOnly: true,
		Expiration:     24 * time.Hour,
	})

	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB limit
	})

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowCredentials: true,
	}))

	// Encrypt Cookies
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		log.Fatal("SESSION_SECRET environment variable is required")
	}
	if len(secret) != 32 {
		log.Fatal("SESSION_SECRET environment variable must be 32 characters long")
	}
	app.Use(encryptcookie.New(encryptcookie.Config{
		Key:    secret,
		Except: []string{"csrf_"}, // Allow frontend to read CSRF token
	}))

	// CSRF Protection
	app.Use(csrf.New(csrf.Config{
		KeyLookup:      "header:X-Csrf-Token",
		CookieName:     "csrf_",
		CookieHTTPOnly: false,
		CookieSameSite: "Lax",
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

	api := app.Group("/api")

	// Public routes
	api.Post("/register", authHandler.Register)
	api.Post("/login", authHandler.Login)
	api.Post("/logout", authHandler.Logout)

	// Protected routes
	protected := api.Use(middleware.AuthRequired(store, authService))
	protected.Get("/me", authHandler.Me)
	protected.Put("/me", authHandler.UpdateProfile)

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

	// Serve static files for uploads
	app.Static("/uploads", "./uploads")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}
