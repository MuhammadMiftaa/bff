package env

import (
	"os"

	"github.com/joho/godotenv"
)

type (
	Server struct {
		Mode     string
		HTTPPort string
	}

	Auth struct {
		JWTSecret string
	}

	GRPCConfig struct {
		DashboardAddress   string
		WalletAddress      string
		TransactionAddress string
	}

	Config struct {
		Server     Server
		Auth       Auth
		GRPCConfig GRPCConfig
	}
)

var Cfg Config

func Load() ([]string, error) {
	var ok bool
	var missing []string

	// Try loading from /app/.env first (Docker), then local .env
	if _, err := os.Stat("/app/.env"); err == nil {
		if err := godotenv.Load("/app/.env"); err != nil {
			return nil, err
		}
	} else {
		_ = godotenv.Load() // best-effort local .env
	}

	// Server
	if Cfg.Server.Mode, ok = os.LookupEnv("MODE"); !ok {
		missing = append(missing, "MODE")
	}
	if Cfg.Server.HTTPPort, ok = os.LookupEnv("HTTP_PORT"); !ok {
		missing = append(missing, "HTTP_PORT")
	}

	// Auth
	if Cfg.Auth.JWTSecret, ok = os.LookupEnv("JWT_SECRET"); !ok {
		missing = append(missing, "JWT_SECRET")
	}

	// gRPC downstream services
	if Cfg.GRPCConfig.DashboardAddress, ok = os.LookupEnv("DASHBOARD_GRPC_ADDRESS"); !ok {
		missing = append(missing, "DASHBOARD_GRPC_ADDRESS")
	}
	if Cfg.GRPCConfig.WalletAddress, ok = os.LookupEnv("WALLET_GRPC_ADDRESS"); !ok {
		missing = append(missing, "WALLET_GRPC_ADDRESS")
	}
	if Cfg.GRPCConfig.TransactionAddress, ok = os.LookupEnv("TRANSACTION_GRPC_ADDRESS"); !ok {
		missing = append(missing, "TRANSACTION_GRPC_ADDRESS")
	}

	return missing, nil
}
