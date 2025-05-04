package app

import (
	"context"
	authhandler "github.com/MxTrap/gophermart/internal/gophermart/controller/http/handlers/auth"
	balancehandler "github.com/MxTrap/gophermart/internal/gophermart/controller/http/handlers/balance"
	orderhandler "github.com/MxTrap/gophermart/internal/gophermart/controller/http/handlers/order"
	withdrawalhandler "github.com/MxTrap/gophermart/internal/gophermart/controller/http/handlers/withdrawal"
	"github.com/MxTrap/gophermart/internal/gophermart/services/accrual"
	"github.com/MxTrap/gophermart/internal/gophermart/services/auth"
	"github.com/MxTrap/gophermart/internal/gophermart/services/balance"
	"github.com/MxTrap/gophermart/internal/gophermart/services/jwt"
	"github.com/MxTrap/gophermart/internal/gophermart/services/order"
	"github.com/MxTrap/gophermart/internal/gophermart/services/orderworker"
	"github.com/MxTrap/gophermart/internal/gophermart/services/storage"
	"github.com/MxTrap/gophermart/internal/gophermart/services/withdrawal"
	"github.com/go-chi/chi/v5/middleware"
	"time"

	"github.com/MxTrap/gophermart/config"

	"github.com/MxTrap/gophermart/internal/gophermart/controller/http"
	"github.com/MxTrap/gophermart/internal/gophermart/controller/http/middlewares"
	"github.com/MxTrap/gophermart/internal/gophermart/migrator"
	"github.com/MxTrap/gophermart/internal/gophermart/repository/postgres"
	balancerepo "github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/balance"
	"github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/combined"
	orderrepo "github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/order"
	userrepo "github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/user"
	withdrawalrepo "github.com/MxTrap/gophermart/internal/gophermart/repository/postgres/withdrawal"
	"github.com/MxTrap/gophermart/logger"
)

type App struct {
	pgStorage      *postgres.Storage
	httpController *http.Controller
	orderWorker    *orderworker.OrderWorkerService
	logger         *logger.Logger
}

func NewApp(ctx context.Context, log *logger.Logger, cfg *config.Config) (*App, error) {
	postgresStorage, err := postgres.NewPostgresStorage(ctx, cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	mgrtr, err := migrator.NewMigrator(postgresStorage.Pool)
	if err != nil {
		return nil, err
	}
	err = mgrtr.InitializeDB()
	if err != nil {
		return nil, err
	}

	userRepo := userrepo.NewUserRepository(postgresStorage.Pool)
	orderRepo := orderrepo.NewOrderRepository(postgresStorage.Pool)
	balanceRepo := balancerepo.NewBalanceRepository(postgresStorage.Pool)
	withdrawalRepo := withdrawalrepo.NewWithdrawnRepo(postgresStorage.Pool)
	orderBalanceRepo := combined.NewOrderBalanceRepo(postgresStorage.Pool, orderRepo, balanceRepo)
	balanceWithdrawalRepo := combined.NewBalanceWithdrawnRepo(postgresStorage.Pool, balanceRepo, withdrawalRepo)

	storageSvc := storage.NewStorageService()
	jwtSvc := jwt.NewJWTService("very secret")
	orderSvc := order.NewOrderService(log, storageSvc, orderRepo)
	balanceSvc := balance.NewBalanceService(log, balanceRepo)
	accrualSvc := accrual.NewAccrualService(log, cfg.AccrualAddress)
	withdrawalSvc := withdrawal.NewWithdrawalService(log, balanceWithdrawalRepo, withdrawalRepo)
	authSvc := auth.NewAuthService(log, userRepo, jwtSvc, 15*time.Hour)
	orderWorkerSvc := orderworker.NewOrderWorkerService(log, accrualSvc, storageSvc, orderBalanceRepo)

	httpController := http.NewController(cfg.HTTPAdress)
	httpController.RegisterMiddlewares(
		middlewares.LoggerMiddleware(log),
		middleware.Compress(5, "application/json"),
	)

	authMiddleware := middlewares.NewAuhtorizationMiddleware(jwtSvc)

	authHandler := authhandler.NewAuthHandler(authSvc)
	ordersHandler := orderhandler.NewOrdersHandler(authMiddleware, orderSvc)
	balanceHandler := balancehandler.NewBalanceHandler(authMiddleware, balanceSvc, withdrawalSvc)
	withdrawalHandler := withdrawalhandler.NewWithdrawalHandler(authMiddleware, withdrawalSvc)

	httpController.AddHandler("/user", authHandler, ordersHandler, balanceHandler, withdrawalHandler)

	return &App{
		pgStorage:      postgresStorage,
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
