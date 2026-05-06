package client

import (
	"context"
	"fmt"
	"time"

	userv1 "github.com/gulmix/hermes/gen/go/user/v1"
	"github.com/gulmix/hermes/pkg/circuitbreaker"
	"github.com/gulmix/hermes/pkg/grpcclient"
	"go.uber.org/zap"
)

type UserClient struct {
	rpc userv1.UserServiceClient
	cb  *circuitbreaker.Breaker
	log *zap.Logger
}

func NewUserClient(conn *grpcclient.Client, log *zap.Logger) *UserClient {
	return &UserClient{
		rpc: userv1.NewUserServiceClient(conn.Conn()),
		cb: circuitbreaker.New(circuitbreaker.Config{
			Threshold:   5,
			Timeout:     10 * time.Second,
			HalfOpenMax: 2,
		}),
		log: log,
	}
}

func (c *UserClient) GetUser(ctx context.Context, id string) (*userv1.User, error) {
	if err := c.cb.Allow(); err != nil {
		return nil, fmt.Errorf("user service circuit open: %w", err)
	}

	resp, err := c.rpc.GetUser(ctx, &userv1.GetUserRequest{Id: id})
	if err != nil {
		c.cb.RecordFailure()
		c.log.Warn("GetUser failed", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	c.cb.RecordSuccess()
	return resp.GetUser(), nil
}

func (c *UserClient) WatchUsers(ctx context.Context, statusFilter userv1.UserStatus) (userv1.UserService_WatchUsersClient, error) {
	if err := c.cb.Allow(); err != nil {
		return nil, fmt.Errorf("user service circuit open: %w", err)
	}

	stream, err := c.rpc.WatchUsers(ctx, &userv1.WatchUsersRequest{StatusFilter: statusFilter})
	if err != nil {
		c.cb.RecordFailure()
		return nil, err
	}

	c.cb.RecordSuccess()
	return stream, nil
}
