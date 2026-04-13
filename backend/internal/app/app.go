package app

import (
	"crypto/sha256"
	"encoding/base64"
	"log"
	"strings"
	"time"

	"familyjournal/backend/internal/config"
	"familyjournal/backend/internal/handlers"
	"familyjournal/backend/internal/middleware"
	"familyjournal/backend/internal/models"
	"familyjournal/backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/encryptcookie"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func New(cfg config.Config, service *services.Service, store *session.Store) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: handlers.JSONErrorHandler,
	})
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	app.Use(recover.New())
	app.Use(logger.New())
	if len(cfg.CORSOrigins) > 0 {
		for _, origin := range cfg.CORSOrigins {
			if origin == "*" {
				log.Fatal("CORS_ALLOW_ORIGINS cannot contain '*' when AllowCredentials is true")
			}
		}
		app.Use(cors.New(cors.Config{
			AllowOrigins:     strings.Join(cfg.CORSOrigins, ","),
			AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
			AllowHeaders:     "Origin,Content-Type,Accept,X-CSRF-Token",
			AllowCredentials: true,
		}))
	}
	app.Use(encryptcookie.New(encryptcookie.Config{
		Key: deriveCookieKey(cfg.SessionSecret),
		// Keep CSRF cookie readable by the browser so the SPA can send it in X-CSRF-Token.
		Except: []string{"csrf_"},
	}))
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
		CookieName:     "csrf_",
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

	api.Post("/auth/register", authHandler.Register)
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/logout", authHandler.Logout)
	api.Get("/auth/profile", middleware.RequireAuth(store, service), authHandler.Profile)
	api.Put("/auth/profile", middleware.RequireAuth(store, service), authHandler.UpdateProfile)

	admin := api.Group("/admin", middleware.RequireAuth(store, service), middleware.RequireRole(store, models.RoleAdmin))
	admin.Get("/users", adminHandler.ListUsers)
	admin.Patch("/users/:id/role", adminHandler.UpdateRole)
	admin.Patch("/users/:id/active", adminHandler.UpdateActive)

	posts := api.Group("/posts", middleware.RequireAuth(store, service))
	posts.Get("/", postsHandler.List)
	posts.Post("/", postsHandler.Create)
	posts.Get("/:id", postsHandler.Get)
	posts.Put("/:id", postsHandler.Update)
	posts.Delete("/:id", postsHandler.Delete)
	posts.Post("/:id/comments", postsHandler.AddComment)
	posts.Post("/:id/attachments", postsHandler.UploadAttachment)

	api.Put("/comments/:id", middleware.RequireAuth(store, service), postsHandler.UpdateComment)
	api.Delete("/comments/:id", middleware.RequireAuth(store, service), postsHandler.DeleteComment)

	api.Get("/hashtags", middleware.RequireAuth(store, service), postsHandler.ListHashtags)
	api.Get("/attachments/:id/download", middleware.RequireAuth(store, service), postsHandler.DownloadAttachmentByID)
	api.Delete("/attachments/:id", middleware.RequireAuth(store, service), postsHandler.DeleteAttachment)
	api.Get("/persons", middleware.RequireAuth(store, service), personsHandler.List)
	api.Post("/persons", middleware.RequireAuth(store, service), personsHandler.Create)
	api.Put("/persons/:id", middleware.RequireAuth(store, service), personsHandler.Update)
	api.Delete("/persons/:id", middleware.RequireAuth(store, service), personsHandler.Delete)

	return app
}

func deriveCookieKey(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return base64.StdEncoding.EncodeToString(hash[:])
}
