package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"refina-web-bff/config/env"
	logger "refina-web-bff/config/log"
	grpcClient "refina-web-bff/interface/grpc/client"
	"refina-web-bff/interface/http/router"
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

func main() {
	startTime := time.Now()

	// Set up gRPC clients
	grpcMgr := grpcClient.GetManager()
	if err := grpcMgr.SetupGRPCClient(); err != nil {
		logger.Fatal(data.LogGRPCClientSetupFailed, map[string]any{
			"service": data.GRPCClientService,
			"error":   err.Error(),
		})
	}

	// Create dashboard client wrapper
	dashboardClient := grpcClient.NewDashboardClient(grpcMgr.GetDashboardClient())

	// Create wallet client wrapper
	walletClient := grpcClient.NewWalletClient(grpcMgr.GetWalletClient())

	// Create transaction client wrapper
	transactionClient := grpcClient.NewTransactionClient(grpcMgr.GetTransactionClient())

	// Set up the HTTP server (Fiber)
	app := router.SetupHTTPServer(dashboardClient, walletClient, transactionClient)
	logger.Info(data.LogHTTPServerStarted, map[string]any{
		"service":  data.HTTPServerService,
		"port":     env.Cfg.Server.HTTPPort,
		"duration": utils.Ms(time.Since(startTime)),
	})

	// Start server in a goroutine
	go func() {
		if err := app.Listen(":" + env.Cfg.Server.HTTPPort); err != nil {
			logger.Fatal(data.LogHTTPServerStartFailed, map[string]any{
				"service": data.HTTPServerService,
				"error":   err.Error(),
			})
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info(data.LogShutdownSignalReceived, map[string]any{"service": data.MainService})

	startTime = time.Now()

	// Close gRPC connections
	grpcMgr.Close()

	if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
		logger.Error(data.LogHTTPServerShutdownFailed, map[string]any{
			"service": data.HTTPServerService,
			"error":   err.Error(),
		})
	}

	logger.Info(data.LogShutdownCompleted, map[string]any{
		"service":  data.MainService,
		"duration": utils.Ms(time.Since(startTime)),
	})
}
