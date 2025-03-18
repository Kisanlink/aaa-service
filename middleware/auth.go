package middleware

import (
	"context"

	"github.com/Kisanlink/aaa-service/helper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string


func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "Metadata is not provided")
		}

		authHeaders := md["aaa-auth-token"]
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "Authorization token is not provided")
		}

		token := authHeaders[0]
		userID, err := helper.ValidateToken(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "Invalid token: %v", err)
		}

		ctx = context.WithValue(ctx,"user_id", userID)
		return handler(ctx, req)
	}
}
