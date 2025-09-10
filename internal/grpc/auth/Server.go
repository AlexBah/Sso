package authgrpc

import (
	"context"
	"errors"

	"sso/internal/services/auth"

	ssov1 "github.com/AlexBah/Protos/gen/go/sso"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	RegisterNewUser(ctx context.Context,
		phone string,
		password string,
	) (userID int64, err error)
	Login(ctx context.Context,
		phone string,
		password string,
		appID int,
	) (name string, email string, token string, user_id int64, err error)
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	GetUser(ctx context.Context, phone string) (name string, err error)
	UpdateUser(ctx context.Context,
		userID int64,
		name string,
		email string,
		phone string,
		password string,
		token string,
	) (success bool, err error)
	DeleteUser(ctx context.Context,
		phone string,
		token string,
	) (success bool, err error)
}

type serverAPI struct {
	ssov1.UnimplementedAuthServer
	auth Auth
}

func Register(gRPC *grpc.Server, auth Auth) {
	ssov1.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
}

const (
	emptyValue = 0
)

func (s *serverAPI) Register(
	ctx context.Context,
	req *ssov1.RegisterRequest,
) (*ssov1.RegisterResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	userID, err := s.auth.RegisterNewUser(ctx, req.GetPhone(), req.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		return nil, status.Error(codes.AlreadyExists, "internal error")
	}

	return &ssov1.RegisterResponse{
		UserId: userID,
	}, nil
}

func (s *serverAPI) Login(
	ctx context.Context,
	req *ssov1.LoginRequest,
) (*ssov1.LoginResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	name, email, token, user_id, err := s.auth.Login(ctx, req.GetPhone(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid phone or password")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.LoginResponse{
		Name:   name,
		Email:  email,
		Token:  token,
		UserId: user_id,
	}, nil
}

func (s *serverAPI) IsAdmin(
	ctx context.Context,
	req *ssov1.IsAdminRequest,
) (*ssov1.IsAdminResponse, error) {
	if err := validateIsAdmin(req); err != nil {
		return nil, err
	}

	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

func (s *serverAPI) GetUser(
	ctx context.Context,
	req *ssov1.GetUserRequest,
) (*ssov1.GetUserResponse, error) {
	if err := validateGetUser(req); err != nil {
		return nil, err
	}

	name, err := s.auth.GetUser(ctx, req.GetPhone())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.GetUserResponse{
		Name: name,
	}, nil
}

func (s *serverAPI) UpdateUser(
	ctx context.Context,
	req *ssov1.UpdateUserRequest,
) (*ssov1.UpdateUserResponse, error) {
	if err := validateUpdateUser(req); err != nil {
		return nil, err
	}

	success, err := s.auth.UpdateUser(
		ctx,
		req.GetUserId(),
		req.GetName(),
		req.GetEmail(),
		req.GetPhone(),
		req.GetPassword(),
		req.GetToken(),
	)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.UpdateUserResponse{
		Success: success,
	}, nil
}

func (s *serverAPI) DeleteUser(
	ctx context.Context,
	req *ssov1.DeleteUserRequest,
) (*ssov1.DeleteUserResponse, error) {
	if err := validateDeleteUser(req); err != nil {
		return nil, err
	}

	success, err := s.auth.DeleteUser(ctx, req.GetPhone(), req.GetToken())
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &ssov1.DeleteUserResponse{
		Success: success,
	}, nil
}

func validateLogin(req *ssov1.LoginRequest) error {
	if req.GetPhone() == "" {
		return status.Error(codes.InvalidArgument, "phone is required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	if req.GetAppId() == emptyValue {
		return status.Error(codes.InvalidArgument, "app_id is required")
	}
	return nil
}

func validateRegister(req *ssov1.RegisterRequest) error {
	if req.GetPhone() == "" {
		return status.Error(codes.InvalidArgument, "phone is required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	return nil
}

func validateIsAdmin(req *ssov1.IsAdminRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "userID is required")
	}
	return nil
}

func validateGetUser(req *ssov1.GetUserRequest) error {
	if req.GetPhone() == "" {
		return status.Error(codes.InvalidArgument, "phone is required")
	}
	return nil
}

func validateUpdateUser(req *ssov1.UpdateUserRequest) error {
	if req.GetUserId() == emptyValue {
		return status.Error(codes.InvalidArgument, "userID is required")
	}
	if req.GetToken() == "" {
		return status.Error(codes.InvalidArgument, "token is required")
	}
	return nil
}

func validateDeleteUser(req *ssov1.DeleteUserRequest) error {
	if req.GetPhone() == "" {
		return status.Error(codes.InvalidArgument, "phone is required")
	}
	if req.GetToken() == "" {
		return status.Error(codes.InvalidArgument, "token is required")
	}
	return nil
}
