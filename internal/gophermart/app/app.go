package app

import (
	"context"
	"time"

	"github.com/MxTrap/gophermart/config"

	"github.com/MxTrap/gophermart/internal/gophermart/controller/http"
	handlers "github.com/MxTrap/gophermart/internal/gophermart/controller/http/handlers"
	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/middlewares"
	"github.com/MxTrap/gophermart/internal/gophermart/migrator"
	"github.com/MxTrap/gophermart/internal/gophermart/repository/postgres"
	order_repo "github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/order"
	user_repo "github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/user"
	"github.com/MxTrap/gophermart/internal/gophermart/services"
	"github.com/MxTrap/gophermart/internal/gophermart/usecase"
	"github.com/MxTrap/gophermart/logger"
)

type App struct {
	pgStorage      *postgres.Storage
	httpController *http.Controller
}

func NewApp(ctx context.Context, log *logger.Logger, cfg *config.Config) (*App, error) {
	storage, err := postgres.NewPostgresStorage(ctx, cfg.DatabaseDSN)
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

	userRepo := user_repo.NewUserRepository(storage.Pool)
	orderRepo := order_repo.NewOrderRepository(storage.Pool)

	jwtService := services.NewJWTService("very secret")

	accrualService := services.NewWorkerService(log, cfg.AccrualAddress)

	authService := usecase.NewAuthService(log, userRepo, jwtService, 15*time.Hour)
	orderService := usecase.NewOrderService(log, accrualService, orderRepo)

	httpController := http.NewController(cfg.HTTPAdress)
	httpController.RegisterMiddlewares(middlewares.LoggerMiddleware(log))
	authHandler := handlers.NewAuthHandler(authService)
	ordersHandler := handlers.NewOrdersHandler(middlewares.NewAuhtorizationMiddleware(jwtService), orderService)

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
