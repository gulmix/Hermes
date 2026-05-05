package handler

import (
    "context"

    "go.uber.org/zap"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    userv1 "github.com/gulmix/hermes/gen/go/user/v1"
)

type UserHandler struct {
    userv1.UnimplementedUserServiceServer
    log *zap.Logger
}

func NewUserHandler(log *zap.Logger) *UserHandler {
    return &UserHandler{log: log}
}

func (h *UserHandler) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
    h.log.Info("GetUser", zap.String("id", req.GetId()))
    return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (h *UserHandler) WatchUsers(req *userv1.WatchUsersRequest, stream userv1.UserService_WatchUsersServer) error {
    return status.Errorf(codes.Unimplemented, "not implemented")
}