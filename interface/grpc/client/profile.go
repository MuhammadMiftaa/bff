package client

import (
	"context"
	"time"

	ppb "github.com/MuhammadMiftaa/Refina-Protobuf/profile"
)

type ProfileClient interface {
	GetProfile(ctx context.Context, userID string) (*ppb.Profile, error)
	UpdateProfile(ctx context.Context, userID string, fullname string) (*ppb.Profile, error)
	UploadProfilePhoto(ctx context.Context, userID string, base64Image string) (*ppb.UploadProfilePhotoResponse, error)
	DeleteProfilePhoto(ctx context.Context, userID string) (*ppb.DeleteProfilePhotoResponse, error)
}

type profileClientImpl struct {
	client ppb.ProfileServiceClient
}

func NewProfileClient(grpcClient ppb.ProfileServiceClient) ProfileClient {
	return &profileClientImpl{
		client: grpcClient,
	}
}

func (p *profileClientImpl) GetProfile(ctx context.Context, userID string) (*ppb.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return p.client.GetProfile(ctx, &ppb.GetProfileRequest{UserId: userID})
}

func (p *profileClientImpl) UpdateProfile(ctx context.Context, userID string, fullname string) (*ppb.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return p.client.UpdateProfile(ctx, &ppb.UpdateProfileRequest{
		UserId:   userID,
		Fullname: fullname,
	})
}

func (p *profileClientImpl) UploadProfilePhoto(ctx context.Context, userID string, base64Image string) (*ppb.UploadProfilePhotoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second) // longer timeout for photo upload
	defer cancel()
	return p.client.UploadProfilePhoto(ctx, &ppb.UploadProfilePhotoRequest{
		UserId:      userID,
		Base64Image: base64Image,
	})
}

func (p *profileClientImpl) DeleteProfilePhoto(ctx context.Context, userID string) (*ppb.DeleteProfilePhotoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return p.client.DeleteProfilePhoto(ctx, &ppb.DeleteProfilePhotoRequest{UserId: userID})
}
