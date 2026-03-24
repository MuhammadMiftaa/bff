package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	redisSetup "refina-web-bff/config/cache"
	"refina-web-bff/config/env"
	logger "refina-web-bff/config/log"
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/interface/http/router"
	"refina-web-bff/internal/cache"
	"refina-web-bff/internal/utils"
	"refina-web-bff/internal/utils/data"
)

func init() {
	missing, err := env.Load()
	if err != nil {
		log.Fatalf("Failed to load environment variables: %v", err)
	}

	logger.SetupLogger()

	if len(missing) > 0 {
		for _, envVar := range missing {
			logger.Warn(data.LogEnvVarMissing, map[string]any{"service": data.EnvService, "key": envVar})
		}
	}
}

// shutdownHTTP gracefully shuts down the Fiber HTTP server.
// Extracted to reduce cognitive complexity (go:S3776).
func shutdownHTTP(app interface{ ShutdownWithTimeout(time.Duration) error }, errors map[string]any) {
	if app == nil {
		return
	}
	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		logger.Error(data.LogHTTPServerShutdownFailed, map[string]any{
			"service": data.HTTPServerService,
			"error":   err.Error(),
		})
		errors["http_error"] = true
	}
}

// shutdownGRPC gracefully closes gRPC client connections.
// Extracted to reduce cognitive complexity (go:S3776).
func shutdownGRPC(ctx context.Context, grpcMgr *grpcClient.GRPCClientManager, errors map[string]any) {
	if grpcMgr == nil {
		return
	}
	if err := grpcMgr.Shutdown(ctx); err != nil {
		logger.Error(data.LogGRPCClientShutdownFailed, map[string]any{
			"service": data.GRPCClientService,
			"error":   err.Error(),
		})
		errors["grpc_error"] = true
	}
}

// shutdownCache closes the Redis cache connection.
// Extracted to reduce cognitive complexity (go:S3776).
func shutdownCache(appCache cache.Cache, errors map[string]any) {
	if appCache == nil {
		return
	}
	if err := appCache.Close(); err != nil {
		logger.Error(data.LogRedisCloseFailed, map[string]any{
			"service": data.CacheService,
			"error":   err.Error(),
		})
		errors["cache_error"] = true
	}
}

func main() {
	// Setup Redis Cache
	startTime := time.Now()
	redisClient, err := redisSetup.NewRedisClient(redisSetup.RedisConfig{
		Address:  env.Cfg.RedisConfig.Address,
		Password: env.Cfg.RedisConfig.Password,
		DB:       redisSetup.ParseRedisDB(env.Cfg.RedisConfig.DB),
	})
	if err != nil {
		logger.Fatal(data.LogRedisSetupFailed, map[string]any{
			"service": data.CacheService,
			"error":   err.Error(),
		})
	}
	appCache := cache.NewRedisCache(redisClient)
	logger.Info(data.LogRedisSetupSuccess, map[string]any{
		"service":  data.CacheService,
		"duration": utils.Ms(time.Since(startTime)),
	})

	// Setup gRPC Clients
	startTime = time.Now()
	grpcMgr := grpcClient.GetManager()
	if err := grpcMgr.SetupGRPCClient(); err != nil {
		logger.Fatal(data.LogGRPCClientSetupFailed, map[string]any{
			"service": data.GRPCClientService,
			"error":   err.Error(),
		})
	}
	logger.Info(data.LogGRPCClientSetupSuccess, map[string]any{
		"service":  data.GRPCClientService,
		"duration": utils.Ms(time.Since(startTime)),
	})

	// Create gRPC client wrappers
	dashboardClient := grpcClient.NewDashboardClient(grpcMgr.GetDashboardClient())
	walletClient := grpcClient.NewWalletClient(grpcMgr.GetWalletClient())
	transactionClient := grpcClient.NewTransactionClient(grpcMgr.GetTransactionClient())
	investmentClient := grpcClient.NewInvestmentClient(grpcMgr.GetInvestmentClient())
	profileClient := grpcClient.NewProfileClient(grpcMgr.GetProfileClient())

	// Setup HTTP Server
	startTime = time.Now()
	app := router.SetupHTTPServer(dashboardClient, walletClient, transactionClient, investmentClient, profileClient, appCache, redisClient)
	if app != nil {
		go func() {
			if err := app.Listen(":" + env.Cfg.Server.HTTPPort); err != nil {
				logger.Fatal(data.LogHTTPServerStartFailed, map[string]any{
					"service": data.HTTPServerService,
					"error":   err.Error(),
				})
			}
		}()
		logger.Info(data.LogHTTPServerStarted, map[string]any{
			"service":  data.HTTPServerService,
			"port":     env.Cfg.Server.HTTPPort,
			"duration": utils.Ms(time.Since(startTime)),
		})
	}

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info(data.LogShutdownSignalReceived, map[string]any{"service": data.MainService})

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	startTime = time.Now()
	shutdownErrors := map[string]any{"service": data.MainService}

	shutdownHTTP(app, shutdownErrors)
	shutdownGRPC(shutdownCtx, grpcMgr, shutdownErrors)
	shutdownCache(appCache, shutdownErrors)

	if len(shutdownErrors) > 1 {
		logger.Info(data.LogShutdownCompletedWithErrors, shutdownErrors)
	} else {
		logger.Info(data.LogShutdownCompleted, map[string]any{
			"service":  data.MainService,
			"duration": utils.Ms(time.Since(startTime)),
		})
	}
}