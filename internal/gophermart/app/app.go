package app

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"time"

	"github.com/MxTrap/gophermart/config"

	"github.com/MxTrap/gophermart/internal/gophermart/controller/http"
	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/handlers"
	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/middlewares"
	"github.com/MxTrap/gophermart/internal/gophermart/migrator"
	"github.com/MxTrap/gophermart/internal/gophermart/repository/postgres"
	balance_repo "github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/balance"
	"github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/combined"
	order_repo "github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/order"
	user_repo "github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/user"
	withdrawal_repo "github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/withdrawal"
	"github.com/MxTrap/gophermart/internal/gophermart/services"
	"github.com/MxTrap/gophermart/logger"
)

type App struct {
	pgStorage      *postgres.Storage
	httpController *http.Controller
	orderWorker    *services.OrderWorkerService
	logger         *logger.Logger
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
	balanceRepo := balance_repo.NewBalanceRepository(storage.Pool)
	withdrawalRepo := withdrawal_repo.NewWithdrawnRepo(storage.Pool)
	orderBalanceRepo := combined.NewOrderBalanceRepo(storage.Pool, orderRepo, balanceRepo)
	balanceWithdrawalRepo := combined.NewBalanceWithdrawnRepo(storage.Pool, balanceRepo, withdrawalRepo)

	storageSvc := services.NewStorageService()
	jwtSvc := services.NewJWTService("very secret")
	orderSvc := services.NewOrderService(log, storageSvc, orderRepo)
	balanceSvc := services.NewBalanceService(log, balanceRepo)
	accrualSvc := services.NewWorkerService(log, cfg.AccrualAddress)
	withdrawalSvc := services.NewWithdrawalService(log, balanceWithdrawalRepo, withdrawalRepo)
	authSvc := services.NewAuthService(log, userRepo, jwtSvc, 15*time.Hour)
	orderWorkerSvc := services.NewOrderWorkerService(log, accrualSvc, storageSvc, orderBalanceRepo)

	httpController := http.NewController(cfg.HTTPAdress)
	httpController.RegisterMiddlewares(
		middlewares.LoggerMiddleware(log),
		middleware.Compress(5, "application/json"),
	)

	authMiddleware := middlewares.NewAuhtorizationMiddleware(jwtSvc)

	authHandler := handlers.NewAuthHandler(authSvc)
	ordersHandler := handlers.NewOrdersHandler(authMiddleware, orderSvc)
	balanceHandler := handlers.NewBalanceHandler(authMiddleware, balanceSvc, withdrawalSvc)
	withdrawalHandler := handlers.NewWithdrawalHandler(authMiddleware, withdrawalSvc)

	httpController.AddHandler("/user", authHandler, ordersHandler, balanceHandler, withdrawalHandler)

	return &App{
		pgStorage:      storage,
		httpController: httpController,
		orderWorker:    orderWorkerSvc,
		logger:         log,
	}, nil
}

func (a *App) Run(ctx context.Context) {
	go func() {
		err := a.httpController.Start()
		if err != nil {
			a.logger.Fatal(err)
		}
	}()

	go a.orderWorker.Run(ctx)
	a.logger.Info("App started")
}

func (a *App) Stop(ctx context.Context) error {
	err := a.httpController.Stop(ctx)
	if err != nil {
		return err
	}
	a.pgStorage.Stop()
	a.logger.Info("App stopped")

	return nil
}
