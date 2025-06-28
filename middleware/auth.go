package middleware

import (
	"context"
	"log"
	"net/http"

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

		ctx = context.WithValue(ctx, "user_id", userID)
		return handler(ctx, req)
	}
}

func ErrorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		// Log the error
		log.Printf("Method %s failed: %v", info.FullMethod, err)

		// Convert to status error if not already
		if _, ok := status.FromError(err); !ok {
			err = status.Error(codes.Unknown, err.Error())
		}
	}
	return resp, err
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the token from the header
		token := r.Header.Get("aaa-auth-token")
		if token == "" {
			http.Error(w, "Authorization token is not provided", http.StatusUnauthorized)
			return
		}

		// Validate the token
		userID, err := helper.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Add userID to the context
		ctx := context.WithValue(r.Context(), "user_id", userID)

		// Call the next handler with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
