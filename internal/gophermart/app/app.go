package app

import (
	"context"
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
	"github.com/MxTrap/gophermart/internal/gophermart/usecase"
	"github.com/MxTrap/gophermart/logger"
)

type App struct {
	pgStorage      *postgres.Storage
	httpController *http.Controller
	orderWorker    *usecase.OrderUsecase
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

	storageService := services.NewStorageService()
	jwtService := services.NewJWTService("very secret")
	orderService := services.NewOrderService(log, storageService, orderRepo)
	balanceService := services.NewBalanceService(log, balanceRepo)
	accrualService := services.NewWorkerService(log, cfg.AccrualAddress)
	withdrawalService := services.NewWithdrawalService(log, balanceWithdrawalRepo, withdrawalRepo)

	authUsecase := usecase.NewAuthUsecase(log, userRepo, jwtService, 15*time.Hour)
	orderUsecase := usecase.NewOrderUsecase(log, accrualService, storageService, orderBalanceRepo)

	httpController := http.NewController(cfg.HTTPAdress)
	httpController.RegisterMiddlewares(middlewares.LoggerMiddleware(log))

	authMiddleware := middlewares.NewAuhtorizationMiddleware(jwtService)

	authHandler := handlers.NewAuthHandler(authUsecase)
	ordersHandler := handlers.NewOrdersHandler(authMiddleware, orderService)
	balanceHandler := handlers.NewBalanceHandler(authMiddleware, balanceService, withdrawalService)
	withdrawalHander := handlers.NewWithdrawalHandler(authMiddleware, withdrawalService)

	httpController.AddHandler("/user", authHandler, ordersHandler, balanceHandler, withdrawalHander)

	return &App{
		pgStorage:      storage,
		httpController: httpController,
		orderWorker:    orderUsecase,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	go a.httpController.Start()

	go a.orderWorker.Run(ctx)

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	err := a.httpController.Stop(ctx)
	if err != nil {
		return err
	}
	a.pgStorage.Stop()

	return nil
}
