package handler

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	userv1 "github.com/gulmix/hermes/gen/go/user/v1"
)

type UserHandler struct {
	userv1.UnimplementedUserServiceServer
	log *zap.Logger
}

func NewUserHandler(log *zap.Logger) *UserHandler {
	return &UserHandler{log: log}
}

// GetUser — Unary RPC.
func (h *UserHandler) GetUser(_ context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	h.log.Info("GetUser", zap.String("id", req.GetId()))

	now := timestamppb.Now()
	user := &userv1.User{
		Id:          req.GetId(),
		Email:       fmt.Sprintf("%s@example.com", req.GetId()),
		DisplayName: fmt.Sprintf("User %s", req.GetId()),
		Status:      userv1.UserStatus_USER_STATUS_ACTIVE,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return &userv1.GetUserResponse{User: user}, nil
}

// ListUsers — Unary RPC.
func (h *UserHandler) ListUsers(_ context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	h.log.Info("ListUsers", zap.String("status_filter", req.GetStatusFilter().String()))

	var users []*userv1.User
	for i := 1; i <= 3; i++ {
		now := timestamppb.Now()
		users = append(users, &userv1.User{
			Id:          fmt.Sprintf("user-%d", i),
			Email:       fmt.Sprintf("user%d@example.com", i),
			DisplayName: fmt.Sprintf("User %d", i),
			Status:      userv1.UserStatus_USER_STATUS_ACTIVE,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}
	return &userv1.ListUsersResponse{Users: users}, nil
}

// CreateUser — Unary RPC.
func (h *UserHandler) CreateUser(_ context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	now := timestamppb.Now()
	user := &userv1.User{
		Id:          fmt.Sprintf("user-%d", now.AsTime().UnixNano()),
		Email:       req.GetEmail(),
		DisplayName: req.GetDisplayName(),
		Status:      userv1.UserStatus_USER_STATUS_ACTIVE,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	h.log.Info("CreateUser", zap.String("email", user.GetEmail()), zap.String("id", user.GetId()))
	return &userv1.CreateUserResponse{User: user}, nil
}

// UpdateUser — Unary RPC.
func (h *UserHandler) UpdateUser(_ context.Context, req *userv1.UpdateUserRequest) (*userv1.UpdateUserResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	now := timestamppb.Now()
	user := &userv1.User{
		Id:          req.GetId(),
		Email:       fmt.Sprintf("%s@example.com", req.GetId()),
		DisplayName: req.GetDisplayName(),
		Status:      req.GetStatus(),
		UpdatedAt:   now,
	}

	h.log.Info("UpdateUser", zap.String("id", user.GetId()))
	return &userv1.UpdateUserResponse{User: user}, nil
}

// DeleteUser — Unary RPC.
func (h *UserHandler) DeleteUser(_ context.Context, req *userv1.DeleteUserRequest) (*userv1.DeleteUserResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	h.log.Info("DeleteUser", zap.String("id", req.GetId()))
	return &userv1.DeleteUserResponse{}, nil
}

// WatchUsers — Server-side streaming RPC.
// Сервер держит поток открытым и пушит события. Клиент читает до ctx.Done() или EOF.
func (h *UserHandler) WatchUsers(req *userv1.WatchUsersRequest, stream userv1.UserService_WatchUsersServer) error {
	h.log.Info("WatchUsers started", zap.String("status_filter", req.GetStatusFilter().String()))

	ctx := stream.Context()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	eventTypes := []userv1.UserEvent_EventType{
		userv1.UserEvent_EVENT_TYPE_CREATED,
		userv1.UserEvent_EVENT_TYPE_UPDATED,
		userv1.UserEvent_EVENT_TYPE_DELETED,
	}

	userStatus := userv1.UserStatus_USER_STATUS_ACTIVE
	if req.GetStatusFilter() != userv1.UserStatus_USER_STATUS_UNSPECIFIED {
		userStatus = req.GetStatusFilter()
	}

	seq := 0
	for {
		select {
		case <-ctx.Done():
			h.log.Info("WatchUsers: client disconnected", zap.Error(ctx.Err()))
			return ctx.Err()

		case t := <-ticker.C:
			seq++
			event := &userv1.UserEvent{
				Type: eventTypes[seq%len(eventTypes)],
				User: &userv1.User{
					Id:          fmt.Sprintf("user-%d", seq),
					Email:       fmt.Sprintf("user%d@example.com", seq),
					DisplayName: fmt.Sprintf("User %d", seq),
					Status:      userStatus,
					UpdatedAt:   timestamppb.New(t),
				},
			}

			if err := stream.Send(event); err != nil {
				return err
			}
			h.log.Info("sent UserEvent", zap.Int("seq", seq), zap.String("type", event.GetType().String()))
		}
	}
}
