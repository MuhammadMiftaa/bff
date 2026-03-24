package env

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type (
	Server struct {
		Mode     string `env:"MODE"`
		HTTPPort string `env:"HTTP_PORT"`
	}

	Auth struct {
		JWTSecret string `env:"JWT_SECRET"`
	}

	GRPCConfig struct {
		DashboardAddress   string `env:"DASHBOARD_GRPC_ADDRESS"`
		WalletAddress      string `env:"WALLET_GRPC_ADDRESS"`
		TransactionAddress string `env:"TRANSACTION_GRPC_ADDRESS"`
		InvestmentAddress  string `env:"INVESTMENT_GRPC_ADDRESS"`
		ProfileAddress     string `env:"PROFILE_GRPC_ADDRESS"`
	}

	RedisConfig struct {
		Address  string `env:"REDIS_ADDRESS"`
		Password string `env:"REDIS_PASSWORD"`
		DB       string `env:"REDIS_DB"`
	}

	Config struct {
		Server      Server
		Auth        Auth
		GRPCConfig  GRPCConfig
		RedisConfig RedisConfig
	}
)

var Cfg Config

const errEnvNotSet = " env is not set"

// lookupEnv reads an OS environment variable.
// If missing, it appends a message to the missing slice.
func lookupEnv(key string, dest *string, missing *[]string) {
	if val, ok := os.LookupEnv(key); ok {
		*dest = val
	} else {
		*missing = append(*missing, key+errEnvNotSet)
	}
}

// lookupEnvInt reads an integer environment variable (optional for Redis DB).
// Currently not used in this config, but kept for consistency with the pattern.
func lookupEnvInt(key string, dest *int, missing *[]string) {
	val, ok := os.LookupEnv(key)
	if !ok {
		*missing = append(*missing, key+errEnvNotSet)
		return
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		*missing = append(*missing, key+" must be integer, got "+val)
		return
	}
	*dest = n
}

func Load() ([]string, error) {
	// Try loading from /app/.env first (Docker), then local .env
	if _, err := os.Stat("/app/.env"); err == nil {
		if err := godotenv.Load("/app/.env"); err != nil {
			return nil, err
		}
	} else {
		_ = godotenv.Load() // best-effort for local development
	}

	var missing []string

	// Server
	lookupEnv("MODE", &Cfg.Server.Mode, &missing)
	lookupEnv("HTTP_PORT", &Cfg.Server.HTTPPort, &missing)

	// Auth
	lookupEnv("JWT_SECRET", &Cfg.Auth.JWTSecret, &missing)

	// gRPC downstream services
	lookupEnv("DASHBOARD_GRPC_ADDRESS", &Cfg.GRPCConfig.DashboardAddress, &missing)
	lookupEnv("WALLET_GRPC_ADDRESS", &Cfg.GRPCConfig.WalletAddress, &missing)
	lookupEnv("TRANSACTION_GRPC_ADDRESS", &Cfg.GRPCConfig.TransactionAddress, &missing)
	lookupEnv("INVESTMENT_GRPC_ADDRESS", &Cfg.GRPCConfig.InvestmentAddress, &missing)
	lookupEnv("PROFILE_GRPC_ADDRESS", &Cfg.GRPCConfig.ProfileAddress, &missing)

	// Redis
	lookupEnv("REDIS_ADDRESS", &Cfg.RedisConfig.Address, &missing)
	Cfg.RedisConfig.Password, _ = os.LookupEnv("REDIS_PASSWORD")
	Cfg.RedisConfig.DB, _ = os.LookupEnv("REDIS_DB")

	return missing, nil
}
