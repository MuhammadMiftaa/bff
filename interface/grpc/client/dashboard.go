package client

import (
	"context"
	"time"

	dpb "github.com/MuhammadMiftaa/Refina-Protobuf/dashboard"
)

type DashboardClient interface {
	GetUserTransactions(ctx context.Context, req *dpb.GetUserTransactionsRequest) (*dpb.GetUserTransactionsResponse, error)
	GetUserBalance(ctx context.Context, req *dpb.GetUserBalanceRequest) (*dpb.GetUserBalanceResponse, error)
	GetUserFinancialSummary(ctx context.Context, req *dpb.GetUserFinancialSummaryRequest) (*dpb.GetUserFinancialSummaryResponse, error)
	GetUserNetWorthComposition(ctx context.Context, req *dpb.GetUserNetWorthCompositionRequest) (*dpb.NetWorthComposition, error)
	GetUserWallets(ctx context.Context, userID string) (*dpb.GetUserWalletsResponse, error)
	GetCategoryTransactions(ctx context.Context, req *dpb.GetCategoryTransactionsRequest) (*dpb.GetCategoryTransactionsResponse, error)
}

type dashboardClientImpl struct {
	client dpb.DashboardServiceClient
}

func NewDashboardClient(grpcClient dpb.DashboardServiceClient) DashboardClient {
	return &dashboardClientImpl{
		client: grpcClient,
	}
}

func (d *dashboardClientImpl) GetUserTransactions(ctx context.Context, req *dpb.GetUserTransactionsRequest) (*dpb.GetUserTransactionsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return d.client.GetUserTransactions(ctx, req)
}

func (d *dashboardClientImpl) GetUserBalance(ctx context.Context, req *dpb.GetUserBalanceRequest) (*dpb.GetUserBalanceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return d.client.GetUserBalance(ctx, req)
}

func (d *dashboardClientImpl) GetUserFinancialSummary(ctx context.Context, req *dpb.GetUserFinancialSummaryRequest) (*dpb.GetUserFinancialSummaryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return d.client.GetUserFinancialSummary(ctx, req)
}

func (d *dashboardClientImpl) GetUserNetWorthComposition(ctx context.Context, req *dpb.GetUserNetWorthCompositionRequest) (*dpb.NetWorthComposition, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return d.client.GetUserNetWorthComposition(ctx, req)
}

func (d *dashboardClientImpl) GetUserWallets(ctx context.Context, userID string) (*dpb.GetUserWalletsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return d.client.GetUserWallets(ctx, &dpb.UserID{Id: userID})
}

func (d *dashboardClientImpl) GetCategoryTransactions(ctx context.Context, req *dpb.GetCategoryTransactionsRequest) (*dpb.GetCategoryTransactionsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	return d.client.GetCategoryTransactions(ctx, req)
}
