package compression

import (
	_ "google.golang.org/grpc/encoding/gzip"
)

const (
	Gzip   = "gzip"
	Snappy = "snappy"
)
