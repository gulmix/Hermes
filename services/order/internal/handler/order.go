package handler

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderv1 "github.com/gulmix/hermes/gen/go/order/v1"
	"github.com/gulmix/hermes/services/order/internal/client"
)

type OrderHandler struct {
	orderv1.UnimplementedOrderServiceServer
	log  *zap.Logger
	user *client.UserClient
}

func NewOrderHandler(log *zap.Logger, user *client.UserClient) *OrderHandler {
	return &OrderHandler{log: log, user: user}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	user, err := h.user.GetUser(ctx, req.GetUserId())
	if err != nil {
		h.log.Warn("user lookup failed", zap.String("user_id", req.GetUserId()), zap.Error(err))
		return nil, status.Errorf(codes.FailedPrecondition, "user not found: %v", err)
	}

	h.log.Info("creating order", zap.String("user_id", user.GetId()))
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	h.log.Info("GetOrder", zap.String("id", req.GetId()))
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (h *OrderHandler) StreamOrderUpdates(stream orderv1.OrderService_StreamOrderUpdatesServer) error {
	return status.Errorf(codes.Unimplemented, "not implemented")
}
