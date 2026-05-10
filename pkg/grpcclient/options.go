package grpcclient

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type Options struct {
	defaultTimeout time.Duration
	maxRetries     int
	retryInitial   time.Duration
	retryMax       time.Duration
	tlsCreds       credentials.TransportCredentials
	extraDialOpts  []grpc.DialOption
}

type Option func(*Options)

func defaultOptions() Options {
	return Options{
		defaultTimeout: 5 * time.Second,
		maxRetries:     3,
		retryInitial:   100 * time.Millisecond,
		retryMax:       2 * time.Second,
		tlsCreds:       insecure.NewCredentials(),
	}
}

func WithTimeout(d time.Duration) Option {
	return func(o *Options) { o.defaultTimeout = d }
}

func WithRetry(maxRetries int, initial, max time.Duration) Option {
	return func(o *Options) {
		o.maxRetries = maxRetries
		o.retryInitial = initial
		o.retryMax = max
	}
}

func WithTLS(cfg TLSConfig) Option {
	return func(o *Options) {
		creds, err := loadClientTLS(cfg)
		if err != nil {
			panic("grpcclient: " + err.Error())
		}
		o.tlsCreds = creds
	}
}

func WithDialOpts(opts ...grpc.DialOption) Option {
	return func(o *Options) {
		o.extraDialOpts = append(o.extraDialOpts, opts...)
	}
}

func WithKeepalive(pingTime, timeout time.Duration, permitWithoutStream bool) Option {
	return func(o *Options) {
		o.extraDialOpts = append(o.extraDialOpts,
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                pingTime,
				Timeout:             timeout,
				PermitWithoutStream: permitWithoutStream,
			}),
		)
	}
}

func WithCompression(codec string) Option {
	return func(o *Options) {
		o.extraDialOpts = append(o.extraDialOpts,
			grpc.WithDefaultCallOptions(grpc.UseCompressor(codec)),
		)
	}
}
