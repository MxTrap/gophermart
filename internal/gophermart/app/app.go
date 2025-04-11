package app

import (
	"context"

	"github.com/MxTrap/gophermart/config"

	"github.com/MxTrap/gophermart/internal/gophermart/controller/http"
	handlers "github.com/MxTrap/gophermart/internal/gophermart/controller/http/handlers"
	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/middlewares"
	"github.com/MxTrap/gophermart/internal/gophermart/migrator"
	"github.com/MxTrap/gophermart/internal/gophermart/repository/postgres"
	token_repo "github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/token"
	user_repo "github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/user"
	"github.com/MxTrap/gophermart/internal/gophermart/services"
	"github.com/MxTrap/gophermart/logger"
)

type App struct {
	pgStorage      *postgres.Storage
	httpController *http.Controller
}

func NewApp(ctx context.Context, log *logger.Logger, cfg *config.Config) (*App, error) {
	storage, err := postgres.NewPostgresStorage(ctx, cfg.Postgres.GetConnectionString())
	if err != nil {
		return nil, err
	}

	mgrtr, err := migrator.NewMigrator(storage.Pool)
	if err != nil {
		return nil, err
	}
	err = mgrtr.InitializeDB()
	if err != nil {
		return nil, err
	}

	tokenRepo := token_repo.NewTokenRepo(storage.Pool)
	userRepo := user_repo.NewUserRepository(storage.Pool)

	jwtService := services.NewJWTService(cfg.JWTSecret)
	authService := services.NewAuthService(log, userRepo, tokenRepo, jwtService, services.TokenConfig(cfg.Token))

	httpController := http.NewController(cfg.HTTP.GetAddress())
	httpController.RegisterMiddlewares(middlewares.LoggerMiddleware(log))
	authHandler := handlers.NewAuthHandler(authService)
	ordersHandler := handlers.NewOrdersHandler(middlewares.NewAuhtorizationMiddleware(jwtService))

	httpController.AddHandler("/user", authHandler, ordersHandler)

	return &App{pgStorage: storage, httpController: httpController}, nil
}

func (a *App) Run(ctx context.Context) error {
	return a.httpController.Start()
}

func (a *App) Stop(ctx context.Context) error {
	err := a.httpController.Stop(ctx)
	if err != nil {
		return err
	}
	a.pgStorage.Stop()

	return nil
}
