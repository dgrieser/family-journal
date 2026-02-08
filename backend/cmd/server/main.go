package main

import (
	"log"
	"os"
	"strings"
	"time"

	"familyjournal/backend/internal/config"
	"familyjournal/backend/internal/db"
	"familyjournal/backend/internal/handlers"
	"familyjournal/backend/internal/middleware"
	"familyjournal/backend/internal/models"
	"familyjournal/backend/internal/repositories"
	"familyjournal/backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	mysqlstorage "github.com/gofiber/storage/mysql/v2"
)

func main() {
	cfg := config.Load()
	if cfg.SessionSecret == "" {
		log.Fatal("SESSION_SECRET must be set")
	}
	if cfg.DatabaseDSN == "" {
		log.Fatal("MYSQL_DSN must be set")
	}
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatal(err)
	}
	database, err := db.New(cfg.DatabaseDSN, db.Config{
		MaxOpen:     cfg.DBMaxOpen,
		MaxIdle:     cfg.DBMaxIdle,
		MaxLifetime: time.Duration(cfg.DBMaxLifetime) * time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}
	migrationsDir := "./migrations"
	if _, err := os.Stat(migrationsDir); err != nil {
		migrationsDir = "/app/migrations"
	}
	if err := db.RunMigrations(database, migrationsDir); err != nil {
		log.Fatal(err)
	}
	repo := repositories.New(database)
	service := services.New(repo, repo, repo, repo, repo, repo)

	storage := mysqlstorage.New(mysqlstorage.Config{
		DSN:        cfg.DatabaseDSN,
		Table:      "sessions",
		GCInterval: 1 * time.Hour,
	})

	store := session.New(session.Config{
		CookieHTTPOnly: true,
		CookieSecure:   cfg.CookieSecure,
		CookieSameSite: "Lax",
		Expiration:     24 * time.Hour,
		KeyLookup:      "cookie:fj_session",
		Storage:        storage,
	})

	app := fiber.New()
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Use(recover.New())
	app.Use(logger.New())
	if cfg.RateLimitMax > 0 {
		app.Use(limiter.New(limiter.Config{
			Max:        cfg.RateLimitMax,
			Expiration: time.Duration(cfg.RateLimitTTL) * time.Second,
			KeyGenerator: func(c *fiber.Ctx) string {
				forwardedFor := c.Get(fiber.HeaderXForwardedFor)
				if forwardedFor != "" {
					parts := strings.Split(forwardedFor, ",")
					if len(parts) > 0 {
						clientIP := strings.TrimSpace(parts[0])
						if clientIP != "" {
							return clientIP
						}
					}
				}
				realIP := strings.TrimSpace(c.Get("X-Real-IP"))
				if realIP != "" {
					return realIP
				}
				return c.IP()
			},
		}))
	}
	app.Use(csrf.New(csrf.Config{
		KeyLookup:      "header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookieSecure:   cfg.CookieSecure,
		CookieSameSite: "Lax",
		Expiration:     24 * time.Hour,
	}))

	api := app.Group("/api/v1")
	authHandler := &handlers.AuthHandler{Service: service, Store: store}
	adminHandler := &handlers.AdminHandler{Service: service}
	postsHandler := &handlers.PostsHandler{
		Service:      service,
		Store:        store,
		UploadDir:    cfg.UploadDir,
		MaxUploadMB:  cfg.MaxUploadMB,
		AllowedTypes: cfg.AllowedTypes,
	}
	personsHandler := &handlers.PersonsHandler{Service: service, Store: store}

	app.Get("/uploads/:name", middleware.RequireAuth(store), postsHandler.DownloadAttachment)

	api.Post("/auth/register", authHandler.Register)
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/logout", authHandler.Logout)
	api.Get("/auth/profile", middleware.RequireAuth(store), authHandler.Profile)
	api.Put("/auth/profile", middleware.RequireAuth(store), authHandler.UpdateProfile)

	admin := api.Group("/admin", middleware.RequireAuth(store), middleware.RequireRole(store, models.RoleAdmin))
	admin.Get("/users", adminHandler.ListUsers)
	admin.Patch("/users/:id/role", adminHandler.UpdateRole)
	admin.Patch("/users/:id/active", adminHandler.UpdateActive)

	posts := api.Group("/posts", middleware.RequireAuth(store))
	posts.Get("/", postsHandler.List)
	posts.Post("/", postsHandler.Create)
	posts.Get("/:id", postsHandler.Get)
	posts.Put("/:id", postsHandler.Update)
	posts.Delete("/:id", postsHandler.Delete)
	posts.Post("/:id/comments", postsHandler.AddComment)
	posts.Post("/:id/attachments", postsHandler.UploadAttachment)

	api.Put("/comments/:id", middleware.RequireAuth(store), postsHandler.UpdateComment)
	api.Delete("/comments/:id", middleware.RequireAuth(store), postsHandler.DeleteComment)

	api.Get("/hashtags", middleware.RequireAuth(store), postsHandler.ListHashtags)
	api.Get("/persons", middleware.RequireAuth(store), personsHandler.List)
	api.Post("/persons", middleware.RequireAuth(store), personsHandler.Create)
	api.Put("/persons/:id", middleware.RequireAuth(store), personsHandler.Update)
	api.Delete("/persons/:id", middleware.RequireAuth(store), personsHandler.Delete)

	log.Fatal(app.Listen(":" + cfg.Port))
}
