package main

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	appbuilder "familyjournal/backend/internal/app"
	"familyjournal/backend/internal/config"
	"familyjournal/backend/internal/db"
	"familyjournal/backend/internal/models"
	"familyjournal/backend/internal/repositories"
	"familyjournal/backend/internal/services"
	"familyjournal/backend/internal/sessionstore"

	"github.com/gofiber/fiber/v2/middleware/session"
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
	if cfg.AdminEmail != "" {
		user, err := repo.GetUserByEmail(cfg.AdminEmail)
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("admin promotion: no user with email %q exists, skipping", cfg.AdminEmail)
		} else if err != nil {
			log.Printf("admin promotion: failed to retrieve user %q: %v", cfg.AdminEmail, err)
		} else if user.Role == models.RoleAdmin {
			log.Printf("admin promotion: user %q is already admin", cfg.AdminEmail)
		} else {
			if err := repo.UpdateUserRole(user.ID, models.RoleAdmin); err != nil {
				log.Printf("admin promotion: failed to promote %q: %v", cfg.AdminEmail, err)
			} else {
				log.Printf("admin promotion: promoted %q to admin", cfg.AdminEmail)
			}
		}
	}
	service := services.New(repo, repo, repo, repo, repo, repo)
	storage := sessionstore.NewMySQLStore(database)

	store := session.New(session.Config{
		Storage:        storage,
		CookieHTTPOnly: true,
		CookieSecure:   cfg.CookieSecure,
		CookieSameSite: "Lax",
		Expiration:     24 * time.Hour,
		KeyLookup:      "cookie:fj_session",
	})

	app := appbuilder.New(cfg, service, store)

	listenErrCh := make(chan error, 1)
	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			listenErrCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
	case sig := <-sigCh:
		log.Printf("shutdown signal received: %s", sig.String())
	case err := <-listenErrCh:
		log.Printf("server stopped: %v", err)
	}
	if err := app.Shutdown(); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	if err := storage.Close(); err != nil {
		log.Printf("session store close error: %v", err)
	}
	if err := database.Close(); err != nil {
		log.Printf("db close error: %v", err)
	}
}
