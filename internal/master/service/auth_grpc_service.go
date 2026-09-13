package service

import (
	"context"

	"mallekoppie/ChaosGenerator/internal/contracts"
	"mallekoppie/ChaosGenerator/internal/master/logic"

	"github.com/Mallekoppie/goslow/platform"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer implements the contract.Auth gRPC service. Login and RegisterUser
// are reachable without a token; everything else is guarded by JwtInterceptor.
type AuthServer struct {
	contracts.UnimplementedAuthServer
}

func (s *AuthServer) Register(server *grpc.Server) {
	contracts.RegisterAuthServer(server, s)
}

func (s *AuthServer) Login(ctx context.Context, req *contracts.LoginRequest) (*contracts.LoginResponse, error) {
	platform.Log.Info("Received Login request", zap.String("username", req.Username))

	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	token, err := logic.LoginUser(req.Username, req.Password)
	if err != nil {
		platform.Log.Error("Login failed", zap.String("username", req.Username), zap.Error(err))
		return &contracts.LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		}, nil
	}

	return &contracts.LoginResponse{
		Token:   token,
		Success: true,
		Message: "Login successful",
	}, nil
}

func (s *AuthServer) RegisterUser(ctx context.Context, req *contracts.RegisterRequest) (*contracts.RegisterResponse, error) {
	platform.Log.Info("Received RegisterUser request", zap.String("username", req.Username))

	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	token, err := logic.RegisterUser(req.Username, req.Password, req.RegisterSecret)
	if err != nil {
		platform.Log.Error("Registration failed", zap.String("username", req.Username), zap.Error(err))
		return &contracts.RegisterResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &contracts.RegisterResponse{
		Token:   token,
		Success: true,
		Message: "User registered successfully",
	}, nil
}

func (s *AuthServer) GetAllUsers(ctx context.Context, req *contracts.GetAllUsersRequest) (*contracts.GetAllUsersResponse, error) {
	platform.Log.Info("Received GetAllUsers request")

	users, err := logic.GetAllUsers()
	if err != nil {
		platform.Log.Error("Error getting all users", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get users: %v", err)
	}

	pbUsers := make([]*contracts.User, len(users))
	for i, user := range users {
		pbUsers[i] = &contracts.User{
			Id:        user.Id,
			Username:  user.Username,
			CreatedAt: user.CreatedAt.Unix(),
		}
	}

	return &contracts.GetAllUsersResponse{Users: pbUsers}, nil
}

func (s *AuthServer) DeleteUser(ctx context.Context, req *contracts.DeleteUserRequest) (*contracts.DeleteUserResponse, error) {
	platform.Log.Info("Received DeleteUser request", zap.String("id", req.Id))

	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	err := logic.DeleteUser(req.Id)
	if err != nil {
		platform.Log.Error("Error deleting user", zap.String("id", req.Id), zap.Error(err))
		return &contracts.DeleteUserResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &contracts.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}
