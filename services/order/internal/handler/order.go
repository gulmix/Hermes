package handler

import (
	"context"
	"errors"
	"fmt"
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

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

// GetOrder — Unary RPC.
func (h *OrderHandler) GetOrder(_ context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	h.log.Info("GetOrder", zap.String("id", req.GetId()))

	now := timestamppb.Now()
	order := &orderv1.Order{
		Id:     req.GetId(),
		UserId: "user-42",
		Status: orderv1.OrderStatus_ORDER_STATUS_CONFIRMED,
		Items: []*orderv1.OrderItem{
			{ProductId: "prod-1", Quantity: 2, PriceCents: 1500},
		},
		TotalCents: 3000,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	return &orderv1.GetOrderResponse{Order: order}, nil
}

// CreateOrder — Unary RPC; проверяет пользователя через User-сервис.
// Deadline propagation: ctx уже содержит deadline от входящего запроса —
// interceptor TimeoutUnary на клиенте не перезапишет его.
func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if len(req.GetItems()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items must not be empty")
	}

	user, err := h.user.GetUser(ctx, req.GetUserId())
	if err != nil {
		h.log.Warn("user lookup failed", zap.String("user_id", req.GetUserId()), zap.Error(err))
		return nil, status.Errorf(codes.FailedPrecondition, "user not found: %v", err)
	}

	var total int64
	for _, item := range req.GetItems() {
		total += int64(item.GetQuantity()) * item.GetPriceCents()
	}

	now := timestamppb.Now()
	order := &orderv1.Order{
		Id:         fmt.Sprintf("order-%s-%d", user.GetId(), now.AsTime().UnixNano()),
		UserId:     user.GetId(),
		Items:      req.GetItems(),
		Status:     orderv1.OrderStatus_ORDER_STATUS_PENDING,
		TotalCents: total,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	h.log.Info("order created", zap.String("order_id", order.GetId()))
	return &orderv1.CreateOrderResponse{Order: order}, nil
}

// UploadOrderEvents — Client-side streaming RPC.
// Клиент шлёт поток событий; сервер принимает их все и возвращает один итоговый ответ.
func (h *OrderHandler) UploadOrderEvents(stream orderv1.OrderService_UploadOrderEventsServer) error {
	var accepted, rejected int32

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			h.log.Info("UploadOrderEvents done", zap.Int32("accepted", accepted), zap.Int32("rejected", rejected))
			return stream.SendAndClose(&orderv1.OrderEventResponse{
				Accepted: accepted,
				Rejected: rejected,
			})
		}
		if err != nil {
			return err
		}

		if req.GetOrderId() == "" || req.GetNewStatus() == orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED {
			h.log.Warn("invalid event, skipping", zap.String("order_id", req.GetOrderId()))
			rejected++
			continue
		}

		h.log.Info("event received",
			zap.String("order_id", req.GetOrderId()),
			zap.String("status", req.GetNewStatus().String()),
			zap.String("note", req.GetNote()),
		)
		accepted++
	}
}

// StreamOrderUpdates — Bidirectional streaming RPC.
// Клиент шлёт order_id, сервер отвечает текущим статусом для каждого запроса.
// Оба направления независимы; цикл завершается когда клиент закрывает свою сторону (EOF).
func (h *OrderHandler) StreamOrderUpdates(stream orderv1.OrderService_StreamOrderUpdatesServer) error {
	ctx := stream.Context()

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		if ctx.Err() != nil {
			return status.FromContextError(ctx.Err()).Err()
		}

		h.log.Info("status request", zap.String("order_id", req.GetOrderId()))

		if err := stream.Send(&orderv1.OrderStatusResponse{
			OrderId:   req.GetOrderId(),
			Status:    orderv1.OrderStatus_ORDER_STATUS_SHIPPED,
			UpdatedAt: timestamppb.Now(),
		}); err != nil {
			return err
		}
	}
}
