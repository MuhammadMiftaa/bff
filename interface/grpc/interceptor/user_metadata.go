package interceptor

import (
	"context"

	"refina-web-bff/internal/types/dto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Metadata keys for user data propagation
const (
	MDKeyUserID         = "x-user-id"
	MDKeyUserEmail      = "x-user-email"
	MDKeyUserProvider   = "x-user-provider"
	MDKeyProviderUserID = "x-provider-user-id"
)

// userDataKey is a context key for storing user data before sending gRPC calls
type userDataKey struct{}

// ContextWithUserData attaches UserData to a context so the interceptor can
// read it and inject into gRPC metadata.
func ContextWithUserData(ctx context.Context, ud dto.UserData) context.Context {
	return context.WithValue(ctx, userDataKey{}, ud)
}

// UserDataFromContext retrieves UserData from context, if present.
func UserDataFromContext(ctx context.Context) (dto.UserData, bool) {
	ud, ok := ctx.Value(userDataKey{}).(dto.UserData)
	return ud, ok
}

// UnaryClientInterceptor returns a gRPC UnaryClientInterceptor that reads
// UserData from the context and injects it into outgoing gRPC metadata.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = injectUserMetadata(ctx)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor returns a gRPC StreamClientInterceptor that reads
// UserData from the context and injects it into outgoing gRPC metadata.
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		ctx = injectUserMetadata(ctx)
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func injectUserMetadata(ctx context.Context) context.Context {
	ud, ok := UserDataFromContext(ctx)
	if !ok {
		return ctx
	}

	md := metadata.Pairs(
		MDKeyUserID, ud.ID,
		MDKeyUserEmail, ud.Email,
		MDKeyUserProvider, ud.UserAuthProvider.Provider,
		MDKeyProviderUserID, ud.UserAuthProvider.ProviderUserId,
	)

	// Merge with existing outgoing metadata if any
	existingMD, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		md = metadata.Join(existingMD, md)
	}

	return metadata.NewOutgoingContext(ctx, md)
}
