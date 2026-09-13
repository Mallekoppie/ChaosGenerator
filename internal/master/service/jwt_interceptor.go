package service

import (
	"context"
	"strings"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnauthenticatedMethods lists the fully qualified gRPC methods that do not
// require a valid JWT. Every other method must carry an
// "authorization: Bearer <token>" metadata entry.
var UnauthenticatedMethods = map[string]bool{
	"/contract.Auth/Login":        true,
	"/contract.Auth/RegisterUser": true,
}

// JwtInterceptor validates the local JWT on every unary call and injects the
// claims into the context under platform.ContextLocalJwtClaims.
func JwtInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if UnauthenticatedMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		platform.Log.Error("Missing metadata in gRPC request", zap.String("method", info.FullMethod))
		return nil, status.Error(codes.Unauthenticated, "missing authorization metadata")
	}

	authorization := md["authorization"]
	if len(authorization) < 1 {
		platform.Log.Error("Missing authorization metadata", zap.String("method", info.FullMethod))
		return nil, status.Error(codes.Unauthenticated, "missing authorization metadata")
	}

	token := strings.TrimSpace(strings.TrimPrefix(authorization[0], "Bearer "))
	if token == "" {
		platform.Log.Error("Empty bearer token supplied", zap.String("method", info.FullMethod))
		return nil, status.Error(codes.Unauthenticated, "invalid authorization metadata")
	}

	claims, err := platform.LocalJwt.ValidateLocalJwtToken(token)
	if err != nil {
		platform.Log.Error("Failed to validate JWT", zap.String("method", info.FullMethod), zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	ctx = context.WithValue(ctx, platform.ContextLocalJwtClaims, claims)

	return handler(ctx, req)
}
