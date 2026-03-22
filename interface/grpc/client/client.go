package client

import (
	"fmt"
	"sync"
	"time"

	"refina-web-bff/config/env"
	logger "refina-web-bff/config/log"
	"refina-web-bff/interface/grpc/interceptor"
	"refina-web-bff/internal/utils/data"

	dpb "github.com/MuhammadMiftaa/Refina-Protobuf/dashboard"
	ipb "github.com/MuhammadMiftaa/Refina-Protobuf/investment"
	ppb "github.com/MuhammadMiftaa/Refina-Protobuf/profile"
	tpb "github.com/MuhammadMiftaa/Refina-Protobuf/transaction"
	wpb "github.com/MuhammadMiftaa/Refina-Protobuf/wallet"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type GRPCClientManager struct {
	dashboardClient   dpb.DashboardServiceClient
	walletClient      wpb.WalletServiceClient
	transactionClient tpb.TransactionServiceClient
	investmentClient  ipb.InvestmentServiceClient
	profileClient     ppb.ProfileServiceClient

	connections []*grpc.ClientConn
	mu          sync.RWMutex
}

var (
	manager *GRPCClientManager
	once    sync.Once
)

// GetManager returns singleton instance of GRPCClientManager
func GetManager() *GRPCClientManager {
	once.Do(func() {
		manager = &GRPCClientManager{
			connections: make([]*grpc.ClientConn, 0),
		}
	})
	return manager
}

// SetupGRPCClient sets up all gRPC clients
func (m *GRPCClientManager) SetupGRPCClient() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	logger.Info(data.LogGRPCClientSetupStarted, map[string]any{
		"service":             data.GRPCClientService,
		"dashboard_address":   env.Cfg.GRPCConfig.DashboardAddress,
		"wallet_address":      env.Cfg.GRPCConfig.WalletAddress,
		"transaction_address": env.Cfg.GRPCConfig.TransactionAddress,
		"investment_address":  env.Cfg.GRPCConfig.InvestmentAddress,
		"profile_address":     env.Cfg.GRPCConfig.ProfileAddress,
	})

	if err := m.setupDashboardClient(env.Cfg.GRPCConfig.DashboardAddress); err != nil {
		return fmt.Errorf("failed to setup dashboard client: %w", err)
	}

	if err := m.setupWalletClient(env.Cfg.GRPCConfig.WalletAddress); err != nil {
		return fmt.Errorf("failed to setup wallet client: %w", err)
	}

	if err := m.setupTransactionClient(env.Cfg.GRPCConfig.TransactionAddress); err != nil {
		return fmt.Errorf("failed to setup transaction client: %w", err)
	}

	if err := m.setupInvestmentClient(env.Cfg.GRPCConfig.InvestmentAddress); err != nil {
		return fmt.Errorf("failed to setup investment client: %w", err)
	}

	if err := m.setupProfileClient(env.Cfg.GRPCConfig.ProfileAddress); err != nil {
		return fmt.Errorf("failed to setup profile client: %w", err)
	}

	logger.Info(data.LogGRPCClientSetupSuccess, map[string]any{
		"service": data.GRPCClientService,
	})

	return nil
}

func (m *GRPCClientManager) setupDashboardClient(address string) error {
	conn, err := m.createConnection(address)
	if err != nil {
		return err
	}

	m.dashboardClient = dpb.NewDashboardServiceClient(conn)
	m.connections = append(m.connections, conn)
	return nil
}

func (m *GRPCClientManager) setupWalletClient(address string) error {
	conn, err := m.createConnection(address)
	if err != nil {
		return err
	}

	m.walletClient = wpb.NewWalletServiceClient(conn)
	m.connections = append(m.connections, conn)
	return nil
}

func (m *GRPCClientManager) setupTransactionClient(address string) error {
	conn, err := m.createConnection(address)
	if err != nil {
		return err
	}

	m.transactionClient = tpb.NewTransactionServiceClient(conn)
	m.connections = append(m.connections, conn)
	return nil
}

func (m *GRPCClientManager) setupInvestmentClient(address string) error {
	conn, err := m.createConnection(address)
	if err != nil {
		return err
	}

	m.investmentClient = ipb.NewInvestmentServiceClient(conn)
	m.connections = append(m.connections, conn)
	return nil
}

func (m *GRPCClientManager) setupProfileClient(address string) error {
	conn, err := m.createConnection(address)
	if err != nil {
		return err
	}

	m.profileClient = ppb.NewProfileServiceClient(conn)
	m.connections = append(m.connections, conn)
	return nil
}

func (m *GRPCClientManager) createConnection(address string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             3 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024),
			grpc.MaxCallSendMsgSize(10*1024*1024),
		),
		grpc.WithUnaryInterceptor(interceptor.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(interceptor.StreamClientInterceptor()),
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	return conn, nil
}

// GetDashboardClient returns the dashboard gRPC client
func (m *GRPCClientManager) GetDashboardClient() dpb.DashboardServiceClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dashboardClient
}

// GetWalletClient returns the wallet gRPC client
func (m *GRPCClientManager) GetWalletClient() wpb.WalletServiceClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.walletClient
}

// GetTransactionClient returns the transaction gRPC client
func (m *GRPCClientManager) GetTransactionClient() tpb.TransactionServiceClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.transactionClient
}

// GetInvestmentClient returns the investment gRPC client
func (m *GRPCClientManager) GetInvestmentClient() ipb.InvestmentServiceClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.investmentClient
}

// GetProfileClient returns the profile gRPC client
func (m *GRPCClientManager) GetProfileClient() ppb.ProfileServiceClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.profileClient
}

// Close closes all gRPC connections
func (m *GRPCClientManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, conn := range m.connections {
		if err := conn.Close(); err != nil {
			logger.Error("failed to close gRPC connection", map[string]any{
				"service": data.GRPCClientService,
				"error":   err.Error(),
			})
		}
	}

	logger.Info(data.LogGRPCClientClosed, map[string]any{
		"service":     data.GRPCClientService,
		"connections": len(m.connections),
	})
}
