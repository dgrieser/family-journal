package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	appbuilder "familyjournal/backend/internal/app"
	"familyjournal/backend/internal/config"
	"familyjournal/backend/internal/db"
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
