package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	handler "pullrequest-service/internal/api/http/handlers"
	"pullrequest-service/internal/api/http/router"
	"pullrequest-service/internal/config"
	"pullrequest-service/internal/repository/postgres"
	"pullrequest-service/internal/usecase"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

func applySchema(db *sql.DB, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	_, err = db.Exec(string(data))
	return err
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("starting service")

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger.Info("config loaded")

	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		logger.Error("failed to open postgres", "error", err)
		os.Exit(1)
	}

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping postgres", "error", err)
		os.Exit(1)
	}

	logger.Info("connected to postgres")

	if err := applySchema(db, "./sql/schema.sql"); err != nil {
		logger.Error("failed to apply schema", "error", err)
		os.Exit(1)
	}

	logger.Info("schema applied")

	teamRepo := postgres.NewPostgresTeamRepository(db)
	userRepo := postgres.NewPostgresUserRepository(db)
	prRepo := postgres.NewPostgresPRRepository(db)
	txMgr := postgres.NewTxManager(db)

	teamUsecase := usecase.NewTeamUsecase(teamRepo, userRepo, txMgr, logger)
	userUsecase := usecase.NewUserUsecase(userRepo, prRepo, logger)
	prUsecase := usecase.NewPRUsecase(prRepo, userRepo, teamRepo, txMgr, logger)

	teamHandler := handler.NewTeamHandler(teamUsecase)
	userHandler := handler.NewUserHandler(userUsecase)
	prHandler := handler.NewPRHandler(prUsecase)

	r := router.NewRouter(teamHandler, userHandler, prHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		logger.Info("server started", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server crashed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	logger.Warn("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	} else {
		logger.Info("server stopped cleanly")
	}

}
