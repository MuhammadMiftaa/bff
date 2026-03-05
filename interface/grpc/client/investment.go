package client

import (
	"context"
	"time"

	ipb "github.com/MuhammadMiftaa/Refina-Protobuf/investment"
)

type InvestmentClient interface {
	GetUserInvestmentList(ctx context.Context, req *ipb.GetUserInvestmentListRequest) (*ipb.GetUserInvestmentListResponse, error)
	GetInvestmentDetail(ctx context.Context, req *ipb.GetInvestmentDetailRequest) (*ipb.Investment, error)
	CreateInvestment(ctx context.Context, req *ipb.CreateInvestmentRequest) (*ipb.Investment, error)
	SellInvestment(ctx context.Context, req *ipb.SellInvestmentRequest) (*ipb.SellInvestmentResponse, error)
	GetInvestmentSummary(ctx context.Context, userID string) (*ipb.InvestmentSummaryResponse, error)
	GetAssetCodes(ctx context.Context) (*ipb.GetAssetCodesResponse, error)
}

type investmentClientImpl struct {
	client ipb.InvestmentServiceClient
}

func NewInvestmentClient(grpcClient ipb.InvestmentServiceClient) InvestmentClient {
	return &investmentClientImpl{client: grpcClient}
}

func (i *investmentClientImpl) GetUserInvestmentList(ctx context.Context, req *ipb.GetUserInvestmentListRequest) (*ipb.GetUserInvestmentListResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return i.client.GetUserInvestmentList(ctx, req)
}

func (i *investmentClientImpl) GetInvestmentDetail(ctx context.Context, req *ipb.GetInvestmentDetailRequest) (*ipb.Investment, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return i.client.GetInvestmentDetail(ctx, req)
}

func (i *investmentClientImpl) CreateInvestment(ctx context.Context, req *ipb.CreateInvestmentRequest) (*ipb.Investment, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return i.client.CreateInvestment(ctx, req)
}

func (i *investmentClientImpl) SellInvestment(ctx context.Context, req *ipb.SellInvestmentRequest) (*ipb.SellInvestmentResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return i.client.SellInvestment(ctx, req)
}

func (i *investmentClientImpl) GetInvestmentSummary(ctx context.Context, userID string) (*ipb.InvestmentSummaryResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return i.client.GetInvestmentSummary(ctx, &ipb.UserID{Id: userID})
}

func (i *investmentClientImpl) GetAssetCodes(ctx context.Context) (*ipb.GetAssetCodesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return i.client.GetAssetCodes(ctx, &ipb.Empty{})
}
