package client

import (
	"context"

	orderv1 "github.com/gulmix/hermes/gen/go/order/v1"
	"google.golang.org/grpc"
)

type OrderClient struct {
	svc orderv1.OrderServiceClient
}

func NewOrderClient(conn *grpc.ClientConn) *OrderClient {
	return &OrderClient{svc: orderv1.NewOrderServiceClient(conn)}
}

func (c *OrderClient) GetOrder(ctx context.Context, id string) (*orderv1.Order, error) {
	resp, err := c.svc.GetOrder(ctx, &orderv1.GetOrderRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return resp.GetOrder(), nil
}
